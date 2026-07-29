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
const goExecutable = "go";

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
      const response = await fetch(`${baseURL}/components/interactive/radar`);
      if (response.ok && (await response.text()).includes('data-radar-variant="legend-single"')) return;
    } catch { /* test-owned server still starting */ }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`interactive Radar verification server did not start at ${baseURL}`);
}

async function spawnServer(port, executable = goExecutable) {
  return new Promise((resolve, reject) => {
    const child = spawn(executable, ["run", "./cmd/server", "-port", String(port)], {
      cwd: path.resolve(__dirname, ".."), detached: true, stdio: "pipe",
    });
    child.once("error", reject);
    child.once("spawn", () => resolve(child));
  });
}

before(async () => {
  const port = await randomPort();
  assert.notEqual(port, 8091);
  assert.notEqual(port, 8096);
  baseURL = `http://127.0.0.1:${port}`;
  server = await spawnServer(port);
  await ready();
  browser = await chromium.launch({ headless: true });
});

test("interactive Radar browser harness resolves Go from PATH", () => {
  assert.equal(path.isAbsolute(goExecutable), false);
  assert.equal(goExecutable, "go");
});

test("interactive Radar browser harness reports an unavailable Go executable", async () => {
  await assert.rejects(spawnServer(0, "goshtoso-charts-missing-go-executable"), { code: "ENOENT" });
});

after(async () => {
  await browser?.close();
  if (server?.pid) {
    try { process.kill(-server.pid, "SIGTERM"); } catch { /* already stopped */ }
  }
});

function variant(page, name) {
  return page.locator(`[data-radar-variant="${name}"]`);
}

async function optionFor(page, name) {
  return variant(page, name).locator("[_echarts_instance_]").evaluate((host) => globalThis.echarts.getInstanceByDom(host).getOption());
}

for (const [name, viewport, theme, dark] of [
  ["wide-light", { width: 1440, height: 900 }, "goshtoso", false],
  ["wide-dark", { width: 1440, height: 900 }, "araihu", true],
  ["narrow-light", { width: 390, height: 844 }, "araihu", false],
  ["narrow-dark", { width: 390, height: 844 }, "goshtoso", true],
]) {
  test(`interactive Radar ${name} preserves upstream treatments and responsive theme layout`, async () => {
    const page = await browser.newPage({ viewport, colorScheme: dark ? "dark" : "light" });
    const failures = [];
    page.on("console", (message) => { if (message.type() === "error") failures.push(message.text()); });
    page.on("pageerror", (error) => failures.push(error.message));
    page.on("response", (response) => { if (response.status() >= 400) failures.push(`${response.status()} ${response.url()}`); });
    try {
      await page.goto(`${baseURL}/components/interactive/radar`, { waitUntil: "networkidle" });
      await page.evaluate(({ selected, darkMode }) => {
        document.documentElement.dataset.theme = selected;
        document.documentElement.classList.toggle("dark", darkMode);
      }, { selected: theme, darkMode: dark });
      await page.waitForFunction(() => {
        const hosts = [...document.querySelectorAll("[data-radar-variant] [_echarts_instance_]")];
        return hosts.length === 4 && hosts.every((host) => Boolean(globalThis.echarts?.getInstanceByDom(host)));
      });
      await page.waitForTimeout(350);

      const base = await optionFor(page, "base");
      assert.equal(base.series.length, 1);
      assert.equal(base.series[0].name, "Beijing");
      assert.equal(base.series[0].data.length, 21);
      assert.deepEqual(base.series[0].data[0].value, [55, 9, 56, 0.46, 18, 6]);
      assert.deepEqual(base.radar[0].indicator.map(({ name, max }) => [name, max]), [
        ["AQI", 300], ["PM2.5", 250], ["PM10", 300], ["CO", 5], ["NO2", 200], ["SO2", 100],
      ]);
      assert.equal(base.radar[0].splitArea.show, true);
      assert.equal(base.radar[0].splitLine.show, true);

      const style = await optionFor(page, "style");
      assert.equal(style.radar[0].shape, "circle");
      assert.equal(style.radar[0].splitNumber, 5);
      assert.equal(style.radar[0].splitLine.lineStyle.opacity, 0.1);
      assert.equal(style.series[0].lineStyle.opacity, 0.5);
      assert.equal(style.series[0].areaStyle.opacity, 0.1);

      const multiple = await optionFor(page, "legend-multiple");
      const single = await optionFor(page, "legend-single");
      assert.deepEqual(multiple.series.map(({ name, data }) => [name, data.length]), [["Beijing", 21], ["Guangzhou", 21], ["Shanghai", 21]]);
      assert.equal(multiple.legend[0].selectedMode, "multiple");
      assert.equal(single.legend[0].selectedMode, "single");
      assert.equal(single.series[0].areaStyle.opacity, 0.5);

      for (const [variantName, rows] of [["base", 21], ["style", 21], ["legend-multiple", 63], ["legend-single", 63]]) {
        const current = variant(page, variantName);
        const wrapper = current.locator("[data-goshtoso-chart-wrapper]");
        const details = current.locator("details[data-radar-exact-values]");
        assert.match(await details.locator("summary").textContent(), /Exact radar values/);
        if (variantName === "legend-multiple" && process.env.GOSHTOSO_SCREENSHOT_DIR) {
          await fs.mkdir(process.env.GOSHTOSO_SCREENSHOT_DIR, { recursive: true });
          await wrapper.screenshot({ path: path.join(process.env.GOSHTOSO_SCREENSHOT_DIR, `interactive-radar-${name}.png`) });
        }
        await details.locator("summary").click();
        assert.equal(await details.locator("tbody tr").count(), rows);
        const geometry = await wrapper.evaluate((element) => {
          const content = element.querySelector("[data-goshtoso-chart-content]");
          const host = element.querySelector("[_echarts_instance_]");
          const chart = globalThis.echarts.getInstanceByDom(host);
          const hostRect = host.getBoundingClientRect();
          const contentRect = content.getBoundingClientRect();
          return {
            hostWidth: host.clientWidth, hostHeight: host.clientHeight,
            chartWidth: chart.getWidth(), chartHeight: chart.getHeight(),
            centered: Math.abs((hostRect.left + hostRect.right) / 2 - (contentRect.left + contentRect.right) / 2) < 2,
            colors: chart.getOption().color,
            background: chart.getOption().backgroundColor,
          };
        });
        assert.deepEqual([geometry.chartWidth, geometry.chartHeight], [geometry.hostWidth, geometry.hostHeight]);
        assert.equal(geometry.hostHeight, 520);
        assert.equal(geometry.centered, true);
        assert.ok(geometry.hostWidth > 0 && geometry.hostWidth <= 1024, JSON.stringify(geometry));
        assert.ok(geometry.colors.length >= 3, JSON.stringify(geometry));
        assert.ok(geometry.background, JSON.stringify(geometry));
      }

      assert.equal(await variant(page, "base").locator("tbody tr").last().textContent(), "BeijingDay 213915360.612913");
      assert.deepEqual(await page.evaluate(() => ({ client: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth })), { client: viewport.width, scroll: viewport.width });
      assert.deepEqual(failures, []);
    } finally {
      await page.close();
    }
  });
}

