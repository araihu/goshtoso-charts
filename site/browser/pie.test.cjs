const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const path = require("node:path");
const { spawn } = require("node:child_process");
const net = require("node:net");
const { chromium } = require("playwright");
const sharp = require("sharp");

const candidateMarker = 'data-goshtoso-candidate="pie-doughnut-b97bca2322e90e2f"';
let baseURL;
const screenshotDirectory = process.env.SCREENSHOT_DIR;
let browser;
let server;

async function ready() {
  for (let attempt = 0; attempt < 120; attempt += 1) {
    try {
      const response = await fetch(`${baseURL}/components/pie`);
      if (response.ok && (await response.text()).includes(candidateMarker)) return;
    } catch {
      // Test-owned server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Pie verification server did not start at ${baseURL}`);
}

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
  if (screenshotDirectory) await fs.mkdir(screenshotDirectory, { recursive: true });
});

test("Pie browser server is exact candidate worktree build on a test-owned non-8091 port", async () => {
  assert.notEqual(new URL(baseURL).port, "8091");
  for (const route of [
    "/components/pie",
    "/attributions",
    "/search/assets/search.js",
    "/charts/assets/js/controls/3/controls.js",
  ]) {
    const response = await fetch(`${baseURL}${route}`);
    assert.equal(response.status, 200, route);
    if (route === "/components/pie") {
      assert.match(await response.text(), /data-goshtoso-candidate="pie-doughnut-b97bca2322e90e2f"/);
    }
  }
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

async function piePage(viewport = { width: 1440, height: 900 }) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  await page.addInitScript(() => {
    const createObjectURL = URL.createObjectURL.bind(URL);
    globalThis.__pieBlobTypes = [];
    URL.createObjectURL = (blob) => {
      globalThis.__pieBlobTypes.push(blob.type);
      return createObjectURL(blob);
    };
  });
  const failed = [];
  page.on("response", (response) => {
    if (response.status() >= 400) failed.push(`${response.status()} ${response.url()}`);
  });
  await page.goto(`${baseURL}/components/pie`);
  const figure = page.locator('figure[aria-label="Doughnut Chart"]');
  await figure.waitFor();
  await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
  return { page, failed, wrapper: figure.locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]") };
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

async function download(page, wrapper, format) {
  await page.evaluate(() => { globalThis.__pieBlobTypes.length = 0; });
  const pending = page.waitForEvent("download", { timeout: 10000 });
  await wrapper.getByRole("button", { name: "Export Doughnut Chart" }).click();
  const menu = wrapper.locator('[role="menu"]:visible');
  await menu.waitFor({ state: "visible" });
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
    types: await page.evaluate(() => [...globalThis.__pieBlobTypes]),
  };
}

for (const width of [390, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const mode of ["light", "dark"]) {
      test(`${width}px ${theme} ${mode} keeps Doughnut responsive, contrasting, exact, and modal-contained`, async () => {
        const { page, failed, wrapper } = await piePage({ width, height: 900 });
        const errors = [];
        page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
        page.on("pageerror", (error) => errors.push(error.message));
        try {
          await page.evaluate(({ selected, dark }) => {
            document.documentElement.dataset.theme = selected;
            document.documentElement.classList.toggle("dark", dark);
          }, { selected: theme, dark: mode === "dark" });
          assert.equal(await wrapper.locator("button:visible").count(), 2);
          assert.deepEqual(await wrapper.locator("table tbody th").allTextContents(), [
            "Search Engine", "Direct", "Email", "Union Ads", "Video Ads",
          ]);
          assert.deepEqual(await wrapper.locator("table tbody td:nth-child(2)").allTextContents(), [
            "1048", "735", "580", "484", "300",
          ]);
          const layout = await wrapper.evaluate((element) => {
            const resolveRGB = (color) => {
              const canvas = document.createElement("canvas");
              canvas.width = 1;
              canvas.height = 1;
              const context = canvas.getContext("2d");
              context.fillStyle = color;
              context.fillRect(0, 0, 1, 1);
              return [...context.getImageData(0, 0, 1, 1).data.slice(0, 3)];
            };
            const content = element.querySelector("[data-goshtoso-chart-content]");
            const viewport = content.querySelector(".goshtoso-charts-pie__viewport");
            const svg = viewport.querySelector("svg");
            const rect = svg.getBoundingClientRect();
            const texts = [...svg.querySelectorAll("text")];
            const title = texts.find((node) => node.textContent === "Doughnut Chart");
            const mark = [...svg.querySelectorAll("path")].find((node) => node.getAttribute("style")?.includes("--color-chart-series-1"));
            const surfaceProbe = document.createElement("span");
            surfaceProbe.style.color = "var(--color-chart-surface)";
            element.querySelector("figure").appendChild(surfaceProbe);
            const surfaceColor = getComputedStyle(surfaceProbe).color;
            surfaceProbe.remove();
            return {
              documentClient: document.documentElement.clientWidth,
              documentScroll: document.documentElement.scrollWidth,
              contentClient: content.clientWidth,
              contentScroll: content.scrollWidth,
              viewportClient: viewport.clientWidth,
              viewportScroll: viewport.scrollWidth,
              svgWidth: rect.width,
              svgHeight: rect.height,
              viewBox: svg.getAttribute("viewBox"),
              titleColor: resolveRGB(getComputedStyle(title).fill),
              markColor: resolveRGB(getComputedStyle(mark).fill),
              surfaceColor: resolveRGB(surfaceColor),
            };
          });
          assert.equal(layout.documentScroll, layout.documentClient);
          assert.ok(layout.contentScroll <= layout.contentClient + 1);
          assert.ok(layout.viewportScroll <= layout.viewportClient + 1);
          assert.equal(layout.viewBox, "0 0 600 400");
          assert.ok(layout.svgWidth <= layout.viewportClient + 1);
          assert.ok(Math.abs(layout.svgWidth / layout.svgHeight - 1.5) < 0.02);
          assert.ok(contrast(layout.titleColor, layout.surfaceColor) >= 4.5, JSON.stringify(layout));
          assert.ok(contrast(layout.markColor, layout.surfaceColor) >= 2, JSON.stringify(layout));
          if (screenshotDirectory) {
            await page.screenshot({ path: path.join(screenshotDirectory, `pie-${width}-${theme}-${mode}.png`), fullPage: true });
          }

          await wrapper.evaluate((element) => { element.__pieContent = element.querySelector("[data-goshtoso-chart-content]"); });
          await openExpand(wrapper);
          const dialog = wrapper.getByRole("dialog", { name: "Doughnut Chart" });
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
              bodyHeight: bodyRect.height,
              chartWidth: svgRect.width,
              chartHeight: svgRect.height,
              panelContained: panelRect.left >= 0 && panelRect.right <= innerWidth + 1 && panelRect.top >= 0 && panelRect.bottom <= innerHeight + 1,
              centered: Math.abs((panelRect.left + panelRect.right) / 2 - innerWidth / 2) < 4,
              chartContained: svgRect.left >= bodyRect.left - 1 && svgRect.right <= bodyRect.right + 1 && svgRect.top >= bodyRect.top - 1 && svgRect.bottom <= bodyRect.bottom + 1,
              uniformScale: Math.abs(Math.abs(matrix.a) - Math.abs(matrix.d)) < 0.01,
              sameContent: panel.closest("[data-goshtoso-chart-wrapper]").__pieContent === body.querySelector("[data-goshtoso-chart-content]"),
            };
          });
          assert.equal(geometry.panelContained, true);
          assert.equal(geometry.centered, true);
          assert.equal(geometry.chartContained, true);
          assert.equal(geometry.uniformScale, true);
          assert.equal(geometry.sameContent, true);
          assert.ok(geometry.panelWidth >= (width === 1440 ? 1000 : width * 0.9));
          assert.ok(geometry.chartWidth >= geometry.bodyWidth * 0.9);
          assert.ok(geometry.chartHeight >= geometry.bodyHeight * 0.9);
          await page.keyboard.press("Escape");
          await dialog.waitFor({ state: "hidden" });
          assert.deepEqual(errors, []);
          assert.deepEqual(failed, []);
        } finally {
          await page.close();
        }
      });
    }
  }
}

test("Doughnut exports parseable 600x400 SVG and opaque PNG", async () => {
  const { page, wrapper } = await piePage();
  try {
    const svg = await download(page, wrapper, "SVG");
    assert.equal(svg.filename, "doughnut-chart.svg");
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
    assert.equal(png.filename, "doughnut-chart.png");
    assert.equal(png.types.at(-1), "image/png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: 600, height: 400 });
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
  } finally {
    await page.close();
  }
});
