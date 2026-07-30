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

const builtInThemes = [
  { value: "araihu", label: "Arai Hû" },
  { value: "goshtoso", label: "Goshtoso" },
  { value: "arctic", label: "Arctic" },
  { value: "high-contrast", label: "High Contrast" },
  { value: "minimal", label: "Minimal" },
  { value: "modern", label: "Modern" },
  { value: "neo-brutalism", label: "Neo Brutalism" },
  { value: "halloween", label: "Halloween" },
  { value: "zombie", label: "Zombie" },
  { value: "pastel", label: "Pastel" },
  { value: "90s", label: "90s" },
  { value: "christmas", label: "Christmas" },
  { value: "prototype", label: "Prototype" },
  { value: "news", label: "News" },
  { value: "industrial", label: "Industrial" },
  { value: "dracula", label: "Dracula" },
];

const chartTokens = [
  "--color-chart-surface",
  "--color-chart-surface-alt",
  "--color-chart-outline",
  "--color-chart-grid",
  "--color-chart-text",
  "--color-chart-text-strong",
  "--color-chart-text-muted",
  ...Array.from({ length: 8 }, (_, index) => `--color-chart-series-${index + 1}`),
  "--color-chart-scale-low",
  "--color-chart-scale-mid",
  "--color-chart-scale-high",
  "--color-chart-increasing",
  "--color-chart-decreasing",
  "--color-chart-bollinger-upper",
  "--color-chart-bollinger-middle",
  "--color-chart-bollinger-lower",
];

const visualStates = new Set([
  "araihu:light",
  "high-contrast:dark",
  "pastel:light",
  "dracula:dark",
]);

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

async function selectTheme(frame, theme) {
  await frame.getByRole("combobox", { name: "Theme" }).click();
  await frame.getByRole("option", { name: theme.label, exact: true }).click();
  await frame.waitForFunction((value) => document.documentElement.dataset.theme === value, theme.value);
}

async function setColorScheme(frame, scheme) {
  await frame.evaluate((dark) => document.documentElement.classList.toggle("dark", dark), scheme === "dark");
}

async function themeState(frame, theme, scheme) {
  return frame.evaluate(({ expectedTheme, expectedScheme, tokens }) => {
    const root = document.documentElement;
    const figures = [...document.querySelectorAll(".goshtoso-charts-palette")];
    const cards = [...document.querySelectorAll("[data-theme-playground-chart]")];
    const interactiveHosts = [...document.querySelectorAll('[data-theme-playground-chart="interactive"] [_echarts_instance_]')];
    const canvas = document.createElement("canvas");
    canvas.width = 1;
    canvas.height = 1;
    const context = canvas.getContext("2d", { willReadFrequently: true });
    const rgba = (color) => {
      context.clearRect(0, 0, 1, 1);
      context.fillStyle = "rgba(1, 2, 3, 0.25)";
      context.fillStyle = color;
      context.fillRect(0, 0, 1, 1);
      return [...context.getImageData(0, 0, 1, 1).data];
    };
    const computedPaint = (node, property) => rgba(getComputedStyle(node)[property]);
    const tokenState = (figure) => Object.fromEntries(tokens.map((token) => {
      const raw = getComputedStyle(figure).getPropertyValue(token).trim();
      const probe = document.createElement("span");
      probe.hidden = true;
      probe.style.color = `var(${token})`;
      figure.appendChild(probe);
      const color = getComputedStyle(probe).color;
      probe.remove();
      return [token, { raw, color, rgba: rgba(color) }];
    }));
    const resolved = figures.map(tokenState);
    const first = resolved[0];
    const vector = figures.slice(0, 2).map((figure) => {
      const seriesNode = figure.querySelector('[style*="var(--color-chart-series-1)"]');
      const seriesProperty = getComputedStyle(seriesNode).fill === "none" ? "stroke" : "fill";
      return {
        series: computedPaint(seriesNode, seriesProperty),
        surface: computedPaint(figure.querySelector('[style*="var(--color-chart-surface)"]'), "fill"),
        text: computedPaint(figure.querySelector('text[style*="var(--color-chart-text)"]'), "fill"),
      };
    });
    const interactive = interactiveHosts.map((host) => {
      const option = window.echarts.getInstanceByDom(host).getOption();
      const scalar = (value) => Array.isArray(value) ? value[0] : value;
      return {
        series: rgba(scalar(option.color)),
        surface: rgba(scalar(option.backgroundColor)),
        text: rgba(scalar(option.textStyle).color),
      };
    });
    const cardPaints = cards.map((card) => ({
      background: computedPaint(card, "backgroundColor"),
      heading: computedPaint(card.querySelector("h2"), "color"),
    }));
    const geometry = cards.map((card) => {
      const visual = card.querySelector("svg, [_echarts_instance_]");
      const cardBounds = card.getBoundingClientRect();
      const visualBounds = visual.getBoundingClientRect();
      return {
        cardWidth: cardBounds.width,
        visualWidth: visualBounds.width,
        scrollWidth: card.scrollWidth,
        clientWidth: card.clientWidth,
      };
    });
    return {
      theme: root.dataset.theme,
      scheme: root.classList.contains("dark") ? "dark" : "light",
      expectedTheme,
      expectedScheme,
      ids: interactiveHosts.map((host) => host.getAttribute("_echarts_instance_")),
      resolved,
      vector,
      interactive,
      cardPaints,
      geometry,
      pageWidth: { scroll: root.scrollWidth, client: root.clientWidth },
    };
  }, { expectedTheme: theme, expectedScheme: scheme, tokens: chartTokens });
}

