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
      const response = await fetch(`${baseURL}/components/interactive/pie`);
      if (response.ok && (await response.text()).includes('data-pie-variant="rose-area"')) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`interactive Pie verification server did not start at ${baseURL}`);
}

before(async () => {
  const port = await randomPort();
  assert.notEqual(port, 8091);
  assert.notEqual(port, 8096);
  baseURL = `http://127.0.0.1:${port}`;
  server = spawn("go", ["run", "./cmd/server", "-port", String(port)], {
    cwd: path.resolve(__dirname, ".."), detached: true, stdio: "pipe",
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
  return page.locator(`[data-pie-variant="${name}"]`);
}

function contrast(first, second) {
  const luminance = (rgb) => {
    const linear = rgb.map((value) => value / 255).map((value) => value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4);
    return 0.2126 * linear[0] + 0.7152 * linear[1] + 0.0722 * linear[2];
  };
  const values = [luminance(first), luminance(second)].sort((left, right) => right - left);
  return (values[0] + 0.05) / (values[1] + 0.05);
}

async function optionFor(page, name) {
  return variant(page, name).locator("[_echarts_instance_]").evaluate((host) => {
    return globalThis.echarts.getInstanceByDom(host).getOption();
  });
}

for (const [name, viewport, theme, dark] of [
  ["wide-light", { width: 1440, height: 900 }, "goshtoso", false],
  ["wide-dark", { width: 1440, height: 900 }, "goshtoso", true],
  ["narrow-light", { width: 390, height: 844 }, "araihu", false],
  ["narrow-dark", { width: 390, height: 844 }, "araihu", true],
]) {
  test(`interactive Pie ${name} preserves variants, exact values, theme, and layout`, async () => {
    const page = await browser.newPage({ viewport, colorScheme: dark ? "dark" : "light" });
    const failures = [];
    page.on("console", (message) => { if (message.type() === "error") failures.push(message.text()); });
    page.on("pageerror", (error) => failures.push(error.message));
    page.on("response", (response) => { if (response.status() >= 400) failures.push(`${response.status()} ${response.url()}`); });
    try {
      await page.goto(`${baseURL}/components/interactive/pie`, { waitUntil: "networkidle" });
      await page.evaluate(({ selected, darkMode }) => {
        document.documentElement.dataset.theme = selected;
        document.documentElement.classList.toggle("dark", darkMode);
      }, { selected: theme, darkMode: dark });
      await page.waitForTimeout(350);
      await page.waitForFunction(() => {
        const hosts = [...document.querySelectorAll("[data-pie-variant] [_echarts_instance_]")];
        return hosts.length === 4 && hosts.every((host) => Boolean(globalThis.echarts?.getInstanceByDom(host)));
      });

      const base = await optionFor(page, "base");
      assert.equal(base.series.length, 1);
      assert.deepEqual(base.series[0].radius, ["0%", "75%"]);
      assert.deepEqual(base.series[0].data.map(({ name, value }) => [name, value]), [
        ["Spring", 81], ["Summer", 87], ["Autumn", 47], ["Winter", 59],
      ]);

      const area = await optionFor(page, "rose-area");
      assert.equal(area.legend[0].bottom, "0");
      assert.equal(area.series[0].roseType, "area");
      assert.deepEqual(area.series[0].radius, ["40%", "75%"]);
      assert.deepEqual(area.series[0].center, ["50%", "50%"]);
      assert.equal(area.series[0].label.formatter, "{b}: {c}");

      const radius = await optionFor(page, "rose-radius");
      assert.equal(radius.legend[0].bottom, "0");
      assert.equal(radius.series[0].roseType, "radius");
      assert.deepEqual(radius.series[0].radius, ["30%", "75%"]);
      assert.deepEqual(radius.series[0].center, ["50%", "50%"]);
      assert.equal(radius.series[0].label.formatter, "{b}: {c}");

      const nested = await optionFor(page, "nested");
      assert.equal(nested.legend[0].bottom, "0");
      assert.deepEqual(nested.series.map((series) => series.roseType), ["area", "radius"]);
      assert.deepEqual(nested.series.map((series) => series.radius), [["50%", "55%"], ["0%", "45%"]]);
      assert.deepEqual(nested.series.map((series) => series.center), [["50%", "50%"], ["50%", "50%"]]);

      for (const [variantName, rows] of [["base", 4], ["rose-area", 4], ["rose-radius", 4], ["nested", 8]]) {
        const current = variant(page, variantName);
        const details = current.locator("details");
        assert.match(await details.locator("summary").textContent(), /Exact pie values/);
        if (process.env.GOSHTOSO_SCREENSHOT_DIR) {
          await fs.mkdir(process.env.GOSHTOSO_SCREENSHOT_DIR, { recursive: true });
          await current.locator("[data-goshtoso-chart-wrapper]").screenshot({
            path: path.join(process.env.GOSHTOSO_SCREENSHOT_DIR, `interactive-pie-${name}-${variantName}.png`),
          });
        }
        await details.locator("summary").click();
        assert.equal(await details.locator("tbody tr").count(), rows);
        const wrapper = current.locator("[data-goshtoso-chart-wrapper]");
        assert.equal(await wrapper.locator("button:visible").count(), 2);
        const layout = await wrapper.evaluate((element) => {
          const content = element.querySelector("[data-goshtoso-chart-content]");
          const host = element.querySelector("[_echarts_instance_]");
          const instance = globalThis.echarts.getInstanceByDom(host);
          const hostRect = host.getBoundingClientRect();
          const contentRect = content.getBoundingClientRect();
          const option = instance.getOption();
          const resolve = (value) => {
            const canvas = document.createElement("canvas");
            canvas.width = 1;
            canvas.height = 1;
            const context = canvas.getContext("2d");
            context.fillStyle = value;
            context.fillRect(0, 0, 1, 1);
            return [...context.getImageData(0, 0, 1, 1).data.slice(0, 3)];
          };
          const figure = element.querySelector("figure");
          const probe = document.createElement("span");
          probe.style.color = "var(--color-chart-text)";
          probe.style.backgroundColor = "var(--color-chart-surface)";
          figure.appendChild(probe);
          const computed = getComputedStyle(probe);
          const surface = resolve(computed.backgroundColor);
          const text = resolve(computed.color);
          probe.remove();
          return {
            hostWidth: host.clientWidth,
            hostHeight: host.clientHeight,
            chartWidth: instance.getWidth(),
            chartHeight: instance.getHeight(),
            centered: Math.abs((hostRect.left + hostRect.right) / 2 - (contentRect.left + contentRect.right) / 2) < 2,
            surface,
            text,
            firstSeries: resolve(option.color[0]),
            colors: option.color,
          };
        });
        assert.deepEqual([layout.chartWidth, layout.chartHeight], [layout.hostWidth, layout.hostHeight]);
        assert.equal(layout.centered, true);
        assert.equal(layout.hostHeight, variantName === "nested" ? 440 : 420, JSON.stringify(layout));
        assert.ok(layout.hostWidth > 0, JSON.stringify(layout));
        assert.ok(layout.colors.length >= 4);
        assert.ok(contrast(layout.surface, layout.text) >= 4.5, JSON.stringify(layout));
        assert.ok(contrast(layout.surface, layout.firstSeries) >= 2, JSON.stringify(layout));
      }

      assert.deepEqual(await page.evaluate(() => ({ client: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth })), { client: viewport.width, scroll: viewport.width });
      assert.equal(await variant(page, "rose-radius").locator("tbody tr").first().textContent(), "RadiusSpring9538.5%");
      assert.equal(await variant(page, "nested").locator("tbody tr").last().textContent(), "Inner radiusWinter2614.5%");
      assert.deepEqual(failures, []);
    } finally {
      await page.close();
    }
  });
}
