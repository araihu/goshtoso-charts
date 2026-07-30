const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

const candidateMarker = 'data-static-heatmap-exhaustion="1fe31b06"';
const screenshotDirectory = process.env.GOSHTOSO_SCREENSHOT_DIR;
let baseURL;
let browser;
let server;
let serverStderr = "";

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
      const response = await fetch(`${baseURL}/components/heatmap`);
      if (response.ok && (await response.text()).includes(candidateMarker)) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`HeatMap verification server did not start at ${baseURL}: ${serverStderr.trim() || "no stderr output"}`);
}

before(async () => {
  const port = await randomPort();
  assert.notEqual(port, 8091);
  assert.notEqual(port, 8096);
  baseURL = `http://127.0.0.1:${port}`;
  server = spawn("go", ["run", "./cmd/server", "-port", String(port)], {
    cwd: path.resolve(__dirname, ".."), detached: true, stdio: ["ignore", "ignore", "pipe"],
  });
  server.stderr.setEncoding("utf8");
  server.stderr.on("data", (chunk) => { serverStderr += chunk; });
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

async function heatMapPage(viewport = { width: 1440, height: 900 }) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  const failures = [];
  const errors = [];
  page.on("response", (response) => {
    if (response.status() >= 400) failures.push(`${response.status()} ${response.url()}`);
  });
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  page.on("pageerror", (error) => errors.push(error.message));
  await page.goto(`${baseURL}/components/heatmap`);
  await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
  const basic = page.locator(`[${candidateMarker}][aria-label="Basic heat map"]`);
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

test("HeatMap browser server is exact candidate build on a test-owned port", async () => {
  assert.notEqual(new URL(baseURL).port, "8091");
  assert.notEqual(new URL(baseURL).port, "8096");
  const response = await fetch(`${baseURL}/components/heatmap`);
  assert.equal(response.status, 200);
  const markup = await response.text();
  assert.match(markup, /data-goshtoso-candidate="heatmap-basic-c39a3d85a0df126d"/);
  assert.match(markup, /data-static-heatmap-source="c39a3d85a0df126d"/);
  assert.match(markup, /data-static-heatmap-api="dd9b80660b9e0c0b"/);
});

for (const width of [390, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const mode of ["light", "dark"]) {
      test(`${width}px ${theme} ${mode} keeps both HeatMap treatments legible and contained`, async () => {
        const { page, failures, errors } = await heatMapPage({ width, height: 900 });
        try {
          await page.evaluate(({ selected, dark }) => {
            document.documentElement.dataset.theme = selected;
            document.documentElement.classList.toggle("dark", dark);
          }, { selected: theme, dark: mode === "dark" });
          await page.waitForTimeout(100);
          const state = await page.evaluate(() => {
            const figures = [...document.querySelectorAll(".goshtoso-charts-heatmap")];
            const resolveColor = (value, root) => {
              const probe = document.createElement("span");
              probe.style.color = value;
              root.appendChild(probe);
              const result = getComputedStyle(probe).color;
              probe.remove();
              return result;
            };
            return {
              documentClient: document.documentElement.clientWidth,
              documentScroll: document.documentElement.scrollWidth,
              surface: resolveColor("var(--color-chart-surface)", figures[0]),
              text: resolveColor("var(--color-chart-text)", figures[0]),
              figures: figures.map((figure) => {
                const viewport = figure.querySelector(".goshtoso-charts-heatmap__viewport");
                const svg = viewport.querySelector("svg");
                const rect = svg.getBoundingClientRect();
                const cells = [...svg.querySelectorAll("path")].filter((node) =>
                  [...node.classList].some((name) => name.includes("heatmap-stop") || name.startsWith("scale-")));
                return {
                  label: figure.getAttribute("aria-label"),
                  role: figure.getAttribute("role"),
                  caption: Boolean(figure.querySelector("figcaption")),
                  client: viewport.clientWidth,
                  scroll: viewport.scrollWidth,
                  width: rect.width,
                  height: rect.height,
                  viewBox: svg.getAttribute("viewBox"),
                  cells: cells.length,
                  distinctFills: new Set(cells.map((node) => getComputedStyle(node).fill)).size,
                  title: [...svg.querySelectorAll("text")].some((node) => node.textContent === "Heat Map Chart"),
                  scale: getComputedStyle(figure.querySelector(".goshtoso-charts-heatmap__scale-gradient")).backgroundImage,
                };
              }),
            };
          });
          assert.equal(state.documentScroll, state.documentClient, JSON.stringify(state));
          assert.deepEqual(state.figures.map((figure) => figure.label), [
            "Basic heat map", "Caller-styled pinned heat map with value labels",
          ]);
          for (const figure of state.figures) {
            assert.equal(figure.role, "img");
            assert.equal(figure.caption, true);
            assert.equal(figure.viewBox, "0 0 600 400");
            assert.equal(figure.cells, 25);
            assert.ok(figure.distinctFills >= 5, JSON.stringify(figure));
            assert.equal(figure.title, true);
            assert.match(figure.scale, /linear-gradient/);
            assert.ok(figure.width > 0 && figure.height > 0, JSON.stringify(figure));
            assert.ok(figure.width <= figure.scroll + 1, JSON.stringify(figure));
          }
          if (width === 390) assert.ok(state.figures.every((figure) => figure.scroll > figure.client));
          else assert.ok(state.figures.every((figure) => figure.scroll <= figure.client + 1));
          assert.ok(contrast(state.surface, state.text) >= 4.5, JSON.stringify(state));
          assert.deepEqual(failures, []);
          assert.deepEqual(errors, []);
          if (screenshotDirectory) {
            await page.screenshot({ path: path.join(screenshotDirectory, `heatmap-${width}-${theme}-${mode}.png`), fullPage: true });
            const wrappers = page.locator("[data-goshtoso-chart-wrapper]").filter({ has: page.locator(".goshtoso-charts-heatmap") });
            await wrappers.nth(0).screenshot({ path: path.join(screenshotDirectory, `heatmap-${width}-${theme}-${mode}-basic.png`) });
            await wrappers.nth(1).screenshot({ path: path.join(screenshotDirectory, `heatmap-${width}-${theme}-${mode}-caller.png`) });
          }
        } finally {
          await page.close();
        }
      });
    }
  }
}

test("basic HeatMap preserves the pinned matrix and exact accessible values", async () => {
  const { page, failures, errors, wrapper } = await heatMapPage();
  try {
    assert.deepEqual(await wrapper.locator("table thead th").allTextContents(), ["Y", "X", "Value"]);
    assert.deepEqual(await wrapper.locator("table tbody tr").allTextContents(), [
      "004.4", "014.9", "027", "037.5", "044.3",
      "102.6", "115.9", "129", "136.4", "142.3",
      "203.3", "216.4", "227", "234.9", "243.2",
      "301.9", "316", "329", "335.9", "342.6",
      "404.4", "415.9", "427", "436.4", "444.6",
    ]);
    assert.deepEqual(failures, []);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});

test("caller-styled HeatMap shows all values without creating another source treatment", async () => {
  const { page, failures, errors } = await heatMapPage();
  try {
    const figure = page.locator('[aria-label="Caller-styled pinned heat map with value labels"]');
    assert.equal(await figure.getAttribute("data-static-heatmap-source"), "c39a3d85a0df126d");
    const labels = await figure.locator("svg text").allTextContents();
    for (const value of ["4.4", "7.5", "1.9", "9.0"]) assert.ok(labels.includes(value), `${value} missing from ${labels}`);
    assert.deepEqual(failures, []);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});

test("basic HeatMap expands into a centered contained modal", async () => {
  const { page, failures, errors, wrapper } = await heatMapPage({ width: 390, height: 844 });
  try {
    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "Basic heat map" });
    await dialog.waitFor({ state: "visible" });
    await dialog.locator(".goshtoso-charts-heatmap__viewport svg").waitFor({ state: "visible" });
    const geometry = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
      const svg = panel.querySelector(".goshtoso-charts-heatmap__viewport svg");
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

test("HeatMap exports resolved SVG and opaque PNG", async () => {
  const { page, failures, errors, wrapper } = await heatMapPage();
  try {
    const svg = await download(page, wrapper, "SVG");
    assert.equal(svg.filename, "basic-heat-map.svg");
    const markup = svg.bytes.toString("utf8");
    assert.match(markup, /^<svg\b/);
    assert.match(markup, /viewBox="0 0 600 400"/);
    assert.match(markup, /Heat Map Chart/);
    assert.doesNotMatch(markup, /(?:var\(|color-mix\(|url\(|@import)/i);
    if (screenshotDirectory) await fs.writeFile(path.join(screenshotDirectory, svg.filename), svg.bytes);

    const png = await download(page, wrapper, "PNG");
    assert.equal(png.filename, "basic-heat-map.png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: 600, height: 400 });
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    let transparentAt = -1;
    for (let index = 3; index < pixels.length; index += 4) {
      if (pixels[index] !== 255) {
        transparentAt = index;
        break;
      }
    }
    const transparentAlpha = transparentAt >= 0 ? pixels[transparentAt] : 255;
    assert.equal(transparentAt, -1, `PNG has a transparent pixel at byte ${transparentAt} (alpha ${transparentAlpha})`);
    if (screenshotDirectory) await fs.writeFile(path.join(screenshotDirectory, png.filename), png.bytes);
    assert.deepEqual(failures, []);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});
