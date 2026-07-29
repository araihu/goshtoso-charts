const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const crypto = require("node:crypto");
const fs = require("node:fs/promises");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

const candidateMarker = 'data-goshtoso-candidate="line-dual-axis-78a3edd9aa356dc7"';
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
  throw new Error(`Line verification server did not start at ${baseURL}`);
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

async function linePage(viewport = { width: 1440, height: 900 }) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  await page.addInitScript(() => {
    const createObjectURL = URL.createObjectURL.bind(URL);
    globalThis.__lineBlobTypes = [];
    URL.createObjectURL = (blob) => {
      globalThis.__lineBlobTypes.push(blob.type);
      return createObjectURL(blob);
    };
  });
  const failures = [];
  page.on("response", (response) => {
    if (response.status() >= 400) failures.push(`${response.status()} ${response.url()}`);
  });
  await page.goto(`${baseURL}/components/line`);
  const candidate = page.locator('[data-goshtoso-candidate="line-dual-axis-78a3edd9aa356dc7"]');
  const figure = candidate.locator('figure[aria-label="Dual Axis Line"]');
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

function rgb(value) {
  return (String(value).match(/[\d.]+/g) || []).slice(0, 3).map(Number);
}

function luminance(value) {
  return rgb(value).map((channel) => {
    channel /= 255;
    return channel <= 0.03928 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4;
  }).reduce((sum, channel, index) => sum + channel * [0.2126, 0.7152, 0.0722][index], 0);
}

function contrast(first, second) {
  const [bright, dark] = [luminance(first), luminance(second)].sort((a, b) => b - a);
  return (bright + 0.05) / (dark + 0.05);
}

async function download(page, wrapper, format) {
  await page.evaluate(() => { globalThis.__lineBlobTypes.length = 0; });
  const pending = page.waitForEvent("download", { timeout: 10000 });
  await wrapper.getByRole("button", { name: "Export Dual Axis Line" }).click();
  const menu = wrapper.locator('[role="menu"]:visible');
  await menu.waitFor();
  await menu.getByRole("menuitem", { name: format, exact: true }).click();
  const outcome = await Promise.race([
    pending.then((artifact) => ({ artifact })),
    wrapper.locator("[data-goshtoso-chart-export-status]").evaluateHandle((status) => new Promise((resolve) => {
      const settled = () => status.textContent.startsWith("Download failed:") && resolve(status.textContent);
      new MutationObserver(settled).observe(status, { childList: true, characterData: true, subtree: true });
      settled();
    })).then(async (handle) => ({ failure: await handle.jsonValue() })),
  ]);
  if (outcome.failure) throw new Error(outcome.failure);
  const artifact = outcome.artifact;
  const artifactPath = await artifact.path();
  assert.ok(artifactPath);
  if (await menu.isVisible()) await page.keyboard.press("Escape");
  await menu.waitFor({ state: "hidden" });
  return {
    filename: artifact.suggestedFilename(),
    bytes: await fs.readFile(artifactPath),
    types: await page.evaluate(() => [...globalThis.__lineBlobTypes]),
  };
}

test("test-owned candidate routes, search, and assets are exact and healthy", async () => {
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
  const { page } = await linePage();
  try {
    const search = page.locator("#docs-search");
    await search.fill("line");
    assert.equal(await page.locator("[data-docs-search-item]:visible").filter({ hasText: "Line chart" }).count() >= 1, true);
  } finally {
    await page.close();
  }
});

