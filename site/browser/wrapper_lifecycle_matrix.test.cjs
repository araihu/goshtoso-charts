const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const path = require("node:path");
const { execFileSync, spawn } = require("node:child_process");
const { chromium } = require("playwright");

const testPort = process.env.TEST_PORT || String(20000 + Math.floor(Math.random() * 30000));
const baseURL = process.env.BASE_URL || `http://127.0.0.1:${testPort}`;
let browser;
let server;

const staticRoutes = [
  "/components/line", "/components/bar", "/components/pie", "/components/scatter",
  "/components/radar", "/components/candlestick", "/components/funnel", "/components/heatmap",
  "/components/table", "/components/violin",
];

const interactiveRoutes = [
  "/components/interactive/bar", "/components/interactive/line", "/components/interactive/scatter",
  "/components/interactive/scatter-3d", "/components/interactive/bar-3d",
  "/components/interactive/surface-3d", "/components/interactive/line-3d",
  "/components/interactive/pie", "/components/interactive/radar", "/components/interactive/heatmap",
  "/components/interactive/boxplot", "/components/interactive/gauge",
  "/components/interactive/funnel", "/components/interactive/graph",
  "/components/interactive/sankey", "/components/interactive/tree",
  "/components/interactive/sunburst", "/components/interactive/treemap",
  "/components/interactive/parallel", "/components/interactive/theme-river",
  "/components/interactive/candlestick", "/components/interactive/word-cloud",
  "/components/interactive/map", "/components/interactive/geo",
];

