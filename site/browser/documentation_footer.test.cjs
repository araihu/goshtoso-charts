const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");

const screenshotDirectory = process.env.GOSHTOSO_SCREENSHOT_DIR;
const routes = [
  ["bar", "/components/bar"],
  ["line", "/components/line"],
  ["sunburst", "/components/interactive/sunburst"],
  ["map", "/components/interactive/map"],
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
      if ((await fetch(`${baseURL}/components/bar`)).ok) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Documentation verification server did not start at ${baseURL}`);
}

function rgb(value) {
  return value.match(/\d+(?:\.\d+)?/g).slice(0, 3).map(Number);
}

function contrastRatio(first, second) {
  const luminance = (value) => rgb(value).map((channel) => {
    const normalized = channel / 255;
    return normalized <= 0.04045 ? normalized / 12.92 : ((normalized + 0.055) / 1.055) ** 2.4;
  }).reduce((sum, channel, index) => sum + channel * [0.2126, 0.7152, 0.0722][index], 0);
  const [bright, dark] = [luminance(first), luminance(second)].sort((a, b) => b - a);
  return (bright + 0.05) / (dark + 0.05);
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
    for (const [name, route] of routes) {
      test(`${name} ${width}px ${mode} page keeps guidance and Go API footer readable`, async () => {
        const page = await browser.newPage({ viewport: { width, height: 900 }, colorScheme: mode });
        try {
          await page.goto(`${baseURL}${route}`);
          await page.evaluate((dark) => document.documentElement.classList.toggle("dark", dark), mode === "dark");
          const footer = page.locator("[data-go-api-reference]");
          await footer.waitFor();
          assert.equal(await page.locator("[data-visualization-guidance]").count(), 1);
          await expectVisibleText(page, "Purpose");
          await expectVisibleText(page, "Use when");
          await expectVisibleText(page, "Avoid when");
          await expectVisibleText(footer, "Open v0.0.1 API");
          assert.equal(await footer.locator("[data-shared-chart-guidance]").count(), 1);
          assert.equal(await footer.locator('a[href="/docs/chart-controls"]').count(), 1);
          assert.equal(await footer.locator('a[href="/docs/chart-modes"]').count(), 1);
          const state = await footer.evaluate((element) => {
            const link = element.querySelector("[data-go-api-link]");
            const footerStyle = getComputedStyle(element);
            const linkStyle = getComputedStyle(link);
            return {
              disabled: link.matches(":disabled") || link.getAttribute("aria-disabled") === "true",
              documentClient: document.documentElement.clientWidth,
              documentScroll: document.documentElement.scrollWidth,
              linkText: linkStyle.color,
              linkBackground: linkStyle.backgroundColor,
              footerText: footerStyle.color,
              footerBackground: footerStyle.backgroundColor,
              apiTarget: link.getAttribute("target"),
              apiRel: link.getAttribute("rel"),
            };
          });
          assert.equal(state.disabled, false);
          assert.equal(state.documentScroll, state.documentClient);
          assert.equal(state.apiTarget, "_blank");
          assert.match(state.apiRel || "", /noopener/);
          assert.ok(contrastRatio(state.linkText, state.linkBackground) >= 4.5, JSON.stringify(state));
          if (name === "sunburst") {
            for (const text of ["shallow hierarchy", "Deep hierarchies", "keyboard navigation", "path-and-value table"]) await expectVisibleText(page, text);
          }
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

async function expectVisibleText(locator, text) {
  await locator.getByText(text, { exact: false }).first().waitFor();
}
