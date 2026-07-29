const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

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
      const response = await fetch(`${baseURL}/components/interactive/bar`);
      if (response.ok && (await response.text()).includes('data-bar-variant="mark-lines"')) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`interactive Bar verification server did not start at ${baseURL}`);
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

async function openBarPage(viewport) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  const failures = [];
  page.on("console", (message) => { if (["error", "warning"].includes(message.type())) failures.push(message.text()); });
  page.on("pageerror", (error) => failures.push(error.message));
  page.on("response", (response) => { if (response.status() >= 400) failures.push(`${response.status()} ${response.url()}`); });
  await page.goto(`${baseURL}/components/interactive/bar`, { waitUntil: "networkidle" });
  await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
  await page.locator('[data-bar-variant="mark-lines"] figure').waitFor();
  return { page, failures };
}

async function chartOptions(page, label) {
  return page.locator(`figure[aria-label="${label}"] div[_echarts_instance_]`).evaluate((node) =>
    globalThis.echarts.getInstanceByDom(node).getOption());
}

test("all feasible upstream Bar treatments render through typed contracts", async () => {
  const { page, failures } = await openBarPage({ width: 1440, height: 900 });
  try {
    assert.equal(await page.locator("figure.goshtoso-charts-interactive").count(), 12);
    assert.equal(await page.locator("[data-bar-exact-values]").count(), 12);

    const basic = await chartOptions(page, "Basic bar example");
    assert.equal(basic.title[0].text, "basic bar example");
    assert.equal(basic.title[0].subtext, "This is the subtitle.");
    assert.equal(basic.tooltip[0].trigger, "axis");
    assert.equal(basic.series.length, 2);

    const labels = await chartOptions(page, "Visible value labels");
    assert.equal(labels.series[0].label.show, true);
    assert.equal(labels.series[0].label.position, "top");

    const axes = await chartOptions(page, "Named axes with literal units");
    assert.equal(axes.xAxis[0].name, "XAxisName");
    assert.equal(axes.yAxis[0].name, "YAxisName");
    assert.equal(axes.xAxis[0].axisLabel.formatter, "{value} x-unit");
    assert.equal(axes.yAxis[0].axisLabel.formatter, "{value} y-unit");
    assert.equal(axes.xAxis[0].splitLine.show, true);
    assert.equal(axes.yAxis[0].splitLine.show, true);

    assert.deepEqual((await chartOptions(page, "Explicit series colors")).color.slice(0, 2), ["#2563eb", "#db2777"]);
    const widths = (await chartOptions(page, "Bar widths and gap")).series;
    assert.equal(widths[0].barWidth, "35");
    assert.equal(widths[1].barWidth, "15%");
    assert.ok(widths.some((series) => series.barGap === "150%"));
    const horizontal = await chartOptions(page, "Horizontal bar orientation");
    assert.deepEqual(horizontal.yAxis[0].data, ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]);
    assert.equal((await chartOptions(page, "Stacked bar series")).series[0].stack, "stackA");

    const insideZoom = await chartOptions(page, "Inside category zoom");
    assert.deepEqual([insideZoom.dataZoom[0].type, insideZoom.dataZoom[0].start, insideZoom.dataZoom[0].end], ["inside", 10, 50]);
    const sliderZoom = await chartOptions(page, "Slider category zoom");
    assert.deepEqual([sliderZoom.dataZoom[0].type, sliderZoom.dataZoom[0].start, sliderZoom.dataZoom[0].end], ["slider", 10, 50]);

    const points = (await chartOptions(page, "Bar point references")).series[0].markPoint.data;
    assert.deepEqual(points.slice(0, 2).map(({ name, type }) => [name, type]), [["Maximum", "max"], ["Minimum", "min"]]);
    assert.equal(points[2].name, "special mark");
    assert.deepEqual(points[2].coord, ["Mon", 100]);
    const lines = (await chartOptions(page, "Bar guide references")).series[0].markLine.data;
    assert.deepEqual(lines.map(({ name, type }) => [name, type]), [["Maximum", "max"], ["Average", "average"]]);

    const largeNode = page.locator('figure[aria-label="Large bar canvas"] div[_echarts_instance_]');
    assert.equal(await largeNode.evaluate((node) => node.style.height), "600px");
    assert.match(await page.locator("main").textContent(), /Mixed-series composition.*renderer-neutral composite chart API/s);
    assert.deepEqual(failures, []);
  } finally {
    await page.close();
  }
});

