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
      const port = listener.address().port;
      listener.close((error) => error ? reject(error) : resolve(port));
    });
  });
}

async function ready() {
  for (let attempt = 0; attempt < 120; attempt += 1) {
    try {
      const response = await fetch(`${baseURL}/components/interactive/line`);
      if (response.ok && (await response.text()).includes('data-line-variant="demo"')) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`interactive Line verification server did not start at ${baseURL}`);
}

before(async () => {
  const port = await randomPort();
  assert.notEqual(port, 8091);
  assert.notEqual(port, 8096);
  baseURL = `http://127.0.0.1:${port}`;
  server = spawn("/opt/homebrew/bin/go", ["run", "./cmd/server", "-port", String(port)], {
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

async function openLinePage(viewport) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  const failures = [];
  page.on("console", (message) => { if (message.type() === "error") failures.push(message.text()); });
  page.on("pageerror", (error) => failures.push(error.message));
  page.on("response", (response) => { if (response.status() >= 400) failures.push(`${response.status()} ${response.url()}`); });
  await page.goto(`${baseURL}/components/interactive/line`, { waitUntil: "networkidle" });
  await page.locator('[data-line-variant="demo"] figure').waitFor();
  return { page, failures };
}

async function chartOptions(page, label) {
  return page.locator(`figure[aria-label="${label}"] div[_echarts_instance_]`).evaluate((node) =>
    globalThis.echarts.getInstanceByDom(node).getOption());
}

test("all feasible upstream Line treatments render through typed contracts", async () => {
  const { page, failures } = await openLinePage({ width: 1440, height: 900 });
  try {
    assert.equal(await page.locator("figure.goshtoso-charts-interactive").count(), 13);
    assert.equal(await page.locator("[data-line-exact-values]").count(), 13);
    for (const label of [
      "Basic line example", "Visible point labels", "Diamond symbols", "Calculated point references",
      "Visible split lines", "Numerical x axis with guides", "Step line", "Smooth line", "Area line",
      "Smooth area", "Four line comparison", "Search time comparison", "Temporal X axis",
    ]) {
      assert.equal(await page.locator(`figure[aria-label="${label}"]`).count(), 1, label);
	  assert.equal((await chartOptions(page, label)).legend[0].bottom, "0", `${label} responsive legend`);
    }

    const numerical = await chartOptions(page, "Numerical x axis with guides");
    assert.equal(numerical.xAxis[0].type, "value");
    assert.equal(numerical.visualMap[0].type, "piecewise");
    assert.deepEqual(numerical.visualMap[0].pieces.map(({ gt, lt }) => [gt, lt]), [[1, 7], [10, 15]]);
    assert.equal(numerical.series[0].markArea.data[0][0].name, "Danger zone");
    assert.equal(numerical.series[0].markLine.data[0][0].name, "Danger level");
    assert.equal(numerical.series[0].markLine.data[1].name, "Line of no return");

    const references = await chartOptions(page, "Calculated point references");
    assert.deepEqual(references.series[0].markPoint.data.map(({ name, type }) => [name, type]), [
      ["Maximum", "max"], ["Average", "average"], ["Minimum", "min"],
    ]);
    const step = await chartOptions(page, "Step line");
    assert.equal(step.series[0].step, "end");
    assert.equal((await chartOptions(page, "Smooth line")).series[0].smooth, true);
    assert.equal((await chartOptions(page, "Area line")).series[0].areaStyle.opacity, 0.5);
    const smoothArea = (await chartOptions(page, "Smooth area")).series[0];
    assert.equal(smoothArea.smooth, true);
    assert.equal(smoothArea.areaStyle.opacity, 0.2);
    assert.equal((await chartOptions(page, "Four line comparison")).series.length, 4);
    assert.equal((await chartOptions(page, "Search time comparison")).series.length, 2);
    assert.match(await page.locator("main").textContent(), /Mixed-series composition.*separate renderer-neutral composite chart API/s);
    assert.deepEqual(failures, []);
  } finally {
    await page.close();
  }
});

for (const width of [390, 1440]) {
  for (const [theme, dark] of [["goshtoso", false], ["goshtoso", true], ["araihu", false], ["araihu", true]]) {
    test(`${width}px ${theme} ${dark ? "dark" : "light"} keeps Line variants themed, centered, and responsive`, async () => {
      const { page, failures } = await openLinePage({ width, height: 900 });
      try {
        await page.evaluate(({ themeName, darkMode }) => {
          document.documentElement.dataset.theme = themeName;
          document.documentElement.classList.toggle("dark", darkMode);
          document.documentElement.dispatchEvent(new CustomEvent("goshtoso-theme-change"));
        }, { themeName: theme, darkMode: dark });
        await page.waitForTimeout(150);
        const state = await page.evaluate(() => {
          const figures = [...document.querySelectorAll("figure.goshtoso-charts-interactive")];
          const geometry = figures.map((figure) => {
            const chart = figure.querySelector("div[_echarts_instance_]");
            const host = figure.parentElement;
            const chartBox = chart.getBoundingClientRect();
            const hostBox = host.getBoundingClientRect();
            return {
              width: chartBox.width,
              hostWidth: hostBox.width,
              centerDelta: Math.abs((chartBox.left + chartBox.width / 2) - (hostBox.left + hostBox.width / 2)),
            };
          });
          const option = (label) => {
            const node = document.querySelector(`figure[aria-label="${label}"] div[_echarts_instance_]`);
            return globalThis.echarts.getInstanceByDom(node).getOption();
          };
          const numerical = option("Numerical x axis with guides").series[0];
          const points = option("Calculated point references").series[0];
          return {
            geometry,
            pageOverflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
            background: option("Basic line example").backgroundColor,
            rangeColor: numerical.markArea.itemStyle?.color || "",
            rangeOpacity: numerical.markArea.itemStyle?.opacity,
            guideColor: numerical.markLine.lineStyle?.color || "",
            pointColor: points.markPoint.itemStyle?.color || "",
          };
        });
        assert.ok(state.background);
        assert.ok(state.rangeColor, "marked range must use theme scale color");
        assert.ok(state.rangeOpacity > 0 && state.rangeOpacity <= 0.25, String(state.rangeOpacity));
        assert.ok(state.guideColor, "guide lines must use theme outline color");
        assert.ok(state.pointColor, "reference points must use series palette color");
        assert.ok(state.pageOverflow <= 1, `page overflow ${state.pageOverflow}`);
        for (const item of state.geometry) {
          assert.ok(item.width > 0 && item.width <= item.hostWidth + 1, JSON.stringify(item));
          assert.ok(item.centerDelta <= 1, JSON.stringify(item));
        }
        for (const wrapper of await page.locator("[data-goshtoso-chart-wrapper]").all()) {
          assert.ok(await wrapper.getByRole("button", { name: /Expand/ }).count() >= 1);
          assert.equal(await wrapper.locator("[data-goshtoso-chart-export-status]").count(), 1);
        }
        if (process.env.GOSHTOSO_SCREENSHOT_DIR) {
          await fs.mkdir(process.env.GOSHTOSO_SCREENSHOT_DIR, { recursive: true });
          await page.screenshot({ path: path.join(process.env.GOSHTOSO_SCREENSHOT_DIR, `interactive-line-${width}-${theme}-${dark ? "dark" : "light"}.png`), fullPage: true });
        }
        assert.deepEqual(failures, []);
      } finally {
        await page.close();
      }
    });
  }
}
