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
      if ((await fetch(`${baseURL}/components/interactive/gauge`)).ok) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Gauge verification server did not start at ${baseURL}`);
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
    try {
      process.kill(-server.pid, "SIGTERM");
    } catch {
      // Test-owned process already stopped.
    }
  }
});

async function pageAt(viewport, route = "/components/interactive/gauge") {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  await page.addInitScript(() => {
    const createObjectURL = URL.createObjectURL.bind(URL);
    globalThis.__chartBlobTypes = [];
    URL.createObjectURL = (blob) => {
      globalThis.__chartBlobTypes.push(blob.type);
      return createObjectURL(blob);
    };
  });
  await page.goto(`${baseURL}${route}`);
  await page.waitForFunction((selector) => {
    const hosts = [...document.querySelectorAll(selector)];
    return hosts.length > 0 && hosts.every((host) => Boolean(window.echarts.getInstanceByDom(host)));
  }, route.includes("word-cloud") ? "[data-word-cloud-variant] [_echarts_instance_]" : "[data-gauge-liquid-variant] [_echarts_instance_]");
  return page;
}

function liquidWrapper(page, variant = "basic") {
  return page.locator(`[data-gauge-liquid-variant="${variant}"] [data-goshtoso-chart-wrapper]`).first();
}

async function liquidState(wrapper) {
  return wrapper.evaluate((element) => {
    const host = element.querySelector("[_echarts_instance_]");
    const instance = window.echarts.getInstanceByDom(host);
    const canvas = host.querySelector("canvas");
    const series = instance.getOption().series[0];
    return {
      sameInstance: !element.__liquidInstance || instance === element.__liquidInstance,
      hostWidth: host.clientWidth,
      hostHeight: host.clientHeight,
      chartWidth: instance.getWidth(),
      chartHeight: instance.getHeight(),
      canvasWidth: Math.round(canvas.getBoundingClientRect().width),
      canvasHeight: Math.round(canvas.getBoundingClientRect().height),
      values: series.data.map((item) => item.value),
      shape: series.shape,
      waveAnimation: series.waveAnimation,
      colors: series.color,
    };
  });
}

test("progress stays unchanged and canonical liquid variants preserve exact semantics", async () => {
  const page = await pageAt({ width: 1440, height: 900 });
  const errors = [];
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  page.on("pageerror", (error) => errors.push(error.message));
  try {
    const progress = await page.locator('[data-gauge-variant="progress"] [data-goshtoso-chart-wrapper]').evaluate((element) => {
      const instance = window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]"));
      const series = instance.getOption().series[0];
      return { type: series.type, value: series.data[0].value, progressWidth: series.progress.width, pointer: series.pointer.show };
    });
    assert.deepEqual(progress, { type: "gauge", value: 73, progressWidth: 6, pointer: false });

    assert.equal(await page.locator("[data-gauge-liquid-variant]").count(), 8);
    const basic = await liquidState(liquidWrapper(page));
    assert.deepEqual(basic.values, [0.3, 0.4, 0.5]);
    assert.equal(basic.shape, "circle");
    assert.equal((await liquidState(liquidWrapper(page, "diamond"))).shape, "diamond");
    assert.equal((await liquidState(liquidWrapper(page, "pin"))).shape, "pin");
    assert.equal((await liquidState(liquidWrapper(page, "arrow"))).shape, "arrow");
    assert.equal((await liquidState(liquidWrapper(page, "triangle"))).shape, "triangle");
    assert.equal((await liquidState(liquidWrapper(page, "static"))).waveAnimation, false);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});

test("390 and 1440 liquid layouts stay contained and distinct across themes", async () => {
  const themeColors = new Map();
  for (const width of [390, 1440]) {
    for (const theme of ["goshtoso", "araihu"]) {
      for (const dark of [false, true]) {
        const page = await pageAt({ width, height: 900 });
        const errors = [];
        page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
        page.on("pageerror", (error) => errors.push(error.message));
        try {
          await page.evaluate(({ theme, dark }) => {
            document.documentElement.dataset.theme = theme;
            document.documentElement.classList.toggle("dark", dark);
          }, { theme, dark });
          await page.waitForTimeout(400);
          const state = await liquidState(liquidWrapper(page));
          assert.deepEqual({ chart: state.chartWidth, canvas: state.canvasWidth }, { chart: state.hostWidth, canvas: state.hostWidth });
          assert.deepEqual({ chart: state.chartHeight, canvas: state.canvasHeight }, { chart: state.hostHeight, canvas: state.hostHeight });
          assert.deepEqual(state.values, [0.3, 0.4, 0.5]);
          assert.ok(new Set(state.colors).size >= 3, `wave colors ${state.colors}`);
          assert.deepEqual(await page.evaluate(() => ({ client: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth })), { client: width, scroll: width });
          themeColors.set(`${theme}-${dark}`, state.colors[0]);
          assert.deepEqual(errors, []);
        } finally {
          await page.close();
        }
      }
    }
  }
  assert.equal(new Set(themeColors.values()).size, 4, `theme colors ${JSON.stringify([...themeColors])}`);
});

test("flex, theme, modal, and PNG reuse one liquid instance", async () => {
  const page = await pageAt({ width: 1440, height: 900 });
  const errors = [];
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  page.on("pageerror", (error) => errors.push(error.message));
  try {
    const wrapper = liquidWrapper(page);
    const initial = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      element.__liquidInstance = window.echarts.getInstanceByDom(host);
      element.__viewport = innerWidth;
      element.__resizeEvents = 0;
      addEventListener("resize", () => { element.__resizeEvents += 1; });
      return host.clientWidth;
    });
    await wrapper.evaluate((element) => {
      const flexParent = element.closest('[data-gauge-liquid-variant="basic"]');
      flexParent.style.flex = "0 0 607px";
      flexParent.style.width = "607px";
    });
    await page.waitForFunction((initialWidth) => {
      const wrapper = document.querySelector('[data-gauge-liquid-variant="basic"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth !== initialWidth && instance.getWidth() === host.clientWidth && Math.round(host.querySelector("canvas").getBoundingClientRect().width) === host.clientWidth;
    }, initial);
    let state = await liquidState(wrapper);
    assert.notEqual(state.hostWidth, initial);
    assert.equal(await wrapper.evaluate((element) => Math.round(element.closest('[data-gauge-liquid-variant="basic"]').getBoundingClientRect().width)), 607);
    assert.deepEqual({ same: state.sameInstance, chart: state.chartWidth, canvas: state.canvasWidth }, { same: true, chart: state.hostWidth, canvas: state.hostWidth });
    assert.deepEqual(await wrapper.evaluate((element) => ({ viewport: innerWidth, before: element.__viewport, resizeEvents: element.__resizeEvents })), { viewport: 1440, before: 1440, resizeEvents: 0 });

    const beforeTheme = state.colors;
    await page.evaluate(() => document.documentElement.classList.toggle("dark"));
    await page.waitForFunction((before) => {
      const host = document.querySelector('[data-gauge-liquid-variant="basic"] [_echarts_instance_]');
      return JSON.stringify(window.echarts.getInstanceByDom(host).getOption().series[0].color) !== JSON.stringify(before);
    }, beforeTheme);
    state = await liquidState(wrapper);
    assert.equal(state.sameInstance, true);

    await wrapper.locator("[data-goshtoso-chart-expand] > div > button").first().click();
    const dialog = wrapper.getByRole("dialog", { name: "basic liquid example" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForFunction(() => {
      const wrapper = document.querySelector('[data-gauge-liquid-variant="basic"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth > 607 && instance.getWidth() === host.clientWidth && Math.round(host.querySelector("canvas").getBoundingClientRect().width) === host.clientWidth;
    });
    state = await liquidState(wrapper);
    assert.equal(state.sameInstance, true);
    assert.deepEqual(await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
      const rect = panel.getBoundingClientRect();
      return {
        centered: Math.abs((rect.left + rect.right) / 2 - innerWidth / 2) < 4,
        large: rect.width >= innerWidth * 0.9 && rect.height >= innerHeight * 0.8,
        contained: rect.left >= 0 && rect.right <= innerWidth + 1 && rect.top >= 0 && rect.bottom <= innerHeight + 1,
      };
    }), { centered: true, large: true, contained: true });
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    const pending = page.waitForEvent("download");
    await wrapper.getByRole("button", { name: "Download basic liquid example as PNG" }).click();
    const artifact = await pending;
    const bytes = await fs.readFile(await artifact.path());
    assert.equal(artifact.suggestedFilename(), "basic-liquid-example.png");
    assert.equal(await page.evaluate(() => globalThis.__chartBlobTypes.at(-1)), "image/png");
    assert.deepEqual([...bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(bytes).metadata();
    state = await liquidState(wrapper);
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: state.chartWidth, height: state.chartHeight });
    const pixels = await sharp(bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});

test("WordCloud remains registered, themed, and exportable after liquid runtime", async () => {
  const page = await pageAt({ width: 1280, height: 800 }, "/components/interactive/word-cloud");
  try {
    const wrapper = page.locator('[data-word-cloud-variant="circle"] [data-goshtoso-chart-wrapper]').first();
    const state = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      const series = instance.getOption().series[0];
      return { type: series.type, names: series.data.slice(0, 2).map((item) => item.name), colors: series.data.slice(0, 2).map((item) => item.textStyle.color) };
    });
    assert.deepEqual({ type: state.type, names: state.names }, { type: "wordCloud", names: ["Sam S Club", "Macys"] });
    assert.equal(new Set(state.colors).size, 2);
    const pending = page.waitForEvent("download");
    await wrapper.getByRole("button", { name: "Download basic WordCloud example as PNG" }).click();
    const artifact = await pending;
    const bytes = await fs.readFile(await artifact.path());
    assert.deepEqual([...bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
  } finally {
    await page.close();
  }
});
