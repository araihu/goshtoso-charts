const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

let baseURL;
let browser;
let server;

const screenshotDirectory = process.env.GOSHTOSO_SCREENSHOT_DIR;
if (screenshotDirectory) fs.mkdirSync(screenshotDirectory, { recursive: true });

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
      if (response.ok && markup.includes("Weekly activity by hour") && markup.includes("Calendar activity")) return;
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
  const page = await browser.newPage({ viewport: { width, height: 1000 }, acceptDownloads: true });
  const failures = [];
  page.on("console", (message) => {
    if (message.type() === "error") failures.push(message.text());
  });
  page.on("pageerror", (error) => failures.push(error.message));
  await page.goto(`${baseURL}/components/interactive/heatmap`);
  await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
  const category = page.locator('figure[aria-label="Weekly activity by hour"]');
  const calendar = page.locator('figure[aria-label="Calendar activity"]');
  await category.waitFor();
  await calendar.waitFor();
  return { page, failures, category, calendar };
}

function wrapperFor(figure) {
  return figure.locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]");
}

async function openExpand(wrapper) {
  const trigger = wrapper.locator('[id$="-stacked"] > button:visible, [data-action-group-primary] button:visible').first();
  const stacked = await trigger.evaluate((button) => Boolean(button.closest('[id$="-stacked"]')));
  await trigger.click();
  if (stacked) {
    const action = wrapper.locator('[id$="-chart-expand-action"]').first();
    await action.waitFor({ state: "visible" });
    await action.click();
  }
}

async function enterFullscreen(wrapper) {
  await wrapper.locator('[id$="-stacked"] > button:visible, [data-action-group-primary] button:visible').first().click();
  const action = wrapper.locator('[id$="-fullscreen-action"]').first();
  await action.waitFor({ state: "visible" });
  await action.click();
}

async function closeExpand(page, wrapper, label) {
  await page.keyboard.press("Escape");
  await wrapper.getByRole("dialog", { name: label }).waitFor({ state: "hidden" });
}

async function waitForChartGeometry(figure) {
  await figure.evaluate((element) => new Promise((resolve) => {
    let last = "";
    let stableFrames = 0;
    const check = () => {
      const host = element.querySelector("[_echarts_instance_]");
      const chart = host && window.echarts.getInstanceByDom(host);
      const canvas = host && host.querySelector("canvas");
      if (!chart || !canvas) {
        requestAnimationFrame(check);
        return;
      }
      const bounds = canvas.getBoundingClientRect();
      const key = `${chart.getWidth()}x${chart.getHeight()}:${Math.round(bounds.width)}x${Math.round(bounds.height)}`;
      stableFrames = key === last && chart.getWidth() === Math.round(bounds.width) && chart.getHeight() === Math.round(bounds.height)
        ? stableFrames + 1
        : 0;
      last = key;
      if (stableFrames >= 2) resolve();
      else requestAnimationFrame(check);
    };
    requestAnimationFrame(check);
  }));
}

async function waitForDialogSettled(dialog) {
  const panel = dialog.locator(".goshtoso-charts-expand-panel");
  await panel.evaluate((element) => new Promise((resolve) => {
    let stableFrames = 0;
    const check = () => {
      const style = getComputedStyle(element);
      const matrix = style.transform === "none" ? new DOMMatrixReadOnly() : new DOMMatrixReadOnly(style.transform);
      const settled = Math.abs(matrix.a - 1) < 0.001 && Math.abs(matrix.d - 1) < 0.001 && Number(style.opacity) > 0.999;
      stableFrames = settled ? stableFrames + 1 : 0;
      if (stableFrames >= 2) resolve();
      else requestAnimationFrame(check);
    };
    requestAnimationFrame(check);
  }));
}