async function ready() {
  for (let attempt = 0; attempt < 120; attempt += 1) {
    try {
      if ((await fetch(`${baseURL}/components/line`)).ok) return;
    } catch {
      // Test-owned server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Chart verification server did not start at ${baseURL}`);
}

before(async () => {
  if (!process.env.BASE_URL) {
    try {
      if ((await fetch(`${baseURL}/components/line`)).ok) {
        throw new Error(`Refusing to reuse or stop an existing server at ${baseURL}`);
      }
    } catch (error) {
      if (String(error.message).includes("Refusing")) throw error;
    }
    server = spawn("go", ["run", "./cmd/server", "-port", new URL(baseURL).port], {
      cwd: path.resolve(__dirname, ".."),
      detached: true,
      stdio: "pipe",
    });
    server.stdout.resume();
    server.stderr.resume();
  }
  await ready();
  browser = await chromium.launch({ headless: true });
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

test("all 34 public chart pages preserve one renderer-neutral wrapper lifecycle", async () => {
  assert.equal(staticRoutes.length, 10);
  assert.equal(interactiveRoutes.length, 24);

  for (const route of [...staticRoutes, ...interactiveRoutes]) {
    const page = await browser.newPage({ viewport: { width: 768, height: 900 } });
    const errors = [];
    page.on("pageerror", (error) => errors.push(String(error)));
    page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
    try {
      await page.goto(`${baseURL}${route}`);
      const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
      await wrapper.waitFor({ state: "visible" });
      await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
      await page.waitForFunction(() => document.querySelector("[data-goshtoso-chart-wrapper]")?.dataset.goshtosoChartWrapperInitialized === "true");
      if (route.includes("/interactive/")) {
        await page.waitForFunction(() => {
          const host = document.querySelector("[data-goshtoso-chart-wrapper] [_echarts_instance_]");
          return Boolean(host && window.echarts?.getInstanceByDom(host));
        });
      }

      await wrapper.evaluate((element, interactive) => {
        element.__wrapperMatrixContent = element.querySelector("[data-goshtoso-chart-content]");
        element.__wrapperMatrixFigure = element.__wrapperMatrixContent.querySelector("figure");
        element.__wrapperMatrixSVG = element.querySelector("[data-goshtoso-chart-content] svg");
        element.__wrapperMatrixChanges = [];
        element.addEventListener("goshtoso-charts:wrapper-mode-change", (event) => {
          element.__wrapperMatrixChanges.push([event.detail.previousMode, event.detail.mode]);
        });
        if (interactive) {
          const host = element.querySelector("[_echarts_instance_]");
          element.__wrapperMatrixHost = host;
          element.__wrapperMatrixInstance = window.echarts.getInstanceByDom(host);
        }
      }, route.includes("/interactive/"));

      assert.equal(await wrapper.evaluate((element) => window.__goshtosoChartsControls.setWrapperMode(element, "disabled")), true, `${route} disabled transition`);
      const disabled = await wrapper.evaluate((element) => {
        const fieldset = element.querySelector("[data-goshtoso-chart-actions-fieldset]");
        const action = fieldset && fieldset.querySelector("button, [role=menuitem], a");
        if (action) window.__goshtosoChartsControls.expandFromMenu(action);
        return {
          mode: element.dataset.goshtosoChartWrapperMode,
          visible: element.getBoundingClientRect().height > 0,
          fieldsetDisabled: Boolean(fieldset && fieldset.disabled),
          expanded: element.classList.contains("goshtoso-charts-expanded"),
          sameContent: element.__wrapperMatrixContent === element.querySelector("[data-goshtoso-chart-content]"),
          sameFigure: element.__wrapperMatrixFigure === element.querySelector("[data-goshtoso-chart-content] figure"),
        };
      });
      assert.deepEqual(disabled, {
        mode: "disabled", visible: true, fieldsetDisabled: true, expanded: false,
        sameContent: true, sameFigure: true,
      }, `${route} disabled state`);

      await wrapper.evaluate((element) => element.dispatchEvent(new CustomEvent("goshtoso-charts:set-wrapper-mode", {
        bubbles: true,
        detail: { mode: "hidden" },
      })));
      assert.deepEqual(await wrapper.evaluate((element) => ({
        mode: element.dataset.goshtosoChartWrapperMode,
        hidden: element.hidden,
        inert: element.hasAttribute("inert"),
        ariaHidden: element.getAttribute("aria-hidden"),
        sameContent: element.__wrapperMatrixContent === element.querySelector("[data-goshtoso-chart-content]"),
        sameFigure: element.__wrapperMatrixFigure === element.querySelector("[data-goshtoso-chart-content] figure"),
      })), {
        mode: "hidden", hidden: true, inert: true, ariaHidden: "true",
        sameContent: true, sameFigure: true,
      }, `${route} hidden state`);

      assert.equal(await wrapper.evaluate((element) => window.__goshtosoChartsControls.setWrapperMode(element, "enabled")), true, `${route} enabled transition`);
      await page.waitForTimeout(300);
      const enabled = await wrapper.evaluate((element, interactive) => {
        const host = interactive && element.querySelector("[_echarts_instance_]");
        return {
          mode: element.dataset.goshtosoChartWrapperMode,
          hidden: element.hidden,
          inert: element.hasAttribute("inert"),
          ariaHidden: element.hasAttribute("aria-hidden"),
          sameContent: element.__wrapperMatrixContent === element.querySelector("[data-goshtoso-chart-content]"),
          sameFigure: element.__wrapperMatrixFigure === element.querySelector("[data-goshtoso-chart-content] figure"),
          sameStaticSVG: interactive || element.__wrapperMatrixSVG === element.querySelector("[data-goshtoso-chart-content] svg"),
          sameInteractiveHost: !interactive || element.__wrapperMatrixHost === host,
          sameInteractiveInstance: !interactive || element.__wrapperMatrixInstance === window.echarts.getInstanceByDom(host),
          chartWidth: interactive ? window.echarts.getInstanceByDom(host).getWidth() : element.querySelector("[data-goshtoso-chart-content] svg").getBoundingClientRect().width,
          changes: element.__wrapperMatrixChanges,
        };
      }, route.includes("/interactive/"));
      assert.deepEqual({
        mode: enabled.mode, hidden: enabled.hidden, inert: enabled.inert, ariaHidden: enabled.ariaHidden,
        sameContent: enabled.sameContent, sameFigure: enabled.sameFigure, sameStaticSVG: enabled.sameStaticSVG,
        sameInteractiveHost: enabled.sameInteractiveHost, sameInteractiveInstance: enabled.sameInteractiveInstance,
      }, {
        mode: "enabled", hidden: false, inert: false, ariaHidden: false,
        sameContent: true, sameFigure: true, sameStaticSVG: true,
        sameInteractiveHost: true, sameInteractiveInstance: true,
      }, `${route} enabled state`);
      assert.ok(enabled.chartWidth > 0, `${route} zero-width chart after reveal`);
      assert.deepEqual(enabled.changes, [
        ["enabled", "disabled"], ["disabled", "hidden"], ["hidden", "enabled"],
      ], `${route} stable change events`);
      assert.equal(await wrapper.evaluate((element) => window.__goshtosoChartsControls.setWrapperMode(element, "omitted")), false, `${route} omitted must require server replacement`);
      assert.equal(await wrapper.getAttribute("data-goshtoso-chart-wrapper-mode"), "enabled");
      assert.deepEqual(errors, [], `${route} browser errors`);
    } finally {
      await page.close();
    }
  }
});

test("actionless wrapper still accepts plain-JS lifecycle changes", async () => {
  const page = await browser.newPage({ viewport: { width: 390, height: 600 } });
  const errors = [];
  page.on("pageerror", (error) => errors.push(String(error)));
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  try {
    const rendered = execFileSync("go", ["run", "./browser/fixtures/actionless-wrapper"], {
      cwd: path.resolve(__dirname, ".."),
      encoding: "utf8",
    });
    await page.setContent(`<!doctype html><html><head><base href="${baseURL}/"></head><body>${rendered}</body></html>`, { waitUntil: "load" });
    await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]");
    assert.equal(await wrapper.locator('script[src="/charts/assets/js/controls/4/controls.js"]').count(), 1);
    assert.equal(await wrapper.locator("[data-goshtoso-chart-actions-fieldset]").count(), 0);
    assert.equal(await wrapper.evaluate((element) => window.__goshtosoChartsControls.setWrapperMode(element, "hidden")), true);
    assert.deepEqual(await wrapper.evaluate((element) => ({
      mode: element.dataset.goshtosoChartWrapperMode,
      hidden: element.hidden,
      inert: element.hasAttribute("inert"),
      ariaHidden: element.getAttribute("aria-hidden"),
    })), { mode: "hidden", hidden: true, inert: true, ariaHidden: "true" });
    await wrapper.evaluate((element) => element.dispatchEvent(new CustomEvent("goshtoso-charts:set-wrapper-mode", {
      bubbles: true,
      detail: { mode: "enabled" },
    })));
    assert.deepEqual(await wrapper.evaluate((element) => ({
      mode: element.dataset.goshtosoChartWrapperMode,
      hidden: element.hidden,
      inert: element.hasAttribute("inert"),
      ariaHidden: element.hasAttribute("aria-hidden"),
    })), { mode: "enabled", hidden: false, inert: false, ariaHidden: false });
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});
