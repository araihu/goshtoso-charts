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
      if ((await fetch(`${baseURL}/components/interactive/scatter-3d`)).ok) return;
    } catch {
      // Test-owned server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Scatter3D verification server did not start at ${baseURL}`);
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
  await page.goto(`${baseURL}/components/interactive/scatter-3d`);
  await page.locator("[data-scatter3d-variant]").first().waitFor();
  await page.waitForFunction(() => {
    const hosts = [...document.querySelectorAll("[data-scatter3d-variant] [_echarts_instance_]")];
    return hosts.length === 2 && hosts.every((host) => Boolean(window.echarts.getInstanceByDom(host)));
  });
  return page;
}

function wrapperFor(page, variant = "basic") {
  return page.locator(`[data-scatter3d-variant="${variant}"] [data-goshtoso-chart-wrapper]`).first();
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
    const values = option.series[0].data.map((point) => point.value.slice(0, 3));
    const bytes = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(JSON.stringify(values)));
    return {
      sameInstance: !element.__scatter3DInstance || instance === element.__scatter3DInstance,
      hostWidth: host.clientWidth,
      hostHeight: host.clientHeight,
      chartWidth: instance.getWidth(),
      chartHeight: instance.getHeight(),
      canvasCount: host.querySelectorAll("canvas").length,
      seriesType: option.series[0].type,
      count: values.length,
      hash: [...new Uint8Array(bytes)].map((value) => value.toString(16).padStart(2, "0")).join(""),
      values,
      xName: option.xAxis3D?.[0]?.name,
      xShow: option.xAxis3D?.[0]?.show,
      yName: option.yAxis3D?.[0]?.name,
      zName: option.zAxis3D?.[0]?.name,
      pointColors: option.series[0].data.map((point) => point.itemStyle?.color),
      background: option.backgroundColor,
      axisText: option.xAxis3D?.[0]?.axisLabel?.color,
      grid: option.xAxis3D?.[0]?.splitLine?.lineStyle?.color,
      scaleColors: option.visualMap?.[0]?.inRange?.color || [],
    };
  });
}

test("local extension order, deterministic 80 points, source styles, and 3D canvas", async () => {
  const page = await pageAt({ width: 1440, height: 900 });
  try {
    const resources = await page.evaluate(() => [...document.scripts].map((script) => script.src).filter(Boolean));
    const names = ["echarts/5.6.0", "word-cloud/2.1.0", "liquid/3.1.0", "three-d/2.0.9", "maps/ibge-mmd-2025"];
    const indexes = names.map((name) => resources.findIndex((url) => url.includes(name)));
    assert.ok(indexes.every((index) => index >= 0), JSON.stringify({ resources, indexes }));
    assert.ok(indexes[0] < indexes[1] && indexes[1] < indexes[2] && indexes[2] < indexes[3] && indexes[3] < indexes[4], JSON.stringify(indexes));

    const basic = await measure(wrapperFor(page));
    assert.equal(basic.seriesType, "scatter3D");
    assert.ok(basic.canvasCount >= 1);
    assert.equal(basic.count, 80);
    assert.equal(basic.hash, "f01d67d48ef648c1b32e94ae2940e38b5c66037219516ab9d204e43819cc482e");
    assert.ok(basic.values.flat().every((value) => Number.isInteger(value) && value >= 0 && value <= 99));
    assert.equal(basic.scaleColors.length, 10);

    const styled = await measure(wrapperFor(page, "item-style"));
    assert.deepEqual(styled.values, [[10, 10, 10], [15, 15, 15], [20, 20, 20]]);
    assert.deepEqual([styled.xName, styled.yName, styled.zName], ["MY-X-AXIS", "MY-Y-AXIS", "MY-Z-AXIS"]);
    assert.equal(styled.xShow, true);
    assert.deepEqual(styled.pointColors, ["rgba(0, 128, 0, 1)", "rgba(0, 0, 255, 1)", "rgba(255, 0, 0, 1)"]);
  } finally {
    await page.close();
  }
});

