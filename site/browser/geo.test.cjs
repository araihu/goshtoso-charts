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
      if ((await fetch(`${baseURL}/components/interactive/geo`)).ok) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Geo verification server did not start at ${baseURL}`);
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
  await page.goto(`${baseURL}/components/interactive/geo`);
  await page.locator("[data-geo-variant]").first().waitFor();
  await page.waitForFunction(() => {
    const hosts = [...document.querySelectorAll("[data-geo-variant] [_echarts_instance_]")];
    return hosts.length === 2 && hosts.every((host) => Boolean(window.echarts.getInstanceByDom(host)));
  });
  return page;
}

function wrapperFor(page, variant) {
  return page.locator(`[data-geo-variant="${variant}"] [data-goshtoso-chart-wrapper]`).first();
}

async function measure(wrapper) {
  return wrapper.evaluate((element) => {
    const host = element.querySelector("[_echarts_instance_]");
    const instance = window.echarts.getInstanceByDom(host);
    const canvas = host.querySelector("canvas");
    const option = instance.getOption();
    const series = option.series[0];
    return {
      sameInstance: !element.__geoInstance || instance === element.__geoInstance,
      hostWidth: host.clientWidth,
      hostHeight: host.clientHeight,
      chartWidth: instance.getWidth(),
      chartHeight: instance.getHeight(),
      canvasWidth: Math.round(canvas.getBoundingClientRect().width),
      canvasHeight: Math.round(canvas.getBoundingClientRect().height),
      type: series.type,
      coordinates: series.data.map((point) => point.value),
      ripple: series.rippleEffect || {},
      visualRange: option.visualMap?.[0] || null,
      geometry: option.geo?.[0]?.map,
      areaColor: option.geo?.[0]?.itemStyle?.areaColor,
      pointColor: series.itemStyle?.color,
	  pointColors: series.data.map((point) => point.itemStyle?.color || ""),
      background: option.backgroundColor,
      text: option.textStyle?.color,
    };
  });
}

function channels(color) {
  const match = String(color).match(/[\d.]+/g);
  if (!match || match.length < 3) throw new Error(`Unsupported color ${color}`);
  return match.slice(0, 3).map(Number);
}

function contrast(first, second) {
  const luminance = (color) => {
    const values = channels(color).map((value) => {
      const normalized = value / 255;
      return normalized <= 0.03928 ? normalized / 12.92 : ((normalized + 0.055) / 1.055) ** 2.4;
    });
    return 0.2126 * values[0] + 0.7152 * values[1] + 0.0722 * values[2];
  };
  const [light, dark] = [luminance(first), luminance(second)].sort((a, b) => b - a);
  return (light + 0.05) / (dark + 0.05);
}

test("both variants preserve exact coordinates, ripple, visual range, and reused local geometry", async () => {
  const page = await pageAt({ width: 1440, height: 900 });
  try {
    const resources = await page.evaluate(() => ({
	      brazil: Boolean(window.echarts.getMap("brazil")),
	      saoPaulo: Boolean(window.echarts.getMap("brazil-sao-paulo")),
      maps: performance.getEntriesByType("resource").map((entry) => entry.name).filter((name) => name.includes("/charts/assets/js/maps/")),
    }));
	    assert.deepEqual({ brazil: resources.brazil, saoPaulo: resources.saoPaulo }, { brazil: true, saoPaulo: true });
    assert.equal(resources.maps.length, 2);
	    assert.ok(resources.maps.every((url) => url.includes("/js/maps/ibge-mmd-2025/")));

    const national = await measure(wrapperFor(page, "effect-scatter"));
    assert.equal(national.type, "effectScatter");
	    assert.equal(national.geometry, "brazil");
	    assert.deepEqual(national.coordinates, [
	      [-60.02, -3.12, 81], [-34.88, -8.05, 27], [-47.88, -15.79, 47],
	      [-43.17, -22.91, 59], [-46.63, -23.55, 18], [-51.23, -30.03, 63],
    ]);
    assert.deepEqual(
      { period: national.ripple.period, scale: national.ripple.scale, brushType: national.ripple.brushType },
      { period: 4, scale: 6, brushType: "stroke" },
    );
	assert.deepEqual(channels(national.pointColor), [124, 58, 237]);
	assert.notDeepEqual(channels(national.pointColors[0]), channels(national.pointColor));
	assert.deepEqual(channels(national.pointColors[1]), [220, 38, 38]);

    const regional = await measure(wrapperFor(page, "scatter"));
    assert.equal(regional.type, "scatter");
	    assert.equal(regional.geometry, "brazil-sao-paulo");
	    assert.deepEqual(regional.coordinates, [
	      [-46.63, -23.55, 12], [-47.06, -22.91, 76], [-47.81, -21.18, 41],
    ]);
    assert.deepEqual(
      { min: regional.visualRange.min, max: regional.visualRange.max, calculable: regional.visualRange.calculable },
      { min: 0, max: 100, calculable: true },
    );
    assert.equal(regional.visualRange.inRange.color.length, 3);
	assert.deepEqual(channels(regional.areaColor), [226, 232, 240]);
	assert.notDeepEqual(channels(regional.pointColor), [124, 58, 237]);
	assert.deepEqual(channels(regional.pointColors[1]), [37, 99, 235]);
	assert.deepEqual(channels(regional.pointColors[2]), channels(regional.pointColor));
  } finally {
    await page.close();
  }
});

test("320, 390, 768, and 1440 Goshtoso/AraiHu light/dark layouts stay centered, contained, responsive, and contrast-safe", async () => {
  const themeStates = new Map();
	  for (const width of [320, 390, 768, 1440]) {
    for (const theme of ["goshtoso", "araihu"]) {
      for (const dark of [false, true]) {
        const page = await pageAt({ width, height: 900 });
        try {
          await page.evaluate(({ theme, dark }) => {
            document.documentElement.dataset.theme = theme;
            document.documentElement.classList.toggle("dark", dark);
          }, { theme, dark });
          await page.waitForTimeout(450);
          const wrapper = wrapperFor(page, "effect-scatter");
          const state = await measure(wrapper);
          assert.deepEqual(
            { chartWidth: state.chartWidth, canvasWidth: state.canvasWidth, chartHeight: state.chartHeight, canvasHeight: state.canvasHeight },
            { chartWidth: state.hostWidth, canvasWidth: state.hostWidth, chartHeight: state.hostHeight, canvasHeight: state.hostHeight },
          );
          assert.deepEqual(
            await page.evaluate(() => ({ client: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth })),
            { client: width, scroll: width },
          );
          const centered = await wrapper.evaluate((element) => {
            const host = element.querySelector("[_echarts_instance_]").getBoundingClientRect();
            const content = element.querySelector("[data-goshtoso-chart-content]").getBoundingClientRect();
            return Math.abs((host.left + host.right) / 2 - (content.left + content.right) / 2) < 2;
          });
          assert.equal(centered, true);
          assert.ok(contrast(state.text, state.background) >= 4.5, `${theme}/${dark} text contrast`);
          assert.ok(contrast(state.pointColor, state.background) >= 3, `${theme}/${dark} point contrast`);
          themeStates.set(`${theme}-${dark}`, `${state.areaColor}|${state.pointColor}`);
        } finally {
          await page.close();
        }
      }
    }
  }
  assert.equal(new Set(themeStates.values()).size, 4, JSON.stringify([...themeStates]));
});

test("320, 390, 768, and 1440 modals keep the same geo instance centered and contained", async () => {
  for (const width of [320, 390, 768, 1440]) {
    const page = await pageAt({ width, height: 900 });
    try {
      const wrapper = wrapperFor(page, "effect-scatter");
      await wrapper.evaluate((element) => {
        const host = element.querySelector("[_echarts_instance_]");
        element.__geoInstance = window.echarts.getInstanceByDom(host);
      });
      await wrapper.locator("[data-goshtoso-chart-expand] > div > button").click();
      const dialog = wrapper.getByRole("dialog", { name: "Brazil capitals" });
      await dialog.waitFor({ state: "visible" });
      await page.waitForTimeout(450);
      const state = await measure(wrapper);
      assert.equal(state.sameInstance, true, `same instance at ${width}`);
      assert.equal(state.chartWidth, state.hostWidth, `chart width at ${width}`);
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

test("Goshtoso modal resizes same observed instance and direct PNG is fully opaque", async () => {
  const page = await pageAt({ width: 1440, height: 900 });
  const errors = [];
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  page.on("pageerror", (error) => errors.push(error.message));
  try {
    const wrapper = wrapperFor(page, "effect-scatter");
    await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      element.__geoInstance = window.echarts.getInstanceByDom(host);
    });
    const initial = await measure(wrapper);
    await wrapper.locator("[data-goshtoso-chart-expand] > div > button").click();
	    const dialog = wrapper.getByRole("dialog", { name: "Brazil capitals" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForFunction((initialWidth) => {
      const wrapper = document.querySelector('[data-geo-variant="effect-scatter"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth > initialWidth && instance.getWidth() === host.clientWidth;
    }, initial.hostWidth);
    const expanded = await measure(wrapper);
    assert.equal(expanded.sameInstance, true);
    const panel = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((element) => {
      const rect = element.getBoundingClientRect();
      return {
        centered: Math.abs((rect.left + rect.right) / 2 - innerWidth / 2) < 4,
        contained: rect.left >= 0 && rect.right <= innerWidth + 1 && rect.top >= 0 && rect.bottom <= innerHeight + 1,
      };
    });
    assert.deepEqual(panel, { centered: true, contained: true });
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    const pending = page.waitForEvent("download");
	    await wrapper.getByRole("button", { name: "Download Brazil capitals as PNG" }).click();
    const artifact = await pending;
    const bytes = await fs.readFile(await artifact.path());
	    assert.equal(artifact.suggestedFilename(), "brazil-capitals.png");
    assert.equal(await page.evaluate(() => globalThis.__chartBlobTypes.at(-1)), "image/png");
    assert.deepEqual([...bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const image = sharp(bytes);
    const metadata = await image.metadata();
    const raw = await image.ensureAlpha().raw().toBuffer();
    for (let index = 3; index < raw.length; index += 4) {
      assert.equal(raw[index], 255, `non-opaque alpha at pixel ${Math.floor(index / 4)}`);
    }
    const settled = await measure(wrapper);
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: settled.chartWidth, height: settled.chartHeight });
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});

test("explicit CDN runtime keeps pinned Brazil geometries on application-owned local paths", async () => {
  const page = await browser.newPage();
  const errors = [];
  page.on("pageerror", (error) => errors.push(error.message));
  try {
    await page.setContent(`<!doctype html><html><head>
      <script src="https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js" integrity="sha384-BQKzmHvQLMCAnL3UtDBA1Al5tFjsCz1wrMlIUA1wkzo14DYkRWjywW+p9pCj0cwd" crossorigin="anonymous"></script>
	      <script src="${baseURL}/charts/assets/js/maps/ibge-mmd-2025/brazil.js"></script>
	      <script src="${baseURL}/charts/assets/js/maps/ibge-mmd-2025/sao-paulo.js"></script>
    </head><body></body></html>`, { waitUntil: "networkidle" });
    assert.deepEqual(
	      await page.evaluate(() => ({ brazil: Boolean(window.echarts?.getMap("brazil")), saoPaulo: Boolean(window.echarts?.getMap("brazil-sao-paulo")) })),
	      { brazil: true, saoPaulo: true },
    );
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});
