const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
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
      const response = await fetch(`${baseURL}/components/interactive/heatmap`);
      const markup = await response.text();
      if (response.ok && markup.includes('"containLabel":true') && markup.includes('"left":"8"')) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Heatmap verification server did not start at ${baseURL}`);
}

before(async () => {
  const port = await randomPort();
  assert.notEqual(port, 8091);
  baseURL = `http://127.0.0.1:${port}`;
  server = spawn("go", ["run", "./cmd/server", "-port", String(port)], {
    cwd: path.resolve(__dirname, ".."),
    detached: true,
    stdio: "pipe",
  });
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

async function heatmapPage(width) {
  const page = await browser.newPage({ viewport: { width, height: 900 } });
  const failures = [];
  page.on("console", (message) => {
    if (message.type() === "error") failures.push(message.text());
  });
  page.on("pageerror", (error) => failures.push(error.message));
  await page.goto(`${baseURL}/components/interactive/heatmap`);
  const figure = page.locator('figure[aria-label="Deployment activity"]');
  await figure.waitFor();
  await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
  return {
    page,
    failures,
    figure,
    wrapper: figure.locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]"),
  };
}

async function openExpand(wrapper) {
  const trigger = wrapper.locator('[id$="-stacked"] > button:visible, [data-action-group-primary] button:visible').first();
  const stacked = await trigger.evaluate((button) => Boolean(button.closest('[id$="-stacked"]')));
  await trigger.click();
  const action = wrapper.locator('[id$="-chart-expand-action"]').first();
  if (stacked) {
    await action.waitFor({ state: "visible" });
    await action.click();
  }
}

async function measure(figure) {
  return figure.evaluate((element) => {
    const host = element.querySelector("[_echarts_instance_]");
    const chart = window.echarts.getInstanceByDom(host);
    const visualModel = chart.getModel().getComponent("visualMap");
    const visualView = chart.getViewOfComponentModel(visualModel);
    const visual = visualView.group.getBoundingRect().clone();
    visual.applyTransform(visualView.group.getComputedTransform());
    const categoryNames = new Set(["Development", "Staging", "Production"]);
    const labels = chart.getZr().storage.getDisplayList()
      .filter((item) => item.type === "tspan" && categoryNames.has(item.style?.text))
      .map((item) => {
        const bounds = item.getBoundingRect().clone();
        bounds.applyTransform(item.getComputedTransform());
        return bounds;
      });
    const labelBounds = {
      left: Math.min(...labels.map((bounds) => bounds.x)),
      right: Math.max(...labels.map((bounds) => bounds.x + bounds.width)),
      top: Math.min(...labels.map((bounds) => bounds.y)),
      bottom: Math.max(...labels.map((bounds) => bounds.y + bounds.height)),
    };
    const plot = chart.getModel().getComponent("grid").coordinateSystem.getRect();
    const option = chart.getOption();
    const colors = option.visualMap[0].inRange.color;
    return {
      instanceID: chart.id,
      chart: { width: chart.getWidth(), height: chart.getHeight() },
      canvas: {
        width: Math.round(host.querySelector("canvas").getBoundingClientRect().width),
        height: Math.round(host.querySelector("canvas").getBoundingClientRect().height),
      },
      visual: {
        left: visual.x,
        right: visual.x + visual.width,
        top: visual.y,
        bottom: visual.y + visual.height,
      },
      labels: labelBounds,
      margin: labelBounds.left - (visual.x + visual.width),
      plot: { left: plot.x, width: plot.width },
      colors,
      data: option.series[0].data.map((item) => item.value),
      range: [visualModel.option.min, visualModel.option.max],
      pageOverflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
    };
  });
}

