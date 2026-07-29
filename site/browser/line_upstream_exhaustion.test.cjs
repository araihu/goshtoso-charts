const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

const customMarker = 'data-goshtoso-candidate="line-custom-67a9f05aba96e970"';
const gradientMarker = 'data-goshtoso-candidate="line-gradient-labels-21d84540f36ecdfb"';
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
      const response = await fetch(`${baseURL}/components/line`);
      if (response.ok && (await response.text()).includes(customMarker)) return;
    } catch {
      // Test-owned server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Line verification server did not start at ${baseURL}`);
}

before(async () => {
  const port = await randomPort();
  assert.notEqual(port, 8091);
  baseURL = `http://127.0.0.1:${port}`;
  server = spawn("/opt/homebrew/bin/go", ["run", "./cmd/server", "-port", String(port)], {
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
      // Test-owned server already stopped.
    }
  }
});

async function linePage(viewport = { width: 1440, height: 900 }) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  const failures = [];
  const errors = [];
  page.on("response", (response) => {
    if (response.status() >= 400) failures.push(`${response.status()} ${response.url()}`);
  });
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  page.on("pageerror", (error) => errors.push(error.message));
  await page.goto(`${baseURL}/components/line`);
  await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
  const custom = page.locator(`[${customMarker}] figure`).first();
  const gradient = page.locator(`[${gradientMarker}] figure`).first();
  await custom.waitFor();
  await gradient.waitFor();
  return {
    page,
    failures,
    errors,
    custom,
    gradient,
    wrapper: custom.locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]"),
  };
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

async function download(page, wrapper, format) {
  const staleMenu = wrapper.locator('[role="menu"]:visible');
  if (await staleMenu.count()) {
    await page.keyboard.press("Escape");
    await staleMenu.waitFor({ state: "hidden" });
  }
  const pending = page.waitForEvent("download", { timeout: 10000 });
  await wrapper.getByRole("button", { name: /Export/ }).click();
  const menu = wrapper.locator('[role="menu"]:visible');
  await menu.waitFor();
  await menu.getByRole("menuitem", { name: format, exact: true }).click();
  const artifact = await pending;
  const artifactPath = await artifact.path();
  assert.ok(artifactPath);
  if (await menu.isVisible()) await page.keyboard.press("Escape");
  await menu.waitFor({ state: "hidden" });
  return { filename: artifact.suggestedFilename(), bytes: await fs.readFile(artifactPath) };
}

for (const width of [390, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const mode of ["light", "dark"]) {
      test(`${width}px ${theme} ${mode} renders dense and gradient Line treatments responsively`, async () => {
        const { page, failures, errors, custom, gradient } = await linePage({ width, height: 900 });
        try {
          await page.evaluate(({ selected, dark }) => {
            document.documentElement.dataset.theme = selected;
            document.documentElement.classList.toggle("dark", dark);
          }, { selected: theme, dark: mode === "dark" });
          await page.waitForTimeout(100);
          const state = await page.evaluate(({ customSelector, gradientSelector }) => {
            const inspect = (selector) => {
              const figure = document.querySelector(selector);
              const viewport = figure.querySelector(".goshtoso-charts-line__viewport");
              const svg = figure.querySelector("svg");
              const rect = svg.getBoundingClientRect();
              return {
                client: viewport.clientWidth,
                scroll: viewport.scrollWidth,
                width: rect.width,
                height: rect.height,
                textCount: svg.querySelectorAll("text").length,
              };
            };
            const gradientFigure = document.querySelector(gradientSelector);
            const scaleLabels = [...gradientFigure.querySelectorAll("svg text")]
              .filter((node) => node.getAttribute("style")?.includes("--color-chart-scale"))
              .map((node) => getComputedStyle(node).fill);
            return {
              documentClient: document.documentElement.clientWidth,
              documentScroll: document.documentElement.scrollWidth,
              custom: inspect(customSelector),
              gradient: inspect(gradientSelector),
              scaleLabels,
            };
          }, {
            customSelector: `[${customMarker}] figure`,
            gradientSelector: `[${gradientMarker}] figure`,
          });
          assert.equal(state.documentScroll, state.documentClient, JSON.stringify(state));
          for (const chart of [state.custom, state.gradient]) {
            assert.ok(chart.scroll <= chart.client + 1, JSON.stringify(chart));
            assert.ok(chart.width <= chart.client + 1, JSON.stringify(chart));
            assert.ok(chart.width > 0 && chart.height > 0 && chart.textCount > 10, JSON.stringify(chart));
          }
          assert.ok(state.scaleLabels.length >= 10, JSON.stringify(state));
          assert.ok(new Set(state.scaleLabels).size >= 3, JSON.stringify(state.scaleLabels));
          const screenshot = await page.screenshot();
          assert.ok(screenshot.length > 10000);
          assert.deepEqual(failures, []);
          assert.deepEqual(errors, []);
        } finally {
          await page.close();
        }
      });
    }
  }
}

test("dense Line opens a centered responsive modal", async () => {
  const { page, failures, errors, custom, wrapper } = await linePage({ width: 390, height: 844 });
  try {
    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "Canon RF zoom-lens aperture ranges" });
    await dialog.waitFor({ state: "visible" });
    await dialog.locator('svg[viewBox="0 0 600 400"]').waitFor({ state: "visible" });
    const geometry = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
      const body = panel.children[1];
      const svg = body.querySelector("svg");
      const panelRect = panel.getBoundingClientRect();
      const bodyRect = body.getBoundingClientRect();
      const svgRect = svg.getBoundingClientRect();
      return {
        panelContained: panelRect.left >= 0 && panelRect.right <= innerWidth + 1 && panelRect.top >= 0 && panelRect.bottom <= innerHeight + 1,
        chartContained: svgRect.left >= bodyRect.left - 1 && svgRect.right <= bodyRect.right + 1 && svgRect.top >= bodyRect.top - 1 && svgRect.bottom <= bodyRect.bottom + 1,
        resolvedViewBox: svg.getAttribute("viewBox"),
      };
    });
    assert.deepEqual(geometry, { panelContained: true, chartContained: true, resolvedViewBox: "0 0 600 400" });
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });
    assert.deepEqual(failures, []);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});

test("dense Line exports resolved SVG and opaque PNG", async () => {
  const { page, failures, errors, wrapper } = await linePage();
  try {
    const svg = await download(page, wrapper, "SVG");
    assert.equal(svg.filename, "canon-rf-zoom-lenses.svg");
    const markup = svg.bytes.toString("utf8");
    assert.match(markup, /^<svg\b/);
    assert.match(markup, /viewBox="0 0 600 400"/);
    assert.match(markup, /100-500mm f\/4\.5-7\.1/);
    assert.doesNotMatch(markup, /(?:var\(|color-mix\(|url\(|@import)/i);

    const png = await download(page, wrapper, "PNG");
    assert.equal(png.filename, "canon-rf-zoom-lenses.png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: 600, height: 400 });
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
    assert.deepEqual(failures, []);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});
