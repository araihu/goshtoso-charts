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
      if ((await fetch(`${baseURL}/components/interactive/bar-3d`)).ok) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Bar3D verification server did not start at ${baseURL}`);
}

before(async () => {
  if (!baseURL) {
    const port = await freePort();
    assert.notEqual(port, 8091);
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
  await page.goto(`${baseURL}/components/interactive/bar-3d`);
  await page.locator("[data-bar3d-variant]").first().waitFor();
  await page.waitForFunction(() => {
    const hosts = [...document.querySelectorAll("[data-bar3d-variant] [_echarts_instance_]")];
    return hosts.length === 4 && hosts.every((host) => Boolean(window.echarts.getInstanceByDom(host)));
  });
  return page;
}

function wrapperFor(page, variant = "base") {
  return page.locator(`[data-bar3d-variant="${variant}"] [data-goshtoso-chart-wrapper]`).first();
}

async function measure(wrapper) {
  return wrapper.evaluate(async (element) => {
    const host = element.querySelector("[_echarts_instance_]");
    const instance = window.echarts.getInstanceByDom(host);
    const option = instance.getOption();
    const values = option.series[0].data.map((cell) => cell.value.slice(0, 3));
    const bytes = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(JSON.stringify(values)));
    return {
      sameInstance: !element.__bar3DInstance || instance === element.__bar3DInstance,
      hostWidth: host.clientWidth,
      hostHeight: host.clientHeight,
      chartWidth: instance.getWidth(),
      chartHeight: instance.getHeight(),
      canvasCount: host.querySelectorAll("canvas").length,
      seriesType: option.series[0].type,
      count: values.length,
      hash: [...new Uint8Array(bytes)].map((value) => value.toString(16).padStart(2, "0")).join(""),
      values,
      xCategories: option.xAxis3D?.[0]?.data,
      yCategories: option.yAxis3D?.[0]?.data,
      boxWidth: option.grid3D?.[0]?.boxWidth,
      boxDepth: option.grid3D?.[0]?.boxDepth,
      autoRotate: option.grid3D?.[0]?.viewControl?.autoRotate,
      autoRotateSpeed: option.grid3D?.[0]?.viewControl?.autoRotateSpeed,
      shading: option.series[0].shading,
      range: option.visualMap?.[0]?.range,
      calculable: option.visualMap?.[0]?.calculable,
      scaleColors: option.visualMap?.[0]?.inRange?.color || [],
      background: option.backgroundColor,
      axisText: option.xAxis3D?.[0]?.axisLabel?.color,
      grid: option.xAxis3D?.[0]?.splitLine?.lineStyle?.color,
    };
  });
}

test("four variants preserve source data, index swap, palette, grid, rotation, shading, and local dependency order", async () => {
  const page = await pageAt({ width: 1440, height: 900 });
  try {
    const resources = await page.evaluate(() => [...document.scripts].map((script) => script.src).filter(Boolean));
    const names = ["echarts/5.4.3", "word-cloud/2.1.0", "liquid/3.1.0", "three-d/2.0.9", "maps/41f247b1cbb6"];
    const indexes = names.map((name) => resources.findIndex((url) => url.includes(name)));
    assert.ok(indexes.every((index) => index >= 0), JSON.stringify({ resources, indexes }));
    assert.ok(indexes.every((index, position) => position === 0 || indexes[position - 1] < index), JSON.stringify(indexes));

    const expectedHours = ["12a", "1a", "2a", "3a", "4a", "5a", "6a", "7a", "8a", "9a", "10a", "11a", "12p", "1p", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "9p", "10p", "11p"];
    const expectedDays = ["Saturday", "Friday", "Thursday", "Wednesday", "Tuesday", "Monday", "Sunday"];
    const base = await measure(wrapperFor(page));
    assert.equal(base.seriesType, "bar3D");
    assert.ok(base.canvasCount >= 1);
    assert.equal(base.count, 168);
    assert.equal(base.hash, "580250773b8f88507e97adbdf56b90d3b0f6e5cb13e7e13a3b8e7c7f377e8e94");
    assert.deepEqual(base.xCategories, expectedHours);
    assert.deepEqual(base.yCategories, expectedDays);
    assert.deepEqual(base.values.slice(0, 3), [[0, 0, 5], [1, 0, 1], [2, 0, 0]]);
    assert.deepEqual(base.values.slice(24, 27), [[0, 1, 7], [1, 1, 0], [2, 1, 0]]);
    assert.deepEqual([base.boxWidth, base.boxDepth], [200, 80]);
    assert.deepEqual(base.range, [0, 30]);
    assert.equal(base.calculable, true);
    assert.equal(base.scaleColors.length, 10);

    const rotating = await measure(wrapperFor(page, "auto-rotate"));
    assert.deepEqual([rotating.boxWidth, rotating.boxDepth, rotating.autoRotate], [160, 80, true]);
    const faster = await measure(wrapperFor(page, "faster"));
    assert.deepEqual([faster.boxWidth, faster.boxDepth, faster.autoRotate, faster.autoRotateSpeed], [160, 80, true, 200]);
    const lambert = await measure(wrapperFor(page, "lambert"));
    assert.deepEqual([lambert.boxWidth, lambert.boxDepth, lambert.shading], [200, 80, "lambert"]);
    for (const state of [rotating, faster, lambert]) {
      assert.equal(state.count, 168);
      assert.equal(state.hash, base.hash);
      assert.deepEqual(state.xCategories, expectedHours);
      assert.deepEqual(state.yCategories, expectedDays);
    }

    for (const asset of ["/charts/assets/js/runtime/three-d/2.0.9/runtime.min.js", "/charts/assets/js/runtime/echarts/5.4.3/echarts.min.js"]) {
      const response = await page.request.get(`${baseURL}${asset}`);
      assert.equal(response.status(), 200);
    }
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
          assert.ok(state.canvasCount >= 1);
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
      element.__bar3DInstance = window.echarts.getInstanceByDom(host);
    });
    const initial = await measure(wrapper);
    await wrapper.locator("[data-goshtoso-chart-expand] > div > button").click();
    const dialog = wrapper.getByRole("dialog", { name: "basic bar3d example" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForTimeout(350);
    await page.waitForFunction((initialWidth) => {
      const wrapper = document.querySelector('[data-bar3d-variant="base"] [data-goshtoso-chart-wrapper]');
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
    await wrapper.getByRole("button", { name: "Download basic bar3d example as PNG" }).click();
    const artifact = await pending;
    const bytes = await fs.readFile(await artifact.path());
    assert.equal(artifact.suggestedFilename(), "basic-bar3d-example.png");
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

test("explicit CDN core and 3D extension are pinned with SRI and register Bar3D in order", async () => {
  const page = await browser.newPage();
  const errors = [];
  page.on("pageerror", (error) => errors.push(error.message));
  try {
    await page.setContent(`<!doctype html><html><head>
      <script src="https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js" integrity="sha384-BQKzmHvQLMCAnL3UtDBA1Al5tFjsCz1wrMlIUA1wkzo14DYkRWjywW+p9pCj0cwd" crossorigin="anonymous"></script>
      <script src="https://cdn.jsdelivr.net/npm/echarts-gl@2.0.9/dist/echarts-gl.min.js" integrity="sha384-f4gAUkb5Y6LE9n50CbiH1hCBCw7021OeJu0ZrgRpgW6G1CZjPR8cu33e8rCFLqCl" crossorigin="anonymous"></script>
    </head><body><div id="chart" style="width:400px;height:300px"></div>
    <script>const chart=echarts.init(document.getElementById("chart"));chart.setOption({xAxis3D:{type:"category",data:["x"]},yAxis3D:{type:"category",data:["y"]},zAxis3D:{},grid3D:{},series:[{type:"bar3D",data:[[0,0,1]]}]});</script>
    </body></html>`, { waitUntil: "networkidle" });
    assert.deepEqual(await page.evaluate(() => {
      const instance = window.echarts.getInstanceByDom(document.getElementById("chart"));
      return { type: instance.getOption().series[0].type, canvases: document.querySelectorAll("#chart canvas").length };
    }), { type: "bar3D", canvases: 1 });
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});