async function measure(figure) {
  return figure.evaluate((element) => {
    const host = element.querySelector("[_echarts_instance_]");
    const chart = window.echarts.getInstanceByDom(host);
    const option = chart.getOption();
    const visualModel = chart.getModel().getComponent("visualMap");
    const visualView = chart.getViewOfComponentModel(visualModel);
    const visual = visualView.group.getBoundingRect().clone();
    visual.applyTransform(visualView.group.getComputedTransform());
    const componentBounds = (type) => {
      const model = chart.getModel().getComponent(type);
      if (!model) return null;
      const view = chart.getViewOfComponentModel(model);
      const bounds = view.group.getBoundingRect().clone();
      bounds.applyTransform(view.group.getComputedTransform());
      return { left: bounds.x, right: bounds.x + bounds.width, top: bounds.y, bottom: bounds.y + bounds.height };
    };
    const titleBounds = componentBounds("title");
    const legendBounds = componentBounds("legend");
    const coordinate = option.series[0].coordinateSystem || "cartesian2d";
    const data = option.series[0].data.map((item) => item.value);
    const missing = data.filter((value) => value[value.length - 1] === "-").length;
    let coordinateRect;
    if (coordinate === "calendar") {
      const rect = chart.getModel().getComponent("calendar").coordinateSystem.getRect();
      coordinateRect = { left: rect.x, top: rect.y, width: rect.width, height: rect.height };
    } else {
      const rect = chart.getModel().getComponent("grid").coordinateSystem.getRect();
      coordinateRect = { left: rect.x, top: rect.y, width: rect.width, height: rect.height };
    }
    const canvasBounds = host.querySelector("canvas").getBoundingClientRect();
    return {
      instanceID: chart.id,
      coordinate,
      chart: { width: chart.getWidth(), height: chart.getHeight() },
      canvas: { width: Math.round(canvasBounds.width), height: Math.round(canvasBounds.height) },
      visual: { left: visual.x, right: visual.x + visual.width, top: visual.y, bottom: visual.y + visual.height },
      coordinateRect,
      colors: option.visualMap[0].inRange.color,
      calculable: Boolean(visualModel.option.calculable),
      range: [visualModel.option.min, visualModel.option.max],
      dataLength: data.length,
      missing,
      maxValue: Math.max(...data.map((value) => value[value.length - 1]).filter((value) => typeof value === "number")),
      splitArea: coordinate === "calendar" ? null : Boolean(option.xAxis[0].splitArea?.show && option.yAxis[0].splitArea?.show),
      calendarBorder: coordinate === "calendar" ? option.calendar[0].itemStyle?.borderWidth : null,
      calendarMonthFontSize: coordinate === "calendar" ? option.calendar[0].monthLabel?.fontSize : null,
      calendarMonthColor: coordinate === "calendar" ? option.calendar[0].monthLabel?.color : null,
      titleLegendOverlap: titleBounds && legendBounds
        ? !(titleBounds.right < legendBounds.left || legendBounds.right < titleBounds.left || titleBounds.bottom < legendBounds.top || legendBounds.bottom < titleBounds.top)
        : false,
      pageOverflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
    };
  });
}

function assertCommonLayout(layout, context) {
  assert.ok(layout.visual.left >= -1 && layout.visual.top >= -1, `${context} visual scale starts outside chart`);
  assert.ok(layout.visual.right <= layout.chart.width + 1 && layout.visual.bottom <= layout.chart.height + 1, `${context} visual scale clipped`);
  assert.ok(layout.coordinateRect.left >= 0 && layout.coordinateRect.top >= 0, `${context} coordinate system starts outside chart`);
  assert.ok(layout.coordinateRect.left + layout.coordinateRect.width <= layout.chart.width + 1, `${context} coordinate system horizontally clipped`);
  assert.ok(layout.coordinateRect.top + layout.coordinateRect.height <= layout.chart.height + 1, `${context} coordinate system vertically clipped`);
  assert.ok(layout.coordinateRect.width / layout.chart.width >= 0.45, `${context} coordinate width ${layout.coordinateRect.width}/${layout.chart.width}`);
  assert.deepEqual(layout.canvas, layout.chart, `${context} canvas and chart size diverged`);
  assert.equal(layout.colors.length, 3);
  assert.equal(new Set(layout.colors).size, 3);
  assert.equal(layout.pageOverflow, 0);
  assert.equal(layout.titleLegendOverlap, false, `${context} title and legend overlap`);
}