test("pinned values, hash, title, labels, names, axes, and default Line compatibility stay exact", async () => {
  const { page, failures, figure, wrapper } = await linePage();
  try {
    const rows = await wrapper.locator("tbody tr").evaluateAll((items) => items.map((row) =>
      [...row.querySelectorAll("th,td")].map((cell) => cell.textContent.trim())));
    assert.deepEqual(rows, [
      ["A", "Left Series", "Left Y axis", "120"], ["A", "Right Series", "Right Y axis", "820"],
      ["B", "Left Series", "Left Y axis", "132"], ["B", "Right Series", "Right Y axis", "932"],
      ["C", "Left Series", "Left Y axis", "101"], ["C", "Right Series", "Right Y axis", "901"],
      ["D", "Left Series", "Left Y axis", "134"], ["D", "Right Series", "Right Y axis", "934"],
      ["E", "Left Series", "Left Y axis", "90"], ["E", "Right Series", "Right Y axis", "1290"],
      ["F", "Left Series", "Left Y axis", "230"], ["F", "Right Series", "Right Y axis", "1330"],
      ["G", "Left Series", "Left Y axis", "210"], ["G", "Right Series", "Right Y axis", "1320"],
    ]);
    const normalized = rows.map((row) => `${row.join("|")}\n`).join("");
    assert.equal(crypto.createHash("sha256").update(normalized).digest("hex"), "02df454d5ef671a0852ba86c212038cf0184f52fd96931d260d8f9b90d693355");
    assert.equal(await wrapper.locator("table").getAttribute("aria-label"), "Dual Axis Line exact series values and Y axis mapping");
    const svg = figure.locator("svg");
    assert.equal(await svg.getAttribute("viewBox"), "0 0 600 400");
    const text = await svg.locator("text").allTextContents();
    for (const want of ["Dual Axis Line", "A", "B", "C", "D", "E", "F", "G", "Left Series", "Right Series"]) {
      assert.ok(text.includes(want), want);
    }
    const axes = await svg.evaluate((root) => ({
      left: [...root.querySelectorAll("text")].filter((node) => node.getAttribute("style")?.includes("--color-chart-series-1")).length,
      right: [...root.querySelectorAll("text")].filter((node) => node.getAttribute("style")?.includes("--color-chart-series-2")).length,
    }));
    assert.ok(axes.left >= 2);
    assert.ok(axes.right >= 2);

    const defaultFigure = page.locator('figure[aria-label="HTTPS monitor latency in milliseconds"]');
    const defaultWrapper = defaultFigure.locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]");
    assert.equal(await defaultFigure.count(), 1);
    assert.equal(await defaultFigure.locator("svg").getAttribute("viewBox"), "0 0 720 320");
    assert.deepEqual(await defaultWrapper.locator("tbody tr").evaluateAll((items) => items.map((row) =>
      [...row.querySelectorAll("th,td")].map((cell) => cell.textContent.trim()))), [
      ["08:00", "Latency (ms)", "Y axis", "42"],
      ["08:01", "Latency (ms)", "Y axis", "47"],
      ["08:02", "Latency (ms)", "Y axis", "900"],
      ["08:03", "Latency (ms)", "Y axis", "51"],
      ["08:04", "Latency (ms)", "Y axis", "2000"],
      ["08:05", "Latency (ms)", "Y axis", "44"],
      ["08:06", "Latency (ms)", "Y axis", "46"],
    ]);
    assert.deepEqual(failures, []);
  } finally {
    await page.close();
  }
});

test("caller Color and Class overrides apply independently per series and axis", async () => {
  const { page, failures } = await linePage();
  try {
    const figure = page.locator('figure[aria-label="Dual Axis Line caller presentation overrides"]');
    const state = await figure.locator("svg").evaluate((svg) => {
      const computed = (selector, property) => {
        const element = svg.querySelector(selector);
        return element ? getComputedStyle(element)[property] : "";
      };
      return {
        leftSeries: computed('path[style*="#14532d"]', "stroke"),
        rightSeriesClass: svg.querySelectorAll(".caller-right-series").length,
        leftAxisClass: svg.querySelectorAll(".caller-left-axis").length,
        rightAxis: computed('text[style*="#7e22ce"]', "fill"),
      };
    });
    assert.deepEqual(rgb(state.leftSeries), [20, 83, 45]);
    assert.ok(state.rightSeriesClass >= 1);
    assert.ok(state.leftAxisClass >= 2);
    assert.deepEqual(rgb(state.rightAxis), [126, 34, 206]);
    assert.deepEqual(failures, []);
  } finally {
    await page.close();
  }
});

