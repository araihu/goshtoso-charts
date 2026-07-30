const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

let baseURL = process.env.BASE_URL;
let browser;
let server;

async function freePort() {
  return new Promise((resolve, reject) => {
    const listener = net.createServer();
    listener.once("error", reject);
    listener.listen(0, "127.0.0.1", () => {
      const { port } = listener.address();
      listener.close((error) => error ? reject(error) : resolve(port));
    });
  });
}

async function ready() {
  for (let attempt = 0; attempt < 120; attempt += 1) {
    try {
      if ((await fetch(`${baseURL}/components/interactive/map`)).ok) return;
    } catch {
      // Test-owned server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Map verification server did not start at ${baseURL}`);
}

before(async () => {
  if (!baseURL) {
    const port = await freePort();
    baseURL = `http://127.0.0.1:${port}`;
    server = spawn("go", ["run", "./cmd/server", "-port", String(port)], {
      cwd: path.resolve(__dirname, ".."), detached: true, stdio: "pipe",
    });
  }
  await ready();
  browser = await chromium.launch({ headless: true });
});

after(async () => {
  await browser?.close();
  if (server?.pid) {
    try { process.kill(-server.pid, "SIGTERM"); } catch { /* already stopped */ }
  }
});

async function pageAt(viewport) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  await page.addInitScript(() => {
    const createObjectURL = URL.createObjectURL.bind(URL);
    globalThis.__chartBlobTypes = [];
    URL.createObjectURL = (blob) => {
      globalThis.__chartBlobTypes.push(blob.type);
      return createObjectURL(blob);
    };
  });
  await page.goto(`${baseURL}/components/interactive/map`);
  await page.locator("[data-map-variant]").first().waitFor();
  await page.waitForFunction(() => {
    const hosts = [...document.querySelectorAll("[data-map-variant] [_echarts_instance_]")];
    return hosts.length === 5 && hosts.every((host) => Boolean(window.echarts.getInstanceByDom(host)));
  });
  return page;
}

function wrapperFor(page, variant = "") {
  return page.locator(`[data-map-variant="${variant}"] [data-goshtoso-chart-wrapper]`).first();
}

async function openExpand(wrapper) {
  const trigger = wrapper.locator('[id$="-stacked"] > button:visible, [data-action-group-primary] button:visible').first();
  const stacked = await trigger.evaluate((button) => Boolean(button.closest('[id$="-stacked"]')));
  await trigger.click();
  const action = wrapper.locator('[id$="-chart-expand-action"]').first();
  if (stacked) {
    await action.waitFor({ state: "visible" });
    await action.click();
  }
}

async function clickPNG(wrapper, label) {
  const direct = wrapper.getByRole("button", { name: `Download ${label} as PNG` });
  if (await direct.count() && await direct.isVisible()) {
    await direct.click();
    return;
  }
  await wrapper.getByRole("button", { name: /More .* chart actions/ }).click();
  await wrapper.locator('[id$="-export-png-action"]').first().click();
}

async function measure(wrapper) {
  return wrapper.evaluate((element) => {
    const host = element.querySelector("[_echarts_instance_]");
    const instance = window.echarts.getInstanceByDom(host);
    const canvas = host.querySelector("canvas");
    const option = instance.getOption();
    const displayList = instance.getZr().storage.getDisplayList();
    const regionRects = displayList
      .filter((item) => item.type === "compound" && item.z === 2)
      .map((item) => item.getBoundingRect());
    const mapBounds = {
      left: Math.min(...regionRects.map((rect) => rect.x)),
      right: Math.max(...regionRects.map((rect) => rect.x + rect.width)),
      top: Math.min(...regionRects.map((rect) => rect.y)),
      bottom: Math.max(...regionRects.map((rect) => rect.y + rect.height)),
    };
    const transformedRect = (item) => {
      const rect = item.getBoundingRect();
      const matrix = item.transform || [1, 0, 0, 1, 0, 0];
      const points = [
        [rect.x, rect.y], [rect.x + rect.width, rect.y],
        [rect.x + rect.width, rect.y + rect.height], [rect.x, rect.y + rect.height],
      ].map(([x, y]) => [
        matrix[0] * x + matrix[2] * y + matrix[4],
        matrix[1] * x + matrix[3] * y + matrix[5],
      ]);
      return {
        left: Math.min(...points.map((point) => point[0])),
        right: Math.max(...points.map((point) => point[0])),
        top: Math.min(...points.map((point) => point[1])),
        bottom: Math.max(...points.map((point) => point[1])),
      };
    };
    const scaleRects = displayList.filter((item) => item.z === 4).map(transformedRect);
    const scaleBounds = scaleRects.length ? {
      left: Math.min(...scaleRects.map((rect) => rect.left)),
      right: Math.max(...scaleRects.map((rect) => rect.right)),
      top: Math.min(...scaleRects.map((rect) => rect.top)),
      bottom: Math.max(...scaleRects.map((rect) => rect.bottom)),
    } : null;
    return {
      sameInstance: !element.__mapInstance || instance === element.__mapInstance,
      hostWidth: host.clientWidth,
      hostHeight: host.clientHeight,
      chartWidth: instance.getWidth(),
      chartHeight: instance.getHeight(),
      canvasWidth: Math.round(canvas.getBoundingClientRect().width),
      canvasHeight: Math.round(canvas.getBoundingClientRect().height),
      names: option.series[0].data.map((region) => region.name),
      values: option.series[0].data.map((region) => region.value),
	  regionColors: option.series[0].data.map((region) => region.itemStyle?.color || ""),
      areaColor: option.series[0].itemStyle.areaColor,
      scaleColors: option.visualMap?.[0]?.inRange?.color || [],
      labelShow: option.series[0].label?.show || false,
      showLegendSymbol: option.series[0].showLegendSymbol,
      renderedTexts: instance.getZr().storage.getDisplayList().filter((item) => item.type === "tspan").map((item) => item.style?.text || ""),
      mapBounds,
      mapAspect: (mapBounds.right - mapBounds.left) / (mapBounds.bottom - mapBounds.top),
      mapCenterErrorX: Math.abs((mapBounds.left + mapBounds.right) / 2 - host.clientWidth / 2),
      mapCenterErrorY: Math.abs((mapBounds.top + mapBounds.bottom) / 2 - host.clientHeight * 0.46),
      mapUtilization: Math.max(mapBounds.right - mapBounds.left, mapBounds.bottom - mapBounds.top) / Math.min(host.clientWidth, host.clientHeight),
      scaleBounds,
      scaleOverlap: scaleBounds ? !(
        mapBounds.right < scaleBounds.left || mapBounds.left > scaleBounds.right ||
        mapBounds.bottom < scaleBounds.top || mapBounds.top > scaleBounds.bottom
      ) : false,
    };
  });
}

test("local resources register both geometries before five map variants initialize", async () => {
  const page = await pageAt({ width: 1440, height: 900 });
  try {
    const state = await page.evaluate(() => ({
	      brazil: Boolean(window.echarts.getMap("brazil")),
	      saoPaulo: Boolean(window.echarts.getMap("brazil-sao-paulo")),
      localResources: performance.getEntriesByType("resource").map((entry) => entry.name).filter((name) => name.includes("/charts/assets/js/maps/")),
    }));
	    assert.equal(state.brazil, true);
	    assert.equal(state.saoPaulo, true);
    assert.equal(state.localResources.length, 2);
	    assert.ok(state.localResources.every((url) => url.includes("/js/maps/ibge-mmd-2025/")));
	    const basic = await measure(wrapperFor(page));
	    assert.deepEqual(basic.names, ["Rondônia", "Acre", "Amazonas", "Roraima", "Pará", "Amapá", "Tocantins", "Maranhão", "Piauí", "Ceará", "Rio Grande do Norte", "Paraíba", "Pernambuco", "Alagoas", "Sergipe", "Bahia", "Minas Gerais", "Espírito Santo", "Rio de Janeiro", "São Paulo", "Paraná", "Santa Catarina", "Rio Grande do Sul", "Mato Grosso do Sul", "Mato Grosso", "Goiás", "Distrito Federal"]);
	    assert.deepEqual(basic.values, [42, 28, 81, 19, 96, 24, 37, 73, 48, 102, 55, 61, 118, 52, 35, 134, 126, 58, 121, 146, 109, 87, 112, 46, 64, 92, 76]);
	    assert.equal(basic.showLegendSymbol, false);
	    const regional = await measure(wrapperFor(page, "regional"));
	    assert.deepEqual(regional.names, basic.names);
	    assert.match(regional.regionColors[18], /220|dc2626/i);
	    assert.notEqual(regional.regionColors[19], "");
	    const labels = await measure(wrapperFor(page, "labels"));
	    assert.equal(labels.labelShow, true);
	    assert.deepEqual(labels.names, basic.names);
	    for (const code of ["RO", "AC", "AM", "RR", "PA", "AP", "TO", "MA", "PI", "CE", "RN", "PB", "PE", "AL", "SE", "BA", "MG", "ES", "RJ", "SP", "PR", "SC", "RS", "MS", "MT", "GO", "DF"]) {
	      assert.ok(labels.renderedTexts.includes(code), `missing rendered UF label ${code}`);
	    }
  } finally {
    await page.close();
  }
});

test("flex and centered modal resize one observed map instance and PNG export stays direct", async () => {
  const page = await pageAt({ width: 1440, height: 900 });
  const errors = [];
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  page.on("pageerror", (error) => errors.push(error.message));
  try {
    const wrapper = wrapperFor(page);
    await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      element.__mapInstance = window.echarts.getInstanceByDom(host);
    });
    const initialWidth = (await measure(wrapper)).hostWidth;
    await page.locator('[data-map-variant=""]').evaluate((element) => {
      element.style.flex = "0 0 320px";
      element.style.width = "320px";
      element.style.maxWidth = "320px";
    });
    await page.waitForFunction((initialWidth) => {
      const wrapper = document.querySelector('[data-map-variant=""] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth !== initialWidth && instance.getWidth() === host.clientWidth;
    }, initialWidth);
    let state = await measure(wrapper);
    const flexWidth = state.hostWidth;
    assert.ok(flexWidth > 0 && flexWidth <= 320, `flex width ${flexWidth}`);
    assert.deepEqual({ same: state.sameInstance, chart: state.chartWidth, canvas: state.canvasWidth }, { same: true, chart: flexWidth, canvas: flexWidth });
	await page.locator('[data-map-variant=""]').evaluate((element) => { element.removeAttribute("style"); });

    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "Brazil states" });
    await dialog.waitFor({ state: "visible" });
	await page.waitForTimeout(350);
    await page.waitForFunction((flexWidth) => {
      const wrapper = document.querySelector('[data-map-variant=""] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth > flexWidth && instance.getWidth() === host.clientWidth;
    }, flexWidth);
    state = await measure(wrapper);
    assert.equal(state.sameInstance, true);
    const panel = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((element) => {
      const rect = element.getBoundingClientRect();
      return {
        centered: Math.abs((rect.left + rect.right) / 2 - innerWidth / 2) < 4,
        widthRatio: rect.width / innerWidth,
        heightRatio: rect.height / innerHeight,
        contained: rect.left >= 0 && rect.right <= innerWidth + 1 && rect.top >= 0 && rect.bottom <= innerHeight + 1,
      };
    });
    assert.equal(panel.centered, true);
    assert.equal(panel.contained, true);
    assert.ok(panel.widthRatio >= 0.75 && panel.heightRatio >= 0.55, `modal ratios ${JSON.stringify(panel)}`);
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    const pending = page.waitForEvent("download");
    await clickPNG(wrapper, "Brazil states");
    const artifact = await pending;
    const bytes = await fs.readFile(await artifact.path());
	    assert.equal(artifact.suggestedFilename(), "brazil-states.png");
    assert.equal(await page.evaluate(() => globalThis.__chartBlobTypes.at(-1)), "image/png");
    assert.deepEqual([...bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(bytes).metadata();
    state = await measure(wrapper);
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: state.chartWidth, height: state.chartHeight });
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});