function assertCategory(layout, context) {
  assertCommonLayout(layout, context);
  assert.equal(layout.coordinate, "cartesian2d");
  assert.deepEqual(layout.range, [0, 10]);
  assert.equal(layout.calculable, true);
  assert.equal(layout.splitArea, true);
  assert.equal(layout.dataLength, 168);
  assert.equal(layout.missing, 62);
  assert.equal(layout.maxValue, 14, `${context} values above the scale maximum were not preserved`);
}

function assertCalendar(layout, context) {
  assertCommonLayout(layout, context);
  assert.equal(layout.coordinate, "calendar");
  assert.deepEqual(layout.range, [0, 20]);
  assert.equal(layout.dataLength, 366);
  assert.equal(layout.missing, 30);
  assert.equal(layout.maxValue, 20);
  assert.equal(layout.calendarBorder, 0.5);
  assert.equal(layout.calendarMonthFontSize, 7);
  assert.match(layout.calendarMonthColor, /^(#|rgb)/, `${context} calendar month labels are not theme-colored`);
  assert.ok(layout.visual.left >= layout.coordinateRect.left + layout.coordinateRect.width - 1,
    `${context} visual scale overlaps calendar cells or left-side weekday labels: ${JSON.stringify(layout)}`);
}

test("test-owned random-port HeatMap route and shared assets stay healthy", async () => {
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

for (const width of [390, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const mode of ["light", "dark"]) {
      test(`${width}px ${theme} ${mode} preserves both upstream HeatMap behaviors in page and same-instance modal`, async () => {
        const { page, failures, category, calendar } = await heatmapPage(width);
        try {
          await page.evaluate(({ selected, dark }) => {
            document.documentElement.dataset.theme = selected;
            document.documentElement.classList.toggle("dark", dark);
          }, { selected: theme, dark: mode === "dark" });
          await page.waitForTimeout(300);

          for (const [figure, label, assertion] of [
            [category, "Weekly activity by hour", assertCategory],
            [calendar, "Calendar activity", assertCalendar],
          ]) {
            const wrapper = wrapperFor(figure);
            await waitForChartGeometry(figure);
            const normal = await measure(figure);
            assertion(normal, `${width}px ${theme} ${mode} ${label} page`);
            await openExpand(wrapper);
            const dialog = wrapper.getByRole("dialog", { name: label });
            await dialog.waitFor({ state: "visible" });
            await waitForDialogSettled(dialog);
            await waitForChartGeometry(figure);
            const modal = await measure(figure);
            assertion(modal, `${width}px ${theme} ${mode} ${label} modal`);
            assert.equal(modal.instanceID, normal.instanceID, `${label} modal replaced the chart instance`);
            const panel = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((element) => {
              const bounds = element.getBoundingClientRect();
              return {
                contained: bounds.left >= 0 && bounds.right <= innerWidth + 1 && bounds.top >= 0 && bounds.bottom <= innerHeight + 1,
                centered: Math.abs((bounds.left + bounds.right) / 2 - innerWidth / 2) < 4,
              };
            });
            assert.deepEqual(panel, { contained: true, centered: true });
            await closeExpand(page, wrapper, label);
          }

          const exactTables = page.locator("[data-heatmap-exact-values]");
          assert.equal(await exactTables.count(), 2);
          assert.equal(await exactTables.nth(0).locator('[data-heatmap-missing="true"]').count(), 62);
          assert.equal(await exactTables.nth(1).locator('[data-heatmap-missing="true"]').count(), 30);
          if (screenshotDirectory) {
            await page.screenshot({ path: path.join(screenshotDirectory, `interactive-heatmap-${width}-${theme}-${mode}.png`), fullPage: true });
            await wrapperFor(category).screenshot({ path: path.join(screenshotDirectory, `interactive-heatmap-category-${width}-${theme}-${mode}.png`) });
            await wrapperFor(calendar).screenshot({ path: path.join(screenshotDirectory, `interactive-heatmap-calendar-${width}-${theme}-${mode}.png`) });
          }
          assert.deepEqual(failures, []);
        } finally {
          await page.close();
        }
      });
    }
  }
}

test("both HeatMap instances resize in place and react to live theme changes", async () => {
  const { page, failures, category, calendar } = await heatmapPage(1440);
  try {
    const beforeCategory = await measure(category);
    const beforeCalendar = await measure(calendar);
    await page.setViewportSize({ width: 390, height: 1000 });
    await waitForChartGeometry(category);
    await waitForChartGeometry(calendar);
    const narrowCategory = await measure(category);
    const narrowCalendar = await measure(calendar);
    assert.equal(narrowCategory.instanceID, beforeCategory.instanceID);
    assert.equal(narrowCalendar.instanceID, beforeCalendar.instanceID);
    assert.ok(narrowCategory.chart.width < beforeCategory.chart.width);
    assert.ok(narrowCalendar.chart.width < beforeCalendar.chart.width);
    assertCategory(narrowCategory, "resized category");
    assertCalendar(narrowCalendar, "resized calendar");

    await page.evaluate(() => {
      document.documentElement.dataset.theme = "araihu";
      document.documentElement.classList.add("dark");
    });
    await page.waitForTimeout(350);
    await waitForChartGeometry(category);
    await waitForChartGeometry(calendar);
    const themedCategory = await measure(category);
    const themedCalendar = await measure(calendar);
    assert.equal(themedCategory.instanceID, beforeCategory.instanceID);
    assert.equal(themedCalendar.instanceID, beforeCalendar.instanceID);
    assert.notDeepEqual(themedCategory.colors, beforeCategory.colors);
    assert.notDeepEqual(themedCalendar.colors, beforeCalendar.colors);
    assert.deepEqual(failures, []);
  } finally {
    await page.close();
  }
});

test("both HeatMap instances preserve identity and geometry through native fullscreen", async () => {
  const { page, failures, category, calendar } = await heatmapPage(1440);
  try {
    for (const [figure, label, assertion] of [
      [category, "Weekly activity by hour", assertCategory],
      [calendar, "Calendar activity", assertCalendar],
    ]) {
      const wrapper = wrapperFor(figure);
      await waitForChartGeometry(figure);
      const before = await measure(figure);
      await enterFullscreen(wrapper);
      await page.waitForFunction(() => Boolean(document.fullscreenElement));
      await waitForChartGeometry(figure);
      const fullscreen = await measure(figure);
      assertion(fullscreen, `${label} fullscreen`);
      assert.equal(fullscreen.instanceID, before.instanceID);
      assert.ok(fullscreen.chart.width > before.chart.width, `${label} did not grow in fullscreen`);
      await page.evaluate(() => document.exitFullscreen());
      await page.waitForFunction(() => !document.fullscreenElement);
      await waitForChartGeometry(figure);
      assert.equal((await measure(figure)).instanceID, before.instanceID);
    }
    assert.deepEqual(failures, []);
  } finally {
    await page.close();
  }
});

for (const [label, filename, assertion] of [
  ["Weekly activity by hour", "weekly-activity-heatmap.png", assertCategory],
  ["Calendar activity", "calendar-activity-heatmap.png", assertCalendar],
]) {
  test(`${label} PNG snapshot dimensions match its rendered large-modal instance`, async () => {
    const { page, failures } = await heatmapPage(1440);
    const figure = page.locator(`figure[aria-label="${label}"]`);
    const wrapper = wrapperFor(figure);
    try {
      await openExpand(wrapper);
      const dialog = wrapper.getByRole("dialog", { name: label });
      await dialog.waitFor({ state: "visible" });
      await waitForDialogSettled(dialog);
      await waitForChartGeometry(figure);
      const rendered = await measure(figure);
      assertion(rendered, `${label} PNG source`);
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
      await closeExpand(page, wrapper, label);
      const downloadPromise = page.waitForEvent("download");
      await page.getByRole("button", { name: `Download ${label} as PNG` }).first().click();
      const download = await downloadPromise;
      assert.equal(download.suggestedFilename(), filename);
      const stream = await download.createReadStream();
      const chunks = [];
      for await (const chunk of stream) chunks.push(chunk);
      const exported = Buffer.concat(chunks);
      assert.deepEqual([...exported.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
      assert.deepEqual(failures, []);
    } finally {
      await page.close();
    }
  });
}