for (const width of [390, 1440]) {
  for (const [theme, dark] of [["goshtoso", false], ["goshtoso", true], ["araihu", false], ["araihu", true]]) {
    test(`${width}px ${theme} ${dark ? "dark" : "light"} keeps Bar variants themed, centered, and responsive`, async () => {
      const { page, failures } = await openBarPage({ width, height: 900 });
      try {
        await page.evaluate(({ themeName, darkMode }) => {
          document.documentElement.dataset.theme = themeName;
          document.documentElement.classList.toggle("dark", darkMode);
          document.documentElement.dispatchEvent(new CustomEvent("goshtoso-theme-change"));
        }, { themeName: theme, darkMode: dark });
        await page.waitForTimeout(180);
        const state = await page.evaluate(() => {
          const figures = [...document.querySelectorAll("figure.goshtoso-charts-interactive")];
          return {
            pageOverflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
            geometry: figures.map((figure) => {
              const chart = figure.querySelector("div[_echarts_instance_]");
              const host = figure.parentElement;
              const chartBox = chart.getBoundingClientRect();
              const hostBox = host.getBoundingClientRect();
              return {
                width: chartBox.width,
                hostWidth: hostBox.width,
                centerDelta: Math.abs((chartBox.left + chartBox.width / 2) - (hostBox.left + hostBox.width / 2)),
              };
            }),
            backgrounds: figures.map((figure) => globalThis.echarts.getInstanceByDom(figure.querySelector("div[_echarts_instance_]")).getOption().backgroundColor),
          };
        });
        assert.ok(state.pageOverflow <= 1, `page overflow ${state.pageOverflow}`);
        assert.ok(state.backgrounds.every(Boolean));
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
          await page.screenshot({ path: path.join(process.env.GOSHTOSO_SCREENSHOT_DIR, `interactive-bar-${width}-${theme}-${dark ? "dark" : "light"}.png`), fullPage: true });
        }
        assert.deepEqual(failures, []);
      } finally {
        await page.close();
      }
    });
  }
}

test("Bar resizes in place and exports a PNG snapshot", async () => {
  const { page, failures } = await openBarPage({ width: 1440, height: 900 });
  try {
    const figure = page.locator('figure[aria-label="Basic bar example"]');
    const wrapper = figure.locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]");
    const pending = page.waitForEvent("download", { timeout: 10000 });
    await wrapper.getByRole("button", { name: "Download Basic bar example as PNG", exact: true }).click();
    const download = await pending;
    assert.equal(download.suggestedFilename(), "basic-bar-example.png");
    const artifactPath = await download.path();
    assert.ok(artifactPath);
    const bytes = await fs.readFile(artifactPath);
    assert.deepEqual([...bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(bytes).metadata();
    assert.ok(metadata.width > 0 && metadata.height > 0, JSON.stringify(metadata));

    const before = await figure.locator("div[_echarts_instance_]").evaluate((node) => node.getBoundingClientRect().width);
    await page.setViewportSize({ width: 390, height: 844 });
    await page.waitForTimeout(250);
    const afterResize = await figure.locator("div[_echarts_instance_]").evaluate((node) => node.getBoundingClientRect().width);
    assert.ok(afterResize > 0 && afterResize < before, `${before} -> ${afterResize}`);
    assert.deepEqual(failures, []);
  } finally {
    await page.close();
  }
});