function contrast(left, right) {
  const luminance = (rgba) => {
    const channels = rgba.slice(0, 3).map((channel) => {
      const value = channel / 255;
      return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
    });
    return 0.2126 * channels[0] + 0.7152 * channels[1] + 0.0722 * channels[2];
  };
  const brighter = Math.max(luminance(left), luminance(right));
  const darker = Math.min(luminance(left), luminance(right));
  return (brighter + 0.05) / (darker + 0.05);
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

test("all built-in themes resolve coherent opaque chart tokens without replacing charts", async () => {
  const context = await browser.newContext({ viewport: { width: 1440, height: 1100 }, reducedMotion: "reduce" });
  await context.addInitScript(() => localStorage.setItem("theme", "modern"));
  const page = await context.newPage();
  const browserIssues = [];
  page.on("console", (message) => {
    if (["warning", "error"].includes(message.type())) browserIssues.push(`${message.type()}: ${message.text()}`);
  });
  page.on("pageerror", (error) => browserIssues.push(`pageerror: ${error.message}`));

  try {
    await page.goto(`${baseURL}/docs/theme-playground`);
    const child = page.frames().find((frame) => frame.url().endsWith("/docs/theme-playground/frame"));
    assert.ok(child, "same-origin theme frame missing");
    await child.waitForFunction(() => document.querySelectorAll("[_echarts_instance_]").length === 2);
    await child.getByRole("combobox", { name: "Theme" }).click();
    await child.getByRole("option").first().waitFor();
    const optionLabels = await child.getByRole("option").allTextContents();
    assert.deepEqual(optionLabels.map((label) => label.trim()), builtInThemes.map((theme) => theme.label));
    await child.getByRole("option", { name: "Arai Hû", exact: true }).click();

    const parentBefore = await page.evaluate(() => ({
      theme: document.documentElement.dataset.theme,
      className: document.documentElement.className,
      storedTheme: localStorage.getItem("theme"),
    }));
    const initialIDs = await child.locator("[_echarts_instance_]").evaluateAll((nodes) => nodes.map((node) => node.getAttribute("_echarts_instance_")));
    const signatures = { light: new Map(), dark: new Map() };
    const screenshots = [new Set(), new Set(), new Set(), new Set()];

    for (const theme of builtInThemes) {
      await setColorScheme(child, "light");
      await selectTheme(child, theme);
      for (const scheme of ["light", "dark"]) {
        await setColorScheme(child, scheme);
        await child.waitForFunction(({ expectedTheme, dark }) => {
          if (document.documentElement.dataset.theme !== expectedTheme || document.documentElement.classList.contains("dark") !== dark) return false;
          const figure = document.querySelector(".goshtoso-charts-palette");
          const host = document.querySelector('[_echarts_instance_]');
          if (!figure || !host) return false;
          const chart = window.echarts.getInstanceByDom(host);
          const option = chart && chart.getOption();
          if (!option) return false;
          const probe = document.createElement("span");
          probe.hidden = true;
          probe.style.color = "var(--color-chart-surface)";
          figure.appendChild(probe);
          const expectedSurface = getComputedStyle(probe).color;
          probe.remove();
          const actualSurface = Array.isArray(option.backgroundColor) ? option.backgroundColor[0] : option.backgroundColor;
          const colorCanvas = document.createElement("canvas");
          colorCanvas.width = 1;
          colorCanvas.height = 1;
          const colorContext = colorCanvas.getContext("2d", { willReadFrequently: true });
          const pixel = (color) => {
            colorContext.clearRect(0, 0, 1, 1);
            colorContext.fillStyle = color;
            colorContext.fillRect(0, 0, 1, 1);
            return [...colorContext.getImageData(0, 0, 1, 1).data].join(",");
          };
          return pixel(expectedSurface) === pixel(actualSurface);
        }, { expectedTheme: theme.value, dark: scheme === "dark" });

        const state = await themeState(child, theme.value, scheme);
        assert.equal(state.theme, theme.value);
        assert.equal(state.scheme, scheme);
        assert.deepEqual(state.ids, initialIDs, `${theme.value}:${scheme} replaced an ECharts instance`);
        assert.equal(state.resolved.length, 4);
        for (const [figureIndex, tokenSet] of state.resolved.entries()) {
          for (const token of chartTokens) {
            const value = tokenSet[token];
            assert.ok(value.raw, `${theme.value}:${scheme} figure ${figureIndex + 1} missing ${token}`);
            assert.ok(value.color, `${theme.value}:${scheme} figure ${figureIndex + 1} invalid ${token}: ${value.raw}`);
            assert.equal(value.rgba[3], 255, `${theme.value}:${scheme} figure ${figureIndex + 1} transparent ${token}: ${value.color}`);
            assert.deepEqual(value.rgba, state.resolved[0][token].rgba, `${theme.value}:${scheme} ${token} differs across charts`);
          }
        }

        const tokens = state.resolved[0];
        const paletteSignature = [
          ...Array.from({ length: 8 }, (_, index) => tokens[`--color-chart-series-${index + 1}`].rgba.join(",")),
          tokens["--color-chart-scale-low"].rgba.join(","),
          tokens["--color-chart-scale-mid"].rgba.join(","),
          tokens["--color-chart-scale-high"].rgba.join(","),
        ].join("|");
        assert.equal(signatures[scheme].has(paletteSignature), false, `${theme.value}:${scheme} reused ${signatures[scheme].get(paletteSignature)} fallback palette`);
        signatures[scheme].set(paletteSignature, theme.value);
        assert.ok(contrast(tokens["--color-chart-text"].rgba, tokens["--color-chart-surface"].rgba) >= 4.5, `${theme.value}:${scheme} chart text contrast below 4.5`);
        assert.ok(contrast(tokens["--color-chart-text-strong"].rgba, tokens["--color-chart-surface"].rgba) >= 4.5, `${theme.value}:${scheme} strong chart text contrast below 4.5`);
        assert.ok(contrast(tokens["--color-chart-text-muted"].rgba, tokens["--color-chart-surface"].rgba) >= 3, `${theme.value}:${scheme} muted chart text contrast below 3`);
        assert.ok(contrast(tokens["--color-chart-series-1"].rgba, tokens["--color-chart-surface"].rgba) >= 2, `${theme.value}:${scheme} primary series contrast below 2`);

        for (const sample of [...state.vector, ...state.interactive]) {
          assert.deepEqual(sample.series, tokens["--color-chart-series-1"].rgba, `${theme.value}:${scheme} chart series did not use theme token`);
          assert.deepEqual(sample.surface, tokens["--color-chart-surface"].rgba, `${theme.value}:${scheme} chart surface did not use theme token`);
          assert.deepEqual(sample.text, tokens["--color-chart-text"].rgba, `${theme.value}:${scheme} chart text did not use theme token`);
        }
        for (const paint of state.cardPaints) {
          assert.equal(paint.background[3], 255, `${theme.value}:${scheme} transparent card background`);
          assert.ok(contrast(paint.heading, paint.background) >= 4.5, `${theme.value}:${scheme} card heading contrast below 4.5`);
        }
        assert.equal(state.pageWidth.scroll, state.pageWidth.client, `${theme.value}:${scheme} frame overflow`);
        for (const geometry of state.geometry) {
          assert.equal(geometry.scrollWidth, geometry.clientWidth, `${theme.value}:${scheme} card overflow`);
          assert.ok(geometry.visualWidth <= geometry.cardWidth, `${theme.value}:${scheme} chart exceeds card`);
        }

        if (visualStates.has(`${theme.value}:${scheme}`)) {
          const cards = child.locator("[data-theme-playground-chart]");
          for (let index = 0; index < 4; index += 1) {
            screenshots[index].add(digest(await cards.nth(index).screenshot({ animations: "disabled" })));
          }
        }
      }
    }

    assert.equal(signatures.light.size, builtInThemes.length);
    assert.equal(signatures.dark.size, builtInThemes.length);
    screenshots.forEach((values, index) => assert.equal(values.size, visualStates.size, `representative chart ${index + 1} visuals did not all differ`));
    assert.deepEqual(await page.evaluate(() => ({
      theme: document.documentElement.dataset.theme,
      className: document.documentElement.className,
      storedTheme: localStorage.getItem("theme"),
    })), parentBefore);
    assert.deepEqual(browserIssues, []);
  } finally {
    await context.close();
  }
});