test("390, 768, 1499, and 1440 layouts preserve Brazil geometry, center the plot, reserve scale space, and respond to theme", async () => {
  const themed = new Map();
	  for (const width of [390, 768, 1499, 1440]) {
    for (const theme of ["goshtoso", "araihu"]) {
      for (const dark of [false, true]) {
        const page = await pageAt({ width, height: 900 });
        try {
          const wrapper = wrapperFor(page, "scale");
          await wrapper.evaluate((element) => {
            const host = element.querySelector("[_echarts_instance_]");
            element.__mapInstance = window.echarts.getInstanceByDom(host);
          });
          await page.evaluate(({ theme, dark }) => {
            document.documentElement.dataset.theme = theme;
            document.documentElement.classList.toggle("dark", dark);
          }, { theme, dark });
          await page.waitForTimeout(450);
          const state = await measure(wrapper);
          assert.equal(state.sameInstance, true, `theme change replaced map instance at ${width}`);
          assert.ok(state.hostWidth > 0, `nonzero map width at ${width}`);
	      assert.equal(state.scaleColors.length, 3);
	      assert.notEqual(state.scaleColors[0], state.scaleColors[2]);
          assert.deepEqual({ chart: state.chartWidth, canvas: state.canvasWidth }, { chart: state.hostWidth, canvas: state.hostWidth });
          assert.deepEqual({ chart: state.chartHeight, canvas: state.canvasHeight }, { chart: state.hostHeight, canvas: state.hostHeight });
          assert.ok(state.mapAspect >= 0.98 && state.mapAspect <= 1.03, `Brazil aspect ${state.mapAspect} at ${width}`);
          assert.ok(state.mapCenterErrorX <= 2, `Brazil horizontal center error ${state.mapCenterErrorX} at ${width}`);
          assert.ok(state.mapCenterErrorY <= 2, `Brazil vertical plot center error ${state.mapCenterErrorY} at ${width}`);
          assert.ok(state.mapUtilization >= 0.78, `Brazil canvas utilization ${state.mapUtilization} at ${width}`);
          assert.equal(state.scaleOverlap, false, `Brazil scale overlap at ${width}: ${JSON.stringify({ map: state.mapBounds, scale: state.scaleBounds })}`);
          assert.deepEqual(await page.evaluate(() => ({ client: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth })), { client: width, scroll: width });
          const centered = await wrapper.evaluate((element) => {
            const host = element.querySelector("[_echarts_instance_]").getBoundingClientRect();
            const content = element.querySelector("[data-goshtoso-chart-content]").getBoundingClientRect();
            return Math.abs((host.left + host.right) / 2 - (content.left + content.right) / 2) < 2;
          });
          assert.equal(centered, true);
          themed.set(`${theme}-${dark}`, `${state.areaColor}|${state.scaleColors.join("|")}`);
        } finally {
          await page.close();
        }
      }
    }
  }
  assert.equal(new Set(themed.values()).size, 4, JSON.stringify([...themed]));
});

