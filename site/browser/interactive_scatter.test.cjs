const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");

let baseURL;
let browser;
let server;

async function randomPort() {
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
      const response = await fetch(`${baseURL}/components/interactive/scatter`);
      if (response.ok && (await response.text()).includes('data-scatter-variant="effect-styles"')) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`interactive Scatter verification server did not start at ${baseURL}`);
}

before(async () => {
  const port = await randomPort();
  assert.notEqual(port, 8091);
  assert.notEqual(port, 8096);
  baseURL = `http://127.0.0.1:${port}`;
  server = spawn("go", ["run", "./cmd/server", "-port", String(port)], {
    cwd: path.resolve(__dirname, ".."), detached: true, stdio: "ignore",
  });
  await ready();
  browser = await chromium.launch({ headless: true });
});

after(async () => {
  await browser?.close();
  if (server?.pid) {
    try { process.kill(-server.pid, "SIGTERM"); } catch { /* server stopped */ }
  }
});

function variant(page, name) {
  return page.locator(`[data-scatter-variant="${name}"]`);
}

async function optionFor(page, name) {
  return variant(page, name).locator("[_echarts_instance_]").evaluate((host) => globalThis.echarts.getInstanceByDom(host).getOption());
}

for (const [name, viewport, theme, dark] of [
  ["wide-light", { width: 1440, height: 900 }, "goshtoso", false],
  ["wide-dark", { width: 1440, height: 900 }, "goshtoso", true],
  ["narrow-light", { width: 390, height: 844 }, "araihu", false],
  ["narrow-dark", { width: 390, height: 844 }, "araihu", true],
]) {
  test(`interactive Scatter ${name} preserves all variants, exact data, theme, and layout`, async () => {
    const page = await browser.newPage({ viewport, colorScheme: dark ? "dark" : "light" });
    const failures = [];
    page.on("console", (message) => { if (message.type() === "error") failures.push(message.text()); });
    page.on("pageerror", (error) => failures.push(error.message));
    page.on("response", (response) => { if (response.status() >= 400) failures.push(`${response.status()} ${response.url()}`); });
    try {
      await page.goto(`${baseURL}/components/interactive/scatter`, { waitUntil: "networkidle" });
      await page.evaluate(({ selected, darkMode }) => {
        document.documentElement.dataset.theme = selected;
        document.documentElement.classList.toggle("dark", darkMode);
      }, { selected: theme, darkMode: dark });
      await page.waitForTimeout(350);
      await page.waitForFunction(() => {
        const hosts = [...document.querySelectorAll("[data-scatter-variant] [_echarts_instance_]")];
        return hosts.length === 5 && hosts.every((host) => Boolean(globalThis.echarts?.getInstanceByDom(host)));
      });

      const base = await optionFor(page, "base");
      assert.equal(base.series.length, 2);
      assert.deepEqual(base.xAxis[0].data, ["Swimming", "Surfing", "Shooting", "Skating", "Wrestling", "Diving"]);
      assert.deepEqual(base.series[0].data.map(({ value }) => value), [81, 87, 47, 59, 81, 18]);
      assert.deepEqual(base.series[1].data.map(({ value }) => value), [25, 40, 56, 0, 94, 11]);
      assert.equal(base.series[0].data[0].symbol, "roundRect");
      assert.equal(base.series[0].data[0].symbolSize, 20);
      assert.equal(base.series[0].data[0].symbolRotate, 10);

      const labels = await optionFor(page, "labels");
      assert.equal(labels.series[0].label.show, true);
      assert.equal(labels.series[0].label.position, "right");

      const splitLines = await optionFor(page, "split-lines");
      assert.equal(splitLines.xAxis[0].name, "Sports");
      assert.equal(splitLines.yAxis[0].name, "Score");
      assert.equal(splitLines.xAxis[0].splitLine.show, true);
      assert.equal(splitLines.yAxis[0].splitLine.show, true);

      const effectBase = await optionFor(page, "effect-base");
      assert.equal(effectBase.series[0].type, "effectScatter");
      assert.deepEqual(effectBase.xAxis[0].data, ["Kobe", "Jordan", "Iverson", "LeBron", "Wade", "McGrady"]);
      assert.deepEqual(effectBase.series[0].data.map(({ value }) => value), [37, 31, 85, 26, 13, 90]);

      const effectStyles = await optionFor(page, "effect-styles");
      assert.deepEqual(effectStyles.series.map(({ type }) => type), ["effectScatter", "effectScatter"]);
      assert.deepEqual(effectStyles.series.map(({ rippleEffect }) => [rippleEffect.period, rippleEffect.scale, rippleEffect.brushType]), [[4, 10, "stroke"], [3, 6, "fill"]]);
      assert.deepEqual(effectStyles.series[0].data.map(({ value }) => value), [94, 63, 33, 47, 78, 24]);
      assert.deepEqual(effectStyles.series[1].data.map(({ value }) => value), [59, 53, 57, 21, 89, 99]);

      assert.ok((await page.locator('script[src="/charts/assets/js/runtime/echarts/5.6.0/echarts.min.js"]').count()) >= 1);
      assert.ok((await page.locator('script[src="/charts/assets/js/controls/6/controls.js"]').count()) >= 1);
      for (const [variantName, rows] of [["base", 12], ["labels", 12], ["split-lines", 12], ["effect-base", 6], ["effect-styles", 12]]) {
        const current = variant(page, variantName);
        const details = current.locator("details[data-scatter-exact-values]");
        assert.match(await details.locator("summary").textContent(), /Exact scatter values/);
        await details.locator("summary").click();
        assert.equal(await details.locator("tbody tr").count(), rows);
        const wrapper = current.locator("[data-goshtoso-chart-wrapper]");
        assert.equal(await wrapper.locator("button:visible").count(), 4);
        const geometry = await wrapper.evaluate((element) => {
          const host = element.querySelector("[_echarts_instance_]");
          const chart = globalThis.echarts.getInstanceByDom(host);
          return { hostWidth: host.clientWidth, hostHeight: host.clientHeight, chartWidth: chart.getWidth(), chartHeight: chart.getHeight() };
        });
        assert.ok(geometry.hostWidth > 0, JSON.stringify(geometry));
        assert.deepEqual([geometry.chartWidth, geometry.chartHeight], [geometry.hostWidth, geometry.hostHeight]);
        assert.equal(geometry.hostHeight, 420);
        if (process.env.GOSHTOSO_SCREENSHOT_DIR) {
          await fs.mkdir(process.env.GOSHTOSO_SCREENSHOT_DIR, { recursive: true });
          await wrapper.screenshot({ path: path.join(process.env.GOSHTOSO_SCREENSHOT_DIR, `interactive-scatter-${name}-${variantName}.png`) });
        }
      }
      assert.deepEqual(await page.evaluate(() => ({ client: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth })), { client: viewport.width, scroll: viewport.width });
      assert.deepEqual(failures, []);
    } finally {
      await page.close();
    }
  });
}

