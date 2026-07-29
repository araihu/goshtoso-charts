const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const net = require("node:net");
const fs = require("node:fs/promises");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");

let baseURL;
let browser;
let server;

async function visibleTimeAxisLabelBoxes(chart) {
  return chart.evaluate((node) => {
    const instance = globalThis.echarts?.getInstanceByDom(node);
    const axis = instance?.getModel().getComponent("xAxis").axis;
    const labels = new Set(axis?.getViewLabels().map((label) => label.formattedLabel));
    return instance?.getZr().storage.getDisplayList()
      .filter((element) => element.type === "tspan" && labels.has(element.style?.text))
      .map((element) => {
        const rect = element.getBoundingRect();
        const transform = element.getComputedTransform?.() || [1, 0, 0, 1, 0, 0];
        const x = transform[0] * rect.x + transform[2] * rect.y + transform[4];
        const y = transform[1] * rect.x + transform[3] * rect.y + transform[5];
        return { text: element.style.text, x, y, width: rect.width, height: rect.height };
      })
      .sort((left, right) => left.x - right.x) || [];
  });
}

function assertSeparateAndContained(boxes, width) {
  assert.ok(boxes.length >= 3, `expected readable temporal labels, got ${JSON.stringify(boxes)}`);
  for (const box of boxes) {
    assert.ok(box.x >= 0 && box.x + box.width <= width, `label clamped: ${JSON.stringify(box)}`);
  }
  for (let index = 1; index < boxes.length; index += 1) {
    const previous = boxes[index - 1];
    const current = boxes[index];
    assert.ok(previous.x + previous.width <= current.x, `labels overlap: ${JSON.stringify([previous, current])}`);
  }
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

async function ready() {
  for (let attempt = 0; attempt < 120; attempt += 1) {
    try {
      const response = await fetch(`${baseURL}/components/interactive/line`);
      if (response.ok && (await response.text()).includes("Exact time and values")) return;
    } catch {
      // Test-owned server still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`interactive Line verification server did not start at ${baseURL}`);
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
    try { process.kill(-server.pid, "SIGTERM"); } catch { /* server stopped */ }
  }
});

for (const [name, viewport, dark] of [
  ["wide-light", { width: 1440, height: 900 }, false],
  ["narrow-dark", { width: 390, height: 844 }, true],
]) {
  test(`temporal Line ${name} renders without page or console errors`, async () => {
    const page = await browser.newPage({ viewport, colorScheme: dark ? "dark" : "light" });
    const failures = [];
    page.on("console", (message) => { if (message.type() === "error") failures.push(message.text()); });
    page.on("pageerror", (error) => failures.push(error.message));
    page.on("response", (response) => { if (response.status() >= 400) failures.push(`${response.status()} ${response.url()}`); });
    try {
      await page.goto(`${baseURL}/components/interactive/line`, { waitUntil: "networkidle" });
      const exact = page.locator("[data-line-time-exact-values]");
      await exact.getByText("Exact time and values").waitFor();
      await exact.locator("summary").click();
      assert.equal(await exact.locator("tbody tr").count(), 50);
      assert.equal(await exact.locator("tbody tr").first().textContent(), "2025-01-31T00:00:00ZCategory A118");
      assert.equal(await exact.locator("tbody tr").last().textContent(), "2025-03-21T00:00:00ZCategory A101");
      const figure = page.locator('figure[aria-label="Temporal X axis"]');
      assert.equal(await figure.count(), 1);
      const chart = figure.locator("div[_echarts_instance_]");
      assert.equal(await chart.count(), 1);
      const options = await chart.evaluateAll((nodes) => nodes.map((node) => {
        const instance = globalThis.echarts?.getInstanceByDom(node);
        return instance?.getOption();
      }).filter(Boolean));
      assert.equal(options.length, 1);
      assert.equal(options[0].title[0].text, "temporal X axis");
      assert.equal(options[0].title[0].subtext, "time.Date as X axis values");
      assert.equal(options[0].xAxis[0].type, "time");
      assert.equal(options[0].xAxis[0].min, "2025-01-01T00:00:00Z");
	  assert.equal(options[0].xAxis[0].splitNumber, 4);
	  assert.equal(options[0].xAxis[0].axisLabel.hideOverlap, true);
      assert.equal(options[0].yAxis[0].min, 0);
      assert.equal(options[0].yAxis[0].max, 200);
      assert.equal(options[0].tooltip[0].trigger, "axis");
	  const boxes = await visibleTimeAxisLabelBoxes(chart);
	  assertSeparateAndContained(boxes, await chart.evaluate((node) => node.clientWidth));
	  assert.ok(boxes.some((box) => box.text.includes("Feb")), JSON.stringify(boxes));
	  assert.ok(boxes.some((box) => box.text.includes("Mar")), JSON.stringify(boxes));
	  assert.match(await exact.textContent(), /Evidence runs from 2025-01-31T00:00:00Z to 2025-03-21T00:00:00Z/);
      if (process.env.GOSHTOSO_SCREENSHOT_DIR) {
        await fs.mkdir(process.env.GOSHTOSO_SCREENSHOT_DIR, { recursive: true });
        await page.screenshot({ path: path.join(process.env.GOSHTOSO_SCREENSHOT_DIR, `interactive-line-time-${name}.png`), fullPage: true });
      }
      assert.deepEqual(failures, []);
    } finally {
      await page.close();
    }
  });
}
