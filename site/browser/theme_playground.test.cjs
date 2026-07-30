const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const crypto = require("node:crypto");
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
      if ((await fetch(`${baseURL}/docs/theme-playground`)).ok) return;
    } catch {
      // Test-owned server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Theme playground server did not start at ${baseURL}`);
}

function digest(buffer) {
  return crypto.createHash("sha256").update(buffer).digest("hex");
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
});

after(async () => {
  await browser?.close();
  if (server?.pid) {
    try { process.kill(-server.pid, "SIGTERM"); } catch { /* already stopped */ }
  }
});

for (const width of [390, 1440]) {
  test(`theme playground isolates parent and fits at ${width}px`, async () => {
    const context = await browser.newContext({ viewport: { width, height: 1000 }, reducedMotion: "reduce" });
    await context.addInitScript(() => localStorage.setItem("theme", "modern"));
    const page = await context.newPage();
    const browserIssues = [];
    page.on("console", (message) => {
      if (["warning", "error"].includes(message.type())) browserIssues.push(`${message.type()}: ${message.text()}`);
    });
    page.on("pageerror", (error) => browserIssues.push(`pageerror: ${error.message}`));

    try {
      await page.goto(`${baseURL}/docs/theme-playground`);
      assert.equal(await page.locator("#componentdocshell-theme-trigger").count(), 0);
      assert.equal(await page.locator('[data-docs-search-item][href="/docs/theme-playground"]').count(), 1);
      const parentBefore = await page.evaluate(() => ({
        theme: document.documentElement.dataset.theme,
        className: document.documentElement.className,
        storedTheme: localStorage.getItem("theme"),
        clientWidth: document.documentElement.clientWidth,
        scrollWidth: document.documentElement.scrollWidth,
      }));
      assert.equal(parentBefore.theme, "araihu");
      assert.equal(parentBefore.storedTheme, "modern");
      assert.equal(parentBefore.scrollWidth, parentBefore.clientWidth, JSON.stringify(parentBefore));

      const iframe = page.locator("[data-theme-playground-frame-host]");
      await iframe.waitFor();
      const child = page.frames().find((frame) => frame.url().endsWith("/docs/theme-playground/frame"));
      assert.ok(child, "same-origin theme frame missing");
      assert.equal(new URL(child.url()).origin, new URL(page.url()).origin);
      await child.waitForFunction(() => document.querySelectorAll("[_echarts_instance_]").length === 2);

      const childBefore = await child.evaluate(() => ({
        theme: document.documentElement.dataset.theme,
        clientWidth: document.documentElement.clientWidth,
        scrollWidth: document.documentElement.scrollWidth,
        innerHeight: window.innerHeight,
        scrollHeight: document.documentElement.scrollHeight,
        gridColumns: getComputedStyle(document.querySelector("[data-theme-playground-grid]")).gridTemplateColumns.split(" ").length,
        ids: [...document.querySelectorAll("[_echarts_instance_]")].map((node) => node.getAttribute("_echarts_instance_")),
        colors: [...document.querySelectorAll('[data-theme-playground-chart="interactive"] [_echarts_instance_]')].map((node) => window.echarts.getInstanceByDom(node).getOption().color[0]),
      }));
      assert.equal(childBefore.theme, "araihu");
      assert.equal(childBefore.scrollWidth, childBefore.clientWidth, JSON.stringify(childBefore));
      assert.ok(childBefore.scrollHeight <= childBefore.innerHeight, JSON.stringify(childBefore));
      assert.equal(childBefore.gridColumns, 2);
      assert.equal(childBefore.ids.length, 2);

      const cards = child.locator("[data-theme-playground-chart]");
      assert.equal(await cards.count(), 4);
      const beforeImages = [];
      for (let index = 0; index < 4; index += 1) beforeImages.push(digest(await cards.nth(index).screenshot({ animations: "disabled" })));

      await child.getByRole("combobox", { name: "Theme" }).click();
      await child.getByRole("option", { name: "Minimal", exact: true }).click();
      await child.waitForFunction((previous) => {
        if (document.documentElement.dataset.theme !== "minimal") return false;
        const nodes = [...document.querySelectorAll('[data-theme-playground-chart="interactive"] [_echarts_instance_]')];
        return nodes.length === previous.length && nodes.every((node, index) => window.echarts.getInstanceByDom(node).getOption().color[0] !== previous[index]);
      }, childBefore.colors);

      const childAfter = await child.evaluate(() => ({
        theme: document.documentElement.dataset.theme,
        ids: [...document.querySelectorAll("[_echarts_instance_]")].map((node) => node.getAttribute("_echarts_instance_")),
        scrollWidth: document.documentElement.scrollWidth,
        clientWidth: document.documentElement.clientWidth,
      }));
      assert.equal(childAfter.theme, "minimal");
      assert.deepEqual(childAfter.ids, childBefore.ids, "interactive chart instances were replaced");
      assert.equal(childAfter.scrollWidth, childAfter.clientWidth, JSON.stringify(childAfter));

      for (let index = 0; index < 4; index += 1) {
        const afterImage = digest(await cards.nth(index).screenshot({ animations: "disabled" }));
        assert.notEqual(afterImage, beforeImages[index], `chart ${index + 1} did not visibly react to theme change`);
      }

      const parentAfter = await page.evaluate(() => ({
        theme: document.documentElement.dataset.theme,
        className: document.documentElement.className,
        storedTheme: localStorage.getItem("theme"),
      }));
      assert.deepEqual(parentAfter, {
        theme: parentBefore.theme,
        className: parentBefore.className,
        storedTheme: parentBefore.storedTheme,
      });
      assert.deepEqual(browserIssues, []);
    } finally {
      await context.close();
    }
  });
}
