const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

const candidateMarker = 'data-goshtoso-candidate="line-area-b2d7b87ff675f437"';
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
      if (response.ok && (await response.text()).includes(candidateMarker)) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Area Line verification server did not start at ${baseURL}`);
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
      // Test-owned server already stopped.
    }
  }
});

async function areaPage(viewport = { width: 1440, height: 900 }) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  await page.addInitScript(() => {
    const createObjectURL = URL.createObjectURL.bind(URL);
    globalThis.__areaBlobTypes = [];
    URL.createObjectURL = (blob) => {
      globalThis.__areaBlobTypes.push(blob.type);
      return createObjectURL(blob);
    };
  });
  const failures = [];
  const errors = [];
  page.on("response", (response) => {
    if (response.status() >= 400) failures.push(`${response.status()} ${response.url()}`);
  });
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  page.on("pageerror", (error) => errors.push(error.message));
  await page.goto(`${baseURL}/components/line`);
  const candidate = page.locator(`[${candidateMarker}]`);
  const figure = candidate.locator('figure[aria-label="Line"]');
  await figure.waitFor();
  await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
  return {
    page,
    failures,
    errors,
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

function luminance(rgb) {
  return rgb.map((channel) => {
    channel /= 255;
    return channel <= 0.03928 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4;
  }).reduce((sum, channel, index) => sum + channel * [0.2126, 0.7152, 0.0722][index], 0);
}

function contrast(first, second) {
  const [bright, dark] = [luminance(first), luminance(second)].sort((a, b) => b - a);
  return (bright + 0.05) / (dark + 0.05);
}

async function download(page, wrapper, format) {
  await page.evaluate(() => { globalThis.__areaBlobTypes.length = 0; });
  const pending = page.waitForEvent("download", { timeout: 10000 });
  await wrapper.getByRole("button", { name: "Export Line" }).click();
  const menu = wrapper.locator('[role="menu"]:visible');
  await menu.waitFor();
  await menu.getByRole("menuitem", { name: format, exact: true }).click();
  const artifact = await pending;
  const artifactPath = await artifact.path();
  assert.ok(artifactPath);
  if (await menu.isVisible()) await page.keyboard.press("Escape");
  await menu.waitFor({ state: "hidden" });
  return {
    filename: artifact.suggestedFilename(),
    bytes: await fs.readFile(artifactPath),
    types: await page.evaluate(() => [...globalThis.__areaBlobTypes]),
  };
}

test("area Line routes, search, assets, title, and exact adjacent values stay healthy", async () => {
  assert.notEqual(new URL(baseURL).port, "8091");
  for (const route of [
    "/components/line",
    "/attributions",
    "/search/assets/search.js",
    "/charts/assets/js/controls/4/controls.js",
    "/assets/styles.css",
  ]) {
    const response = await fetch(`${baseURL}${route}`);
    assert.equal(response.status, 200, route);
    if (route === "/components/line") assert.match(await response.text(), new RegExp(candidateMarker));
  }
  const { page, failures, errors, figure, wrapper } = await areaPage();
  try {
    assert.equal(await figure.locator("svg").getAttribute("viewBox"), "0 0 600 400");
    const text = await figure.locator("svg text").allTextContents();
    for (const want of ["Line", "Email", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]) assert.ok(text.includes(want), want);
    assert.deepEqual(await wrapper.locator("tbody tr").evaluateAll((items) => items.map((row) =>
      [...row.querySelectorAll("th,td")].map((cell) => cell.textContent.trim()))), [
      ["Mon", "Email", "Y axis", "120"], ["Tue", "Email", "Y axis", "132"],
      ["Wed", "Email", "Y axis", "101"], ["Thu", "Email", "Y axis", "134"],
      ["Fri", "Email", "Y axis", "90"], ["Sat", "Email", "Y axis", "230"],
      ["Sun", "Email", "Y axis", "210"],
    ]);
    const search = page.locator("#docs-search");
    await search.fill("line");
    assert.ok(await page.locator("[data-docs-search-item]:visible").filter({ hasText: "Line chart" }).count() >= 1);
    assert.deepEqual(failures, []);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});

for (const width of [390, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const mode of ["light", "dark"]) {
      test(`${width}px ${theme} ${mode} shows contrasting filled area and centered responsive modal`, async () => {
        const { page, failures, errors, figure, wrapper } = await areaPage({ width, height: 900 });
        try {
          await page.evaluate(({ selected, dark }) => {
            document.documentElement.dataset.theme = selected;
            document.documentElement.classList.toggle("dark", dark);
          }, { selected: theme, dark: mode === "dark" });
          const state = await figure.evaluate((root) => {
            const svg = root.querySelector("svg");
            const viewport = root.querySelector(".goshtoso-charts-line__viewport");
            const area = [...svg.querySelectorAll("path")].find((node) => node.getAttribute("style")?.includes("color-mix"));
            const surface = [...svg.querySelectorAll("path")].find((node) => node.getAttribute("style")?.includes("--color-chart-surface"));
            const rgba = (color) => {
              const canvas = document.createElement("canvas");
              canvas.width = canvas.height = 1;
              const context = canvas.getContext("2d");
              context.clearRect(0, 0, 1, 1);
              context.fillStyle = color;
              context.fillRect(0, 0, 1, 1);
              return [...context.getImageData(0, 0, 1, 1).data];
            };
            const fill = rgba(getComputedStyle(area).fill);
            const background = rgba(getComputedStyle(surface).fill);
            const alpha = fill[3] / 255;
            const composite = fill.slice(0, 3).map((channel, index) => Math.round(channel * alpha + background[index] * (1 - alpha)));
            const rect = svg.getBoundingClientRect();
            return {
              fill,
              background: background.slice(0, 3),
              composite,
              documentClient: document.documentElement.clientWidth,
              documentScroll: document.documentElement.scrollWidth,
              viewportClient: viewport.clientWidth,
              viewportScroll: viewport.scrollWidth,
              svgWidth: rect.width,
              svgHeight: rect.height,
            };
          });
          assert.ok(state.fill[3] >= 148 && state.fill[3] <= 152, JSON.stringify(state));
          assert.ok(contrast(state.composite, state.background) >= 1.35, JSON.stringify(state));
          assert.equal(state.documentScroll, state.documentClient);
          assert.ok(state.viewportScroll <= state.viewportClient + 1);
          assert.ok(state.svgWidth <= state.viewportClient + 1);
          assert.ok(Math.abs(state.svgWidth / state.svgHeight - 1.5) < 0.02);

          await openExpand(wrapper);
          const dialog = wrapper.getByRole("dialog", { name: "Line" });
          await dialog.waitFor({ state: "visible" });
          await page.waitForTimeout(350);
          const geometry = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
            const body = panel.children[1];
            const svg = body.querySelector("svg");
            const panelRect = panel.getBoundingClientRect();
            const bodyRect = body.getBoundingClientRect();
            const svgRect = svg.getBoundingClientRect();
            const matrix = svg.getScreenCTM();
            return {
              panelContained: panelRect.left >= 0 && panelRect.right <= innerWidth + 1 && panelRect.top >= 0 && panelRect.bottom <= innerHeight + 1,
              centered: Math.abs((panelRect.left + panelRect.right) / 2 - innerWidth / 2) < 4,
              chartContained: svgRect.left >= bodyRect.left - 1 && svgRect.right <= bodyRect.right + 1 && svgRect.top >= bodyRect.top - 1 && svgRect.bottom <= bodyRect.bottom + 1,
              uniformScale: Math.abs(Math.abs(matrix.a) - Math.abs(matrix.d)) < 0.01,
            };
          });
          assert.deepEqual(geometry, { panelContained: true, centered: true, chartContained: true, uniformScale: true });
          await page.keyboard.press("Escape");
          await dialog.waitFor({ state: "hidden" });
          assert.deepEqual(failures, []);
          assert.deepEqual(errors, []);
        } finally {
          await page.close();
        }
      });
    }
  }
}

test("area Line exports resolved 600x400 SVG and opaque PNG", async () => {
  const { page, failures, errors, wrapper } = await areaPage();
  try {
    const svg = await download(page, wrapper, "SVG");
    assert.equal(svg.filename, "filled-area-line.svg");
    assert.equal(svg.types.at(-1), "image/svg+xml;charset=utf-8");
    const markup = svg.bytes.toString("utf8");
    assert.match(markup, /^<svg\b/);
    assert.match(markup, /viewBox="0 0 600 400"/);
    assert.doesNotMatch(markup, /(?:var\(|color-mix\(|url\(|@import)/i);
    const parsed = await page.evaluate((source) => {
      const document = new DOMParser().parseFromString(source, "image/svg+xml");
      const root = document.documentElement;
      return {
        parserErrors: document.querySelectorAll("parsererror").length,
        width: Number(root.getAttribute("width")),
        height: Number(root.getAttribute("height")),
        viewBox: root.getAttribute("viewBox"),
        filledPaths: [...root.querySelectorAll("path")].filter((node) => getComputedStyle(node).fill !== "none").length,
      };
    }, markup);
    assert.equal(parsed.parserErrors, 0);
    assert.deepEqual({ width: parsed.width, height: parsed.height, viewBox: parsed.viewBox }, { width: 600, height: 400, viewBox: "0 0 600 400" });
    assert.ok(parsed.filledPaths >= 2);

    const png = await download(page, wrapper, "PNG");
    assert.equal(png.filename, "filled-area-line.png");
    assert.equal(png.types.at(-1), "image/png");
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
