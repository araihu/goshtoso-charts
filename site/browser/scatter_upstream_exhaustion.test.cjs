const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

const candidateMarker = 'data-static-scatter-exhaustion="1fe31b06"';
const screenshotDirectory = process.env.GOSHTOSO_SCREENSHOT_DIR;
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
      const response = await fetch(`${baseURL}/components/scatter`);
      if (response.ok && (await response.text()).includes(candidateMarker)) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Scatter verification server did not start at ${baseURL}`);
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
  if (screenshotDirectory) await fs.mkdir(screenshotDirectory, { recursive: true });
});

after(async () => {
  await browser?.close();
  if (server?.pid) {
    try { process.kill(-server.pid, "SIGTERM"); } catch { /* already stopped */ }
  }
});

async function scatterPage(viewport = { width: 1440, height: 900 }) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  const failures = [];
  const errors = [];
  page.on("response", (response) => {
    if (response.status() >= 400) failures.push(`${response.status()} ${response.url()}`);
  });
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  page.on("pageerror", (error) => errors.push(error.message));
  await page.goto(`${baseURL}/components/scatter`);
  await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
  const dense = page.locator(`[${candidateMarker}]`);
  await dense.waitFor();
  return {
    page,
    failures,
    errors,
    dense,
    wrapper: dense.locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]"),
  };
}

async function openExpand(wrapper) {
  const trigger = wrapper.locator('[id$="-stacked"] > button:visible, [data-action-group-primary] button:visible').first();
  const stacked = await trigger.evaluate((button) => Boolean(button.closest('[id$="-stacked"]')));
  await trigger.click();
  if (stacked) {
    const action = wrapper.locator('[id$="-chart-expand-action"]');
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
  const pending = page.waitForEvent("download", { timeout: 15000 });
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

test("Scatter browser server is exact candidate build on a test-owned port", async () => {
  assert.notEqual(new URL(baseURL).port, "8091");
  assert.notEqual(new URL(baseURL).port, "8096");
  const response = await fetch(`${baseURL}/components/scatter`);
  assert.equal(response.status, 200);
  const markup = await response.text();
  assert.match(markup, /data-goshtoso-candidate="scatter-dense-0a50b43ccad6a96b"/);
  assert.match(markup, /data-static-scatter-exhaustion="1fe31b06"/);
});

for (const width of [390, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const mode of ["light", "dark"]) {
      test(`${width}px ${theme} ${mode} keeps all Scatter treatments responsive and accessible`, async () => {
        const { page, failures, errors, dense } = await scatterPage({ width, height: 900 });
        try {
          await page.evaluate(({ selected, dark }) => {
            document.documentElement.dataset.theme = selected;
            document.documentElement.classList.toggle("dark", dark);
          }, { selected: theme, dark: mode === "dark" });
          await page.waitForTimeout(100);
          const state = await page.evaluate(() => {
            const figures = [...document.querySelectorAll(".goshtoso-charts-scatter")];
            return {
              documentClient: document.documentElement.clientWidth,
              documentScroll: document.documentElement.scrollWidth,
              figures: figures.map((figure) => {
                const viewport = figure.querySelector(".goshtoso-charts-scatter__viewport");
                const svg = viewport.querySelector("svg");
                const rect = svg.getBoundingClientRect();
                return {
                  label: figure.getAttribute("aria-label"),
                  role: figure.getAttribute("role"),
                  caption: Boolean(figure.querySelector("figcaption")),
                  client: viewport.clientWidth,
                  scroll: viewport.scrollWidth,
                  width: rect.width,
                  height: rect.height,
                };
              }),
              seriesToken: getComputedStyle(figures[0]).getPropertyValue("--color-chart-series-1").trim(),
              topNRows: document.querySelectorAll('table[aria-label*="selected top labels"] tbody tr').length,
            };
          });
          assert.equal(state.documentScroll, state.documentClient, JSON.stringify(state));
          assert.equal(state.figures.length, 5);
          for (const figure of state.figures) {
            assert.equal(figure.role, "img");
            assert.ok(figure.label);
            assert.equal(figure.caption, true);
            assert.ok(figure.scroll <= figure.client + 1, JSON.stringify(figure));
            assert.ok(figure.width <= figure.client + 1, JSON.stringify(figure));
            assert.ok(figure.width > 0 && figure.height > 0, JSON.stringify(figure));
          }
          assert.ok(state.seriesToken && state.seriesToken !== "transparent");
          assert.equal(state.topNRows, 30);
          const denseMarks = await dense.locator("svg circle, svg rect, svg path, svg polygon").count();
          assert.ok(denseMarks > 3000, `dense mark count ${denseMarks}`);
          assert.deepEqual(failures, []);
          assert.deepEqual(errors, []);
          if (screenshotDirectory) {
            await page.screenshot({ path: path.join(screenshotDirectory, `scatter-${width}-${theme}-${mode}.png`), fullPage: true });
          }
        } finally {
          await page.close();
        }
      });
    }
  }
}

test("dense Scatter opens a centered contained modal at 390px", async () => {
  const { page, failures, errors, wrapper } = await scatterPage({ width: 390, height: 844 });
  try {
    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "Dense scatter data" });
    await dialog.waitFor({ state: "visible" });
	await dialog.locator(".goshtoso-charts-scatter__viewport svg").waitFor({ state: "visible" });
    const geometry = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
		  const svg = panel.querySelector(".goshtoso-charts-scatter__viewport svg");
		  const body = svg.closest(".goshtoso-charts-control-content");
      const panelRect = panel.getBoundingClientRect();
      const bodyRect = body.getBoundingClientRect();
      const svgRect = svg.getBoundingClientRect();
      return {
        panelContained: panelRect.left >= 0 && panelRect.right <= innerWidth + 1 && panelRect.top >= 0 && panelRect.bottom <= innerHeight + 1,
        chartContained: svgRect.left >= bodyRect.left - 1 && svgRect.right <= bodyRect.right + 1 && svgRect.top >= bodyRect.top - 1 && svgRect.bottom <= bodyRect.bottom + 1,
        viewBox: svg.getAttribute("viewBox"),
      };
    });
    assert.deepEqual(geometry, { panelContained: true, chartContained: true, viewBox: "0 0 600 400" });
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });
    assert.deepEqual(failures, []);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});

test("Scatter exports resolved SVG plus opaque and transparent PNG", async () => {
  const { page, failures, errors, wrapper } = await scatterPage();
  try {
    const svg = await download(page, wrapper, "SVG");
    assert.equal(svg.filename, "dense-scatter-data.svg");
    const markup = svg.bytes.toString("utf8");
    assert.match(markup, /^<svg\b/);
    assert.match(markup, /viewBox="0 0 600 400"/);
    assert.match(markup, /Dense Scatter Chart Demo/);
    assert.doesNotMatch(markup, /(?:var\(|color-mix\(|url\(|@import)/i);
    if (screenshotDirectory) await fs.writeFile(path.join(screenshotDirectory, svg.filename), svg.bytes);

    const png = await download(page, wrapper, "PNG");
    assert.equal(png.filename, "dense-scatter-data.png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: 600, height: 400 });
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
    if (screenshotDirectory) await fs.writeFile(path.join(screenshotDirectory, png.filename), png.bytes);

    const basicWrapper = page.locator('figure[aria-label="Basic scatter chart with one missing Email observation"]')
      .locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]");
    const transparent = await download(page, basicWrapper, "PNG");
    assert.equal(transparent.filename, "basic-scatter-chart.png");
    const transparentMetadata = await sharp(transparent.bytes).metadata();
    assert.deepEqual({ width: transparentMetadata.width, height: transparentMetadata.height }, { width: 600, height: 400 });
    const transparentPixels = await sharp(transparent.bytes).ensureAlpha().raw().toBuffer();
    let hasTransparency = false;
    for (let index = 3; index < transparentPixels.length; index += 4) {
      if (transparentPixels[index] < 255) {
        hasTransparency = true;
        break;
      }
    }
    assert.equal(hasTransparency, true);
    if (screenshotDirectory) await fs.writeFile(path.join(screenshotDirectory, transparent.filename), transparent.bytes);
    assert.deepEqual(failures, []);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});
