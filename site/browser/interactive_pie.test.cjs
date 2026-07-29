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
        return hosts.length === 10 && hosts.every((host) => Boolean(globalThis.echarts?.getInstanceByDom(host)));
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

	  const labels = await optionFor(page, "labels");
	  assert.equal(labels.series[0].label.show, true);
	  assert.equal(labels.series[0].label.formatter, "{b}: {c}");
	  assert.deepEqual(labels.series[0].data.map(({ value }) => value), [81, 18, 25, 40]);

	  const donut = await optionFor(page, "radius");
	  assert.deepEqual(donut.series[0].radius, ["40%", "75%"]);
	  assert.deepEqual(donut.series[0].data.map(({ value }) => value), [56, 0, 94, 11]);

	  const padded = await optionFor(page, "padded");
	  assert.equal(padded.series[0].padAngle, 5);
	  assert.deepEqual(padded.series[0].center, ["40%", "50%"]);
	  assert.deepEqual(padded.legend[0].padding, [1, 1, 1, 1]);
	  assert.equal(padded.legend[0].orient, "vertical");
	  assert.equal(padded.legend[0].right, "20%");
	  assert.equal(padded.series[0].label.show, false);
	  assert.equal(padded.tooltip[0].formatter, "{b}: {d}%");

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

	  const paired = await optionFor(page, "paired-roses");
	  assert.deepEqual(paired.series.map((series) => series.center), [["25%", "50%"], ["75%", "50%"]]);
	  assert.deepEqual(paired.series.map((series) => series.roseType), ["area", "radius"]);

	  const automatic = await optionFor(page, "auto-emphasis");
	  assert.deepEqual(automatic.series[0].radius, ["0%", "55%"]);
	  assert.deepEqual(automatic.series[0].center, ["50%", "60%"]);
	  assert.equal(automatic.series[0].emphasis.itemStyle.shadowBlur, 10);
	  assert.equal(automatic.tooltip[0].formatter, "{b}: {d}%");

	  const selected = await optionFor(page, "selected");
	  assert.equal(selected.series[0].selectedMode, true);
	  assert.equal(selected.series[0].data[0].selected, true);

	  for (const [variantName, rows] of [
		["base", 4], ["labels", 4], ["radius", 4], ["padded", 4], ["rose-area", 4],
		["rose-radius", 4], ["paired-roses", 8], ["nested", 8], ["auto-emphasis", 4], ["selected", 4],
	  ]) {
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
		assert.equal(layout.hostHeight, ["nested", "paired-roses", "auto-emphasis"].includes(variantName) ? 440 : 420, JSON.stringify(layout));
        assert.ok(layout.hostWidth > 0, JSON.stringify(layout));
        assert.ok(layout.colors.length >= 4);
        assert.ok(contrast(layout.surface, layout.text) >= 4.5, JSON.stringify(layout));
        assert.ok(contrast(layout.surface, layout.firstSeries) >= 2, JSON.stringify(layout));
      }

      assert.deepEqual(await page.evaluate(() => ({ client: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth })), { client: viewport.width, scroll: viewport.width });
      assert.equal(await variant(page, "rose-radius").locator("tbody tr").first().textContent(), "RadiusSpring9538.5%");
      assert.equal(await variant(page, "nested").locator("tbody tr").last().textContent(), "Inner radiusWinter2614.5%");
	  if (name === "wide-light") {
		const autoFigure = variant(page, "auto-emphasis").locator("figure");
		const actionTypes = await autoFigure.evaluate(async (figure) => {
		  const host = figure.querySelector("[_echarts_instance_]");
		  const chart = globalThis.echarts.getInstanceByDom(host);
		  const original = chart.dispatchAction.bind(chart);
		  const actions = [];
		  chart.dispatchAction = (action) => { actions.push(action.type); return original(action); };
		  await new Promise((resolve) => setTimeout(resolve, 1150));
		  chart.dispatchAction = original;
		  return actions;
		});
		assert.ok(actionTypes.includes("highlight"), JSON.stringify(actionTypes));
		assert.ok(actionTypes.includes("showTip"), JSON.stringify(actionTypes));
	  }
      assert.deepEqual(failures, []);
    } finally {
      await page.close();
    }
  });
}