test("responsive themes retain contrast, centered layout, and ResizeObserver convergence", async () => {
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
          assert.ok(state.canvasCount >= 1);
          assert.deepEqual(await page.evaluate(() => ({ client: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth })), { client: width, scroll: width });
          const centered = await wrapper.evaluate((element) => {
            const host = element.querySelector("[_echarts_instance_]").getBoundingClientRect();
            const content = element.querySelector("[data-goshtoso-chart-content]").getBoundingClientRect();
            return Math.abs((host.left + host.right) / 2 - (content.left + content.right) / 2) < 2;
          });
          assert.equal(centered, true);
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
      element.__scatter3DInstance = window.echarts.getInstanceByDom(host);
    });
    const initial = await measure(wrapper);
    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "basic Scatter3D example" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForTimeout(350);
    await page.waitForFunction((initialWidth) => {
      const wrapper = document.querySelector('[data-scatter3d-variant="basic"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth > initialWidth && instance.getWidth() === host.clientWidth;
    }, initial.hostWidth);
    let state = await measure(wrapper);
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
    assert.ok(panel.widthRatio >= 0.75 && panel.heightRatio >= 0.55, JSON.stringify(panel));
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    const pending = page.waitForEvent("download");
    await wrapper.getByRole("button", { name: "Export basic Scatter3D example" }).click();
    await wrapper.locator('[id$="-export-png-action"]:visible').first().click();
    const artifact = await pending;
    const bytes = await fs.readFile(await artifact.path());
    assert.equal(artifact.suggestedFilename(), "basic-scatter3d-example.png");
    assert.equal(await page.evaluate(() => globalThis.__chartBlobTypes.at(-1)), "image/png");
    assert.deepEqual([...bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const image = sharp(bytes).ensureAlpha().raw();
    const { data, info } = await image.toBuffer({ resolveWithObject: true });
    let minimumAlpha = 255;
    for (let index = 3; index < data.length; index += info.channels) minimumAlpha = Math.min(minimumAlpha, data[index]);
    assert.equal(minimumAlpha, 255);
    state = await measure(wrapper);
    assert.equal(state.sameInstance, true);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});

test("explicit CDN extension is pinned with SRI and registers Scatter3D after core", async () => {
  const page = await browser.newPage();
  const errors = [];
  page.on("pageerror", (error) => errors.push(error.message));
  try {
    await page.setContent(`<!doctype html><html><head>
      <script src="https://cdn.jsdelivr.net/npm/echarts@5.6.0/dist/echarts.min.js" integrity="sha384-pPi0zxBAoDu6+JXW/C68UZLvBUUtU+7zonhif43rqj7pxsGyqyqzcian2Rj37Rss" crossorigin="anonymous"></script>
      <script src="https://cdn.jsdelivr.net/npm/echarts-gl@2.0.9/dist/echarts-gl.min.js" integrity="sha384-f4gAUkb5Y6LE9n50CbiH1hCBCw7021OeJu0ZrgRpgW6G1CZjPR8cu33e8rCFLqCl" crossorigin="anonymous"></script>
    </head><body><div id="chart" style="width:400px;height:300px"></div>
    <script>const chart=echarts.init(document.getElementById("chart"));chart.setOption({xAxis3D:{},yAxis3D:{},zAxis3D:{},grid3D:{},series:[{type:"scatter3D",data:[[1,2,3]]}]});</script>
    </body></html>`, { waitUntil: "networkidle" });
    assert.deepEqual(await page.evaluate(() => {
      const instance = window.echarts.getInstanceByDom(document.getElementById("chart"));
      return { type: instance.getOption().series[0].type, canvases: document.querySelectorAll("#chart canvas").length };
    }), { type: "scatter3D", canvases: 1 });
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});
