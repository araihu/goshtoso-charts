const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");

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
      if ((await fetch(`${baseURL}/`)).ok) return;
    } catch {
      // Test-owned server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`App Shell contract server did not start at ${baseURL}`);
}

before(async () => {
  const port = await randomPort();
  assert.notEqual(port, 8091);
  assert.notEqual(port, 8096);
  baseURL = `http://127.0.0.1:${port}`;
  server = spawn("go", ["run", "./cmd/server", "-port", String(port)], {
    cwd: path.resolve(__dirname, ".."),
    detached: true,
    env: { ...process.env, GOWORK: "off" },
    stdio: "pipe",
  });
  await ready();
  browser = await chromium.launch({ headless: true });
});

after(async () => {
  await browser?.close();
  if (server?.pid) {
    try { process.kill(-server.pid, "SIGTERM"); } catch { /* already stopped */ }
  }
});

for (const width of [390, 1440]) {
  test(`released App Shell keeps Charts consumer policy at ${width}px`, async () => {
    const page = await browser.newPage({ viewport: { width, height: 900 }, reducedMotion: "reduce" });
    const browserIssues = [];
    page.on("console", (message) => {
      if (["warning", "error"].includes(message.type())) browserIssues.push(`${message.type()}: ${message.text()}`);
    });
    page.on("pageerror", (error) => browserIssues.push(`pageerror: ${error.message}`));
    try {
      await page.goto(`${baseURL}/`);
      await page.locator(".component-doc-shell__brand-logo--goshtoso").waitFor();
      assert.equal(await page.locator(".component-doc-shell__brand-name").count(), 0);
      assert.equal(await page.locator('[data-site-version]').count(), 1);
      assert.equal((await page.locator('[data-site-version]').textContent()).trim(), "dev");
      assert.equal(await page.locator('[data-site-version] a').count(), 0);
      assert.equal(await page.locator('#componentdocshell-theme-trigger, [aria-label="Theme"]').count(), 0);
      assert.equal(await page.locator("[data-campaign-toggle]").count(), 0);
      assert.equal(await page.locator('link[rel="stylesheet"][href*="/componentdocshell/assets/shell.css?v="]').count(), 1);
      assert.equal(await page.locator('script[src*="/assets/js/action-group.js"]').count(), 0);
      assert.equal(await page.getByRole("link", { name: "Assets", exact: true }).count(), 0);
      const current = page.locator('[aria-current="page"]');
      assert.equal(await current.getAttribute("href"), "/");
      assert.match((await current.textContent()).trim(), /^Getting started(?:\s+active)?$/);
      assert.equal(await page.getByRole("heading", { name: "Getting Started", exact: true }).count(), 1);

      const state = await page.evaluate(() => {
        const logo = document.querySelector(".component-doc-shell__brand-logo--goshtoso").getBoundingClientRect();
        return {
          theme: document.documentElement.dataset.theme,
          logoWidth: logo.width,
          logoHeight: logo.height,
          clientWidth: document.documentElement.clientWidth,
          scrollWidth: document.documentElement.scrollWidth,
        };
      });
      assert.equal(state.theme, "araihu");
      assert.ok(state.logoWidth > 0 && state.logoHeight > 0, JSON.stringify(state));
      assert.equal(state.scrollWidth, state.clientWidth, JSON.stringify(state));
      assert.deepEqual(browserIssues, []);
    } finally {
      await page.close();
    }
  });
}

test("Theme Playground keeps its picker inside the isolated frame", async () => {
  const page = await browser.newPage({ viewport: { width: 1440, height: 1000 }, reducedMotion: "reduce" });
  const browserIssues = [];
  page.on("console", (message) => {
    if (["warning", "error"].includes(message.type())) browserIssues.push(`${message.type()}: ${message.text()}`);
  });
  page.on("pageerror", (error) => browserIssues.push(`pageerror: ${error.message}`));
  try {
    await page.goto(`${baseURL}/docs/theme-playground`);
    assert.equal(await page.locator('#componentdocshell-theme-trigger, [aria-label="Theme"]').count(), 0);
    assert.equal(await page.locator('[data-site-version]').count(), 1);
    assert.equal(await page.locator("[data-campaign-toggle]").count(), 0);
    const frame = page.frames().find((candidate) => candidate.url().endsWith("/docs/theme-playground/frame"));
    assert.ok(frame, "same-origin Theme Playground frame missing");
    assert.equal(new URL(frame.url()).origin, new URL(page.url()).origin);
    await frame.getByRole("combobox", { name: "Theme" }).waitFor();
    assert.equal(await frame.getByRole("combobox", { name: "Theme" }).count(), 1);
    assert.equal(await frame.locator(".component-doc-shell__header, [data-site-version]").count(), 0);
    assert.equal(await frame.locator('script[src*="/assets/js/action-group.js"]').count(), 0);
    assert.deepEqual(browserIssues, []);
  } finally {
    await page.close();
  }
});
