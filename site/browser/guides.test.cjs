const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");

const screenshotDirectory = process.env.GOSHTOSO_SCREENSHOT_DIR;
const routes = [
  ["chart-modes", "/docs/chart-modes", "[data-chart-modes-guide]"],
  ["chart-controls", "/docs/chart-controls", "[data-chart-controls-guide]"],
];
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
      if ((await fetch(`${baseURL}/docs/chart-modes`)).ok) return;
    } catch {
      // Test-owned server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Guide verification server did not start at ${baseURL}`);
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

for (const width of [390, 1440]) {
  for (const mode of ["light", "dark"]) {
    for (const [name, route, selector] of routes) {
      test(`${name} remains readable at ${width}px in ${mode} mode`, async () => {
        const page = await browser.newPage({ viewport: { width, height: 900 }, colorScheme: mode });
        const browserIssues = [];
        page.on("console", (message) => {
          if (["warning", "error"].includes(message.type())) browserIssues.push(`${message.type()}: ${message.text()}`);
        });
        page.on("pageerror", (error) => browserIssues.push(`pageerror: ${error.message}`));
        try {
          await page.goto(`${baseURL}${route}`);
          await page.evaluate((dark) => document.documentElement.classList.toggle("dark", dark), mode === "dark");
          await page.locator(selector).waitFor();
          await page.locator(`[data-chart-guide-nav] [aria-current="page"]`).waitFor();
          await page.getByText("pkg.go.dev is the canonical reference", { exact: false }).waitFor();
          if (name === "chart-controls") {
            await page.locator("[data-wrapper-mode-comparison]").waitFor();
            await page.getByText("goshtoso-charts:set-wrapper-mode", { exact: true }).waitFor();
            await page.getByRole("heading", { name: "No-JavaScript behavior", exact: true }).waitFor();
            await page.getByRole("heading", { name: "Caller responsibilities", exact: true }).waitFor();
            await page.getByText("window.__goshtosoChartsControls.setWrapperMode", { exact: false }).first().waitFor();
          } else {
            await page.getByText("chart controls and wrapper lifecycle", { exact: true }).waitFor();
          }
          const state = await page.evaluate(() => ({
            clientWidth: document.documentElement.clientWidth,
            scrollWidth: document.documentElement.scrollWidth,
            mode: document.documentElement.classList.contains("dark") ? "dark" : "light",
            tableOverflow: [...document.querySelectorAll(".overflow-x-auto")].every((element) => element.scrollWidth >= element.clientWidth),
            activeGuide: document.querySelector("[data-chart-guide-nav] [aria-current='page']")?.textContent.trim(),
            apiLinks: document.querySelectorAll("[data-guide-api-link]").length,
            duplicateIDs: [...document.querySelectorAll("[id]")]
              .map((element) => element.id)
              .filter((id, index, ids) => ids.indexOf(id) !== index),
          }));
          assert.equal(state.scrollWidth, state.clientWidth, JSON.stringify(state));
          assert.equal(state.mode, mode);
          assert.equal(state.tableOverflow, true);
          assert.ok(state.activeGuide.length > 0);
          assert.ok(state.apiLinks >= 3);
          assert.deepEqual(state.duplicateIDs, []);
          assert.deepEqual(browserIssues, []);
          const text = (await page.locator("main").innerText()).toLowerCase();
          for (const forbidden of ["go-echarts", "apache echarts", "go-analyze/charts"]) assert.equal(text.includes(forbidden), false);
          if (screenshotDirectory) {
            await page.screenshot({ path: path.join(screenshotDirectory, `${name}-${width}-${mode}.png`), fullPage: true });
          }
        } finally {
          await page.close();
        }
      });
    }
  }
}

for (const width of [390, 1440]) {
  for (const mode of ["light", "dark"]) {
    test(`Getting Started exposes the shared decisions at ${width}px in ${mode} mode`, async () => {
      const page = await browser.newPage({ viewport: { width, height: 900 }, colorScheme: mode });
      try {
        await page.goto(`${baseURL}/`);
        await page.evaluate((dark) => document.documentElement.classList.toggle("dark", dark), mode === "dark");
        const section = page.locator("#choose-delivery-and-wrapper-behavior").locator("..");
        await section.waitFor();
        assert.equal(await section.locator('a[href="/docs/chart-modes"]').count(), 1);
        assert.equal(await section.locator('a[href="/docs/chart-controls"]').count(), 1);
        const geometry = await page.evaluate(() => ({
          clientWidth: document.documentElement.clientWidth,
          scrollWidth: document.documentElement.scrollWidth,
        }));
        assert.equal(geometry.scrollWidth, geometry.clientWidth, JSON.stringify(geometry));
        if (screenshotDirectory) {
          await page.screenshot({ path: path.join(screenshotDirectory, `getting-started-${width}-${mode}.png`), fullPage: true });
        }
      } finally {
        await page.close();
      }
    });
  }
}

test("guide links navigate with HTMX and update active state", async () => {
  const page = await browser.newPage({ viewport: { width: 960, height: 900 } });
  try {
    await page.goto(`${baseURL}/docs/chart-modes`);
    await page.locator("[data-chart-guide-nav] a", { hasText: "Chart controls" }).click();
    await page.waitForURL(`${baseURL}/docs/chart-controls`);
    await page.locator("[data-chart-controls-guide]").waitFor();
    assert.equal(await page.locator("[data-chart-guide-nav] [aria-current='page']").innerText(), "Chart controls");
    assert.match(await page.title(), /^Chart controls/);
    await page.locator("#wrapper-lifecycle").waitFor();
    assert.equal(await page.locator("[data-wrapper-mode-comparison] tbody tr").count(), 4);
  } finally {
    await page.close();
  }
});

test("documentation search finds shared wrapper and mode behavior", async () => {
  const page = await browser.newPage({ viewport: { width: 1440, height: 900 } });
  try {
    await page.goto(`${baseURL}/docs/chart-modes`);
    const input = page.locator("#docs-search");
    for (const [query, label] of [
      ["disabled hidden omitted", "Chart controls"],
      ["htmx alpine", "Chart controls"],
      ["export svg png", "Chart controls"],
      ["static vector interactive", "Static and interactive"],
    ]) {
      await input.fill(query);
      const visible = page.locator("[data-docs-search-item]:visible");
      assert.ok(await visible.count() >= 1, query);
      assert.ok(await visible.filter({ hasText: label }).count() >= 1, query);
    }
  } finally {
    await page.close();
  }
});
