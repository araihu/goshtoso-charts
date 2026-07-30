const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");

const screenshotDirectory = process.env.GOSHTOSO_SCREENSHOT_DIR;
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
      if ((await fetch(`${baseURL}/docs/chart-controls`)).ok) return;
    } catch {
      // Test-owned server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Chart-control example server did not start at ${baseURL}`);
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

async function setRange(page, selector, value) {
  const control = page.locator(selector);
  await control.focus();
  await control.evaluate((element, next) => {
    element.value = next;
    element.dispatchEvent(new Event("change", { bubbles: true }));
  }, String(value));
}

async function setColor(page, selector, value) {
  const control = page.locator(selector);
  await control.focus();
  await control.evaluate((element, next) => {
    element.value = next;
    element.dispatchEvent(new Event("change", { bubbles: true }));
  }, value);
}

async function afterHTMXSettle(page, action) {
	await page.waitForFunction(() => Boolean(globalThis.htmx));
  const previous = await page.evaluate(() => {
    if (!globalThis.__chartControlSettleListener) {
      globalThis.__chartControlSettleListener = true;
      globalThis.__chartControlSettleCount = 0;
      document.addEventListener("htmx:afterSettle", () => { globalThis.__chartControlSettleCount += 1; });
    }
    return globalThis.__chartControlSettleCount;
  });
  await action();
  await page.waitForFunction((count) => globalThis.__chartControlSettleCount > count, previous);
}

async function waitForInteractiveChart(page) {
  await page.waitForFunction(() => {
    const host = document.querySelector("#interactive-chart-control-example [_echarts_instance_]");
    return Boolean(host && globalThis.echarts?.getInstanceByDom(host));
  });
}

test("static form swaps typed options and every wrapper lifecycle state", async () => {
  const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
  try {
    await page.goto(`${baseURL}/docs/chart-controls`);
    const example = page.locator("#static-chart-control-example");
    await example.waitFor();
    assert.equal(await example.getAttribute("data-chart-control-stroke"), "3");
    assert.equal(await example.getAttribute("data-chart-control-area"), "on");

    await afterHTMXSettle(page, () => setRange(page, "#static-stroke", 7));
    await page.waitForFunction(() => document.querySelector("#static-chart-control-example")?.dataset.chartControlStroke === "7");
    assert.equal(await page.evaluate(() => document.activeElement?.id), "static-stroke");
    assert.match(await page.locator('[data-chart-control-state="static"]').innerText(), /7 px stroke/);

    await afterHTMXSettle(page, () => page.locator("#static-area").uncheck());
    await page.waitForFunction(() => document.querySelector("#static-chart-control-example")?.dataset.chartControlArea === "off");
    assert.equal(await page.evaluate(() => document.activeElement?.id), "static-area");

    await page.locator("#static-mode").focus();
    await afterHTMXSettle(page, () => page.locator("#static-mode").selectOption("disabled"));
    await page.waitForFunction(() => document.querySelector("#static-chart-control-example")?.dataset.chartControlMode === "disabled");
    assert.equal(await page.evaluate(() => document.activeElement?.id), "static-mode");
    assert.equal(await page.locator("[data-chart-control-preview='static'] [data-goshtoso-chart-actions-fieldset]").isDisabled(), true);

    await page.locator("#static-mode").focus();
    await afterHTMXSettle(page, () => page.locator("#static-mode").selectOption("hidden"));
    await page.waitForFunction(() => document.querySelector("#static-chart-control-example")?.dataset.chartControlMode === "hidden");
    const hidden = page.locator("[data-chart-control-preview='static'] [data-goshtoso-chart-wrapper]");
    assert.equal(await hidden.getAttribute("hidden"), "");
    assert.equal(await hidden.getAttribute("inert"), "");
    assert.equal(await hidden.getAttribute("aria-hidden"), "true");

    await page.locator("#static-mode").focus();
    await afterHTMXSettle(page, () => page.locator("#static-mode").selectOption("omitted"));
    await page.waitForFunction(() => document.querySelector("#static-chart-control-example")?.dataset.chartControlMode === "omitted");
    assert.equal(await page.locator("[data-chart-control-preview='static'] [data-goshtoso-chart-wrapper]").count(), 0);
    assert.equal(await page.locator("[data-chart-control-preview='static'] svg").count() > 0, true);

    await page.locator("#static-mode").focus();
    await afterHTMXSettle(page, () => page.locator("#static-mode").selectOption("enabled"));
    await page.waitForFunction(() => document.querySelector("#static-chart-control-example")?.dataset.chartControlMode === "enabled");
    assert.equal(await page.locator("[data-chart-control-preview='static'] [data-goshtoso-chart-wrapper]").count(), 1);
  } finally {
    await page.close();
  }
});

test("interactive form replaces one initialized chart with altered data and options", async () => {
  const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
  try {
    await page.goto(`${baseURL}/docs/chart-controls`);
    await waitForInteractiveChart(page);
    const initialValue = await page.evaluate(() => {
      const host = document.querySelector("#interactive-chart-control-example [_echarts_instance_]");
      return globalThis.echarts.getInstanceByDom(host).getOption().series[0].data[0].value;
    });

    await page.locator("#interactive-orientation").focus();
    await afterHTMXSettle(page, () => page.locator("#interactive-orientation").selectOption("horizontal"));
    await page.waitForFunction(() => document.querySelector("#interactive-chart-control-example")?.dataset.chartControlOrientation === "horizontal");
    await waitForInteractiveChart(page);
    assert.equal(await page.evaluate(() => document.activeElement?.id), "interactive-orientation");
    const orientation = await page.evaluate(() => {
      const host = document.querySelector("#interactive-chart-control-example [_echarts_instance_]");
      const options = globalThis.echarts.getInstanceByDom(host).getOption();
      return { x: options.xAxis[0].type, y: options.yAxis[0].type };
    });
    assert.deepEqual(orientation, { x: "value", y: "category" });

    await afterHTMXSettle(page, () => setRange(page, "#interactive-scale", 150));
    await page.waitForFunction(() => document.querySelector("#interactive-chart-control-example")?.dataset.chartControlScale === "150");
    await waitForInteractiveChart(page);
    const scaledValue = await page.evaluate(() => {
      const host = document.querySelector("#interactive-chart-control-example [_echarts_instance_]");
      return globalThis.echarts.getInstanceByDom(host).getOption().series[0].data[0].value;
    });
    assert.equal(scaledValue, initialValue * 1.5);
    assert.equal(await page.evaluate(() => document.activeElement?.id), "interactive-scale");

    await afterHTMXSettle(page, () => page.locator("#interactive-labels").uncheck());
    await page.waitForFunction(() => document.querySelector("#interactive-chart-control-example")?.dataset.chartControlLabels === "off");
    await waitForInteractiveChart(page);
    const state = await page.evaluate(() => {
      const root = document.querySelector("#interactive-chart-control-example");
      const host = root.querySelector("[_echarts_instance_]");
      const options = globalThis.echarts.getInstanceByDom(host).getOption();
      return {
        labels: options.series[0].label.show,
        hosts: root.querySelectorAll("[_echarts_instance_]").length,
        canvases: root.querySelectorAll("canvas").length,
        exactRows: root.querySelectorAll("[data-bar-exact-values] tbody tr").length,
      };
    });
    assert.deepEqual(state, { labels: false, hosts: 1, canvases: 1, exactRows: 14 });
    assert.equal(await page.evaluate(() => document.activeElement?.id), "interactive-labels");
  } finally {
    await page.close();
  }
});

test("one custom palette updates all four chart semantics through HTMX", async () => {
  const page = await browser.newPage({ viewport: { width: 1440, height: 1100 } });
  try {
    await page.goto(`${baseURL}/docs/chart-controls`);
    await page.locator("#chart-palette").focus();
    await afterHTMXSettle(page, () => page.locator("#chart-palette").selectOption("custom"));
    await page.waitForFunction(() => document.querySelector("#palette-chart-control-example")?.dataset.chartPalette === "custom");
    assert.equal(await page.evaluate(() => document.activeElement?.id), "chart-palette");

    const colors = ["#123456", "#236789", "#b45309", "#be123c"];
    for (let index = 0; index < colors.length; index += 1) {
      await afterHTMXSettle(page, () => setColor(page, `#palette-color-${index + 1}`, colors[index]));
      await page.waitForFunction(({ id, value }) => document.querySelector(id)?.value === value, { id: `#palette-color-${index + 1}`, value: colors[index] });
      assert.equal(await page.evaluate(() => document.activeElement?.id), `palette-color-${index + 1}`);
    }

    const mapping = await page.evaluate((expected) => {
      const markup = (kind) => document.querySelector(`[data-palette-chart='${kind}']`).innerHTML.toLowerCase();
      const line = markup("line"), bar = markup("bar"), pie = markup("pie"), heatmap = markup("heatmap");
      return {
        line: [line.includes(expected[0])],
        bar: expected.slice(0, 2).map((color) => bar.includes(color)),
        pie: expected.map((color) => pie.includes(color)),
        heatmap: expected.map((color) => heatmap.includes(color)),
        wrappers: document.querySelectorAll("#palette-chart-control-example [data-goshtoso-chart-wrapper]").length,
        exportGroups: document.querySelectorAll("#palette-chart-control-example [data-goshtoso-chart-actions]").length,
        state: document.querySelector('[data-chart-control-state="palette-grid"]').textContent.trim(),
      };
    }, colors);
    assert.deepEqual(mapping.line, [true]);
    assert.deepEqual(mapping.bar, [true, true]);
    assert.deepEqual(mapping.pie, [true, true, true, true]);
    assert.deepEqual(mapping.heatmap, [true, true, true, true]);
    assert.equal(mapping.wrappers, 4);
    assert.equal(mapping.exportGroups, 4);
    for (const color of colors) assert.equal(mapping.state.includes(color), true, color);
  } finally {
    await page.close();
  }
});

for (const width of [390, 768, 1440]) {
  for (const mode of ["light", "dark"]) {
    test(`forms remain usable after swaps at ${width}px in ${mode} mode`, async () => {
      const page = await browser.newPage({ viewport: { width, height: 900 }, colorScheme: mode });
      const browserIssues = [];
      page.on("console", (message) => {
        if (["warning", "error"].includes(message.type())) browserIssues.push(`${message.type()}: ${message.text()}`);
      });
      page.on("pageerror", (error) => browserIssues.push(`pageerror: ${error.message}`));
      try {
        await page.goto(`${baseURL}/docs/chart-controls`);
        await page.evaluate((dark) => document.documentElement.classList.toggle("dark", dark), mode === "dark");
        await afterHTMXSettle(page, () => setRange(page, "#static-stroke", 6));
        await page.waitForFunction(() => document.querySelector("#static-chart-control-example")?.dataset.chartControlStroke === "6");
        await afterHTMXSettle(page, () => setRange(page, "#interactive-scale", 110));
        await page.waitForFunction(() => document.querySelector("#interactive-chart-control-example")?.dataset.chartControlScale === "110");
        await waitForInteractiveChart(page);
        await afterHTMXSettle(page, () => page.locator("#chart-palette").selectOption("custom"));
        await page.waitForFunction(() => document.querySelector("#palette-chart-control-example")?.dataset.chartPalette === "custom");
        await afterHTMXSettle(page, () => setColor(page, "#palette-color-1", "#123456"));
        const geometry = await page.evaluate(() => {
          const cards = [...document.querySelectorAll("[data-palette-chart]")].map((card) => card.getBoundingClientRect());
          const formButtons = [
            [document.querySelector("#static-chart-control-form"), document.querySelector("#static-apply")],
            [document.querySelector("#interactive-chart-control-form"), document.querySelector("#interactive-apply")],
          ].map(([form, button]) => {
            const formRect = form.getBoundingClientRect();
            const buttonRect = button.getBoundingClientRect();
            return { formLeft: formRect.left, formRight: formRect.right, buttonLeft: buttonRect.left, buttonRight: buttonRect.right };
          });
          return {
            clientWidth: document.documentElement.clientWidth,
            scrollWidth: document.documentElement.scrollWidth,
            duplicateIDs: [...document.querySelectorAll("[id]")]
              .map((element) => element.id)
              .filter((id, index, ids) => ids.indexOf(id) !== index),
            cards: cards.map(({ x, y, width: cardWidth }) => ({ x: Math.round(x), y: Math.round(y), width: Math.round(cardWidth) })),
            formButtons,
          };
        });
        assert.equal(geometry.scrollWidth, geometry.clientWidth, JSON.stringify(geometry));
        assert.deepEqual(geometry.duplicateIDs, []);
        for (const bounds of geometry.formButtons) {
          assert.ok(bounds.buttonLeft >= bounds.formLeft, JSON.stringify(bounds));
          assert.ok(bounds.buttonRight <= bounds.formRight, JSON.stringify(bounds));
        }
        assert.equal(geometry.cards.length, 4);
        if (width === 1440) {
          assert.equal(geometry.cards[0].y, geometry.cards[1].y);
          assert.ok(geometry.cards[1].x > geometry.cards[0].x);
          assert.equal(geometry.cards[2].y, geometry.cards[3].y);
        } else {
          assert.deepEqual(new Set(geometry.cards.map((card) => card.x)).size, 1);
          assert.ok(geometry.cards.every((card, index) => index === 0 || card.y > geometry.cards[index - 1].y));
        }
        assert.deepEqual(browserIssues, []);
        if (screenshotDirectory) {
          await page.locator("#static-chart-control-form").evaluate((element) => element.scrollIntoView({ block: "start" }));
          await page.screenshot({ path: path.join(screenshotDirectory, `chart-control-static-form-${width}-${mode}.png`) });
          await page.locator("#interactive-chart-control-form").evaluate((element) => element.scrollIntoView({ block: "start" }));
          await page.screenshot({ path: path.join(screenshotDirectory, `chart-control-interactive-form-${width}-${mode}.png`) });
          await page.locator("#palette-chart-control-form").evaluate((element) => element.scrollIntoView({ block: "start" }));
          await page.screenshot({ path: path.join(screenshotDirectory, `chart-control-palette-grid-${width}-${mode}.png`) });
        }
      } finally {
        await page.close();
      }
    });
  }
}

test("static form has a no-JavaScript GET submission fallback", async () => {
  const context = await browser.newContext({ javaScriptEnabled: false, viewport: { width: 390, height: 900 } });
  const page = await context.newPage();
  try {
    await page.goto(`${baseURL}/docs/chart-controls`);
    await page.locator("#static-mode").selectOption("omitted");
    await page.locator("#static-stroke").fill("8");
    await page.locator("#static-area").uncheck();
    await Promise.all([
      page.waitForNavigation(),
      page.locator("#static-apply").click(),
    ]);
    assert.match(page.url(), /static_present=1/);
    assert.match(page.url(), /static_mode=omitted/);
    assert.match(page.url(), /static_stroke=8/);
    const example = page.locator("#static-chart-control-example");
    assert.equal(await example.getAttribute("data-chart-control-mode"), "omitted");
    assert.equal(await example.getAttribute("data-chart-control-stroke"), "8");
    assert.equal(await example.getAttribute("data-chart-control-area"), "off");
    assert.equal(await page.locator("[data-chart-control-preview='static'] [data-goshtoso-chart-wrapper]").count(), 0);
    assert.equal(await page.locator("[data-chart-control-preview='static'] svg").count() > 0, true);
  } finally {
    await context.close();
  }
});

test("palette form has a no-JavaScript custom-color GET fallback", async () => {
  const context = await browser.newContext({ javaScriptEnabled: false, viewport: { width: 390, height: 900 } });
  const page = await context.newPage();
  try {
    await page.goto(`${baseURL}/docs/chart-controls`);
    await page.locator("#chart-palette").selectOption("custom");
    // Disabled inputs are intentionally absent until the custom server state is selected.
    await Promise.all([page.waitForNavigation(), page.locator("#palette-apply").click()]);
    assert.match(page.url(), /chart_palette=custom/);
    for (let index = 0; index < 4; index += 1) {
      await page.locator(`#palette-color-${index + 1}`).fill(["#123456", "#236789", "#b45309", "#be123c"][index]);
    }
    await Promise.all([page.waitForNavigation(), page.locator("#palette-apply").click()]);
    const example = page.locator("#palette-chart-control-example");
    assert.equal(await example.getAttribute("data-chart-palette"), "custom");
    const body = (await example.innerHTML()).toLowerCase();
    for (const color of ["#123456", "#236789", "#b45309", "#be123c"]) assert.equal(body.includes(color), true, color);
    assert.equal(await example.locator("[data-goshtoso-chart-wrapper]").count(), 4);
  } finally {
    await context.close();
  }
});

test("Goshtoso CodeBlock copy action exposes complete matching source", async () => {
  const context = await browser.newContext({ permissions: ["clipboard-read", "clipboard-write"] });
  const page = await context.newPage();
  try {
    await page.goto(`${baseURL}/docs/chart-controls`);
    await page.getByRole("button", { name: "Copy Static form and chart · templ code", exact: true }).click();
    await page.getByText("Copied!", { exact: true }).first().waitFor();
    const copied = await page.evaluate(() => navigator.clipboard.readText());
    for (const marker of ["StaticChartControlsHandler", "templ StaticChartControls", "static_mode", "static_stroke", "static_area", "return line.Config{", "@line.Line(staticConfig(state))"]) {
      assert.equal(copied.includes(marker), true, marker);
    }
    await page.getByRole("button", { name: "Copy Palette form and four charts · templ code", exact: true }).click();
    await page.getByText("Copied!", { exact: true }).last().waitFor();
    const paletteCopied = await page.evaluate(() => navigator.clipboard.readText());
    for (const marker of ["PaletteChartsHandler", "templ PaletteCharts", "chart_palette", "palette_color_%d", "@line.Line(lineConfig(state))", "@bar.Bar(barConfig(state))", "@pie.Pie(pieConfig(state))", "@heatmap.HeatMap(heatMapConfig(state))"]) {
      assert.equal(paletteCopied.includes(marker), true, marker);
    }
  } finally {
    await context.close();
  }
});