function assertLayout(layout, context) {
  assert.ok(layout.margin >= 12, `${context} legend/Y-label margin ${layout.margin}`);
  assert.ok(layout.visual.left >= 0 && layout.visual.top >= 0, `${context} visual scale starts outside chart`);
  assert.ok(layout.visual.right <= layout.chart.width && layout.visual.bottom <= layout.chart.height, `${context} visual scale clipped`);
  assert.ok(layout.labels.left >= 0 && layout.labels.right <= layout.chart.width, `${context} Y labels clipped`);
  assert.ok(layout.labels.top >= 0 && layout.labels.bottom <= layout.chart.height, `${context} Y labels vertically clipped`);
  assert.ok(layout.plot.width >= 160 && layout.plot.width / layout.chart.width >= 0.45, `${context} plot width ${layout.plot.width}/${layout.chart.width}`);
  assert.deepEqual(layout.canvas, layout.chart, `${context} canvas and chart size diverged`);
  assert.deepEqual(layout.range, [0, 20]);
  assert.deepEqual(layout.data, [
    [0, 0, 2], [1, 0, 10], [2, 0, 20],
    [1, 1, 5], [2, 1, 12], [3, 1, 18],
    [2, 2, 0], [3, 2, 8], [4, 2, 16],
  ]);
  assert.equal(layout.colors.length, 3);
  assert.equal(new Set(layout.colors).size, 3);
  assert.equal(layout.pageOverflow, 0);
}

test("test-owned random-port Heatmap route and shared assets stay healthy", async () => {
  assert.notEqual(new URL(baseURL).port, "8091");
  for (const route of [
    "/components/interactive/heatmap",
    "/attributions",
    "/search/assets/search.js",
    "/charts/assets/js/controls/5/controls.js",
    "/assets/styles.css",
  ]) {
    const response = await fetch(`${baseURL}${route}`);
    assert.equal(response.status, 200, route);
  }
});

const combinations = [];
for (const width of [390, 768, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const mode of ["light", "dark"]) combinations.push({ width, theme, mode });
  }
}
combinations.push({ width: 768, theme: "unknown-theme-fallback", mode: "light" });

for (const { width, theme, mode } of combinations) {
  test(`${width}px ${theme} ${mode} keeps Heatmap scale and Y labels separated in page and same-instance modal`, async () => {
    const { page, failures, figure, wrapper } = await heatmapPage(width);
    try {
      await page.evaluate(({ selected, dark }) => {
        document.documentElement.dataset.theme = selected;
        document.documentElement.classList.toggle("dark", dark);
      }, { selected: theme, dark: mode === "dark" });
      await page.waitForTimeout(350);

      const normal = await measure(figure);
      assertLayout(normal, `${width}px ${theme} ${mode} page`);

      await openExpand(wrapper);
      const dialog = wrapper.getByRole("dialog", { name: "Deployment activity" });
      await dialog.waitFor({ state: "visible" });
      await page.waitForTimeout(350);
      const modal = await measure(figure);
      assertLayout(modal, `${width}px ${theme} ${mode} modal`);
      assert.equal(modal.instanceID, normal.instanceID, "modal replaced Heatmap instance");
      const panel = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((element) => {
        const bounds = element.getBoundingClientRect();
        return {
          contained: bounds.left >= 0 && bounds.right <= innerWidth + 1 && bounds.top >= 0 && bounds.bottom <= innerHeight + 1,
          centered: Math.abs((bounds.left + bounds.right) / 2 - innerWidth / 2) < 4,
          width: bounds.width,
        };
      });
      assert.deepEqual({ contained: panel.contained, centered: panel.centered }, { contained: true, centered: true });
      if (width === 1440) assert.ok(panel.width >= 1300, `large modal width ${panel.width}`);
      assert.deepEqual(failures, []);
    } finally {
      await page.close();
    }
  });
}

test("Heatmap PNG snapshot dimensions match the rendered large-modal instance", async () => {
  const { page, figure, wrapper } = await heatmapPage(1440);
  try {
    await openExpand(wrapper);
    await wrapper.getByRole("dialog", { name: "Deployment activity" }).waitFor({ state: "visible" });
    await page.waitForTimeout(350);
    const rendered = await measure(figure);
    const dataURL = await figure.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      return window.echarts.getInstanceByDom(host).getDataURL({
        type: "png",
        pixelRatio: 1,
        backgroundColor: getComputedStyle(element).getPropertyValue("--color-chart-surface").trim() || "#fff",
      });
    });
    const png = Buffer.from(dataURL.split(",", 2)[1], "base64");
    const metadata = await sharp(png).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, rendered.chart);
    assert.deepEqual([...png.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
  } finally {
    await page.close();
  }
});
