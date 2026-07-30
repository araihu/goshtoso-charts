const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

let baseURL;
let browser;
let server;
const goExecutable = "go";

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
      const response = await fetch(`${baseURL}/components/interactive/funnel`);
      if (response.ok && (await response.text()).includes('data-funnel-variant="labels-left"')) return;
    } catch { /* test-owned server still starting */ }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`interactive Funnel verification server did not start at ${baseURL}`);
}

async function spawnServer(port, executable = goExecutable) {
  return new Promise((resolve, reject) => {
    const child = spawn(executable, ["run", "./cmd/server", "-port", String(port)], {
      cwd: path.resolve(__dirname, ".."), detached: true, stdio: "pipe",
    });
    child.stdout.resume();
    child.stderr.resume();
    child.once("error", reject);
    child.once("spawn", () => resolve(child));
  });
}

before(async () => {
  const port = await randomPort();
  assert.notEqual(port, 8091);
  assert.notEqual(port, 8096);
  baseURL = `http://127.0.0.1:${port}`;
  server = await spawnServer(port);
  await ready();
  browser = await chromium.launch({ headless: true });
});

after(async () => {
  await browser?.close();
  if (server?.pid) {
    try { process.kill(-server.pid, "SIGTERM"); } catch { /* already stopped */ }
  }
});

test("interactive Funnel browser harness resolves Go from PATH", () => {
  assert.equal(path.isAbsolute(goExecutable), false);
  assert.equal(goExecutable, "go");
});

test("interactive Funnel browser harness reports an unavailable Go executable", async () => {
  await assert.rejects(spawnServer(0, "goshtoso-charts-missing-go-executable"), { code: "ENOENT" });
});

function variant(page, name) {
  return page.locator(`[data-funnel-variant="${name}"]`);
}

async function optionFor(page, name) {
  return variant(page, name).locator("[_echarts_instance_]").evaluate((host) => globalThis.echarts.getInstanceByDom(host).getOption());
}

function contrast(first, second) {
  const luminance = (rgb) => {
    const linear = rgb.map((value) => value / 255).map((value) => value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4);
    return 0.2126 * linear[0] + 0.7152 * linear[1] + 0.0722 * linear[2];
  };
  const values = [luminance(first), luminance(second)].sort((left, right) => right - left);
  return (values[0] + 0.05) / (values[1] + 0.05);
}

