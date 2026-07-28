const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

const baseURL = process.env.BAR_BASE_URL || "http://127.0.0.1:8104";
const screenshotDirectory = process.env.SCREENSHOT_DIR;
let browser;
let server;

async function ready() {
  for (let attempt = 0; attempt < 120; attempt += 1) {
    try {
      if ((await fetch(`${baseURL}/components/bar`)).ok) return;
    } catch {
      // Test-owned server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Bar verification server did not start at ${baseURL}`);
}

before(async () => {
  if (!process.env.BAR_BASE_URL) {
    try {
      if ((await fetch(`${baseURL}/components/bar`)).ok) {
        throw new Error(`Refusing to reuse or stop an existing server at ${baseURL}`);
      }
    } catch (error) {
      if (String(error.message).includes("Refusing")) throw error;
    }
    server = spawn("go", ["run", "./cmd/server", "-port", "8104"], {
      cwd: path.resolve(__dirname, ".."),
      detached: true,
      stdio: "pipe",
    });
  }
  await ready();
  browser = await chromium.launch({ headless: true });
  if (screenshotDirectory) await fs.mkdir(screenshotDirectory, { recursive: true });
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

async function barPage(viewport = { width: 1440, height: 900 }) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  await page.addInitScript(() => {
    const createObjectURL = URL.createObjectURL.bind(URL);
    globalThis.__barBlobTypes = [];
    URL.createObjectURL = (blob) => {
      globalThis.__barBlobTypes.push(blob.type);
      return createObjectURL(blob);
    };
  });
  await page.goto(`${baseURL}/components/bar`);
  await page.getByRole("img", { name: "World population by reporting series" }).waitFor();
  await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
  return page;
}

function horizontalWrapper(page) {
  return page.getByRole("img", { name: "World population by reporting series" })
    .locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]");
}

async function download(page, format) {
  await page.evaluate(() => { globalThis.__barBlobTypes.length = 0; });
  const wrapper = horizontalWrapper(page);
  const pending = page.waitForEvent("download", { timeout: 10000 });
  await wrapper.getByRole("button", { name: "Export World population by reporting series" }).click();
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
  const openMenu = wrapper.getByRole("menu");
  if (await openMenu.isVisible()) await page.keyboard.press("Escape");
  await openMenu.waitFor({ state: "hidden" });
  return {
    filename: artifact.suggestedFilename(),
    bytes: await fs.readFile(artifactPath),
    types: await page.evaluate(() => [...globalThis.__barBlobTypes]),
  };
}

function contrastRatio(first, second) {
  const parse = (value) => value.match(/\d+(?:\.\d+)?/g).slice(0, 3).map(Number);
  const luminance = (value) => {
    const channels = parse(value).map((channel) => {
      const normalized = channel / 255;
      return normalized <= 0.04045 ? normalized / 12.92 : ((normalized + 0.055) / 1.055) ** 2.4;
    });
    return 0.2126 * channels[0] + 0.7152 * channels[1] + 0.0722 * channels[2];
  };
  const values = [luminance(first), luminance(second)].sort((a, b) => b - a);
  return (values[0] + 0.05) / (values[1] + 0.05);
}

for (const width of [390, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const mode of ["light", "dark"]) {
      test(`${width}px ${theme} ${mode} keeps horizontal Bar responsive, contrasted, accessible, and modal-contained`, async () => {
        const page = await barPage({ width, height: 900 });
        const errors = [];
        page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
        page.on("pageerror", (error) => errors.push(error.message));
        try {
          await page.evaluate(({ selected, dark }) => {
            document.documentElement.dataset.theme = selected;
            document.documentElement.classList.toggle("dark", dark);
          }, { selected: theme, dark: mode === "dark" });
          const wrapper = horizontalWrapper(page);
          assert.equal(await wrapper.getByRole("button").count(), 4);
          assert.equal(await wrapper.locator("table").getAttribute("aria-label"), "World population by reporting series exact category values");
          assert.deepEqual(await wrapper.locator("tbody th").allTextContents(), ["UN", "Brazil", "Indonesia", "USA", "India", "China", "World"]);
          assert.deepEqual(await wrapper.locator("tbody tr").first().locator("td").allTextContents(), ["10", "20"]);
          assert.deepEqual(await wrapper.locator("tbody tr").last().locator("td").allTextContents(), ["130", "140"]);

          const layout = await wrapper.evaluate((element) => {
            const content = element.querySelector("[data-goshtoso-chart-content]");
            const viewport = element.querySelector(".goshtoso-charts-bar__viewport");
            const svg = viewport.querySelector("svg");
            const box = svg.getBoundingClientRect();
            const paths = [...svg.querySelectorAll("path")];
            const rgb = (color) => {
              const canvas = document.createElement("canvas");
              canvas.width = 1;
              canvas.height = 1;
              const context = canvas.getContext("2d");
              context.fillStyle = color;
              context.fillRect(0, 0, 1, 1);
              const [red, green, blue] = context.getImageData(0, 0, 1, 1).data;
              return `rgb(${red}, ${green}, ${blue})`;
            };
            const bars = paths.filter((mark) => {
              const fill = getComputedStyle(mark).fill;
              return fill !== "none" && fill !== "rgba(0, 0, 0, 0)";
            });
            const surface = rgb(getComputedStyle(paths[0]).fill);
            const marks = [];
            const seenFills = new Set();
            for (const mark of bars) {
              const computed = getComputedStyle(mark);
              const fill = rgb(computed.fill);
              if (fill === surface || seenFills.has(fill)) continue;
              seenFills.add(fill);
              marks.push({ fill, stroke: rgb(computed.stroke) });
            }
            return {
              documentClient: document.documentElement.clientWidth,
              documentScroll: document.documentElement.scrollWidth,
              contentClient: content.clientWidth,
              contentScroll: content.scrollWidth,
              viewportClient: viewport.clientWidth,
              viewportScroll: viewport.scrollWidth,
              width: box.width,
              height: box.height,
              viewBox: svg.getAttribute("viewBox"),
              preserveAspectRatio: svg.getAttribute("preserveAspectRatio"),
              surface,
              text: rgb(getComputedStyle(svg.querySelector("text")).fill),
              marks,
            };
          });
          assert.equal(layout.documentScroll, layout.documentClient);
          assert.ok(layout.contentScroll <= layout.contentClient + 1);
          assert.ok(layout.viewportScroll <= layout.viewportClient + 1);
          assert.ok(layout.width <= layout.viewportClient + 1);
          assert.ok(Math.abs(layout.width / layout.height - 1.5) < 0.02);
          assert.equal(layout.viewBox, "0 0 600 400");
          assert.equal(layout.preserveAspectRatio, "xMidYMid meet");
          assert.ok(layout.marks.length >= 2);
          assert.ok(contrastRatio(layout.text, layout.surface) >= 4.5, `${layout.text} text contrast against ${layout.surface}`);
          for (const mark of layout.marks.slice(0, 2)) {
            assert.ok(contrastRatio(mark.fill, layout.surface) >= 1.5, `${mark.fill} differentiation against ${layout.surface}`);
            assert.ok(contrastRatio(mark.stroke, layout.surface) >= 3, `${mark.stroke} boundary contrast against ${layout.surface}`);
          }
          if (screenshotDirectory) {
            await page.screenshot({ path: path.join(screenshotDirectory, `bar-${width}-${theme}-${mode}.png`), fullPage: true });
          }

          await wrapper.evaluate((element) => { element.__barContent = element.querySelector("[data-goshtoso-chart-content]"); });
          await wrapper.locator("[data-goshtoso-chart-expand] > div > button").first().click();
          const dialog = page.getByRole("dialog", { name: "World population by reporting series" });
          await dialog.waitFor({ state: "visible" });
          await page.waitForTimeout(350);
          const geometry = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
            const body = panel.children[1];
            const content = body.querySelector("[data-goshtoso-chart-content]");
            const svg = body.querySelector("svg");
            const panelRect = panel.getBoundingClientRect();
            const bodyRect = body.getBoundingClientRect();
            const svgRect = svg.getBoundingClientRect();
            const matrix = svg.getScreenCTM();
            return {
              panelWidth: panelRect.width,
              panelContained: panelRect.left >= 0 && panelRect.right <= innerWidth + 1 && panelRect.top >= 0 && panelRect.bottom <= innerHeight + 1,
              centered: Math.abs((panelRect.left + panelRect.right) / 2 - innerWidth / 2) < 4,
              chartContained: svgRect.left >= bodyRect.left - 1 && svgRect.right <= bodyRect.right + 1 && svgRect.top >= bodyRect.top - 1 && svgRect.bottom <= bodyRect.bottom + 1,
              uniformScale: Math.abs(Math.abs(matrix.a) - Math.abs(matrix.d)) < 0.01,
              sameContent: panel.closest("[data-goshtoso-chart-wrapper]").__barContent === content,
            };
          });
          assert.deepEqual({
            panelContained: geometry.panelContained,
            centered: geometry.centered,
            chartContained: geometry.chartContained,
            uniformScale: geometry.uniformScale,
            sameContent: geometry.sameContent,
          }, { panelContained: true, centered: true, chartContained: true, uniformScale: true, sameContent: true });
          assert.ok(geometry.panelWidth >= (width === 1440 ? 1000 : width * 0.9));
          await page.keyboard.press("Escape");
          await dialog.waitFor({ state: "hidden" });
          assert.deepEqual(errors, []);
        } finally {
          await page.close();
        }
      });
    }
  }
}

