const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

const port = Number(process.env.CANDLESTICK_TEST_PORT || 18097);
const baseURL = process.env.BASE_URL || `http://127.0.0.1:${port}`;
const screenshotDirectory = process.env.GOSHTOSO_SCREENSHOT_DIR;
let browser;
let server;

async function ready() {
  for (let attempt = 0; attempt < 120; attempt += 1) {
    try {
      if ((await fetch(`${baseURL}/components/interactive/candlestick`)).ok) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Candlestick verification server did not start at ${baseURL}`);
}

before(async () => {
  if (!process.env.BASE_URL) {
    try {
      if ((await fetch(`${baseURL}/components/interactive/candlestick`)).ok) {
        throw new Error(`Refusing to reuse or stop an existing server at ${baseURL}`);
      }
    } catch (error) {
      if (String(error.message).includes("Refusing")) throw error;
    }
    server = spawn("go", ["run", "./cmd/server", "-port", String(port)], {
      cwd: path.resolve(__dirname, ".."),
      detached: true,
      stdio: "pipe",
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

async function candlestickPage(viewport = { width: 1440, height: 900 }) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  await page.goto(`${baseURL}/components/interactive/candlestick`);
  await page.locator("[data-goshtoso-chart-wrapper]").first().waitFor();
  await page.waitForFunction(() => {
    const host = document.querySelector("[_echarts_instance_]");
    return Boolean(host && window.echarts.getInstanceByDom(host));
  });
  return page;
}

async function openExpand(wrapper) {
  const trigger = wrapper.locator('[id$="-primary-action"]:visible').first();
  await trigger.waitFor({ state: "visible" });
  await trigger.click();
  return trigger;
}

async function enterFullscreen(wrapper) {
  await wrapper.locator('[id$="-stacked"] > [data-popover-root] > [data-popover-trigger] > button:visible').first().click();
  await wrapper.locator('[id$="-fullscreen-action"]:visible').first().click();
}

async function selectTheme(page, theme, dark) {
  await page.evaluate(({ selectedTheme, selectedDark }) => {
    const root = document.documentElement;
    root.dataset.theme = selectedTheme;
    root.classList.toggle("dark", selectedDark);
    window.__goshtosoChartsThemeRuntime?.refresh();
  }, { selectedTheme: theme, selectedDark: dark });
  await page.waitForTimeout(100);
}

async function chartState(page) {
  return page.locator("[data-goshtoso-chart-wrapper]").first().evaluate((wrapper) => {
    const host = wrapper.querySelector("[_echarts_instance_]");
    const instance = window.echarts.getInstanceByDom(host);
    const canvas = host.querySelector("canvas").getBoundingClientRect();
    const option = instance.getOption();
    return {
      hostWidth: host.clientWidth,
      hostHeight: host.clientHeight,
      chartWidth: instance.getWidth(),
      chartHeight: instance.getHeight(),
      canvasWidth: Math.round(canvas.width),
      canvasHeight: Math.round(canvas.height),
      rise: option.series[0].itemStyle.color,
      fall: option.series[0].itemStyle.color0,
      background: option.backgroundColor,
      instanceID: host.getAttribute("_echarts_instance_"),
      documentWidth: document.documentElement.scrollWidth,
      viewportWidth: window.innerWidth,
    };
  });
}

test("all five pinned candlestick behaviors preserve the exact dataset and typed options", async () => {
  const page = await candlestickPage();
  try {
    await page.waitForFunction(() => document.querySelectorAll("[data-candlestick-variant] [_echarts_instance_]").length === 5);
    const variants = await page.locator("[data-candlestick-variant]").evaluateAll((figures) => Object.fromEntries(figures.map((figure) => {
      const host = figure.querySelector("[_echarts_instance_]");
      const option = window.echarts.getInstanceByDom(host).getOption();
      return [figure.dataset.candlestickVariant, {
        count: option.series[0].data.length,
        first: option.series[0].data[0].value,
        last: option.series[0].data.at(-1).value,
        splitNumber: option.xAxis[0].splitNumber,
        scale: option.yAxis[0].scale,
        zoom: (option.dataZoom || []).map((zoom) => ({
          type: zoom.type,
          start: zoom.start,
          end: zoom.end,
          orient: zoom.orient,
          x: zoom.xAxisIndex,
          y: zoom.yAxisIndex,
        })),
        marks: option.series[0].markPoint?.data?.map((point) => ({ name: point.name, type: point.type, valueDim: point.valueDim })) || [],
        styles: figure.getAttribute("data-goshtoso-charts-candlestick-styles"),
      }];
    })));

    assert.deepEqual(Object.keys(variants), ["baseline", "inside", "inside-slider", "y-axis", "style"]);
    for (const variant of Object.values(variants)) {
      assert.equal(variant.count, 88);
      assert.deepEqual(variant.first, [2320.26, 2320.26, 2287.3, 2362.94]);
      assert.deepEqual(variant.last, [2190.1, 2148.35, 2126.22, 2190.1]);
      assert.equal(variant.splitNumber, 20);
      assert.equal(variant.scale, true);
    }
    assert.deepEqual(variants.baseline.zoom.map(({ type, start, end, x }) => ({ type, start, end, x })), [
      { type: "", start: 50, end: 100, x: [0] },
    ]);
    assert.deepEqual(variants.inside.zoom.map(({ type, start, end, x }) => ({ type, start, end, x })), [
      { type: "inside", start: 50, end: 100, x: [0] },
    ]);
    assert.deepEqual(variants["inside-slider"].zoom.map(({ type, start, end, x }) => ({ type, start, end, x })), [
      { type: "inside", start: 50, end: 100, x: [0] },
      { type: "", start: 50, end: 100, x: [0] },
    ]);
    assert.deepEqual(variants["y-axis"].zoom.map(({ type, start, end, orient, y }) => ({ type, start, end, orient, y })), [
      { type: "", start: 50, end: 100, orient: "vertical", y: [0] },
    ]);
    assert.deepEqual(variants.style.marks, [
      { name: "highest value", type: "max", valueDim: "highest" },
      { name: "lowest value", type: "min", valueDim: "lowest" },
    ]);
    assert.match(variants.style.styles, /goshtoso-charts-candlestick__direction--decreasing/);
    assert.match(variants.style.styles, /goshtoso-charts-candlestick__direction--increasing/);
    assert.doesNotMatch(variants.style.styles, /#ec0000|#00da3c|#8A0000|#008F28/i);
  } finally {
    await page.close();
  }
});

test("390 and 1440 layouts converge without overflow across Goshtoso and AraiHu light/dark themes", async () => {
  if (screenshotDirectory) await fs.mkdir(screenshotDirectory, { recursive: true });
  const colors = new Map();
  const measurements = [];
  for (const width of [390, 1440]) {
    for (const theme of ["goshtoso", "araihu"]) {
      for (const dark of [false, true]) {
        const page = await candlestickPage({ width, height: 900 });
        try {
          await selectTheme(page, theme, dark);
          const state = await chartState(page);
          assert.ok(state.documentWidth <= state.viewportWidth, `${width}/${theme}/${dark} overflow ${state.documentWidth} > ${state.viewportWidth}`);
          assert.equal(state.hostWidth, state.chartWidth);
          assert.equal(state.hostWidth, state.canvasWidth);
          assert.equal(state.hostHeight, state.chartHeight);
          assert.equal(state.hostHeight, state.canvasHeight);
          assert.notEqual(state.rise, state.fall, `${theme}/${dark} rise and fall colors match`);
          colors.set(`${theme}/${dark}`, `${state.rise}|${state.fall}|${state.background}`);
          measurements.push({ width, theme, dark, host: `${state.hostWidth}x${state.hostHeight}`, documentWidth: state.documentWidth });
          if (screenshotDirectory) {
            await page.screenshot({
              path: path.join(screenshotDirectory, `interactive-candlestick-${width}-${theme}-${dark ? "dark" : "light"}.png`),
              fullPage: true,
            });
            if ((width === 390 && theme === "goshtoso" && !dark) || (width === 1440 && theme === "araihu" && dark)) {
              const wrappers = page.locator("[data-candlestick-variant]");
              for (let index = 0; index < await wrappers.count(); index += 1) {
                const wrapper = wrappers.nth(index);
                const variant = await wrapper.getAttribute("data-candlestick-variant");
                await wrapper.screenshot({
                  path: path.join(screenshotDirectory, `interactive-candlestick-${variant}-${width}-${theme}-${dark ? "dark" : "light"}.png`),
                });
              }
            }
          }
        } finally {
          await page.close();
        }
      }
    }
  }
  assert.notEqual(colors.get("goshtoso/false"), colors.get("goshtoso/true"));
  assert.notEqual(colors.get("araihu/false"), colors.get("araihu/true"));
  console.log("candlestick matrix", JSON.stringify(measurements));
});

test("flex-parent resize, theme, and modal preserve one chart instance with exact convergence", async () => {
  const page = await candlestickPage();
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      element.__candlestickInstance = window.echarts.getInstanceByDom(host);
      const flex = document.createElement("div");
      flex.style.display = "flex";
      flex.style.width = "847px";
      element.parentNode.insertBefore(flex, element);
      flex.appendChild(element);
      element.style.flex = "1 1 auto";
      element.style.minWidth = "0";
      element.__candlestickFlexParent = flex;
    });
    await page.waitForFunction(() => {
      const wrapper = document.querySelector("[data-goshtoso-chart-wrapper]");
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return wrapper.__candlestickFlexParent.clientWidth === 847 &&
        host.clientWidth === instance.getWidth() &&
        host.clientWidth === Math.round(host.querySelector("canvas").getBoundingClientRect().width);
    });
    await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      element.__candlestickWideHost = host.clientWidth;
      element.__candlestickChromeWidth = element.__candlestickFlexParent.clientWidth - host.clientWidth;
      element.__candlestickFlexParent.style.width = "607px";
    });
    await page.waitForFunction(() => {
      const wrapper = document.querySelector("[data-goshtoso-chart-wrapper]");
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      const expected = 607 - wrapper.__candlestickChromeWidth;
      return host.clientWidth === expected && instance.getWidth() === expected &&
        Math.round(host.querySelector("canvas").getBoundingClientRect().width) === expected;
    });
    const resized = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return {
        same: instance === element.__candlestickInstance,
        wide: element.__candlestickWideHost,
        host: host.clientWidth,
        chart: instance.getWidth(),
        canvas: Math.round(host.querySelector("canvas").getBoundingClientRect().width),
      };
    });
    assert.equal(resized.same, true);
    assert.ok(resized.wide > resized.host);
    assert.deepEqual({ host: resized.host, chart: resized.chart, canvas: resized.canvas }, {
      host: resized.host, chart: resized.host, canvas: resized.host,
    });

    assert.equal(await wrapper.evaluate((element) =>
      window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")) === element.__candlestickInstance), true);

    await selectTheme(page, "araihu", true);
    assert.equal(await wrapper.evaluate((element) =>
      window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")) === element.__candlestickInstance), true);

    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "Candlestick example" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForFunction(() => {
      const host = document.querySelector(".goshtoso-charts-expand-panel [_echarts_instance_]");
      const instance = host && window.echarts.getInstanceByDom(host);
      return instance && instance.getWidth() === host.clientWidth && instance.getHeight() === host.clientHeight;
    });
    const geometry = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
      const body = panel.children[1];
      const host = body.querySelector("[_echarts_instance_]");
      const wrapper = panel.closest("[data-goshtoso-chart-wrapper]");
      const instance = window.echarts.getInstanceByDom(host);
      const panelRect = panel.getBoundingClientRect();
      const bodyRect = body.getBoundingClientRect();
      const hostRect = host.getBoundingClientRect();
      return {
        same: instance === wrapper.__candlestickInstance,
        panelCenterX: (panelRect.left + panelRect.right) / 2,
        viewportCenterX: window.innerWidth / 2,
        panelCenterY: (panelRect.top + panelRect.bottom) / 2,
        viewportCenterY: window.innerHeight / 2,
        contained: hostRect.left >= bodyRect.left && hostRect.right <= bodyRect.right + 1 &&
          hostRect.top >= bodyRect.top && hostRect.bottom <= bodyRect.bottom + 1,
        hostWidth: host.clientWidth,
        chartWidth: instance.getWidth(),
        canvasWidth: Math.round(host.querySelector("canvas").getBoundingClientRect().width),
      };
    });
    assert.equal(geometry.same, true);
    assert.ok(Math.abs(geometry.panelCenterX - geometry.viewportCenterX) < 4);
    assert.ok(Math.abs(geometry.panelCenterY - geometry.viewportCenterY) < 4);
    assert.equal(geometry.contained, true);
    assert.equal(geometry.hostWidth, geometry.chartWidth);
    assert.equal(geometry.hostWidth, geometry.canvasWidth);
    console.log("candlestick resize/modal", JSON.stringify({ resized, geometry }));
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });
    await page.waitForTimeout(200);
    assert.equal(await wrapper.evaluate((element) =>
      window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")) === element.__candlestickInstance), true);

    await enterFullscreen(wrapper);
    await page.waitForFunction(() => document.fullscreenElement !== null);
    await page.waitForTimeout(350);
    const fullscreenState = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return {
        same: instance === element.__candlestickInstance,
        host: host.clientWidth,
        chart: instance.getWidth(),
        canvas: Math.round(host.querySelector("canvas").getBoundingClientRect().width),
      };
    });
    assert.equal(fullscreenState.same, true);
    assert.deepEqual({ chart: fullscreenState.chart, canvas: fullscreenState.canvas }, {
      chart: fullscreenState.host, canvas: fullscreenState.host,
    });
    await page.evaluate(() => document.exitFullscreen());
    await page.waitForFunction(() => document.fullscreenElement === null);
    await page.waitForTimeout(200);
    assert.equal(await wrapper.evaluate((element) =>
      window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")) === element.__candlestickInstance), true);
  } finally {
    await page.close();
  }
});

test("direct export downloads a valid opaque PNG from current instance", async () => {
  const page = await candlestickPage();
  try {
    const expected = await chartState(page);
    const pending = page.waitForEvent("download");
    await page.getByRole("button", { name: "Export Candlestick example" }).first().click();
    await page.locator('[id$="-export-png-action"]:visible').first().click();
    const artifact = await pending;
    assert.equal(artifact.suggestedFilename(), "candlestick-example.png");
    const artifactPath = await artifact.path();
    assert.ok(artifactPath);
    const bytes = await fs.readFile(artifactPath);
    assert.deepEqual([...bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, {
      width: expected.chartWidth,
      height: expected.chartHeight,
    });
    const pixels = await sharp(bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) {
      assert.equal(pixels[index], 255);
    }
    console.log("candlestick png", JSON.stringify({ bytes: bytes.length, width: metadata.width, height: metadata.height }));
  } finally {
    await page.close();
  }
});
