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

async function measure(wrapper) {
  return wrapper.evaluate((element) => {
    const host = element.querySelector("[_echarts_instance_]");
    const instance = window.echarts.getInstanceByDom(host);
    const canvas = host.querySelector("canvas");
    const option = instance.getOption();
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
      areaColor: option.series[0].itemStyle.areaColor,
      scaleColors: option.visualMap?.[0]?.inRange?.color || [],
    };
  });
}

test("local resources register both geometries before five map variants initialize", async () => {
  const page = await pageAt({ width: 1440, height: 900 });
  try {
    const state = await page.evaluate(() => ({
      china: Boolean(window.echarts.getMap("china")),
      guangdong: Boolean(window.echarts.getMap("广东")),
      localResources: performance.getEntriesByType("resource").map((entry) => entry.name).filter((name) => name.includes("/charts/assets/js/maps/")),
    }));
    assert.equal(state.china, true);
    assert.equal(state.guangdong, true);
    assert.equal(state.localResources.length, 2);
    assert.ok(state.localResources.every((url) => url.includes("/js/maps/41f247b1cbb6/")));
    const basic = await measure(wrapperFor(page));
    assert.deepEqual(basic.names, ["北京", "上海", "广东", "辽宁", "山东", "山西", "陕西", "新疆", "内蒙古"]);
    assert.deepEqual(basic.values, [101, 72, 134, 53, 96, 42, 68, 29, 81]);
    const regional = await measure(wrapperFor(page, "regional"));
    assert.deepEqual(regional.names, ["深圳市", "广州市", "湛江市", "汕头市", "东莞市", "佛山市", "云浮市", "肇庆市", "梅州市"]);
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

    await wrapper.locator("[data-goshtoso-chart-expand] > div > button").click();
    const dialog = wrapper.getByRole("dialog", { name: "basic map example" });
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
    await wrapper.getByRole("button", { name: "Download basic map example as PNG" }).click();
    const artifact = await pending;
    const bytes = await fs.readFile(await artifact.path());
    assert.equal(artifact.suggestedFilename(), "basic-map-example.png");
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

test("390 and 1440 layouts stay centered, contained, and theme-responsive", async () => {
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
          await page.waitForTimeout(450);
          const wrapper = wrapperFor(page, "scale");
          const state = await measure(wrapper);
          assert.deepEqual({ chart: state.chartWidth, canvas: state.canvasWidth }, { chart: state.hostWidth, canvas: state.hostWidth });
          assert.deepEqual({ chart: state.chartHeight, canvas: state.canvasHeight }, { chart: state.hostHeight, canvas: state.hostHeight });
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

test("explicit CDN scripts are pinned, integrity-checked, and register both geometries", async () => {
  const page = await browser.newPage();
  const errors = [];
  page.on("pageerror", (error) => errors.push(error.message));
  try {
    await page.setContent(`<!doctype html><html><head>
      <script src="https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js" integrity="sha384-BQKzmHvQLMCAnL3UtDBA1Al5tFjsCz1wrMlIUA1wkzo14DYkRWjywW+p9pCj0cwd" crossorigin="anonymous"></script>
      <script src="https://cdn.jsdelivr.net/gh/go-echarts/go-echarts-assets@41f247b1cbb649b029a2d3fffb04f469de372aa7/assets/maps/china.js" integrity="sha384-qwEZxzbtfuBsHahOge6aHnLsYt6NBGcOFoTctegFtOU3h/mWm8PYtRbJ19Xa6B5I" crossorigin="anonymous"></script>
      <script src="https://cdn.jsdelivr.net/gh/go-echarts/go-echarts-assets@41f247b1cbb649b029a2d3fffb04f469de372aa7/assets/maps/guangdong.js" integrity="sha384-Q7MOpZeBbcPxI3hKHud73/Z1PjvChsn12B3IN6NqOj08KXRF1IU2D7LvaY16uV4w" crossorigin="anonymous"></script>
    </head><body></body></html>`, { waitUntil: "networkidle" });
    assert.deepEqual(await page.evaluate(() => ({ china: Boolean(window.echarts?.getMap("china")), guangdong: Boolean(window.echarts?.getMap("广东")) })), { china: true, guangdong: true });
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});