test("horizontal Bar exports self-contained 600x400 SVG and opaque PNG", async () => {
  const page = await barPage();
  try {
    const svg = await download(page, "SVG");
    assert.equal(svg.filename, "world-population.svg");
    assert.equal(svg.types.at(-1), "image/svg+xml;charset=utf-8");
    const markup = svg.bytes.toString("utf8");
    assert.match(markup, /^<svg\b/);
    assert.match(markup, /viewBox="0 0 600 400"/);
    assert.match(markup, /preserveAspectRatio="xMidYMid meet"/);
    assert.doesNotMatch(markup, /(?:var\(|url\(|@import)/i);
    const imageDecode = await page.evaluate(async (source) => {
      const url = URL.createObjectURL(new Blob([source], { type: "image/svg+xml;charset=utf-8" }));
      try {
        const image = new Image();
        image.src = url;
        return await Promise.race([
          image.decode().then(() => ({ state: "decoded", width: image.naturalWidth, height: image.naturalHeight }))
            .catch((error) => ({ state: "error", message: error.message })),
          new Promise((resolve) => setTimeout(() => resolve({ state: "timeout" }), 3000)),
        ]);
      } finally {
        URL.revokeObjectURL(url);
      }
    }, markup);
    assert.deepEqual(imageDecode, { state: "decoded", width: 600, height: 400 });

    const png = await download(page, "PNG");
    assert.equal(png.filename, "world-population.png");
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
