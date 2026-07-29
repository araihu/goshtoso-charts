const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const crypto = require("node:crypto");
const fs = require("node:fs/promises");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

const candidateMarker = 'data-goshtoso-candidate="candlestick-bollinger-fc218c7fedf84c7a"';
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
      const response = await fetch(`${baseURL}/components/candlestick`);
      if (response.ok && (await response.text()).includes(candidateMarker)) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Static Candlestick verification server did not start at ${baseURL}`);
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

async function chartPage(viewport = { width: 1440, height: 900 }) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  await page.addInitScript(() => {
    const createObjectURL = URL.createObjectURL.bind(URL);
    globalThis.__candlestickBlobTypes = [];
    URL.createObjectURL = (blob) => {
      globalThis.__candlestickBlobTypes.push(blob.type);
      return createObjectURL(blob);
    };
  });
  const failures = [];
  page.on("response", (response) => {
    if (response.status() >= 400) failures.push(`${response.status()} ${response.url()}`);
  });
  await page.goto(`${baseURL}/components/candlestick`);
  const figure = page.locator('figure[aria-label="Candlestick Chart with Bollinger Bands"]');
  await figure.waitFor();
  await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
  return { page, failures, figure, wrapper: figure.locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]") };
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
  await page.evaluate(() => { globalThis.__candlestickBlobTypes.length = 0; });
  const pending = page.waitForEvent("download", { timeout: 10000 });
  await wrapper.getByRole("button", { name: "Export Candlestick Chart with Bollinger Bands" }).click();
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
    types: await page.evaluate(() => [...globalThis.__candlestickBlobTypes]),
  };
}

test("test-owned candidate route, search, and assets are exact and healthy", async () => {
  assert.notEqual(new URL(baseURL).port, "8091");
  for (const route of [
    "/components/candlestick",
    "/attributions",
    "/search/assets/search.js",
    "/charts/assets/js/controls/3/controls.js",
    "/assets/styles.css",
  ]) {
    const response = await fetch(`${baseURL}${route}`);
    assert.equal(response.status, 200, route);
    if (route === "/components/candlestick") assert.match(await response.text(), new RegExp(candidateMarker));
  }
  const { page } = await chartPage();
  try {
    const search = page.locator("#docs-search");
    await search.fill("candlestick");
    assert.equal(await page.locator("[data-docs-search-item]:visible").filter({ hasText: "Candlestick" }).count() >= 1, true);
  } finally {
    await page.close();
  }
});

test("source OHLC hash, period-five bands, title, legend, padding, and default example stay exact", async () => {
  const { page, failures, figure, wrapper } = await chartPage();
  try {
    assert.equal(await page.locator('figure[aria-label="Seven-day stock price"]').count(), 1);
    assert.equal(await page.locator('figure[aria-label="Seven-day stock price"]').locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]").locator("tbody tr").count(), 7);
    const rows = await wrapper.locator("tbody tr").evaluateAll((items) => items.map((row) =>
      [...row.querySelectorAll("th,td")].map((cell) => cell.textContent.trim())));
    assert.equal(rows.length, 20);
    const normalized = rows.map((row) => [row[0], row[2], row[3], row[4], row[5]].join("|") + "\n").join("");
    assert.equal(crypto.createHash("sha256").update(normalized).digest("hex"), "fc218c7fedf84c7ac739016015e2508439a9a31d2c68a46fa1bfa84dc0e8f1ef");
    const expectedBands = [
      ["119.046537", "110.666667", "102.286797"], ["123.862780", "113.000000", "102.137220"],
      ["129.058697", "115.400000", "101.741303"], ["133.459862", "120.400000", "107.340138"],
      ["139.142136", "125.000000", "110.857864"], ["144.142136", "130.000000", "115.857864"],
      ["149.142136", "135.000000", "120.857864"], ["154.142136", "140.000000", "125.857864"],
      ["154.525200", "143.600000", "132.674800"], ["152.540920", "145.800000", "139.059080"],
      ["150.908132", "146.600000", "142.291868"], ["151.656854", "146.000000", "140.343146"],
      ["149.656854", "144.000000", "138.343146"], ["147.656854", "142.000000", "136.343146"],
      ["145.656854", "140.000000", "134.343146"], ["143.656854", "138.000000", "132.343146"],
      ["141.656854", "136.000000", "130.343146"], ["139.656854", "134.000000", "128.343146"],
      ["137.472136", "133.000000", "128.527864"], ["135.265986", "132.000000", "128.734014"],
    ];
    assert.deepEqual(rows.map((row) => row.slice(6, 9)), expectedBands);
    const svgState = await figure.locator("svg").evaluate((svg) => {
      const title = [...svg.querySelectorAll("text")].find((node) => node.textContent === "Candlestick Chart with Bollinger Bands");
      return {
        viewBox: svg.getAttribute("viewBox"),
        titleX: title?.getAttribute("x"),
        titleSize: title?.getAttribute("style").match(/font-size:([^;]+)/)?.[1] || "",
        legend: [...svg.querySelectorAll("text")].some((node) => node.textContent === "Price"),
      };
    });
    // Renderer converts the configured 18-point font to 23 CSS pixels.
    assert.deepEqual(svgState, { viewBox: "0 0 800 600", titleX: "20", titleSize: "23px", legend: true });
    assert.deepEqual(failures, []);
  } finally {
    await page.close();
  }
});

for (const width of [390, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const dark of [false, true]) {
      test(`${width}px ${theme} ${dark ? "dark" : "light"} bands remain distinct, contrasting, responsive, and modal-centered`, async () => {
        const { page, failures, figure, wrapper } = await chartPage({ width, height: 900 });
        try {
          await page.evaluate(({ theme, dark }) => {
            document.documentElement.dataset.theme = theme;
            document.documentElement.classList.toggle("dark", dark);
          }, { theme, dark });
          const state = await figure.evaluate((element) => {
            const style = getComputedStyle(element);
            const probe = document.createElement("span");
            element.appendChild(probe);
            const resolve = (token) => {
              probe.style.color = `var(${token})`;
              const canvas = document.createElement("canvas");
              canvas.width = 1;
              canvas.height = 1;
              const context = canvas.getContext("2d");
              context.fillStyle = getComputedStyle(probe).color;
              context.fillRect(0, 0, 1, 1);
              const [red, green, blue] = context.getImageData(0, 0, 1, 1).data;
              return `rgb(${red}, ${green}, ${blue})`;
            };
            const viewport = element.querySelector(".goshtoso-charts-candlestick__viewport");
            const result = {
              surface: resolve("--color-chart-surface"),
              upper: resolve("--color-chart-bollinger-upper"),
              middle: resolve("--color-chart-bollinger-middle"),
              lower: resolve("--color-chart-bollinger-lower"),
              documentWidth: document.documentElement.scrollWidth,
              viewportWidth: window.innerWidth,
              svgWidth: element.querySelector("svg").getBoundingClientRect().width,
              figureWidth: element.getBoundingClientRect().width,
              chartViewportWidth: viewport.clientWidth,
              chartScrollWidth: viewport.scrollWidth,
              rootClass: style.display,
            };
            probe.remove();
            return result;
          });
          assert.ok(state.documentWidth <= state.viewportWidth);
          assert.ok(state.chartViewportWidth <= state.figureWidth + 1);
          assert.equal(state.chartScrollWidth >= state.chartViewportWidth, true);
          assert.equal(new Set([state.upper, state.middle, state.lower]).size, 3);
          for (const color of [state.upper, state.middle, state.lower]) {
            assert.ok(contrast(color, state.surface) >= 3, `${theme}/${dark} ${color} on ${state.surface}`);
          }
          await openExpand(wrapper);
          const dialog = wrapper.getByRole("dialog", { name: "Candlestick Chart with Bollinger Bands" });
          await dialog.waitFor({ state: "visible" });
          await page.waitForTimeout(350);
          const geometry = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
            const chart = panel.querySelector(".goshtoso-charts-candlestick__viewport > svg").getBoundingClientRect();
            const bounds = panel.getBoundingClientRect();
            return {
              centerX: Math.abs((bounds.left + bounds.right) / 2 - window.innerWidth / 2),
              centerY: Math.abs((bounds.top + bounds.bottom) / 2 - window.innerHeight / 2),
              contained: chart.left >= bounds.left && chart.right <= bounds.right + 1 && chart.top >= bounds.top && chart.bottom <= bounds.bottom + 1,
              chartWidth: chart.width,
            };
          });
          assert.ok(geometry.centerX < 6 && geometry.centerY < 6);
          assert.equal(geometry.contained, true);
          assert.ok(geometry.chartWidth >= Math.min(width * 0.8, 700));
          await page.keyboard.press("Escape");
          await dialog.waitFor({ state: "hidden" });
          assert.deepEqual(failures, []);
        } finally {
          await page.close();
        }
      });
    }
  }
}

test("caller classes and colors render; SVG and PNG direct exports preserve 800x600", async () => {
  const { page } = await chartPage();
  try {
    const override = page.locator('figure[aria-label="Candlestick caller presentation overrides"]');
    const overrideSVG = override.locator("svg");
    assert.ok(await overrideSVG.locator('path[style*="#14532d"]').count() > 0);
    assert.ok(await overrideSVG.locator("path.caller-decreasing-candles").count() > 0);
    assert.ok(await overrideSVG.locator('path[style*="#1d4ed8"]').count() > 0);
    assert.ok(await overrideSVG.locator("path.caller-middle-band").count() > 0);

    const wrapper = page.locator('figure[aria-label="Candlestick Chart with Bollinger Bands"]').locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]");
    const svg = await download(page, wrapper, "SVG");
    assert.equal(svg.filename, "candlestick-bollinger-bands.svg");
    assert.equal(svg.types.at(-1), "image/svg+xml;charset=utf-8");
    assert.match(svg.bytes.toString("utf8"), /width="800".*height="600"/s);
    assert.doesNotMatch(svg.bytes.toString("utf8"), /var\(/);
    const png = await download(page, wrapper, "PNG");
    assert.equal(png.filename, "candlestick-bollinger-bands.png");
    assert.equal(png.types.at(-1), "image/png");
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: 800, height: 600 });
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
  } finally {
    await page.close();
  }
});