for (const [name, viewport, theme, dark] of [
  ["wide-light", { width: 1440, height: 900 }, "goshtoso", false],
  ["wide-dark", { width: 1440, height: 900 }, "araihu", true],
  ["narrow-light", { width: 390, height: 844 }, "araihu", false],
  ["narrow-dark", { width: 390, height: 844 }, "goshtoso", true],
]) {
  test(`interactive Funnel ${name} preserves both upstream treatments and contained theme layout`, async () => {
    const page = await browser.newPage({ viewport, colorScheme: dark ? "dark" : "light" });
    const failures = [];
    page.on("console", (message) => { if (message.type() === "error") failures.push(message.text()); });
    page.on("pageerror", (error) => failures.push(error.message));
    page.on("response", (response) => { if (response.status() >= 400) failures.push(`${response.status()} ${response.url()}`); });
    try {
      await page.goto(`${baseURL}/components/interactive/funnel`, { waitUntil: "domcontentloaded" });
      await page.evaluate(({ selected, darkMode }) => {
        document.documentElement.dataset.theme = selected;
        document.documentElement.classList.toggle("dark", darkMode);
        globalThis.__goshtosoChartsThemeRuntime?.refresh();
      }, { selected: theme, darkMode: dark });
      await page.waitForFunction(() => {
        const hosts = [...document.querySelectorAll("[data-funnel-variant] [_echarts_instance_]")];
        return hosts.length === 2 && hosts.every((host) => {
          const chart = globalThis.echarts?.getInstanceByDom(host);
          return chart && chart.getWidth() === host.clientWidth && chart.getHeight() === host.clientHeight;
        });
      });

      const base = await optionFor(page, "base");
      assert.equal(base.title[0].text, "basic funnel example");
      assert.equal(base.series[0].name, "Analytics");
      assert.equal(base.series[0].sort, "descending");
      assert.deepEqual(base.series[0].data.map(({ name: stage, value }) => [stage, value]), [
        ["Visit", 31], ["Add", 37], ["Order", 47], ["Payment", 9], ["Deal", 31],
      ]);

      const labels = await optionFor(page, "labels-left");
      assert.equal(labels.title[0].text, "show label");
      assert.equal(labels.series[0].label.show, true);
      assert.equal(labels.series[0].label.position, "left");
      assert.deepEqual(labels.series[0].data.map(({ name: stage, value }) => [stage, value]), [
        ["Visit", 18], ["Add", 25], ["Order", 40], ["Payment", 6], ["Deal", 0],
      ]);

      for (const [variantName, accessibleName] of [["base", "Basic five-stage funnel"], ["labels-left", "Funnel with left labels"]]) {
        const current = variant(page, variantName);
        const wrapper = current.locator("[data-goshtoso-chart-wrapper]");
        const details = current.locator("details[data-funnel-exact-values]");
        assert.equal(await wrapper.locator("figure").getAttribute("aria-label"), accessibleName);
        assert.equal(await details.locator("table").getAttribute("aria-label"), `${accessibleName} exact funnel values`);
        assert.match(await details.locator("summary").textContent(), /Exact funnel values/);
        await details.locator("summary").click();
        assert.equal(await details.locator("tbody tr").count(), 5);
        assert.deepEqual(await details.locator("tbody th").allTextContents(), ["Visit", "Add", "Order", "Payment", "Deal"]);
        const geometry = await wrapper.evaluate((element) => {
          const content = element.querySelector("[data-goshtoso-chart-content]");
          const figure = content.querySelector("figure");
          const host = figure.querySelector("[_echarts_instance_]");
          const chart = globalThis.echarts.getInstanceByDom(host);
          const hostRect = host.getBoundingClientRect();
          const contentRect = content.getBoundingClientRect();
          const resolve = (value) => {
            const canvas = document.createElement("canvas");
            const context = canvas.getContext("2d");
            context.fillStyle = value;
            context.fillRect(0, 0, 1, 1);
            return [...context.getImageData(0, 0, 1, 1).data.slice(0, 3)];
          };
          const option = chart.getOption();
          return {
            hostWidth: host.clientWidth,
            hostHeight: host.clientHeight,
            chartWidth: chart.getWidth(),
            chartHeight: chart.getHeight(),
            contentWidth: content.clientWidth,
            centered: Math.abs((hostRect.left + hostRect.right) / 2 - (contentRect.left + contentRect.right) / 2) < 2,
            surface: resolve(option.backgroundColor),
            firstSeries: resolve(option.color[0]),
            colors: option.color,
          };
        });
        assert.deepEqual([geometry.chartWidth, geometry.chartHeight], [geometry.hostWidth, geometry.hostHeight]);
        assert.equal(geometry.hostHeight, 420);
        assert.equal(geometry.centered, true);
        assert.ok(geometry.hostWidth > 0 && geometry.hostWidth <= Math.min(1024, geometry.contentWidth), JSON.stringify(geometry));
        assert.ok(geometry.colors.length >= 5, JSON.stringify(geometry));
        assert.ok(contrast(geometry.surface, geometry.firstSeries) >= 1.5, JSON.stringify(geometry));
        if (process.env.GOSHTOSO_SCREENSHOT_DIR) {
          await fs.mkdir(process.env.GOSHTOSO_SCREENSHOT_DIR, { recursive: true });
          await wrapper.screenshot({ path: path.join(process.env.GOSHTOSO_SCREENSHOT_DIR, `interactive-funnel-${name}-${variantName}.png`) });
        }
      }

      assert.equal(await variant(page, "base").locator("tbody tr").first().textContent(), "AnalyticsVisit31");
      assert.equal(await variant(page, "labels-left").locator("tbody tr").last().textContent(), "AnalyticsDeal0");
      const documentWidths = await page.evaluate(() => ({ client: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth }));
      assert.equal(documentWidths.scroll, documentWidths.client, JSON.stringify(documentWidths));
      assert.ok(documentWidths.client <= viewport.width && documentWidths.client >= viewport.width - 20, JSON.stringify(documentWidths));
      assert.deepEqual(failures, []);
    } finally {
      await page.close();
    }
  });
}