test("interactive Radar exports PNG, resizes in place, and responds to wrapper lifecycle events", async () => {
  const page = await browser.newPage({ viewport: { width: 1280, height: 900 }, acceptDownloads: true });
  try {
    await page.goto(`${baseURL}/components/interactive/radar`, { waitUntil: "networkidle" });
    const current = variant(page, "base");
    const wrapper = current.locator("[data-goshtoso-chart-wrapper]");
    await page.waitForFunction(() => Boolean(globalThis.echarts?.getInstanceByDom(document.querySelector('[data-radar-variant="base"] [_echarts_instance_]'))));
    const before = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      element.__radarInstance = globalThis.echarts.getInstanceByDom(host);
      return element.__radarInstance.getWidth();
    });
    await page.setViewportSize({ width: 900, height: 900 });
    await page.waitForFunction((oldWidth) => {
      const wrapper = document.querySelector('[data-radar-variant="base"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const chart = globalThis.echarts.getInstanceByDom(host);
      return chart === wrapper.__radarInstance && chart.getWidth() !== oldWidth && chart.getWidth() === host.clientWidth;
    }, before);

    for (const mode of ["disabled", "hidden", "enabled"]) {
      await wrapper.evaluate((element, nextMode) => element.dispatchEvent(new CustomEvent("goshtoso-charts:set-wrapper-mode", { bubbles: true, detail: { mode: nextMode } })), mode);
      await page.waitForFunction((nextMode) => document.querySelector('[data-radar-variant="base"] [data-goshtoso-chart-wrapper]').dataset.goshtosoChartWrapperMode === nextMode, mode);
      assert.equal(await wrapper.evaluate((element) => globalThis.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")) === element.__radarInstance), true);
    }

    const pending = page.waitForEvent("download", { timeout: 10000 });
    await wrapper.getByRole("button", { name: "Download Daily Beijing air quality as PNG", exact: true }).click();
    const download = await pending;
    assert.equal(download.suggestedFilename(), "daily-beijing-air-quality.png");
  } finally {
    await page.close();
  }
});