test("390, 768, 1499, and 1440 modals keep the same corrected map instance centered and contained", async () => {
  for (const width of [390, 768, 1499, 1440]) {
    const page = await pageAt({ width, height: 900 });
    try {
      const wrapper = wrapperFor(page);
      await wrapper.evaluate((element) => {
        const host = element.querySelector("[_echarts_instance_]");
        element.__mapInstance = window.echarts.getInstanceByDom(host);
        element.__mapHostWidth = host.clientWidth;
      });
      await openExpand(wrapper);
      const dialog = wrapper.getByRole("dialog", { name: "Brazil states" });
      await dialog.waitFor({ state: "visible" });
      await page.waitForTimeout(450);
      const state = await measure(wrapper);
      assert.equal(state.sameInstance, true, `same instance at ${width}`);
      assert.ok(state.hostWidth > 0, `nonzero modal map width at ${width}`);
      assert.equal(state.chartWidth, state.hostWidth, `chart width at ${width}`);
      assert.ok(state.mapAspect >= 0.98 && state.mapAspect <= 1.03, `modal Brazil aspect ${state.mapAspect} at ${width}`);
      assert.ok(state.mapCenterErrorX <= 2, `modal Brazil center error ${state.mapCenterErrorX} at ${width}`);
      assert.equal(state.scaleOverlap, false, `modal Brazil scale overlap at ${width}`);
      const panel = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((element) => {
        const rect = element.getBoundingClientRect();
        return {
          centered: Math.abs((rect.left + rect.right) / 2 - innerWidth / 2) < 4,
          contained: rect.left >= 0 && rect.right <= innerWidth + 1 && rect.top >= 0 && rect.bottom <= innerHeight + 1,
        };
      });
      assert.deepEqual(panel, { centered: true, contained: true }, `modal at ${width}`);
    } finally {
      await page.close();
    }
  }
});