async function openExpand(wrapper) {
  const trigger = wrapper.locator('[id$="-stacked"] > button:visible, [data-action-group-primary] button:visible').first();
  const stacked = await trigger.evaluate((button) => Boolean(button.closest('[id$="-stacked"]')));
  if (stacked) {
    await trigger.evaluate((button) => new Promise((resolve, reject) => {
      const deadline = performance.now() + 5000;
      const ready = () => {
        if (button.hasAttribute("aria-expanded")) return resolve();
        if (performance.now() >= deadline) return reject(new Error("stacked Expand control did not initialize"));
        requestAnimationFrame(ready);
      };
      ready();
    }));
  }
  await trigger.click();
  if (stacked) {
    await trigger.evaluate((button) => new Promise((resolve, reject) => {
      const deadline = performance.now() + 5000;
      const open = () => {
        if (button.getAttribute("aria-expanded") === "true") return resolve();
        if (performance.now() >= deadline) return reject(new Error("stacked Expand control did not open"));
        requestAnimationFrame(open);
      };
      open();
    }));
    const action = wrapper.locator('[id$="-chart-expand-action"]').first();
    await action.waitFor({ state: "visible" });
    await action.click();
  }
}

async function enterFullscreen(wrapper) {
  const trigger = wrapper.locator('[id$="-stacked"] > button:visible, [data-action-group-primary] button:visible').first();
  await trigger.evaluate((button) => new Promise((resolve, reject) => {
    const deadline = performance.now() + 5000;
    const ready = () => {
      if (button.hasAttribute("aria-expanded")) return resolve();
      if (performance.now() >= deadline) return reject(new Error("stacked fullscreen control did not initialize"));
      requestAnimationFrame(ready);
    };
    ready();
  }));
  await trigger.click();
  await trigger.evaluate((button) => new Promise((resolve, reject) => {
    const deadline = performance.now() + 5000;
    const open = () => {
      if (button.getAttribute("aria-expanded") === "true") return resolve();
      if (performance.now() >= deadline) return reject(new Error("stacked fullscreen control did not open"));
      requestAnimationFrame(open);
    };
    open();
  }));
  await wrapper.locator('[id$="-fullscreen-action"]').first().click();
}

