const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

const port = Number(process.env.CANDLESTICK_TEST_PORT || 18097);
const baseURL = process.env.BASE_URL || `http://127.0.0.1:${port}`;
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

test("390 and 1440 layouts converge without overflow across Goshtoso and AraiHu light/dark themes", async () => {
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

    const collapse = wrapper.locator('[data-goshtoso-chart-control="collapse"]');
    await collapse.click();
    assert.equal(await wrapper.locator("[data-goshtoso-chart-content]").getAttribute("hidden"), "");
    await collapse.click();
    await page.waitForTimeout(100);
    assert.equal(await wrapper.evaluate((element) =>
      window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")) === element.__candlestickInstance), true);

    await selectTheme(page, "araihu", true);
    assert.equal(await wrapper.evaluate((element) =>
      window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")) === element.__candlestickInstance), true);

    await wrapper.locator("[data-goshtoso-chart-expand] > div > button").first().click();
    const dialog = wrapper.getByRole("dialog", { name: "Candlestick example" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForTimeout(350);
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

    const fullscreen = wrapper.locator('[data-goshtoso-chart-control="fullscreen"]');
    await fullscreen.click();
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
    await page.getByRole("button", { name: "Download Candlestick example as PNG" }).first().click();
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
