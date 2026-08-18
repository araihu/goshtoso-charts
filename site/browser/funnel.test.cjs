const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

const baseURL = process.env.FUNNEL_BASE_URL || "http://127.0.0.1:8098";
const screenshotDirectory = process.env.SCREENSHOT_DIR;
let browser;
let server;

async function ready() {
  for (let attempt = 0; attempt < 120; attempt += 1) {
    try {
      if ((await fetch(`${baseURL}/components/funnel`)).ok) return;
    } catch {
      // Test-owned server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Funnel verification server did not start at ${baseURL}`);
}

before(async () => {
  if (!process.env.FUNNEL_BASE_URL) {
    try {
      if ((await fetch(`${baseURL}/components/funnel`)).ok) {
        throw new Error(`Refusing to reuse or stop an existing server at ${baseURL}`);
      }
    } catch (error) {
      if (String(error.message).includes("Refusing")) throw error;
    }
    server = spawn("go", ["run", "./cmd/server", "-port", "8098"], {
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

async function funnelPage(viewport = { width: 1440, height: 900 }) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  await page.addInitScript(() => {
    const createObjectURL = URL.createObjectURL.bind(URL);
    globalThis.__funnelBlobTypes = [];
    URL.createObjectURL = (blob) => {
      globalThis.__funnelBlobTypes.push(blob.type);
      return createObjectURL(blob);
    };
  });
  await page.goto(`${baseURL}/components/funnel`);
  await page.locator("[data-goshtoso-chart-wrapper]").first().waitFor();
  await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
  return page;
}

async function openExpand(wrapper) {
  const trigger = wrapper.locator('[id$="-primary-action"]:visible').first();
  await trigger.waitFor({ state: "visible" });
  await trigger.click();
  return trigger;
}

async function download(page, label) {
  await page.evaluate(() => { globalThis.__funnelBlobTypes.length = 0; });
  const pending = page.waitForEvent("download", { timeout: 10000 });
  const match = label.match(/^Download (.+) as (SVG|PNG)$/);
  assert.ok(match);
  const trigger = page.getByRole("button", { name: `Export ${match[1]}` }).first();
  const wrapper = trigger.locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]");
  await trigger.click();
  const menu = wrapper.locator('[role="menu"]:visible');
  await menu.waitFor({ state: "visible" });
  await menu.locator('[role="menuitem"]').filter({ hasText: `Download ${match[2]}` }).first().click();
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
  if (await openMenu.count()) {
    if (await openMenu.isVisible()) await page.keyboard.press("Escape");
    await openMenu.waitFor({ state: "hidden" });
  }
  return {
    filename: artifact.suggestedFilename(),
    bytes: await fs.readFile(artifactPath),
    types: await page.evaluate(() => [...globalThis.__funnelBlobTypes]),
  };
}

for (const width of [390, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const mode of ["light", "dark"]) {
      test(`${width}px ${theme} ${mode} keeps Funnel responsive, themed, accessible, and modal-contained`, async () => {
        const page = await funnelPage({ width, height: 900 });
        const errors = [];
        page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
        page.on("pageerror", (error) => errors.push(error.message));
        try {
          await page.evaluate(({ selected, dark }) => {
            document.documentElement.dataset.theme = selected;
            document.documentElement.classList.toggle("dark", dark);
          }, { selected: theme, dark: mode === "dark" });
          const wrappers = page.locator("[data-goshtoso-chart-wrapper]");
          assert.equal(await wrappers.count(), 2);
          const wrapper = wrappers.filter({ has: page.locator('figure[aria-label="Basic funnel"]') }).first();
          const compact = wrappers.filter({ has: page.locator('figure[aria-label="Compact five-stage funnel"]') }).first();
          assert.equal(await wrapper.getByRole("button").count(), 4);
          assert.equal(await compact.getByRole("button").count(), 4);
          assert.equal(await wrapper.locator("figure").getAttribute("aria-label"), "Basic funnel");
          assert.equal(await wrapper.locator("table").getAttribute("aria-label"), "Basic funnel exact stage values");
          assert.deepEqual(await wrapper.locator("table tbody th").allTextContents(), ["Show", "Click", "Visit", "Inquiry", "Order", "Pay", "Cancel"]);
          assert.deepEqual(await wrapper.locator("table tbody td:nth-child(2)").allTextContents(), ["100", "80", "60", "40", "20", "10", "2"]);
          assert.equal(await compact.locator("figure").getAttribute("aria-label"), "Compact five-stage funnel");
          assert.equal(await compact.locator("table").getAttribute("aria-label"), "Compact five-stage funnel exact stage values");
          assert.deepEqual(await compact.locator("table tbody th").allTextContents(), ["Show", "Click", "Visit", "Inquiry", "Order"]);
          assert.deepEqual(await compact.locator("table tbody td:nth-child(2)").allTextContents(), ["100", "80", "60", "40", "20"]);
          const layouts = await page.evaluate(() => [...document.querySelectorAll("[data-goshtoso-chart-content]")].map((content) => {
            const viewport = content.querySelector(".goshtoso-charts-funnel__viewport");
            const svg = viewport.querySelector("svg");
            const stage = [...svg.querySelectorAll("path")].find((path) => path.getAttribute("style")?.includes("--color-chart-series-1"));
            const rect = svg.getBoundingClientRect();
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
              stageFill: stage && getComputedStyle(stage).fill,
            };
          }));
          assert.equal(layouts.length, 2);
          for (const layout of layouts) {
            assert.equal(layout.documentScroll, layout.documentClient);
            assert.ok(layout.contentScroll <= layout.contentClient + 1);
            assert.ok(layout.viewportScroll <= layout.viewportClient + 1);
            assert.equal(layout.viewBox, "0 0 600 400");
            assert.ok(layout.svgWidth <= layout.viewportClient + 1);
            assert.ok(Math.abs(layout.svgWidth / layout.svgHeight - 1.5) < 0.02);
            assert.ok(layout.stageFill);
          }
          if (screenshotDirectory) await page.screenshot({ path: path.join(screenshotDirectory, `funnel-${width}-${theme}-${mode}.png`), fullPage: true });

          await wrapper.evaluate((element) => { element.__funnelContent = element.querySelector("[data-goshtoso-chart-content]"); });
          await openExpand(wrapper);
          const dialog = wrapper.getByRole("dialog", { name: "Basic funnel" });
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
              sameContent: panel.closest("[data-goshtoso-chart-wrapper]").__funnelContent === body.querySelector("[data-goshtoso-chart-content]"),
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
          if (screenshotDirectory) await page.screenshot({ path: path.join(screenshotDirectory, `funnel-expand-${width}-${theme}-${mode}.png`), fullPage: true });
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

test("Funnel exports parseable deterministic-size SVG and opaque PNG", async () => {
  const page = await funnelPage();
  try {
    const svg = await download(page, "Download Basic funnel as SVG");
    assert.equal(svg.filename, "basic-funnel.svg");
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

    const png = await download(page, "Download Basic funnel as PNG");
    assert.equal(png.filename, "basic-funnel.png");
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