test("interactive Scatter exports PNG, resizes in place, and keeps wrapper lifecycle modes", async () => {
  const page = await browser.newPage({ viewport: { width: 1440, height: 900 }, acceptDownloads: true });
  const failures = [];
  page.on("console", (message) => { if (message.type() === "error") failures.push(message.text()); });
  page.on("pageerror", (error) => failures.push(error.message));
  try {
    await page.goto(`${baseURL}/components/interactive/scatter`, { waitUntil: "networkidle" });
    const current = variant(page, "base");
    const figure = current.locator("figure");
    const wrapper = current.locator("[data-goshtoso-chart-wrapper]");
    const pending = page.waitForEvent("download", { timeout: 10000 });
    await wrapper.getByRole("button", { name: "Export Basic sports scatter" }).click();
    await wrapper.locator('[id$="-export-png-action"]:visible').first().click();
    const download = await pending;
    assert.equal(download.suggestedFilename(), "basic-sports-scatter.png");
    const bytes = await fs.readFile(await download.path());
    assert.deepEqual([...bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);

    const initial = await figure.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const chart = globalThis.echarts.getInstanceByDom(host);
      element.__scatterHost = host;
      element.__scatterChart = chart;
      return { width: chart.getWidth(), height: chart.getHeight() };
    });
    await page.setViewportSize({ width: 390, height: 844 });
    await page.waitForTimeout(350);
    const resized = await figure.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const chart = globalThis.echarts.getInstanceByDom(host);
      return { width: chart.getWidth(), height: chart.getHeight(), hostWidth: host.clientWidth, hostHeight: host.clientHeight, sameHost: host === element.__scatterHost, sameChart: chart === element.__scatterChart };
    });
    assert.ok(resized.width < initial.width, `${initial.width} -> ${resized.width}`);
    assert.deepEqual([resized.width, resized.height], [resized.hostWidth, resized.hostHeight]);
    assert.equal(resized.sameHost, true);
    assert.equal(resized.sameChart, true);

    for (const mode of ["disabled", "hidden", "enabled"]) {
      await wrapper.evaluate((element, nextMode) => element.dispatchEvent(new CustomEvent("goshtoso-charts:set-wrapper-mode", { bubbles: true, detail: { mode: nextMode } })), mode);
      if (mode === "enabled") await page.waitForTimeout(300);
      const state = await wrapper.evaluate((element) => ({
        mode: element.dataset.goshtosoChartWrapperMode,
        hidden: element.hidden,
        disabled: element.querySelector("[data-goshtoso-chart-actions-fieldset]")?.disabled || false,
        width: globalThis.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")).getWidth(),
      }));
      assert.equal(state.mode, mode);
      assert.equal(state.hidden, mode === "hidden");
      assert.equal(state.disabled, mode === "disabled");
      if (mode === "enabled") assert.ok(state.width > 0);
    }
    assert.deepEqual(failures, []);
  } finally {
    await page.close();
  }
});

test("interactive Scatter effect variant honors reduced motion", async () => {
  const page = await browser.newPage({ viewport: { width: 1440, height: 900 }, reducedMotion: "reduce" });
  try {
    await page.goto(`${baseURL}/components/interactive/scatter`, { waitUntil: "networkidle" });
    await page.waitForTimeout(350);
    for (const name of ["effect-base", "effect-styles"]) {
      const animation = await variant(page, name).locator("[_echarts_instance_]").evaluate((host) => globalThis.echarts.getInstanceByDom(host).getOption().animation);
      assert.equal(animation, false);
    }
  } finally {
    await page.close();
  }
});
