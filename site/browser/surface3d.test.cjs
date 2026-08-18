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
  for (let attempt = 0; attempt < 160; attempt += 1) {
    try {
      if ((await fetch(`${baseURL}/components/interactive/surface-3d`)).ok) return;
    } catch {
      // Test-owned server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Surface3D verification server did not start at ${baseURL}`);
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

async function pageAt(viewport, reducedMotion = "no-preference") {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  await page.emulateMedia({ reducedMotion });
  await page.addInitScript(() => {
    const createObjectURL = URL.createObjectURL.bind(URL);
    globalThis.__chartBlobTypes = [];
    URL.createObjectURL = (blob) => {
      globalThis.__chartBlobTypes.push(blob.type);
      return createObjectURL(blob);
    };
  });
  await page.goto(`${baseURL}/components/interactive/surface-3d`, { waitUntil: "domcontentloaded" });
  await page.locator("[data-surface3d-variant]").first().waitFor();
  await page.waitForFunction(() => {
    const hosts = [...document.querySelectorAll("[data-surface3d-variant] [_echarts_instance_]")];
    return hosts.length === 3 && hosts.every((host) => Boolean(window.echarts.getInstanceByDom(host)));
  });
  await page.waitForFunction(() => Boolean(document.documentElement._x_dataStack));
  return page;
}

function wrapperFor(page, variant = "base") {
  return page.locator(`[data-surface3d-variant="${variant}"] [data-goshtoso-chart-wrapper]`).first();
}

async function openExpand(wrapper) {
  const trigger = wrapper.locator('[id$="-primary-action"]:visible').first();
  await trigger.waitFor({ state: "visible" });
  await trigger.click();
  return trigger;
}

async function measure(wrapper) {
  return wrapper.evaluate(async (element) => {
    const host = element.querySelector("[_echarts_instance_]");
    const instance = window.echarts.getInstanceByDom(host);
    const option = instance.getOption();
    const series = option.series[0];
    const values = series.data.map((point) => point.value.slice(0, 3));
    const bytes = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(JSON.stringify(values)));
    const summary = element.parentElement.querySelector("[data-surface3d-exact-data]");
    const seriesModel = instance.getModel().getSeriesByIndex(0);
    const geometry = instance.getViewOfSeriesModel(seriesModel)._surfaceMesh.geometry;
    const canvas = host.querySelector("canvas");
    const gl = canvas && (canvas.getContext("webgl") || canvas.getContext("experimental-webgl"));
    const visualMap = option.visualMap && option.visualMap[0];
    return {
      sameInstance: !element.__surface3DInstance || instance === element.__surface3DInstance,
      hostWidth: host.clientWidth,
      hostHeight: host.clientHeight,
      chartWidth: instance.getWidth(),
      chartHeight: instance.getHeight(),
      canvasCount: host.querySelectorAll("canvas").length,
      seriesType: series.type,
      count: values.length,
      hash: [...new Uint8Array(bytes)].map((value) => value.toString(16).padStart(2, "0")).join(""),
      values,
      range: visualMap ? [visualMap.min, visualMap.max] : null,
      calculable: visualMap ? visualMap.calculable : null,
      scaleColors: visualMap ? visualMap.inRange.color : [],
      dataShape: series.dataShape || null,
      boxSize: [option.grid3D[0].boxWidth, option.grid3D[0].boxHeight, option.grid3D[0].boxDepth],
      wireframe: series.wireframe ? series.wireframe.show : true,
      shading: series.shading,
      autoRotate: option.grid3D[0].viewControl.autoRotate,
      triangleIndexCount: geometry.indices.length,
      webGL: Boolean(gl && !gl.isContextLost()),
      role: element.querySelector("figure[role=img]")?.getAttribute("role") || element.getAttribute("role"),
      ariaLabel: element.querySelector("figure[role=img]")?.getAttribute("aria-label") || element.getAttribute("aria-label"),
      background: option.backgroundColor,
      axisText: option.xAxis3D[0].axisLabel.color,
      grid: option.xAxis3D[0].splitLine.lineStyle.color,
      formula: summary.querySelector("[data-surface3d-formula]").textContent,
      summary: summary.querySelector("[data-surface3d-summary]").textContent.replace(/\s+/g, " ").trim(),
      motion: summary.querySelector("[data-surface3d-motion]").textContent.replace(/\s+/g, " ").trim(),
      tableRows: summary.querySelectorAll("tr").length,
    };
  });
}

test("both exact surfaces, local runtime order, route/search, palette, and CSV access", async () => {
  const page = await pageAt({ width: 1440, height: 900 });
  try {
    const resources = await page.evaluate(() => [...document.scripts].map((script) => script.src).filter(Boolean));
    const names = ["echarts/5.6.0", "word-cloud/2.1.0", "liquid/3.1.0", "three-d/2.0.9", "maps/ibge-mmd-2025"];
    const indexes = names.map((name) => resources.findIndex((url) => url.includes(name)));
    assert.ok(indexes.every((index) => index >= 0), JSON.stringify({ resources, indexes }));
    assert.ok(indexes[0] < indexes[1] && indexes[1] < indexes[2] && indexes[2] < indexes[3] && indexes[3] < indexes[4], JSON.stringify(indexes));

    const base = await measure(wrapperFor(page));
    assert.equal(base.seriesType, "surface");
    assert.ok(base.canvasCount >= 1);
    assert.equal(base.count, 14400);
    assert.equal(base.hash, "5530efe5532a0482cab10949d18ee56cee3eca691ebe2b37f593804f3565e5c3");
    assert.deepEqual(base.values[0].slice(0, 2), [-1, -1]);
    assert.ok(Math.abs(base.values[0][2]) < 1e-15);
    assert.deepEqual(base.values.at(-1).slice(0, 2), [59 / 60, 59 / 60]);
    assert.deepEqual(base.range, [-3, 3]);
    assert.equal(base.calculable, true);
    assert.equal(base.scaleColors.length, 10);
    assert.match(base.formula, /i \/ 60.*j \/ 60.*sin/);
    assert.match(base.summary, /14400 ordered points.*X domain \[-1, 0\.9833333333333333\].*Y domain \[-1, 0\.9833333333333333\]/);
    assert.equal(base.tableRows, 0);

    const rose = await measure(wrapperFor(page, "rose"));
    assert.equal(rose.seriesType, "surface");
    assert.equal(rose.count, 3600);
    assert.equal(rose.hash, "987c440f0454e249cb105bb15ac2356d26989b1dde57b6a8f3197e0e19e91b42");
    assert.deepEqual(rose.values[0].slice(0, 2), [-3, -3]);
    assert.deepEqual(rose.values.at(-1).slice(0, 2), [2.9, 2.9]);
    assert.deepEqual(rose.range, [-3, 3]);
    assert.equal(rose.calculable, true);
    assert.equal(rose.scaleColors.length, 10);
    assert.match(rose.formula, /i \/ 10.*j \/ 10.*sin/);
    assert.match(rose.summary, /3600 ordered points.*X domain \[-3, 2\.9\].*Y domain \[-3, 2\.9\]/);
    assert.equal(rose.tableRows, 0);

    const expectedByVariant = { base, rose };
    for (const variant of ["base", "rose"]) {
      const wrapper = wrapperFor(page, variant);
      const pending = page.waitForEvent("download");
      await wrapper.locator("xpath=..").getByRole("link", { name: "Download all exact points as CSV" }).click();
      const artifact = await pending;
      const csv = await fs.readFile(await artifact.path(), "utf8");
      const lines = csv.trimEnd().split("\n");
      assert.equal(lines[0], "series,x,y,z");
      assert.equal(lines.length - 1, variant === "base" ? 14400 : 3600);
      assert.equal(lines[1].startsWith('"surface3d",'), true);
      assert.equal(lines.at(-1).startsWith('"surface3d",'), true);
      const csvValues = lines.slice(1).map((line) => line.slice('"surface3d",'.length).split(",").map(Number));
      assert.deepEqual(csvValues, expectedByVariant[variant].values);
    }

    const search = await page.request.get(`${baseURL}/components/line`);
    assert.match(await search.text(), /data-search="surface 3d interactive \/ 3d interactive-surface-3d"/);
    for (const asset of ["/charts/assets/js/controls/6/controls.js", "/charts/assets/js/runtime/three-d/2.0.9/runtime.min.js", "/charts/assets/js/runtime/echarts/5.6.0/echarts.min.js"]) {
      assert.equal((await page.request.get(`${baseURL}${asset}`)).status(), 200);
    }
  } finally {
    await page.close();
  }
});

test("closed heart mesh is solid, connected, accessible, and backed by WebGL", async () => {
  const page = await pageAt({ width: 1440, height: 900 });
  try {
    const heart = await measure(wrapperFor(page, "heart"));
    assert.equal(heart.seriesType, "surface");
    assert.equal(heart.count, 49 * 65);
    assert.deepEqual(heart.dataShape, [49, 65]);
    assert.ok(heart.boxSize[2] < heart.boxSize[0] * 0.4, `heart depth ${heart.boxSize[2]} is too close to width ${heart.boxSize[0]}`);
    assert.equal(heart.wireframe, false);
    assert.equal(heart.shading, "lambert");
    assert.equal(heart.autoRotate, true);
    assert.equal(heart.triangleIndexCount, (49 - 1) * (65 - 1) * 6);
    assert.equal(heart.webGL, true);
    assert.equal(heart.role, "img");
    assert.equal(heart.ariaLabel, "Rotating parametric heart");
    assert.match(heart.formula, /θ ∈ \[0, 2π\].*φ ∈ \[0, π\].*sin³/);
    assert.match(heart.summary, /3185 ordered points/);
    assert.match(heart.motion, /rotates automatically.*Reduced-motion/);
    assert.equal(heart.tableRows, 0);

    const close = (left, right) => Math.abs(left - right) < 1e-10;
    for (let column = 0; column < 65; column += 1) {
      const first = heart.values[column];
      const last = heart.values[48 * 65 + column];
      assert.ok(first.every((value, index) => close(value, last[index])), `open outline seam at column ${column}`);
    }
    const frontPole = heart.values[0];
    const backPole = heart.values[64];
    for (let row = 1; row < 49; row += 1) {
      assert.ok(frontPole.every((value, index) => close(value, heart.values[row * 65][index])), `open front pole at row ${row}`);
      assert.ok(backPole.every((value, index) => close(value, heart.values[(row + 1) * 65 - 1][index])), `open back pole at row ${row}`);
    }
    const cleft = heart.values[32];
    const rightLobe = heart.values[5 * 65 + 32];
    const bottom = heart.values[24 * 65 + 32];
    const leftLobe = heart.values[43 * 65 + 32];
    assert.ok(rightLobe[0] > 3 && leftLobe[0] < -3);
    assert.ok(Math.abs((rightLobe[2] - cleft[2]) - 4.678937306500523) < 0.01);
    assert.ok(cleft[2] - bottom[2] > 15);
    const xs = heart.values.map((point) => point[0]);
    const zs = heart.values.map((point) => point[2]);
    assert.ok(Math.abs(Math.min(...xs) + 13.5) < 1e-10 && Math.abs(Math.max(...xs) - 13.5) < 1e-10);
    assert.ok(Math.min(...zs) < -15 && Math.max(...zs) > 10);

    const pending = page.waitForEvent("download");
    await wrapperFor(page, "heart").locator("xpath=..").getByRole("link", { name: "Download all exact points as CSV" }).click();
    const artifact = await pending;
    const csv = await fs.readFile(await artifact.path(), "utf8");
    assert.equal(csv.trimEnd().split("\n").length - 1, 49 * 65);
  } finally {
    await page.close();
  }
});

test("heart auto-rotation obeys reduced motion and resumes", async () => {
  const page = await pageAt({ width: 1440, height: 900 }, "reduce");
  try {
    const heart = await measure(wrapperFor(page, "heart"));
    assert.equal(heart.autoRotate, false);
    assert.match(heart.motion, /Reduced-motion preference disables animation and keeps the same surface stationary/);
    await page.emulateMedia({ reducedMotion: "no-preference" });
    await page.waitForFunction(() => {
      const wrapper = document.querySelector('[data-surface3d-variant="heart"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      return window.echarts.getInstanceByDom(host).getOption().grid3D[0].viewControl.autoRotate === true;
    });
    assert.equal((await measure(wrapperFor(page, "heart"))).autoRotate, true);
  } finally {
    await page.close();
  }
});

test("390 and 1440 widths retain four theme modes, contrast, centering, and resize convergence", async () => {
  const themed = new Map();
  for (const width of [390, 1440]) {
    for (const theme of ["goshtoso", "araihu"]) {
      for (const dark of [false, true]) {
        const page = await pageAt({ width, height: 900 });
        try {
          await page.evaluate(({ theme, dark }) => {
            document.documentElement.dataset.theme = theme;
            document.documentElement.classList.toggle("dark", dark);
          }, { theme, dark });
          await page.waitForTimeout(500);
          const wrapper = wrapperFor(page);
          const state = await measure(wrapper);
          assert.deepEqual({ chartWidth: state.chartWidth, chartHeight: state.chartHeight }, { chartWidth: state.hostWidth, chartHeight: state.hostHeight });
          assert.deepEqual(await page.evaluate(() => ({ client: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth })), { client: width, scroll: width });
          assert.equal(await wrapper.evaluate((element) => {
            const host = element.querySelector("[_echarts_instance_]").getBoundingClientRect();
            const content = element.querySelector("[data-goshtoso-chart-content]").getBoundingClientRect();
            return Math.abs((host.left + host.right) / 2 - (content.left + content.right) / 2) < 2;
          }), true);
          const contrast = await page.evaluate(({ background, foreground }) => {
            const parse = (value) => (value.match(/[\d.]+/g) || []).slice(0, 3).map(Number);
            const luminance = (value) => {
              const rgb = parse(value).map((channel) => channel / 255).map((channel) => channel <= 0.04045 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4);
              return 0.2126 * rgb[0] + 0.7152 * rgb[1] + 0.0722 * rgb[2];
            };
            const first = luminance(background);
            const second = luminance(foreground);
            return (Math.max(first, second) + 0.05) / (Math.min(first, second) + 0.05);
          }, { background: state.background, foreground: state.axisText });
          assert.ok(contrast >= 4.5, `contrast ${contrast}: ${JSON.stringify(state)}`);
          assert.notEqual(state.grid, state.background);
          themed.set(`${theme}-${dark}`, `${state.background}|${state.axisText}|${state.scaleColors.join("|")}`);
        } finally {
          await page.close();
        }
      }
    }
  }
  assert.equal(new Set(themed.values()).size, 4, JSON.stringify([...themed]));
});

test("large centered modal preserves instance and opaque direct PNG", async () => {
  const page = await pageAt({ width: 1440, height: 900 });
  const errors = [];
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  page.on("pageerror", (error) => errors.push(error.message));
  try {
    const wrapper = wrapperFor(page);
    await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      element.__surface3DInstance = window.echarts.getInstanceByDom(host);
    });
    const initial = await measure(wrapper);
    await wrapper.evaluate((element) => {
      element.querySelector(".goshtoso-charts-interactive > .container").style.width = "60%";
    });
    await page.waitForFunction((initialWidth) => {
      const wrapper = document.querySelector('[data-surface3d-variant="base"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth < initialWidth && instance.getWidth() === host.clientWidth;
    }, initial.hostWidth);
    assert.equal((await measure(wrapper)).sameInstance, true);
    await wrapper.evaluate((element) => {
      element.querySelector(".goshtoso-charts-interactive > .container").style.width = "";
    });
    await page.waitForFunction((initialWidth) => {
      const wrapper = document.querySelector('[data-surface3d-variant="base"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth === initialWidth && instance.getWidth() === host.clientWidth;
    }, initial.hostWidth);
    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "basic surface3D example" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForFunction((initialWidth) => {
      const wrapper = document.querySelector('[data-surface3d-variant="base"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth > initialWidth && instance.getWidth() === host.clientWidth;
    }, initial.hostWidth);
    assert.equal((await measure(wrapper)).sameInstance, true);
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
    assert.ok(panel.widthRatio >= 0.75 && panel.heightRatio >= 0.55, JSON.stringify(panel));
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    const pending = page.waitForEvent("download");
    await wrapper.getByRole("button", { name: "Export basic surface3D example" }).click();
    await wrapper.locator('[id$="-export-png-action"]:visible').first().click();
    const artifact = await pending;
    const bytes = await fs.readFile(await artifact.path());
    assert.equal(artifact.suggestedFilename(), "basic-surface3d-example.png");
    assert.equal(await page.evaluate(() => globalThis.__chartBlobTypes.at(-1)), "image/png");
    assert.deepEqual([...bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const image = sharp(bytes).ensureAlpha().raw();
    const { data, info } = await image.toBuffer({ resolveWithObject: true });
    let minimumAlpha = 255;
    for (let index = 3; index < data.length; index += info.channels) minimumAlpha = Math.min(minimumAlpha, data[index]);
    assert.equal(minimumAlpha, 255);
    assert.equal((await measure(wrapper)).sameInstance, true);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});

test("pinned CDN core and 3D extension use SRI and register Surface3D in order", async () => {
  const page = await browser.newPage();
  const errors = [];
  page.on("pageerror", (error) => errors.push(error.message));
  try {
    await page.setContent(`<!doctype html><html><head>
      <script src="https://cdn.jsdelivr.net/npm/echarts@5.6.0/dist/echarts.min.js" integrity="sha384-pPi0zxBAoDu6+JXW/C68UZLvBUUtU+7zonhif43rqj7pxsGyqyqzcian2Rj37Rss" crossorigin="anonymous"></script>
      <script src="https://cdn.jsdelivr.net/npm/echarts-gl@2.0.9/dist/echarts-gl.min.js" integrity="sha384-f4gAUkb5Y6LE9n50CbiH1hCBCw7021OeJu0ZrgRpgW6G1CZjPR8cu33e8rCFLqCl" crossorigin="anonymous"></script>
    </head><body><div id="chart" style="width:400px;height:300px"></div>
    <script>const chart=echarts.init(document.getElementById("chart"));chart.setOption({xAxis3D:{},yAxis3D:{},zAxis3D:{},grid3D:{},series:[{type:"surface",data:[[0,0,0],[1,0,1]]}]});</script>
    </body></html>`, { waitUntil: "networkidle" });
    assert.deepEqual(await page.evaluate(() => {
      const scripts = [...document.scripts].filter((item) => item.src);
      const instance = window.echarts.getInstanceByDom(document.getElementById("chart"));
      return {
        type: instance.getOption().series[0].type,
        canvases: document.querySelectorAll("#chart canvas").length,
        order: scripts.map((item) => item.src),
        integrity: scripts.map((item) => item.integrity),
      };
    }), {
      type: "surface",
      canvases: 1,
      order: [
        "https://cdn.jsdelivr.net/npm/echarts@5.6.0/dist/echarts.min.js",
        "https://cdn.jsdelivr.net/npm/echarts-gl@2.0.9/dist/echarts-gl.min.js",
      ],
      integrity: [
        "sha384-pPi0zxBAoDu6+JXW/C68UZLvBUUtU+7zonhif43rqj7pxsGyqyqzcian2Rj37Rss",
        "sha384-f4gAUkb5Y6LE9n50CbiH1hCBCw7021OeJu0ZrgRpgW6G1CZjPR8cu33e8rCFLqCl",
      ],
    });
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});