for (const width of [390, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const mode of ["light", "dark"]) {
      test(`${width}px ${theme} ${mode} keeps both axes themed, contrasting, responsive, and centered in Expand`, async () => {
        const { page, failures, figure, wrapper } = await linePage({ width, height: 900 });
        const errors = [];
        page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
        page.on("pageerror", (error) => errors.push(error.message));
        try {
          await page.evaluate(({ selected, dark }) => {
            document.documentElement.dataset.theme = selected;
            document.documentElement.classList.toggle("dark", dark);
          }, { selected: theme, dark: mode === "dark" });
          assert.equal(await wrapper.getByRole("button").count(), 2);
          const state = await figure.evaluate((root) => {
            const svg = root.querySelector("svg");
            const viewport = root.querySelector(".goshtoso-charts-line__viewport");
            const resolvedRGB = (color) => {
              const canvas = document.createElement("canvas");
              canvas.width = canvas.height = 1;
              const context = canvas.getContext("2d");
              context.fillStyle = color;
              context.fillRect(0, 0, 1, 1);
              return [...context.getImageData(0, 0, 1, 1).data.slice(0, 3)];
            };
            const pick = (tag, token, property) => {
              const element = [...svg.querySelectorAll(tag)].find((node) => node.getAttribute("style")?.includes(token));
              return element ? resolvedRGB(getComputedStyle(element)[property]) : [];
            };
            const surface = [...svg.querySelectorAll("path")].find((node) => node.getAttribute("style")?.includes("--color-chart-surface"));
            const svgRect = svg.getBoundingClientRect();
            return {
              documentClient: document.documentElement.clientWidth,
              documentScroll: document.documentElement.scrollWidth,
              viewportClient: viewport.clientWidth,
              viewportScroll: viewport.scrollWidth,
              svgWidth: svgRect.width,
              svgHeight: svgRect.height,
              surface: resolvedRGB(getComputedStyle(surface).fill),
              leftSeries: pick("path", "--color-chart-series-1", "stroke"),
              rightSeries: pick("path", "--color-chart-series-2", "stroke"),
              leftAxis: pick("text", "--color-chart-series-1", "fill"),
              rightAxis: pick("text", "--color-chart-series-2", "fill"),
            };
          });
          assert.equal(state.documentScroll, state.documentClient);
          assert.ok(state.viewportScroll <= state.viewportClient + 1);
          assert.ok(state.svgWidth <= state.viewportClient + 1);
          assert.ok(Math.abs(state.svgWidth / state.svgHeight - 1.5) < 0.02);
          assert.deepEqual(rgb(state.leftAxis), rgb(state.leftSeries));
          assert.deepEqual(rgb(state.rightAxis), rgb(state.rightSeries));
          assert.notDeepEqual(rgb(state.leftSeries), rgb(state.rightSeries));
          assert.ok(contrast(state.leftSeries, state.surface) >= 2, JSON.stringify(state));
          assert.ok(contrast(state.rightSeries, state.surface) >= 2, JSON.stringify(state));

          await wrapper.evaluate((element) => { element.__lineContent = element.querySelector("[data-goshtoso-chart-content]"); });
          await openExpand(wrapper);
          const dialog = wrapper.getByRole("dialog", { name: "Dual Axis Line" });
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
              panelWidth: panelRect.width,
              bodyWidth: bodyRect.width,
              panelContained: panelRect.left >= 0 && panelRect.right <= innerWidth + 1 && panelRect.top >= 0 && panelRect.bottom <= innerHeight + 1,
              centered: Math.abs((panelRect.left + panelRect.right) / 2 - innerWidth / 2) < 4,
              chartContained: svgRect.left >= bodyRect.left - 1 && svgRect.right <= bodyRect.right + 1 && svgRect.top >= bodyRect.top - 1 && svgRect.bottom <= bodyRect.bottom + 1,
              uniformScale: Math.abs(Math.abs(matrix.a) - Math.abs(matrix.d)) < 0.01,
              sameContent: panel.closest("[data-goshtoso-chart-wrapper]").__lineContent === body.querySelector("[data-goshtoso-chart-content]"),
            };
          });
          assert.equal(geometry.panelContained, true);
          assert.equal(geometry.centered, true);
          assert.equal(geometry.chartContained, true);
          assert.equal(geometry.uniformScale, true);
          assert.equal(geometry.sameContent, true);
          assert.ok(geometry.panelWidth >= (width === 1440 ? 1000 : width * 0.9));
          assert.ok(geometry.bodyWidth > 0);
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

test("dual-axis Line exports resolved 600x400 SVG and opaque PNG", async () => {
  const { page, failures, wrapper } = await linePage();
  try {
    const svg = await download(page, wrapper, "SVG");
    assert.equal(svg.filename, "dual-axis-line.svg");
    assert.equal(svg.types.at(-1), "image/svg+xml;charset=utf-8");
    const markup = svg.bytes.toString("utf8");
    assert.match(markup, /^<svg\b/);
    assert.match(markup, /viewBox="0 0 600 400"/);
    assert.doesNotMatch(markup, /(?:var\(|url\(|@import)/i);
    const parsed = await page.evaluate((source) => {
      const document = new DOMParser().parseFromString(source, "image/svg+xml");
      const root = document.documentElement;
      return {
        parserErrors: document.querySelectorAll("parsererror").length,
        width: Number(root.getAttribute("width")),
        height: Number(root.getAttribute("height")),
        viewBox: root.getAttribute("viewBox"),
      };
    }, markup);
    assert.deepEqual(parsed, { parserErrors: 0, width: 600, height: 400, viewBox: "0 0 600 400" });

    const png = await download(page, wrapper, "PNG");
    assert.equal(png.filename, "dual-axis-line.png");
    assert.equal(png.types.at(-1), "image/png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: 600, height: 400 });
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
    assert.deepEqual(failures, []);
  } finally {
    await page.close();
  }
});