test("interactive Funnel preserves one instance through resize, lifecycle, modal, theme, fullscreen, and PNG export", async () => {
  const page = await browser.newPage({ viewport: { width: 1280, height: 900 }, acceptDownloads: true });
  const failures = [];
  page.on("console", (message) => { if (message.type() === "error") failures.push(message.text()); });
  page.on("pageerror", (error) => failures.push(error.message));
  try {
    await page.goto(`${baseURL}/components/interactive/funnel`, { waitUntil: "domcontentloaded" });
    const current = variant(page, "base");
    const wrapper = current.locator("[data-goshtoso-chart-wrapper]");
    await page.waitForFunction(() => Boolean(globalThis.echarts?.getInstanceByDom(document.querySelector('[data-funnel-variant="base"] [_echarts_instance_]'))));
    const before = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      element.__funnelInstance = globalThis.echarts.getInstanceByDom(host);
      return { width: element.__funnelInstance.getWidth(), colors: element.__funnelInstance.getOption().color };
    });

    await page.setViewportSize({ width: 900, height: 900 });
    await page.waitForFunction((oldWidth) => {
      const wrapper = document.querySelector('[data-funnel-variant="base"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const chart = globalThis.echarts.getInstanceByDom(host);
      return chart === wrapper.__funnelInstance && chart.getWidth() !== oldWidth && chart.getWidth() === host.clientWidth;
    }, before.width);

    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "Basic five-stage funnel" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForFunction(() => {
      const panel = document.querySelector('[data-funnel-variant="base"] .goshtoso-charts-expand-panel');
      const host = panel?.querySelector("[_echarts_instance_]");
      const chart = host && globalThis.echarts.getInstanceByDom(host);
      if (!chart || chart.getWidth() !== host.clientWidth || chart.getHeight() !== host.clientHeight) return false;
      const rect = panel.getBoundingClientRect();
      return Math.abs((rect.left + rect.right) / 2 - innerWidth / 2) < 4;
    });
    const modal = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
      const body = panel.children[1];
      const host = body.querySelector("[_echarts_instance_]");
      const wrapper = panel.closest("[data-goshtoso-chart-wrapper]");
      const chart = globalThis.echarts.getInstanceByDom(host);
      const panelRect = panel.getBoundingClientRect();
      const bodyRect = body.getBoundingClientRect();
      const hostRect = host.getBoundingClientRect();
      return {
        same: chart === wrapper.__funnelInstance,
        centered: Math.abs((panelRect.left + panelRect.right) / 2 - innerWidth / 2) < 4,
        contained: hostRect.left >= bodyRect.left - 1 && hostRect.right <= bodyRect.right + 1 && hostRect.top >= bodyRect.top - 1 && hostRect.bottom <= bodyRect.bottom + 1,
        exactSize: chart.getWidth() === host.clientWidth && chart.getHeight() === host.clientHeight,
      };
    });
    assert.deepEqual(modal, { same: true, centered: true, contained: true, exactSize: true });
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    await page.evaluate(() => {
      document.documentElement.dataset.theme = "araihu";
      document.documentElement.classList.add("dark");
      globalThis.__goshtosoChartsThemeRuntime.refresh();
    });
    await page.waitForFunction((oldColors) => {
      const wrapper = document.querySelector('[data-funnel-variant="base"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const chart = globalThis.echarts.getInstanceByDom(host);
      return chart === wrapper.__funnelInstance && JSON.stringify(chart.getOption().color) !== JSON.stringify(oldColors);
    }, before.colors);

    await enterFullscreen(wrapper);
    await page.waitForFunction(() => document.fullscreenElement !== null);
    await page.waitForFunction(() => {
      const wrapper = document.querySelector('[data-funnel-variant="base"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const chart = globalThis.echarts.getInstanceByDom(host);
      return chart === wrapper.__funnelInstance && chart.getWidth() === host.clientWidth && chart.getHeight() === host.clientHeight;
    });
    await page.evaluate(() => document.exitFullscreen());
    await page.waitForFunction(() => document.fullscreenElement === null);

    for (const mode of ["disabled", "hidden", "enabled"]) {
      await wrapper.evaluate((element, nextMode) => element.dispatchEvent(new CustomEvent("goshtoso-charts:set-wrapper-mode", { bubbles: true, detail: { mode: nextMode } })), mode);
      await page.waitForFunction((nextMode) => {
        const wrapper = document.querySelector('[data-funnel-variant="base"] [data-goshtoso-chart-wrapper]');
        if (wrapper.dataset.goshtosoChartWrapperMode !== nextMode) return false;
        const fieldset = wrapper.querySelector("[data-goshtoso-chart-actions-fieldset]");
        if (nextMode === "hidden") return wrapper.hidden === true && wrapper.hasAttribute("inert");
        if (nextMode === "disabled") return wrapper.hidden === false && fieldset?.disabled === true;
        return wrapper.hidden === false && !wrapper.hasAttribute("inert") && fieldset?.disabled === false;
      }, mode);
      assert.equal(await wrapper.evaluate((element) => globalThis.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")) === element.__funnelInstance), true);
    }

    const pending = page.waitForEvent("download", { timeout: 10000 });
    await wrapper.getByRole("button", { name: "Download Basic five-stage funnel as PNG", exact: true }).click();
    const download = await pending;
    assert.equal(download.suggestedFilename(), "basic-five-stage-funnel.png");
    const artifactPath = await download.path();
    assert.ok(artifactPath);
    const bytes = await fs.readFile(artifactPath);
    assert.deepEqual([...bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(bytes).metadata();
    assert.ok(metadata.width > 0 && metadata.height > 0, JSON.stringify(metadata));
    const pixels = await sharp(bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
    assert.deepEqual(failures, []);
  } finally {
    await page.close();
  }
});