test("interactive Pie exports PNG, resizes in place, and keeps wrapper lifecycle modes", async () => {
  const page = await browser.newPage({ viewport: { width: 1440, height: 900 }, acceptDownloads: true });
  const failures = [];
  page.on("console", (message) => { if (message.type() === "error") failures.push(message.text()); });
  page.on("pageerror", (error) => failures.push(error.message));
  try {
	await page.goto(`${baseURL}/components/interactive/pie`, { waitUntil: "networkidle" });
	const current = variant(page, "selected");
	const figure = current.locator("figure");
	const wrapper = current.locator("[data-goshtoso-chart-wrapper]");
	await page.waitForFunction(() => Boolean(globalThis.echarts?.getInstanceByDom(document.querySelector('[data-pie-variant="selected"] [_echarts_instance_]'))));
	const pending = page.waitForEvent("download", { timeout: 10000 });
	await wrapper.getByRole("button", { name: "Download Selectable seasonal donut as PNG", exact: true }).click();
	const download = await pending;
	assert.equal(download.suggestedFilename(), "selected-seasonal-sector.png");
	const bytes = await fs.readFile(await download.path());
	assert.deepEqual([...bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);

	const initial = await figure.evaluate((element) => {
	  const host = element.querySelector("[_echarts_instance_]");
	  const chart = globalThis.echarts.getInstanceByDom(host);
	  element.__pieHost = host;
	  element.__pieChart = chart;
	  return { width: chart.getWidth(), height: chart.getHeight() };
	});
	await page.setViewportSize({ width: 390, height: 844 });
	await page.waitForTimeout(350);
	const resized = await figure.evaluate((element) => {
	  const host = element.querySelector("[_echarts_instance_]");
	  const chart = globalThis.echarts.getInstanceByDom(host);
	  return {
		width: chart.getWidth(), height: chart.getHeight(), hostWidth: host.clientWidth, hostHeight: host.clientHeight,
		sameHost: host === element.__pieHost, sameChart: chart === element.__pieChart,
	  };
	});
	assert.ok(resized.width < initial.width, `${initial.width} -> ${resized.width}`);
	assert.deepEqual({ width: resized.width, height: resized.height }, { width: resized.hostWidth, height: resized.hostHeight });
	assert.equal(resized.sameHost, true);
	assert.equal(resized.sameChart, true);

	for (const mode of ["disabled", "hidden", "enabled"]) {
	  await wrapper.evaluate((element, nextMode) => element.dispatchEvent(new CustomEvent("goshtoso-charts:set-wrapper-mode", {
		bubbles: true, detail: { mode: nextMode },
	  })), mode);
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
	assert.deepEqual(await page.evaluate(() => ({ client: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth })), { client: 390, scroll: 390 });
	assert.deepEqual(failures, []);
  } finally {
	await page.close();
  }
});

test("interactive Pie rotating emphasis honors reduced motion", async () => {
  const page = await browser.newPage({ viewport: { width: 1440, height: 900 }, reducedMotion: "reduce" });
  try {
	await page.goto(`${baseURL}/components/interactive/pie`, { waitUntil: "networkidle" });
	const actions = await variant(page, "auto-emphasis").locator("figure").evaluate(async (figure) => {
	  const chart = globalThis.echarts.getInstanceByDom(figure.querySelector("[_echarts_instance_]"));
	  const original = chart.dispatchAction.bind(chart);
	  const captured = [];
	  chart.dispatchAction = (action) => { captured.push(action.type); return original(action); };
	  await new Promise((resolve) => setTimeout(resolve, 1150));
	  chart.dispatchAction = original;
	  return captured;
	});
	assert.equal(actions.includes("highlight"), false, JSON.stringify(actions));
	assert.equal(actions.includes("showTip"), false, JSON.stringify(actions));
  } finally {
	await page.close();
  }
});
