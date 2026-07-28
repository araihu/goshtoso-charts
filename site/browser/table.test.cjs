const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

const baseURL = process.env.TABLE_BASE_URL || "http://127.0.0.1:8097";
const screenshotDirectory = process.env.SCREENSHOT_DIR;
let browser;
let server;

async function ready() {
  for (let attempt = 0; attempt < 120; attempt += 1) {
    try {
      if ((await fetch(`${baseURL}/components/table`)).ok) return;
    } catch {
      // Test-owned server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Table verification server did not start at ${baseURL}`);
}

before(async () => {
  if (!process.env.TABLE_BASE_URL) {
    try {
      if ((await fetch(`${baseURL}/components/table`)).ok) {
        throw new Error(`Refusing to reuse or stop an existing server at ${baseURL}`);
      }
    } catch (error) {
      if (String(error.message).includes("Refusing")) throw error;
    }
    server = spawn("go", ["run", "./cmd/server", "-port", "8097"], {
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

async function tablePage(viewport = { width: 1440, height: 900 }) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  await page.addInitScript(() => {
    const createObjectURL = URL.createObjectURL.bind(URL);
    globalThis.__tableBlobTypes = [];
    URL.createObjectURL = (blob) => {
      globalThis.__tableBlobTypes.push(blob.type);
      return createObjectURL(blob);
    };
  });
  await page.goto(`${baseURL}/components/table`);
  await page.locator("[data-goshtoso-chart-wrapper]").first().waitFor();
  await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
  return page;
}

async function download(page, label) {
  await page.evaluate(() => { globalThis.__tableBlobTypes.length = 0; });
  const pending = page.waitForEvent("download", { timeout: 10000 });
  const match = label.match(/^Download (.+) as (SVG|PNG)$/);
  assert.ok(match);
  const trigger = page.getByRole("button", { name: `Export ${match[1]}` }).first();
  const wrapper = trigger.locator("xpath=ancestor::*[@data-goshtoso-chart-wrapper][1]");
  await trigger.click();
  const menu = wrapper.locator('[role="menu"]:visible');
  await menu.waitFor({ state: "visible" });
  await menu.getByRole("menuitem", { name: match[2], exact: true }).click();
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
  await wrapper.locator("[data-goshtoso-chart-export-status]").evaluate((status) => {
    if (status.textContent.startsWith("Download ready:")) return;
    return new Promise((resolve) => {
      const observer = new MutationObserver(() => {
        if (!status.textContent.startsWith("Download ready:")) return;
        observer.disconnect();
        resolve();
      });
      observer.observe(status, { childList: true, characterData: true, subtree: true });
    });
  });
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
    types: await page.evaluate(() => [...globalThis.__tableBlobTypes]),
  };
}

for (const width of [390, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const mode of ["light", "dark"]) {
      test(`${width}px ${theme} ${mode} keeps Table responsive, themed, accessible, and modal-contained`, async () => {
        const page = await tablePage({ width, height: 900 });
        const errors = [];
        page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
        page.on("pageerror", (error) => errors.push(error.message));
        try {
          await page.evaluate(({ selected, dark }) => {
            document.documentElement.dataset.theme = selected;
            document.documentElement.classList.toggle("dark", dark);
          }, { selected: theme, dark: mode === "dark" });
          const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
          assert.equal(await wrapper.getByRole("button").count(), 4);
          assert.equal(await wrapper.locator("table").getAttribute("aria-label"), "People directory data");
          assert.deepEqual(await wrapper.locator("table thead th").allTextContents(), ["Name", "Age", "Address", "Tag", "Action"]);
          assert.equal(await wrapper.locator("table tbody tr").count(), 3);
          const layout = await page.evaluate(() => {
            const content = document.querySelector("[data-goshtoso-chart-content]");
            const viewport = content.querySelector(".goshtoso-charts-table__viewport");
            const svg = content.querySelector("svg");
            const fill = getComputedStyle(svg.querySelector("path")).fill;
            return {
              documentClient: document.documentElement.clientWidth,
              documentScroll: document.documentElement.scrollWidth,
              contentClient: content.clientWidth,
              contentScroll: content.scrollWidth,
              overflow: getComputedStyle(content).overflowX,
              viewportClient: viewport.clientWidth,
              viewportScroll: viewport.scrollWidth,
              viewportOverflow: getComputedStyle(viewport).overflowX,
              fill,
              viewBox: svg.getAttribute("viewBox"),
            };
          });
          assert.equal(layout.documentScroll, layout.documentClient);
          assert.match(layout.viewBox, /^0 0 810 \d+$/);
          assert.notEqual(layout.fill, "");
          if (width === 390) {
            assert.equal(layout.overflow, "auto");
            assert.equal(layout.viewportOverflow, "auto");
            assert.ok(layout.viewportScroll > layout.viewportClient);
          }
          if (screenshotDirectory) {
            await page.screenshot({ path: path.join(screenshotDirectory, `table-${width}-${theme}-${mode}.png`), fullPage: true });
          }

          await wrapper.locator("[data-goshtoso-chart-expand] > div > button").first().click();
          const dialog = wrapper.getByRole("dialog", { name: "People directory" });
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
              panelContained: panelRect.left >= 0 && panelRect.right <= innerWidth + 1 && panelRect.top >= 0 && panelRect.bottom <= innerHeight + 1,
              centered: Math.abs((panelRect.left + panelRect.right) / 2 - innerWidth / 2) < 4,
              chartContained: svgRect.left >= bodyRect.left && svgRect.right <= bodyRect.right + 1 && svgRect.top >= bodyRect.top && svgRect.bottom <= bodyRect.bottom + 1,
              uniformScale: Math.abs(Math.abs(matrix.a) - Math.abs(matrix.d)) < 0.01,
            };
          });
          assert.equal(geometry.panelContained, true);
          assert.equal(geometry.centered, true);
          assert.equal(geometry.chartContained, true);
          assert.equal(geometry.uniformScale, true);
          assert.ok(geometry.panelWidth >= (width === 1440 ? 1000 : width * 0.9));
          if (screenshotDirectory) {
            await page.screenshot({ path: path.join(screenshotDirectory, `table-expand-${width}-${theme}-${mode}.png`), fullPage: true });
          }
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

test("exported Table SVG is self-contained and independently image-decodable", async () => {
	const page = await tablePage();
	try {
		const svg = await download(page, "Download People directory as SVG");
		const markup = svg.bytes.toString("utf8");
		assert.equal(svg.types.at(-1), "image/svg+xml;charset=utf-8");
		assert.doesNotMatch(markup, /(?:var\(|url\(|@import)/i);
		const viability = await page.evaluate(async (source) => {
			const document = new DOMParser().parseFromString(source, "image/svg+xml");
			const root = document.documentElement;
			const viewBox = root.getAttribute("viewBox").split(/\s+/).map(Number);
			const externalReferences = [...document.querySelectorAll("[href]")]
				.map((element) => element.getAttribute("href"))
				.filter((value) => value && !value.startsWith("#"));
			const fontFamilies = [...document.querySelectorAll("[style]")]
				.map((element) => element.style.fontFamily)
				.filter(Boolean);
			const blob = new Blob([source], { type: "image/svg+xml;charset=utf-8" });
			const url = URL.createObjectURL(blob);
			try {
				const image = new Image();
				const loaded = new Promise((resolve, reject) => {
					image.addEventListener("load", resolve, { once: true });
					image.addEventListener("error", () => reject(new Error("independent SVG image load failed")), { once: true });
				});
				image.src = url;
				await loaded;
				await image.decode();
				return {
					viewBox,
					width: Number(root.getAttribute("width")),
					height: Number(root.getAttribute("height")),
					naturalWidth: image.naturalWidth,
					naturalHeight: image.naturalHeight,
					externalReferences,
					fontFamilies: [...new Set(fontFamilies)],
					parserErrors: document.querySelectorAll("parsererror").length,
				};
			} finally {
				URL.revokeObjectURL(url);
			}
		}, markup);
		assert.equal(viability.parserErrors, 0);
		assert.deepEqual(viability.externalReferences, []);
		assert.deepEqual(viability.viewBox.slice(0, 3), [0, 0, 810]);
		assert.ok(viability.viewBox[3] > 100 && viability.viewBox[3] < 400);
		assert.deepEqual(
			{ width: viability.width, height: viability.height, naturalWidth: viability.naturalWidth, naturalHeight: viability.naturalHeight },
			{ width: 810, height: viability.viewBox[3], naturalWidth: 810, naturalHeight: viability.viewBox[3] },
		);
		assert.ok(viability.fontFamilies.length > 0);
	} finally {
		await page.close();
	}
});

test("Table supports explicit optional controls and SVG/opaque PNG exports", async () => {
  const page = await tablePage();
  try {
    const wrappers = page.locator("[data-goshtoso-chart-wrapper]");
    assert.equal(await wrappers.nth(0).locator('[data-goshtoso-chart-control="collapse"]').count(), 1);
    assert.equal(await wrappers.nth(0).locator('[data-goshtoso-chart-control="fullscreen"]').count(), 1);
    assert.equal(await wrappers.nth(1).locator('[data-goshtoso-chart-control="collapse"]').count(), 0);
    assert.equal(await wrappers.nth(1).locator('[data-goshtoso-chart-control="fullscreen"]').count(), 0);

    const fullscreen = wrappers.nth(0).getByRole("button", { name: /^Enter fullscreen / });
    await fullscreen.click();
    await page.waitForFunction(() => document.fullscreenElement !== null);
    await page.evaluate(() => document.exitFullscreen());
    await page.waitForFunction(() => document.fullscreenElement === null);

    const svg = await download(page, "Download People directory as SVG");
    assert.equal(svg.filename, "people-directory.svg");
    assert.equal(svg.types.at(-1), "image/svg+xml;charset=utf-8");
    const markup = svg.bytes.toString("utf8");
    assert.match(markup, /^<svg\b/);
    assert.match(markup, /viewBox="0 0 810 \d+"/);
    assert.doesNotMatch(markup, /var\(/);

    const png = await download(page, "Download People directory as PNG");
    assert.equal(png.filename, "people-directory.png");
    assert.equal(png.types.at(-1), "image/png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.equal(metadata.width, 810);
    assert.ok(metadata.height > 100 && metadata.height < 400);
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
  } finally {
    await page.close();
  }
});
