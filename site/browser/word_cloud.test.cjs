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
      if ((await fetch(`${baseURL}/components/interactive/word-cloud`)).ok) return;
    } catch {
      // Test-owned server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Word-cloud verification server did not start at ${baseURL}`);
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
  await page.goto(`${baseURL}/components/interactive/word-cloud`);
  await page.locator("[data-word-cloud-variant]").first().waitFor();
  await page.waitForFunction(() => {
    const hosts = [...document.querySelectorAll("[data-word-cloud-variant] [_echarts_instance_]")];
    return hosts.length === 3 && hosts.every((host) => Boolean(window.echarts.getInstanceByDom(host)));
  });
  return page;
}

function wrapperFor(page, variant = "circle") {
  return page.locator(`[data-word-cloud-variant="${variant}"] [data-goshtoso-chart-wrapper]`).first();
}

async function measure(wrapper) {
  return wrapper.evaluate((element) => {
    const host = element.querySelector("[_echarts_instance_]");
    const instance = window.echarts.getInstanceByDom(host);
    const canvas = host.querySelector("canvas");
    const option = instance.getOption();
    return {
      sameInstance: !element.__wordCloudInstance || instance === element.__wordCloudInstance,
      hostWidth: host.clientWidth,
      hostHeight: host.clientHeight,
      chartWidth: instance.getWidth(),
      chartHeight: instance.getHeight(),
      canvasWidth: Math.round(canvas.getBoundingClientRect().width),
      canvasHeight: Math.round(canvas.getBoundingClientRect().height),
      names: option.series[0].data.map((word) => word.name),
      values: option.series[0].data.map((word) => word.value),
      colors: option.series[0].data.map((word) => word.textStyle && word.textStyle.color),
    };
  });
}

test("flex-parent and modal size changes reuse one observed host and chart instance", async () => {
  const page = await pageAt({ width: 1440, height: 900 });
  const errors = [];
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  page.on("pageerror", (error) => errors.push(error.message));
  try {
    const wrapper = wrapperFor(page);
	await page.addStyleTag({ content: ".retail { color: rgb(12, 34, 56); }" });
	await page.evaluate(() => { document.documentElement.dataset.theme = "goshtoso"; });
	await page.waitForTimeout(350);
    const initial = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      element.__wordCloudInstance = window.echarts.getInstanceByDom(host);
      element.__windowResizeEvents = 0;
      addEventListener("resize", () => { element.__windowResizeEvents += 1; });
      return { viewport: innerWidth, hostWidth: host.clientWidth };
    });
	const styled = await measure(wrapper);
	assert.equal(styled.colors[0], "rgba(12, 34, 56, 1)");
	assert.equal(styled.colors[1], "rgba(255, 138, 61, 1)");

    await page.locator("[data-word-cloud-variants]").evaluate((element) => { element.style.width = "480px"; });
    await page.waitForFunction(() => {
      const wrapper = document.querySelector('[data-word-cloud-variant="circle"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      const canvas = host.querySelector("canvas");
      return host.clientWidth === 480 && instance.getWidth() === 480 &&
        Math.round(canvas.getBoundingClientRect().width) === 480;
    });
    let state = await measure(wrapper);
    assert.notEqual(state.hostWidth, initial.hostWidth);
    assert.deepEqual(
      { same: state.sameInstance, host: state.hostWidth, chart: state.chartWidth, canvas: state.canvasWidth },
      { same: true, host: 480, chart: 480, canvas: 480 },
    );
    assert.deepEqual(await wrapper.evaluate((element) => ({ viewport: innerWidth, resizeEvents: element.__windowResizeEvents })), { viewport: initial.viewport, resizeEvents: 0 });

    await wrapper.locator("[data-goshtoso-chart-expand] > div > button").first().click();
    const dialog = wrapper.getByRole("dialog", { name: "basic WordCloud example" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForFunction(() => {
      const wrapper = document.querySelector('[data-word-cloud-variant="circle"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth > 480 && instance.getWidth() === host.clientWidth &&
        Math.round(host.querySelector("canvas").getBoundingClientRect().width) === host.clientWidth;
    });
    state = await measure(wrapper);
    assert.equal(state.sameInstance, true);
    assert.deepEqual({ chart: state.chartWidth, canvas: state.canvasWidth }, { chart: state.hostWidth, canvas: state.hostWidth });
    const modal = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
      const rect = panel.getBoundingClientRect();
      return {
        centered: Math.abs((rect.left + rect.right) / 2 - innerWidth / 2) < 4,
        large: rect.width >= innerWidth * 0.9 && rect.height >= innerHeight * 0.8,
        contained: rect.left >= 0 && rect.right <= innerWidth + 1 && rect.top >= 0 && rect.bottom <= innerHeight + 1,
      };
    });
    assert.deepEqual(modal, { centered: true, large: true, contained: true });
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    const collapse = wrapper.locator('[data-goshtoso-chart-control="collapse"]');
    await collapse.click();
    await collapse.click();
    await page.waitForTimeout(350);
    state = await measure(wrapper);
    assert.equal(state.sameInstance, true);

    const pending = page.waitForEvent("download");
    await wrapper.getByRole("button", { name: "Download basic WordCloud example as PNG" }).click();
    const artifact = await pending;
    const bytes = await fs.readFile(await artifact.path());
    assert.equal(artifact.suggestedFilename(), "basic-wordcloud-example.png");
    assert.equal(await page.evaluate(() => globalThis.__chartBlobTypes.at(-1)), "image/png");
    assert.deepEqual([...bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(bytes).metadata();
    state = await measure(wrapper);
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: state.chartWidth, height: state.chartHeight });
    const pixels = await sharp(bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});

test("390 and 1440 layouts keep every theme mode contained and palette-distinct", async () => {
  const palettes = new Map();
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
          await page.waitForTimeout(500);
          const wrapper = wrapperFor(page);
          const state = await measure(wrapper);
          assert.deepEqual({ chart: state.chartWidth, canvas: state.canvasWidth }, { chart: state.hostWidth, canvas: state.hostWidth });
          assert.deepEqual({ chart: state.chartHeight, canvas: state.canvasHeight }, { chart: state.hostHeight, canvas: state.hostHeight });
          assert.equal(state.names.length, 20);
          assert.deepEqual(state.names.slice(0, 3), ["Sam S Club", "Macys", "Amy Schumer"]);
          assert.deepEqual(state.values.slice(0, 3), [10000, 6181, 4386]);
		  assert.equal(state.colors[1], "rgba(255, 138, 61, 1)");
          assert.ok(new Set(state.colors.slice(0, 8)).size >= 6, `colors ${state.colors.slice(0, 8)}`);
          assert.deepEqual(await page.evaluate(() => ({ client: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth })), { client: width, scroll: width });
          assert.ok(state.hostHeight >= 288 && state.hostHeight <= 500, `host height ${state.hostHeight}`);
          palettes.set(`${theme}-${dark}`, state.colors[8]);
          assert.deepEqual(errors, []);
        } finally {
          await page.close();
        }
      }
    }
  }
  assert.equal(new Set(palettes.values()).size, 4, `theme colors ${JSON.stringify([...palettes])}`);
});
