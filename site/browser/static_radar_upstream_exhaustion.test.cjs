const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

const candidateMarker = 'data-static-radar-exhaustion="1fe31b06"';
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
      const response = await fetch(`${baseURL}/components/radar`);
      if (response.ok && (await response.text()).includes(candidateMarker)) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Radar verification server did not start at ${baseURL}`);
}

before(async () => {
  const port = await randomPort();
  assert.notEqual(port, 8091);
  assert.notEqual(port, 8096);
  baseURL = `http://127.0.0.1:${port}`;
  server = spawn("go", ["run", "./cmd/server", "-port", String(port)], {
    cwd: path.resolve(__dirname, ".."), detached: true, stdio: "ignore",
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

async function radarPage(viewport = { width: 1440, height: 900 }) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  const failures = [];
  const errors = [];
  page.on("response", (response) => {
    if (response.status() >= 400) failures.push(`${response.status()} ${response.url()}`);
  });
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  page.on("pageerror", (error) => errors.push(error.message));
  await page.goto(`${baseURL}/components/radar`);
  await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
  const basic = page.locator(`[${candidateMarker}][aria-label="Basic radar chart"]`);
  await basic.waitFor();
  return {
    page,
    failures,
    errors,
    basic,
    wrapper: basic.locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]"),
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

function rgb(value) {
  return (String(value).match(/[\d.]+/g)?.map(Number) || []).slice(0, 3);
}

function luminance(color) {
  const linear = rgb(color).map((value) => {
    const channel = value / 255;
    return channel <= 0.03928 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * linear[0] + 0.7152 * linear[1] + 0.0722 * linear[2];
}

function contrast(first, second) {
  const [high, low] = [luminance(first), luminance(second)].sort((a, b) => b - a);
  return (high + 0.05) / (low + 0.05);
}

test("Radar browser server is exact candidate build on a test-owned port", async () => {
  assert.notEqual(new URL(baseURL).port, "8091");
  assert.notEqual(new URL(baseURL).port, "8096");
  const response = await fetch(`${baseURL}/components/radar`);
  assert.equal(response.status, 200);
  const markup = await response.text();
  assert.match(markup, /data-goshtoso-candidate="radar-basic-0cf8dbdd72f6a398"/);
  assert.match(markup, /data-static-radar-exhaustion="1fe31b06"/);
});

for (const width of [390, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const mode of ["light", "dark"]) {
      test(`${width}px ${theme} ${mode} keeps both Radar treatments legible and contained`, async () => {
        const { page, failures, errors, basic } = await radarPage({ width, height: 900 });
        try {
          await page.evaluate(({ selected, dark }) => {
            document.documentElement.dataset.theme = selected;
            document.documentElement.classList.toggle("dark", dark);
          }, { selected: theme, dark: mode === "dark" });
          await page.waitForTimeout(100);
          const state = await page.evaluate(() => {
            const figures = [...document.querySelectorAll(".goshtoso-charts-radar")];
            const first = figures[0];
            const resolveRGB = (color) => {
              const canvas = document.createElement("canvas");
              canvas.width = 1;
              canvas.height = 1;
              const context = canvas.getContext("2d");
              context.fillStyle = color;
              context.fillRect(0, 0, 1, 1);
              return `rgb(${[...context.getImageData(0, 0, 1, 1).data.slice(0, 3)].join(", ")})`;
            };
            const resolveColor = (value) => {
              const probe = document.createElement("span");
              probe.style.color = value;
              first.appendChild(probe);
              const result = getComputedStyle(probe).color;
              probe.remove();
              return result;
            };
            const paths = [...first.querySelectorAll("svg path")]
              .filter((node) => node.getAttribute("style")?.includes("--color-chart-series-"));
            const series = [...new Set(paths.map((node) => getComputedStyle(node).stroke)
              .filter((color) => color && color !== "none").map(resolveRGB))];
            return {
              documentClient: document.documentElement.clientWidth,
              documentScroll: document.documentElement.scrollWidth,
              figures: figures.map((figure) => {
                const viewport = figure.querySelector(".goshtoso-charts-radar__viewport");
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
                  viewBox: svg.getAttribute("viewBox"),
                  title: [...svg.querySelectorAll("text")].some((node) => node.textContent === "Basic Radar Chart"),
                };
              }),
              surface: resolveColor("var(--color-chart-surface)"),
              text: resolveColor("var(--color-chart-text)"),
              series,
            };
          });
          assert.equal(state.documentScroll, state.documentClient, JSON.stringify(state));
          assert.deepEqual(state.figures.map((figure) => figure.label), [
            "Basic radar chart", "Readable radar values and compact layout",
          ]);
          for (const figure of state.figures) {
            assert.equal(figure.role, "img");
            assert.equal(figure.caption, true);
            assert.equal(figure.viewBox, "0 0 600 400");
            assert.equal(figure.title, true);
            assert.ok(figure.width > 0 && figure.height > 0, JSON.stringify(figure));
            assert.ok(figure.width <= figure.scroll + 1, JSON.stringify(figure));
          }
          if (width === 390) assert.ok(state.figures.every((figure) => figure.scroll > figure.client));
          else assert.ok(state.figures.every((figure) => figure.scroll <= figure.client + 1));
          assert.ok(contrast(state.surface, state.text) >= 4.5, JSON.stringify(state));
          assert.ok(state.series.length >= 2, JSON.stringify(state));
          assert.notEqual(state.series[0], state.series[1]);
          assert.ok(state.series.slice(0, 2).every((color) => contrast(state.surface, color) >= 2), JSON.stringify(state));
          assert.deepEqual(failures, []);
          assert.deepEqual(errors, []);
          if (screenshotDirectory) {
            await page.screenshot({ path: path.join(screenshotDirectory, `radar-${width}-${theme}-${mode}.png`), fullPage: true });
            const wrappers = page.locator("[data-goshtoso-chart-wrapper]").filter({ has: page.locator(".goshtoso-charts-radar") });
            await wrappers.nth(0).screenshot({ path: path.join(screenshotDirectory, `radar-${width}-${theme}-${mode}-basic.png`) });
            await wrappers.nth(1).screenshot({ path: path.join(screenshotDirectory, `radar-${width}-${theme}-${mode}-readable.png`) });
          }
        } finally {
          await page.close();
        }
      });
    }
  }
}

test("basic Radar preserves the pinned dataset and exact accessible values", async () => {
  const { page, failures, errors, wrapper } = await radarPage();
  try {
    assert.deepEqual(await wrapper.locator("table thead th").allTextContents(), [
      "Series", "Sales (max 6500)", "Administration (max 16000)",
      "Information Technology (max 30000)", "Customer Support (max 38000)",
      "Development (max 52000)", "Marketing (max 25000)",
    ]);
    assert.deepEqual(await wrapper.locator("table tbody tr").allTextContents(), [
      "Allocated Budget4200300020000350005000018000",
      "Actual Spending50001400028000260004200021000",
    ]);
    assert.deepEqual(failures, []);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});

test("basic Radar expands into a centered contained modal", async () => {
  const { page, failures, errors, wrapper } = await radarPage({ width: 390, height: 844 });
  try {
    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "Basic radar chart" });
    await dialog.waitFor({ state: "visible" });
    await dialog.locator(".goshtoso-charts-radar__viewport svg").waitFor({ state: "visible" });
    const geometry = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
      const svg = panel.querySelector(".goshtoso-charts-radar__viewport svg");
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

test("Radar exports resolved SVG plus opaque and transparent PNG", async () => {
  const { page, failures, errors, wrapper } = await radarPage();
  try {
    const svg = await download(page, wrapper, "SVG");
    assert.equal(svg.filename, "basic-radar-chart.svg");
    const markup = svg.bytes.toString("utf8");
    assert.match(markup, /^<svg\b/);
    assert.match(markup, /viewBox="0 0 600 400"/);
    assert.match(markup, /Basic Radar Chart/);
    assert.doesNotMatch(markup, /(?:var\(|color-mix\(|url\(|@import)/i);
    if (screenshotDirectory) await fs.writeFile(path.join(screenshotDirectory, svg.filename), svg.bytes);

    const png = await download(page, wrapper, "PNG");
    assert.equal(png.filename, "basic-radar-chart.png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: 600, height: 400 });
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
    if (screenshotDirectory) await fs.writeFile(path.join(screenshotDirectory, png.filename), png.bytes);

    const readableWrapper = page.locator('figure[aria-label="Readable radar values and compact layout"]')
      .locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]");
    const transparent = await download(page, readableWrapper, "PNG");
    assert.equal(transparent.filename, "readable-radar-values.png");
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