test("390, 768, 1499, and 1440 fullscreen fallback resizes the same corrected map instance", async () => {
  for (const width of [390, 768, 1499, 1440]) {
    const page = await pageAt({ width, height: 900 });
    try {
      const wrapper = wrapperFor(page, "scale");
      await wrapper.evaluate((element) => {
        const host = element.querySelector("[_echarts_instance_]");
        element.__mapInstance = window.echarts.getInstanceByDom(host);
        element.__mapHostWidth = host.clientWidth;
        element.classList.add("goshtoso-charts-fullscreen-fallback");
        document.body.appendChild(element.closest("[data-map-variant]"));
        element.dispatchEvent(new CustomEvent("goshtoso-charts:resize", { bubbles: true }));
      });
      await page.waitForFunction(() => {
        const wrapper = document.querySelector(".goshtoso-charts-fullscreen-fallback");
        const host = wrapper?.querySelector("[_echarts_instance_]");
        const instance = host && window.echarts.getInstanceByDom(host);
        if (!instance || host.clientWidth === wrapper.__mapHostWidth ||
          instance.getWidth() !== host.clientWidth || instance.getHeight() !== host.clientHeight) return false;
        const regions = instance.getZr().storage.getDisplayList().filter((item) => item.type === "compound" && item.z === 2);
        const left = Math.min(...regions.map((item) => item.getBoundingRect().x));
        const right = Math.max(...regions.map((item) => {
          const rect = item.getBoundingRect();
          return rect.x + rect.width;
        }));
        return Math.abs((left + right) / 2 - host.clientWidth / 2) <= 2;
      });
      const state = await measure(wrapper);
      assert.equal(state.sameInstance, true, `fullscreen same instance at ${width}`);
      assert.ok(state.mapAspect >= 0.98 && state.mapAspect <= 1.03, `fullscreen Brazil aspect ${state.mapAspect} at ${width}`);
      assert.ok(state.mapCenterErrorX <= 2, `fullscreen Brazil center error ${state.mapCenterErrorX} at ${width}`);
      assert.equal(state.scaleOverlap, false, `fullscreen Brazil scale overlap at ${width}`);
    } finally {
      await page.close();
    }
  }
});

test("explicit CDN runtime keeps pinned Brazil geometries on application-owned local paths", async () => {
  const page = await browser.newPage();
  const errors = [];
  page.on("pageerror", (error) => errors.push(error.message));
  try {
    await page.setContent(`<!doctype html><html><head>
      <script src="https://cdn.jsdelivr.net/npm/echarts@5.6.0/dist/echarts.min.js" integrity="sha384-pPi0zxBAoDu6+JXW/C68UZLvBUUtU+7zonhif43rqj7pxsGyqyqzcian2Rj37Rss" crossorigin="anonymous"></script>
	      <script src="${baseURL}/charts/assets/js/maps/ibge-mmd-2025/brazil.js"></script>
	      <script src="${baseURL}/charts/assets/js/maps/ibge-mmd-2025/sao-paulo.js"></script>
	    </head><body></body></html>`, { waitUntil: "networkidle" });
	    assert.deepEqual(await page.evaluate(() => ({ brazil: Boolean(window.echarts?.getMap("brazil")), saoPaulo: Boolean(window.echarts?.getMap("brazil-sao-paulo")) })), { brazil: true, saoPaulo: true });
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});
