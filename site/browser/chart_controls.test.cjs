const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

const testPort = process.env.TEST_PORT || String(20000 + Math.floor(Math.random() * 30000));
const baseURL = process.env.BASE_URL || `http://127.0.0.1:${testPort}`;
const screenshotDirectory = process.env.SCREENSHOT_DIR;
let browser;
let server;

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

async function pageAt(route, viewport = { width: 1280, height: 900 }) {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  await page.addInitScript(() => {
    const createObjectURL = URL.createObjectURL.bind(URL);
    globalThis.__chartBlobTypes = [];
    URL.createObjectURL = (blob) => {
      globalThis.__chartBlobTypes.push(blob.type);
      return createObjectURL(blob);
    };
  });
	await page.goto(`${baseURL}${route}`);
	await page.locator("[data-goshtoso-chart-wrapper]").first().waitFor();
	await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
	return page;
}

async function download(page, label) {
  await page.evaluate(() => { globalThis.__chartBlobTypes.length = 0; });
  const pending = page.waitForEvent("download");
  const direct = page.getByRole("button", { name: label });
  if (await direct.count()) {
    await direct.first().click();
	} else {
    const match = label.match(/^Download (.+) as (SVG|PNG)$/);
    assert.ok(match, `no direct export button and unsupported label ${label}`);
		const trigger = page.getByRole("button", { name: `Export ${match[1]}` }).first();
		const menu = trigger.locator("..").getByRole("menu");
		if (!(await menu.isVisible())) await trigger.click();
		await menu.getByRole("menuitem", { name: match[2], exact: true }).first().click();
  }
  const artifact = await pending;
	const artifactPath = await artifact.path();
	assert.ok(artifactPath);
	const openMenu = page.getByRole("menu").filter({ visible: true });
	if (await openMenu.count()) {
		if (await openMenu.isVisible()) await page.keyboard.press("Escape");
		await openMenu.waitFor({ state: "hidden" });
	}
	return {
    filename: artifact.suggestedFilename(),
    bytes: await fs.readFile(artifactPath),
    types: await page.evaluate(() => [...globalThis.__chartBlobTypes]),
  };
}

async function openExpand(wrapper) {
  const trigger = wrapper.locator('[id$="-stacked"] > button:visible, [data-action-group-primary] button:visible').first();
  const stacked = await trigger.evaluate((button) => Boolean(button.closest('[id$="-stacked"]')));
  await trigger.click();
  if (stacked) {
    const menuItem = wrapper.locator('[id$="-chart-expand-action"]').first();
    await menuItem.waitFor({ state: "visible" });
    await menuItem.click();
  }
  return trigger;
}

async function enterFullscreen(wrapper) {
  const trigger = wrapper.locator('[id$="-stacked"] > button:visible, [data-action-group-primary] button:visible').first();
  await trigger.click();
  const action = wrapper.locator('[id$="-fullscreen-action"]').first();
  await action.click();
  return { action, trigger };
}

test("responsive controls keep one primary Expand and flatten overflow at 320, 390, 768, and 1440", async () => {
  const page = await pageAt("/components/line", { width: 320, height: 800 });
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    for (const width of [320, 390, 768, 1440]) {
      await page.setViewportSize({ width, height: 900 });
      await page.waitForTimeout(100);
      const state = await wrapper.evaluate((element) => {
        const root = element.querySelector("[data-goshtoso-action-group]");
        const expand = Array.from(root.querySelectorAll('[id$="-stacked"] > button, [data-action-group-primary] button'))
          .find((button) => button.getBoundingClientRect().width > 0 && getComputedStyle(button).visibility !== "hidden");
        const secondaries = Array.from(root.querySelectorAll(":scope > [data-action-group-secondary]"));
        const overflow = root.querySelector(":scope > [data-action-group-overflow]");
        return {
          wrapperWidth: element.clientWidth,
          pageOverflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
          primaryVisible: Boolean(expand),
          primaryText: expand && expand.textContent.trim(),
          secondaryVisible: secondaries.filter((secondary) => secondary.getBoundingClientRect().width > 0).length,
          overflowVisible: Boolean(overflow && overflow.getBoundingClientRect().width),
          collapseCount: Array.from(element.querySelectorAll("button")).filter((button) => /collapse/i.test(button.textContent)).length,
        };
      });
      assert.equal(state.pageOverflow, 0, `${width}px page overflow`);
      assert.equal(state.primaryVisible, true, `${width}px primary Expand hidden`);
      assert.match(state.primaryText, /^Expand/);
      assert.equal(state.collapseCount, 0);
      const expectedSecondary = width === 320 ? 0 : width === 390 ? 1 : 2;
      assert.equal(state.secondaryVisible, expectedSecondary, `${width}px secondary count at ${state.wrapperWidth}px wrapper`);
      assert.equal(state.overflowVisible, width <= 390, `${width}px overflow visibility at ${state.wrapperWidth}px wrapper`);

      const primary = wrapper.locator('[id$="-stacked"] > button:visible, [data-action-group-primary] button:visible').first();
      await primary.focus();
      assert.equal(await primary.evaluate((button) => button === document.activeElement), true);
      if (width > 320) {
        await primary.press("Enter");
        const primaryMenu = wrapper.locator('[id$="-stacked"] [role=menu]');
        await primaryMenu.waitFor({ state: "visible" });
        assert.deepEqual(
          (await primaryMenu.getByRole("menuitem").allTextContents()).map((text) => text.trim()),
          ["Expand", "Fullscreen"],
        );
        await page.keyboard.press("Escape");
        await primaryMenu.waitFor({ state: "hidden" });
      }

      if (state.overflowVisible) {
        const overflow = wrapper.getByRole("button", { name: /More .* chart actions/ });
        await overflow.click();
        const menu = wrapper.locator("[data-action-group-overflow] [role=menu]");
        await menu.waitFor({ state: "visible" });
        const expected = width === 320
          ? ["Expand", "Expand", "Fullscreen", "Export", "SVG", "PNG"]
          : ["Export", "SVG", "PNG"];
        assert.deepEqual(
          (await menu.getByRole("menuitem").filter({ visible: true }).allTextContents()).map((text) => text.trim()),
          expected,
        );
        await page.mouse.click(1, 1);
        await menu.waitFor({ state: "hidden" });
      }
    }
  } finally {
    await page.close();
  }
});

test("Word Cloud, Line, Gauge, Pie, Map, and Tree controls stay usable across the full width and theme matrix", async () => {
  const routes = [
    ["/components/interactive/word-cloud", "word-cloud"],
    ["/components/line", "line"],
    ["/components/interactive/gauge", "gauge"],
    ["/components/pie", "pie"],
    ["/components/interactive/map", "map"],
    ["/components/interactive/tree", "tree"],
  ];
  const page = await browser.newPage({ viewport: { width: 320, height: 900 } });
  try {
    for (const width of [320, 390, 768, 1440]) {
      await page.setViewportSize({ width, height: 900 });
      for (const theme of ["goshtoso", "araihu"]) {
        for (const mode of ["light", "dark"]) {
          for (const [route, slug] of routes) {
            await page.goto(`${baseURL}${route}`);
            const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
            await wrapper.waitFor();
            await page.waitForFunction(() => Boolean(window.__goshtosoChartsControls));
            await page.waitForFunction(() => document.querySelector('[data-goshtoso-action-group][data-action-group-initialized="true"]'));
            await page.evaluate(({ selected, dark }) => {
              document.documentElement.dataset.theme = selected;
              document.documentElement.classList.toggle("dark", dark);
            }, { selected: theme, dark: mode === "dark" });
            await page.waitForTimeout(75);
            const state = await wrapper.evaluate((element) => {
              const group = element.querySelector("[data-goshtoso-action-group]");
              const visibleExpand = Array.from(group.querySelectorAll("button")).filter((button) =>
                button.textContent.trim() === "Expand" &&
                button.getBoundingClientRect().width > 0 &&
                getComputedStyle(button).visibility !== "hidden");
              return {
                initialized: group.dataset.actionGroupInitialized === "true",
                aria: group.getAttribute("aria-label"),
                collapse: Array.from(group.querySelectorAll("button")).filter((button) => /collapse/i.test(button.textContent)).length,
                expandCount: visibleExpand.length,
                small: getComputedStyle(visibleExpand[0]).fontSize === "12px",
                pageOverflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
              };
            });
            assert.equal(state.initialized, true, `${width}px ${theme} ${mode} ${slug} ActionGroup not initialized`);
            assert.match(state.aria, / chart controls$/, `${width}px ${theme} ${mode} ${slug} group label`);
            assert.equal(state.collapse, 0, `${width}px ${theme} ${mode} ${slug} leaked Collapse`);
            assert.equal(state.expandCount, 1, `${width}px ${theme} ${mode} ${slug} primary Expand count`);
            assert.equal(state.small, true, `${width}px ${theme} ${mode} ${slug} Expand is not small`);
            assert.equal(state.pageOverflow, 0, `${width}px ${theme} ${mode} ${slug} page overflow`);
            if (screenshotDirectory && (width === 320 || width === 1440)) {
              await page.screenshot({
                path: path.join(screenshotDirectory, `${slug}-${width}-${theme}-${mode}.png`),
                fullPage: true,
              });
            }
          }
        }
      }
    }
  } finally {
    await page.close();
  }
});

test("tree expansion state and instance survive fullscreen", async () => {
  const page = await pageAt("/components/interactive/tree");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    await wrapper.scrollIntoViewIfNeeded();
    const state = await wrapper.evaluate((element) => {
      const instance = window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]"));
      const data = instance.getModel().getSeriesByIndex(0).getData();
      let dataIndex = -1;
      for (let index = 0; index < data.count(); index += 1) {
        if (data.getName(index) === "Node3") dataIndex = index;
      }
      if (dataIndex < 0) throw new Error("Node3 data index not found");
      const node = data.tree.getNodeByDataIndex(dataIndex);
      const initiallyExpanded = node.isExpand;
      element.__treeInstance = instance;
      element.__treeNodeIndex = dataIndex;
      instance.dispatchAction({ type: "treeExpandAndCollapse", seriesIndex: 0, dataIndex });
      return { initiallyExpanded, expanded: node.isExpand };
    });
    assert.deepEqual(state, { initiallyExpanded: false, expanded: true });

    await enterFullscreen(wrapper);
    await page.waitForFunction(() => Boolean(document.fullscreenElement));
	await page.waitForTimeout(350);
	const fullscreenGeometry = await wrapper.evaluate((element) => {
	  const host = element.querySelector("[_echarts_instance_]");
	  const instance = window.echarts.getInstanceByDom(host);
	  const symbols = instance.getZr().storage.getDisplayList().filter((item) => item.type === "path" && item.z2 === 100 && item.shape && item.shape.symbolType === "circle");
	  const xs = symbols.map((item) => item.getComputedTransform()[4]);
	  return {
	    centerDelta: Math.abs((Math.min(...xs) + Math.max(...xs)) / 2 - instance.getWidth() / 2),
	    chartWidth: instance.getWidth(),
	    hostWidth: host.clientWidth,
	    canvasWidth: Math.round(host.querySelector("canvas").getBoundingClientRect().width),
	  };
	});
	assert.ok(fullscreenGeometry.centerDelta <= 2, `fullscreen tree center delta ${fullscreenGeometry.centerDelta}`);
	assert.deepEqual(fullscreenGeometry.chartWidth, fullscreenGeometry.hostWidth);
	assert.deepEqual(fullscreenGeometry.canvasWidth, fullscreenGeometry.hostWidth);
    await page.evaluate(() => document.exitFullscreen());
    await page.waitForFunction(() => !document.fullscreenElement);
	await page.waitForTimeout(350);

    const restored = await wrapper.evaluate((element) => {
      const instance = window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]"));
      const node = instance.getModel().getSeriesByIndex(0).getData().tree.getNodeByDataIndex(element.__treeNodeIndex);
	  const symbols = instance.getZr().storage.getDisplayList().filter((item) => item.type === "path" && item.z2 === 100 && item.shape && item.shape.symbolType === "circle");
	  const xs = symbols.map((item) => item.getComputedTransform()[4]);
	  return { sameInstance: instance === element.__treeInstance, expanded: node.isExpand, centerDelta: Math.abs((Math.min(...xs) + Math.max(...xs)) / 2 - instance.getWidth() / 2) };
	});
	assert.deepEqual({ sameInstance: restored.sameInstance, expanded: restored.expanded }, { sameInstance: true, expanded: true });
	assert.ok(restored.centerDelta <= 2, `restored tree center delta ${restored.centerDelta}`);
  } finally {
    await page.close();
  }
});

test("shared ResizeObserver converges a responsive interactive canvas after its consumer container shrinks without re-init", async () => {
  const page = await pageAt("/components/interactive/tree");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    const sizes = await wrapper.evaluate((element) => {
      const content = element.querySelector("[data-goshtoso-chart-content]");
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      const data = instance.getModel().getSeriesByIndex(0).getData();
      let nodeIndex = -1;
      for (let index = 0; index < data.count(); index += 1) if (data.getName(index) === "Node3") nodeIndex = index;
      instance.dispatchAction({ type: "treeExpandAndCollapse", seriesIndex: 0, dataIndex: nodeIndex });
      element.__flexInstance = instance;
      element.__flexNodeIndex = nodeIndex;
      const large = Math.min(847, content.clientWidth);
      const small = Math.max(320, large - 240);
      content.style.width = `${large}px`;
      return { large, small };
    });
    await page.waitForFunction((width) => document.querySelector("[_echarts_instance_]").clientWidth === width, sizes.large);
    await wrapper.evaluate((element, width) => {
      element.querySelector("[data-goshtoso-chart-content]").style.width = `${width}px`;
    }, sizes.small);
    await page.waitForFunction((width) => {
      const host = document.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth === width && instance.getWidth() === width && Math.round(host.querySelector("canvas").getBoundingClientRect().width) === width;
    }, sizes.small);
    assert.deepEqual(await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      const node = instance.getModel().getSeriesByIndex(0).getData().tree.getNodeByDataIndex(element.__flexNodeIndex);
      return { sameInstance: instance === element.__flexInstance, expanded: node.isExpand, host: host.clientWidth, chart: instance.getWidth(), canvas: Math.round(host.querySelector("canvas").getBoundingClientRect().width) };
    }), { sameInstance: true, expanded: true, host: sizes.small, chart: sizes.small, canvas: sizes.small });
  } finally {
    await page.close();
  }
});

test("Bar, Line, Pie, and Tree settle across narrow/dark and wide/light modal geometry without animation reset", async () => {
  const cases = [
    ["/components/interactive/bar", "Weekly deployments by environment"],
    ["/components/interactive/line", "Weekly latency trend"],
    ["/components/interactive/pie", "Incident states"],
    ["/components/interactive/tree", "Basic tree example"],
  ];
  const displays = [
    { viewport: { width: 390, height: 844 }, dark: true },
    { viewport: { width: 1440, height: 1000 }, dark: false },
  ];
  for (const [route, label] of cases) {
    for (const display of displays) {
      const page = await pageAt(route, display.viewport);
      try {
        await page.evaluate((dark) => document.documentElement.classList.toggle("dark", dark), display.dark);
        const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
        await wrapper.evaluate((element) => {
          const content = element.querySelector("[data-goshtoso-chart-content]");
          const host = element.querySelector("[_echarts_instance_]");
          const instance = window.echarts.getInstanceByDom(host);
          element.__responsiveContent = content;
          element.__responsiveHost = host;
          element.__responsiveInstance = instance;
          element.__responsiveResizeOptions = [];
          const resize = instance.resize;
          instance.resize = function (options) {
            element.__responsiveResizeOptions.push(options || null);
            return resize.apply(this, arguments);
          };
        });
        await page.waitForFunction(() => {
          const host = document.querySelector("[data-goshtoso-chart-wrapper] [_echarts_instance_]");
          const chart = window.echarts.getInstanceByDom(host);
          return chart.getWidth() === host.clientWidth && chart.getHeight() === host.clientHeight;
        });
        const initial = await wrapper.evaluate((element) => {
          const content = element.querySelector("[data-goshtoso-chart-content]");
          const host = element.querySelector("[_echarts_instance_]");
          const hostRect = host.getBoundingClientRect();
          const contentRect = content.getBoundingClientRect();
          return {
            contained: hostRect.width <= contentRect.width + 1,
            centerDelta: Math.abs((hostRect.left + hostRect.right - contentRect.left - contentRect.right) / 2),
          };
        });
        assert.equal(initial.contained, true, `${route} overflowed ${display.viewport.width}px preview`);
        assert.ok(initial.centerDelta <= 1, `${route} center delta ${initial.centerDelta}`);

        await openExpand(wrapper);
        const dialog = wrapper.getByRole("dialog", { name: label });
        await dialog.waitFor({ state: "visible" });
        await page.waitForTimeout(350);
        await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
          panel.style.width = "70vw";
          window.dispatchEvent(new Event("resize"));
        });
        await page.waitForTimeout(250);
        const expanded = await wrapper.evaluate((element) => {
          const host = element.querySelector("[_echarts_instance_]");
          const instance = window.echarts.getInstanceByDom(host);
          const canvas = host.querySelector("canvas");
          return {
            sameContent: element.__responsiveContent === element.querySelector("[data-goshtoso-chart-content]"),
            sameHost: element.__responsiveHost === host,
            sameInstance: element.__responsiveInstance === instance,
            chartWidth: instance.getWidth(),
            chartHeight: instance.getHeight(),
            hostWidth: host.clientWidth,
            hostHeight: host.clientHeight,
            canvasWidth: Math.round(canvas.getBoundingClientRect().width),
            canvasHeight: Math.round(canvas.getBoundingClientRect().height),
            instanceHosts: element.querySelectorAll("[_echarts_instance_]").length,
          };
        });
        assert.equal(expanded.sameContent, true);
        assert.equal(expanded.sameHost, true);
        assert.equal(expanded.sameInstance, true);
        assert.equal(expanded.instanceHosts, 1);
        assert.deepEqual(
          { width: expanded.chartWidth, height: expanded.chartHeight, canvasWidth: expanded.canvasWidth, canvasHeight: expanded.canvasHeight },
          { width: expanded.hostWidth, height: expanded.hostHeight, canvasWidth: expanded.hostWidth, canvasHeight: expanded.hostHeight },
        );

        await dialog.getByRole("button", { name: "close modal" }).click();
        await dialog.waitFor({ state: "hidden" });
        await page.waitForTimeout(350);
        const restored = await wrapper.evaluate((element) => ({
          sameContent: element.__responsiveContent === element.querySelector("[data-goshtoso-chart-content]"),
          sameHost: element.__responsiveHost === element.querySelector("[_echarts_instance_]"),
          sameInstance: element.__responsiveInstance === window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")),
          resizeOptions: element.__responsiveResizeOptions,
        }));
        assert.equal(restored.sameContent, true);
        assert.equal(restored.sameHost, true);
        assert.equal(restored.sameInstance, true);
        assert.ok(restored.resizeOptions.length >= 2, `${route} observed ${restored.resizeOptions.length} resize calls`);
        for (const options of restored.resizeOptions) {
          assert.equal(options?.animation?.duration, 0, `${route} resize restarted animation`);
        }
      } finally {
        await page.close();
      }
    }
  }
});

test("responsive registration is idempotent and unobserves host and container after removal", async () => {
  const page = await browser.newPage({ viewport: { width: 900, height: 800 } });
  try {
    await page.addInitScript(() => {
      const NativeResizeObserver = window.ResizeObserver;
      globalThis.__responsiveObserved = [];
      globalThis.__responsiveUnobserved = [];
      window.ResizeObserver = class extends NativeResizeObserver {
        observe(target) {
          globalThis.__responsiveObserved.push(target);
          return super.observe(target);
        }
        unobserve(target) {
          globalThis.__responsiveUnobserved.push(target);
          return super.unobserve(target);
        }
      };
    });
    await page.goto(`${baseURL}/components/interactive/bar`);
    await page.locator("[data-goshtoso-chart-wrapper]").first().waitFor();
    await page.waitForFunction(() => Boolean(window.__goshtosoChartsThemeRuntime));
    const observed = await page.evaluate(() => {
      const figure = document.querySelector(".goshtoso-charts-interactive");
      const host = figure.querySelector("[_echarts_instance_]");
      host.dataset.resizeProbe = "host";
      host.parentElement.dataset.resizeProbe = "container";
      const before = globalThis.__responsiveObserved.filter((target) => target === host || target === host.parentElement).length;
      window.__goshtosoChartsThemeRuntime.register(figure);
      const after = globalThis.__responsiveObserved.filter((target) => target === host || target === host.parentElement).length;
      return { before, after };
    });
    assert.deepEqual(observed, { before: 2, after: 2 });
    await page.locator("[data-goshtoso-chart-wrapper]").first().evaluate((element) => element.remove());
    await page.waitForFunction(() => {
      const probes = globalThis.__responsiveUnobserved.map((target) => target.dataset.resizeProbe);
      return probes.includes("host") && probes.includes("container");
    });
  } finally {
    await page.close();
  }
});

test("Tree node hit targets settle cleanly and plot bounds remain centered", async () => {
  const page = await pageAt("/components/interactive/tree");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    await wrapper.scrollIntoViewIfNeeded();
    const inspect = async (name) => wrapper.evaluate((element, nodeName) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      const data = instance.getModel().getSeriesByIndex(0).getData();
      let index = -1;
      const xs = [], ys = [];
      for (let candidate = 0; candidate < data.count(); candidate += 1) {
        if (data.getName(candidate) === nodeName) index = candidate;
      }
      const node = data.tree.getNodeByDataIndex(index);
      const display = instance.getZr().storage.getDisplayList();
		const symbols = display.filter((item) => item.type === "path" && item.z2 === 100 && item.shape && item.shape.symbolType === "circle");
		for (const symbol of symbols) {
			const transform = symbol.getComputedTransform();
			xs.push(transform[4]);
			ys.push(transform[5]);
		}
		const selectedLayout = data.getItemLayout(index);
		const seriesTransform = display.find((item) => item.type === "bezier-curve")?.getComputedTransform() || [1, 0, 0, 1, 0, 0];
		const selectedPixel = [selectedLayout.x + seriesTransform[4], selectedLayout.y + seriesTransform[5]];
		return { index, expanded: node.isExpand, x: selectedPixel[0], y: selectedPixel[1],
        hostLeft: host.getBoundingClientRect().left, hostTop: host.getBoundingClientRect().top,
		labels: display.filter((item) => item.type === "text" || item.type === "tspan").length,
        paths: display.filter((item) => item.type === "path" || item.type === "bezier-curve").length,
		nodeLabel: display.filter((item) => (item.type === "text" || item.type === "tspan") && item.style && item.style.text === nodeName).length,
        centerDelta: Math.abs((Math.min(...xs) + Math.max(...xs)) / 2 - instance.getWidth() / 2) };
    }, name);
    const clickNode = async (name, expected) => {
      const before = await inspect(name);
      await page.mouse.click(before.hostLeft + before.x, before.hostTop + before.y);
      await page.waitForFunction(({ nodeName, value }) => {
        const host = document.querySelector("[_echarts_instance_]");
        const instance = window.echarts.getInstanceByDom(host);
        const data = instance.getModel().getSeriesByIndex(0).getData();
        for (let index = 0; index < data.count(); index += 1) if (data.getName(index) === nodeName) return data.tree.getNodeByDataIndex(index).isExpand === value;
        return false;
      }, { nodeName: name, value: expected });
      await page.waitForTimeout(150);
      return inspect(name);
    };
    const initial = await inspect("Node3");
    assert.equal(initial.expanded, false);
    assert.equal(initial.nodeLabel, 1);
    assert.ok(initial.centerDelta <= 2, `tree center delta ${initial.centerDelta}`);
	if (screenshotDirectory) await page.screenshot({ path: path.join(screenshotDirectory, "tree-normal-centered.png"), fullPage: true });
    const opened = await clickNode("Node3", true);
    assert.equal(opened.nodeLabel, 1);
    assert.ok(opened.labels > initial.labels && opened.paths > initial.paths);
    const closed = await clickNode("Node3", false);
    assert.deepEqual({ labels: closed.labels, paths: closed.paths, nodeLabel: closed.nodeLabel }, { labels: initial.labels, paths: initial.paths, nodeLabel: 1 });
    const node2Closed = await clickNode("Node2", false);
    assert.equal(node2Closed.nodeLabel, 1);
    assert.ok(node2Closed.labels < closed.labels && node2Closed.paths < closed.paths);
    const node2Opened = await clickNode("Node2", true);
    assert.equal(node2Opened.nodeLabel, 1);
    assert.deepEqual({ labels: node2Opened.labels, paths: node2Opened.paths }, { labels: closed.labels, paths: closed.paths });
  } finally {
    await page.close();
  }
});

test("Expand uses Goshtoso Modal, scales static Scatter, and restores focus", async () => {
  const page = await pageAt("/components/scatter");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    await wrapper.evaluate((element) => {
      element.__expandedContent = element.querySelector("[data-goshtoso-chart-content]");
    });
    const trigger = await openExpand(wrapper);
		const dialog = wrapper.getByRole("dialog", { name: "Dense scatter data" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForTimeout(350);
    assert.equal(await dialog.getAttribute("aria-modal"), "true");
    const geometry = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
      const body = panel.children[1];
      const content = body.querySelector("[data-goshtoso-chart-content]");
      const svg = content.querySelector("svg");
      const panelRect = panel.getBoundingClientRect();
      const bodyRect = body.getBoundingClientRect();
      const svgRect = svg.getBoundingClientRect();
      return {
        panel: { left: panelRect.left, top: panelRect.top, right: panelRect.right, bottom: panelRect.bottom, width: panelRect.width, height: panelRect.height },
        body: { left: bodyRect.left, top: bodyRect.top, right: bodyRect.right, bottom: bodyRect.bottom },
        svg: { left: svgRect.left, top: svgRect.top, right: svgRect.right, bottom: svgRect.bottom, width: svgRect.width, height: svgRect.height },
        contentMoved: content === panel.closest("[data-goshtoso-chart-wrapper]").__expandedContent,
      };
    });
    assert.ok(geometry.panel.width >= 1000);
    assert.ok(geometry.panel.height >= 700);
    assert.ok(Math.abs((geometry.panel.left + geometry.panel.right) / 2 - 640) < 4);
    assert.ok(geometry.svg.left >= geometry.body.left && geometry.svg.right <= geometry.body.right + 1);
    assert.ok(geometry.svg.top >= geometry.body.top && geometry.svg.bottom <= geometry.body.bottom + 1);
    assert.ok(geometry.svg.width > 900 && geometry.svg.height > 500);
    assert.equal(geometry.contentMoved, true);
    await page.keyboard.press("Tab");
    assert.equal(await dialog.evaluate((element) => element.contains(document.activeElement)), true);
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });
    await page.waitForTimeout(350);
    assert.deepEqual(await wrapper.evaluate((element) => ({
      sameContent: element.__expandedContent === element.querySelector("[data-goshtoso-chart-content]"),
      focusReturned: document.activeElement === element.querySelector('[id$="-stacked"] > button'),
    })), { sameContent: true, focusReturned: true });
  } finally {
    await page.close();
  }
});

test("Expand scales interactive Tree and preserves renderer, hierarchy state, and theme", async () => {
  const page = await pageAt("/components/interactive/tree");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    const before = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      const data = instance.getModel().getSeriesByIndex(0).getData();
      let nodeIndex = -1;
      for (let index = 0; index < data.count(); index += 1) if (data.getName(index) === "Node3") nodeIndex = index;
      instance.dispatchAction({ type: "treeExpandAndCollapse", seriesIndex: 0, dataIndex: nodeIndex });
      element.__modalTreeInstance = instance;
      element.__modalTreeNodeIndex = nodeIndex;
      return { width: instance.getWidth(), height: instance.getHeight(), background: instance.getOption().backgroundColor };
    });
    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "Basic tree example" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForTimeout(350);
    const expanded = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      const canvas = host.querySelector("canvas").getBoundingClientRect();
      const node = instance.getModel().getSeriesByIndex(0).getData().tree.getNodeByDataIndex(element.__modalTreeNodeIndex);
	  const symbols = instance.getZr().storage.getDisplayList().filter((item) => item.type === "path" && item.z2 === 100 && item.shape && item.shape.symbolType === "circle");
	  const xs = symbols.map((item) => item.getComputedTransform()[4]);
      return {
        sameInstance: instance === element.__modalTreeInstance,
        nodeExpanded: node.isExpand,
        width: instance.getWidth(), height: instance.getHeight(),
        hostWidth: host.clientWidth, hostHeight: host.clientHeight,
        canvasWidth: Math.round(canvas.width), canvasHeight: Math.round(canvas.height),
	    centerDelta: Math.abs((Math.min(...xs) + Math.max(...xs)) / 2 - instance.getWidth() / 2),
      };
    });
    assert.equal(expanded.sameInstance, true);
    assert.equal(expanded.nodeExpanded, true);
	assert.ok(expanded.centerDelta <= 2, `modal tree center delta ${expanded.centerDelta}`);
    assert.ok(expanded.width > before.width && expanded.height > before.height);
    assert.deepEqual(
      { width: expanded.width, height: expanded.height, canvasWidth: expanded.canvasWidth, canvasHeight: expanded.canvasHeight },
      { width: expanded.hostWidth, height: expanded.hostHeight, canvasWidth: expanded.hostWidth, canvasHeight: expanded.hostHeight },
    );
	if (screenshotDirectory) await page.screenshot({ path: path.join(screenshotDirectory, "tree-expand-centered.png"), fullPage: true });
    await page.evaluate(() => document.documentElement.classList.add("dark"));
    await page.waitForFunction((value) => {
      const element = document.querySelector("[data-goshtoso-chart-wrapper]");
      return window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")).getOption().backgroundColor !== value;
    }, before.background);
    assert.equal(await wrapper.evaluate((element) => {
      const instance = window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]"));
      const node = instance.getModel().getSeriesByIndex(0).getData().tree.getNodeByDataIndex(element.__modalTreeNodeIndex);
      return instance === element.__modalTreeInstance && node.isExpand;
    }), true);
    await dialog.getByRole("button", { name: "close modal" }).click();
    await dialog.waitFor({ state: "hidden" });
    await page.waitForTimeout(350);
    assert.deepEqual(await wrapper.evaluate((element) => ({
      sameInstance: window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")) === element.__modalTreeInstance,
      focusReturned: document.activeElement === element.querySelector('[id$="-stacked"] > button'),
    })), { sameInstance: true, focusReturned: true });
  } finally {
    await page.close();
  }
});

test("native fullscreen enters and exits, preserves instance, resizes, and returns focus", async () => {
  const page = await pageAt("/components/interactive/line");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    const initial = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      element.__instanceIdentity = instance;
      element.__resizeEvents = 0;
      element.addEventListener("goshtoso-charts:resize", () => { element.__resizeEvents += 1; });
      return { width: instance.getWidth(), height: instance.getHeight() };
    });
    const { action: button } = await enterFullscreen(wrapper);
    await page.waitForFunction(() => Boolean(document.fullscreenElement));
    await page.waitForTimeout(350);
    assert.equal(await button.getAttribute("aria-pressed"), "true");
    const full = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      const canvas = host.querySelector("canvas");
      return {
        chartWidth: instance.getWidth(), chartHeight: instance.getHeight(),
        hostWidth: host.clientWidth, hostHeight: host.clientHeight,
        canvasWidth: Math.round(canvas.getBoundingClientRect().width),
        canvasHeight: Math.round(canvas.getBoundingClientRect().height),
      };
    });
    assert.ok(full.chartWidth > initial.width, `${full.chartWidth} <= ${initial.width}`);
    assert.ok(full.chartHeight > initial.height, `${full.chartHeight} <= ${initial.height}`);
    assert.equal(full.chartWidth, full.hostWidth);
    assert.equal(full.chartHeight, full.hostHeight);
    assert.equal(full.canvasWidth, full.hostWidth);
    assert.equal(full.canvasHeight, full.hostHeight);
    await page.evaluate(() => document.exitFullscreen());
    await page.waitForFunction(() => !document.fullscreenElement);
    await page.waitForTimeout(350);
    const restored = await wrapper.evaluate((element) => ({
      sameInstance: element.__instanceIdentity === window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")),
      resizeEvents: element.__resizeEvents,
      focusReturned: document.activeElement === element.querySelector('[id$="-stacked"] > button'),
    }));
    assert.equal(restored.sameInstance, true);
    assert.equal(restored.focusReturned, true);
    assert.ok(restored.resizeEvents >= 4, `resize events = ${restored.resizeEvents}`);
  } finally {
    await page.close();
  }
});

test("Escape exits the safe fullscreen fallback and returns focus", async () => {
  const page = await pageAt("/components/interactive/bar");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    const initial = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      element.__fallbackInstance = instance;
      return { width: instance.getWidth(), height: instance.getHeight() };
    });
    await wrapper.evaluate((element) => { element.requestFullscreen = undefined; });
    const { action: button } = await enterFullscreen(wrapper);
    await page.waitForTimeout(350);
    assert.equal(await wrapper.evaluate((element) => element.classList.contains("goshtoso-charts-fullscreen-fallback")), true);
    assert.equal(await button.getAttribute("aria-pressed"), "true");
    const full = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return { width: instance.getWidth(), height: instance.getHeight(), hostWidth: host.clientWidth, hostHeight: host.clientHeight };
    });
    assert.ok(full.width > initial.width);
    assert.ok(full.height > initial.height);
    assert.deepEqual({ width: full.width, height: full.height }, { width: full.hostWidth, height: full.hostHeight });
    await page.keyboard.press("Escape");
    await page.waitForTimeout(350);
    assert.deepEqual(await wrapper.evaluate((element) => ({
      active: element.classList.contains("goshtoso-charts-fullscreen-fallback"),
      focusReturned: document.activeElement === element.querySelector('[id$="-stacked"] > button'),
      sameInstance: element.__fallbackInstance === window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")),
    })), { active: false, focusReturned: true, sameInstance: true });
  } finally {
    await page.close();
  }
});

test("theme changes and live SSE updates preserve the interactive instance", async () => {
  const page = await pageAt("/examples/live-availability");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    const initial = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      element.__instanceIdentity = instance;
      return {
        background: instance.getOption().backgroundColor,
        categories: instance.getOption().xAxis[0].data.join("|"),
      };
    });
    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "Live availability from server-sent events" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForTimeout(350);
    await page.evaluate(() => document.documentElement.classList.add("dark"));
    await page.waitForFunction((value) => {
      const wrapper = document.querySelector("[data-goshtoso-chart-wrapper]");
      const instance = window.echarts.getInstanceByDom(wrapper.querySelector("[_echarts_instance_]"));
      return instance.getOption().backgroundColor !== value;
    }, initial.background);
    await page.waitForFunction((value) => {
      const wrapper = document.querySelector("[data-goshtoso-chart-wrapper]");
      const instance = window.echarts.getInstanceByDom(wrapper.querySelector("[_echarts_instance_]"));
      return instance.getOption().xAxis[0].data.join("|") !== value;
    }, initial.categories, { timeout: 5000 });
    assert.equal(await wrapper.evaluate((element) =>
      element.__instanceIdentity === window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]"))), true);
    assert.equal(await dialog.locator("[data-goshtoso-chart-content]").count(), 1);
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });
  } finally {
    await page.close();
  }
});

test("Sunburst drill-down and back survive modal, theme, resize, and fullscreen on one instance", async () => {
  const page = await pageAt("/components/interactive/sunburst");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    const initial = await wrapper.evaluate((element) => {
      element.style.width = "607px";
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      element.__sunburstInstance = instance;
      const series = instance.getModel().getSeriesByIndex(0);
      const target = series.getData().tree.root.children[0];
      instance.dispatchAction({ type: "sunburstRootToNode", seriesIndex: 0, targetNode: target });
      return { root: series.getData().tree.root.name, target: target.name };
    });
    await page.waitForTimeout(350);
    const measure = async () => wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return {
        sameInstance: instance === element.__sunburstInstance,
        viewRoot: instance.getModel().getSeriesByIndex(0).getViewRoot().name,
        hostWidth: host.clientWidth,
        canvasWidth: instance.getWidth(),
      };
    });
    let state = await measure();
    assert.equal(state.sameInstance, true);
    assert.equal(state.viewRoot, initial.target);
    assert.equal(state.canvasWidth, state.hostWidth);

    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "Basic sunburst example" });
    await dialog.waitFor({ state: "visible" });
    await page.evaluate(() => document.documentElement.classList.add("dark"));
    await page.waitForTimeout(400);
    state = await measure();
    assert.equal(state.sameInstance, true);
    assert.equal(state.viewRoot, initial.target);
    assert.equal(state.canvasWidth, state.hostWidth);
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    await enterFullscreen(wrapper);
    await page.waitForFunction(() => document.fullscreenElement !== null);
    await page.waitForTimeout(350);
    state = await measure();
    assert.equal(state.sameInstance, true);
    assert.equal(state.viewRoot, initial.target);
    assert.equal(state.canvasWidth, state.hostWidth);
    await page.evaluate(() => document.exitFullscreen());
    await page.waitForFunction(() => document.fullscreenElement === null);
    await page.waitForTimeout(350);

    const restored = await wrapper.evaluate((element, rootName) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      const series = instance.getModel().getSeriesByIndex(0);
      instance.dispatchAction({ type: "sunburstRootToNode", seriesIndex: 0, targetNode: series.getData().tree.root });
      return {
        sameInstance: instance === element.__sunburstInstance,
        viewRoot: series.getViewRoot().name,
        rootName,
      };
    }, initial.root);
    assert.deepEqual(restored, { sameInstance: true, viewRoot: initial.root, rootName: initial.root });

    const expectedPNG = await wrapper.evaluate((element) => {
      const instance = window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]"));
      return { width: instance.getWidth(), height: instance.getHeight() };
    });
    const png = await download(page, "Download Basic sunburst example as PNG");
    assert.equal(png.filename, "basic-sunburst-example.png");
    assert.equal(png.types.at(-1), "image/png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, expectedPNG);
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
  } finally {
    await page.close();
  }
});

test("Treemap focus and native breadcrumb back survive modal, theme, resize, and fullscreen on one instance", async () => {
  const page = await pageAt("/components/interactive/treemap");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    const host = wrapper.locator("[_echarts_instance_]");
    await host.scrollIntoViewIfNeeded();
    await wrapper.evaluate((element) => {
      const chartHost = element.querySelector("[_echarts_instance_]");
      element.__treemapInstance = window.echarts.getInstanceByDom(chartHost);
    });
    await host.click({ position: { x: 100, y: 188 } });
    await page.waitForFunction(() => {
      const chartHost = document.querySelector("[_echarts_instance_]");
      return window.echarts.getInstanceByDom(chartHost).getModel().getSeriesByIndex(0).getViewRoot().name === "d1";
    });
    const measure = () => wrapper.evaluate((element) => {
      const chartHost = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(chartHost);
      return {
        sameInstance: instance === element.__treemapInstance,
        viewRoot: instance.getModel().getSeriesByIndex(0).getViewRoot().name,
        hostWidth: chartHost.clientWidth,
        chartWidth: instance.getWidth(),
        canvasWidth: Math.round(chartHost.querySelector("canvas").getBoundingClientRect().width),
      };
    });

	let state = await measure();
	assert.equal(state.sameInstance, true);
	assert.equal(state.viewRoot, "d1");
	assert.deepEqual({ chart: state.chartWidth, canvas: state.canvasWidth }, { chart: state.hostWidth, canvas: state.hostWidth });

    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "Basic treemap example" });
    await dialog.waitFor({ state: "visible" });
    await page.evaluate(() => document.documentElement.classList.add("dark"));
    await page.waitForTimeout(450);
	state = await measure();
    assert.equal(state.sameInstance, true);
    assert.equal(state.viewRoot, "d1");
    assert.deepEqual({ chart: state.chartWidth, canvas: state.canvasWidth }, { chart: state.hostWidth, canvas: state.hostWidth });
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    await wrapper.evaluate((element) => { element.style.width = "607px"; });
    await page.waitForFunction(() => {
      const chartHost = document.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(chartHost);
      return chartHost.clientWidth === 607 && instance.getWidth() === 607 &&
        Math.round(chartHost.querySelector("canvas").getBoundingClientRect().width) === 607;
    });
    state = await measure();
    assert.deepEqual(state, { sameInstance: true, viewRoot: "d1", hostWidth: 607, chartWidth: 607, canvasWidth: 607 });

    await enterFullscreen(wrapper);
    await page.waitForFunction(() => document.fullscreenElement !== null);
    await page.waitForTimeout(350);
    state = await measure();
    assert.equal(state.sameInstance, true);
    assert.equal(state.viewRoot, "d1");
    assert.deepEqual({ chart: state.chartWidth, canvas: state.canvasWidth }, { chart: state.hostWidth, canvas: state.hostWidth });
    await page.evaluate(() => document.exitFullscreen());
    await page.waitForFunction(() => document.fullscreenElement === null);
    await page.waitForTimeout(350);

    const back = await wrapper.evaluate((element) => {
      const instance = window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]"));
      const rootBreadcrumb = instance.getZr().storage.getDisplayList().find((item) => item.z2 === 100000 && item.cursor === "pointer");
      if (!rootBreadcrumb || typeof rootBreadcrumb.onclick !== "function") throw new Error("native root breadcrumb not found");
      rootBreadcrumb.onclick();
      return { sameInstance: instance === element.__treemapInstance, nativePointer: rootBreadcrumb.cursor };
    });
    assert.deepEqual(back, { sameInstance: true, nativePointer: "pointer" });
    await page.waitForFunction(() => {
      const chartHost = document.querySelector("[_echarts_instance_]");
      return window.echarts.getInstanceByDom(chartHost).getModel().getSeriesByIndex(0).getViewRoot().name === "Basic treemap example";
    });

    const expected = await measure();
    assert.equal(await page.getByRole("button", { name: /Basic treemap example as SVG/ }).count(), 0);
    const png = await download(page, "Download Basic treemap example as PNG");
    assert.equal(png.filename, "basic-treemap-example.png");
    assert.equal(png.types.at(-1), "image/png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: expected.chartWidth, height: 500 });
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
  } finally {
    await page.close();
  }
});

test("Parallel coordinates preserve instance, data, theme lines, resize, controls, and opaque PNG", async () => {
  const page = await pageAt("/components/interactive/parallel");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      element.__parallelInstance = window.echarts.getInstanceByDom(host);
    });
    const measure = () => wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      const option = instance.getOption();
      return {
        sameInstance: instance === element.__parallelInstance,
        hostWidth: host.clientWidth,
        chartWidth: instance.getWidth(),
        chartHeight: instance.getHeight(),
        canvasWidth: Math.round(host.querySelector("canvas").getBoundingClientRect().width),
        seriesRows: option.series.map((series) => series.data.length),
        seriesNames: option.series.map((series) => series.name),
        colors: option.series.map((series) => series.lineStyle.color),
      };
    });

    let state = await measure();
    assert.deepEqual(state.seriesRows, [21, 21, 21]);
    assert.deepEqual(state.seriesNames, ["Beijing", "Guangzhou", "Shanghai"]);
    assert.equal(new Set(state.colors).size, 3);

    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "Multi Series parallel coordinates" });
    await dialog.waitFor({ state: "visible" });
    await page.evaluate(() => {
      document.documentElement.dataset.theme = "araihu";
      document.documentElement.classList.add("dark");
    });
    await page.waitForTimeout(450);
    const themed = await measure();
    assert.equal(themed.sameInstance, true);
    assert.equal(new Set(themed.colors).size, 3);
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    await wrapper.evaluate((element) => { element.querySelector("[_echarts_instance_]").style.width = "607px"; });
    await page.waitForFunction(() => {
      const host = document.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth === 607 && instance.getWidth() === 607 &&
        Math.round(host.querySelector("canvas").getBoundingClientRect().width) === 607;
    });
    state = await measure();
    assert.deepEqual(
      { same: state.sameInstance, host: state.hostWidth, chart: state.chartWidth, canvas: state.canvasWidth },
      { same: true, host: 607, chart: 607, canvas: 607 },
    );

    await enterFullscreen(wrapper);
    await page.waitForFunction(() => document.fullscreenElement !== null);
    await page.waitForTimeout(350);
    state = await measure();
    assert.equal(state.sameInstance, true);
    assert.deepEqual({ chart: state.chartWidth, canvas: state.canvasWidth }, { chart: state.hostWidth, canvas: state.hostWidth });
    await page.evaluate(() => document.exitFullscreen());
    await page.waitForFunction(() => document.fullscreenElement === null);
    await page.waitForTimeout(350);

    state = await measure();
    assert.equal(state.sameInstance, true);
    assert.deepEqual(state.seriesRows, [21, 21, 21]);
    assert.equal(await page.getByRole("button", { name: /Multi Series parallel coordinates as SVG/ }).count(), 0);
    const png = await download(page, "Download Multi Series parallel coordinates as PNG");
    assert.equal(png.filename, "multi-series-parallel-coordinates.png");
    assert.equal(png.types.at(-1), "image/png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: state.chartWidth, height: state.chartHeight });
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
  } finally {
    await page.close();
  }
});

test("Parallel flex consumer layout and expand modal resize the existing host instance", async () => {
  const page = await pageAt("/components/interactive/parallel");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    await wrapper.evaluate((element) => {
      const content = element.querySelector("[data-goshtoso-chart-content]");
      const host = element.querySelector("[_echarts_instance_]");
      content.style.display = "flex";
      content.style.width = "700px";
      content.style.maxWidth = "100%";
      host.style.width = "100%";
      host.style.minWidth = "0";
      host.style.flex = "1 1 auto";
      element.__parallelFlexInstance = window.echarts.getInstanceByDom(host);
    });
    await page.waitForTimeout(500);
    const initialFlex = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return { host: host.clientWidth, chart: instance.getWidth(), canvas: Math.round(host.querySelector("canvas").getBoundingClientRect().width) };
    });
    assert.ok(initialFlex.host > 400 && initialFlex.host < 900, `initial flex host width ${initialFlex.host}`);
    assert.deepEqual({ chart: initialFlex.chart, canvas: initialFlex.canvas }, { chart: initialFlex.host, canvas: initialFlex.host });

    await wrapper.evaluate((element) => {
      element.querySelector("[data-goshtoso-chart-content]").style.width = "400px";
    });
    await page.waitForFunction((previousWidth) => {
      const host = document.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth < previousWidth && instance.getWidth() === host.clientWidth &&
        Math.round(host.querySelector("canvas").getBoundingClientRect().width) === host.clientWidth;
    }, initialFlex.host);
    assert.equal(await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      return window.echarts.getInstanceByDom(host) === element.__parallelFlexInstance;
    }), true);

    await wrapper.evaluate((element) => {
      element.querySelector("[data-goshtoso-chart-content]").style.width = "100%";
    });
    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "Multi Series parallel coordinates" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForFunction(() => {
      const wrapperElement = document.querySelector("[data-goshtoso-chart-wrapper]");
      const host = wrapperElement.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return instance === wrapperElement.__parallelFlexInstance && host.clientWidth > 400 &&
        instance.getWidth() === host.clientWidth &&
        Math.round(host.querySelector("canvas").getBoundingClientRect().width) === host.clientWidth;
    });
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });
    assert.equal(await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      return window.echarts.getInstanceByDom(host) === element.__parallelFlexInstance;
    }), true);
  } finally {
    await page.close();
  }
});

test("Candlestick shared controls export exact 600x400 opaque SVG and PNG artifacts", async () => {
  const page = await pageAt("/components/candlestick");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    assert.deepEqual(await wrapper.locator("tbody tr").count(), 7);
	assert.deepEqual(await wrapper.locator("tbody").getByText("Decrease", { exact: false }).count(), 1);
    const svg = await download(page, "Download Seven-day stock price as SVG");
    assert.equal(svg.filename, "basic-candlestick-chart.svg");
    assert.equal(svg.types.at(-1), "image/svg+xml;charset=utf-8");
    const markup = svg.bytes.toString("utf8");
    assert.match(markup, /^<svg\b/);
    assert.match(markup, /width="600"/);
    assert.match(markup, /height="400"/);
    assert.doesNotMatch(markup, /var\(/);

    const png = await download(page, "Download Seven-day stock price as PNG");
    assert.equal(png.filename, "basic-candlestick-chart.png");
    assert.equal(png.types.at(-1), "image/png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: 600, height: 400 });
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
  } finally {
    await page.close();
  }
});

test("Violin exports self-contained 1200x800 SVG and opaque PNG artifacts", async () => {
  const page = await pageAt("/components/violin");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    assert.equal(await wrapper.locator("tbody tr").count(), 8);
    assert.equal(await wrapper.getByText("Normal", { exact: true }).count() >= 1, true);
    const svg = await download(page, "Download Distribution shapes from deterministic samples as SVG");
    assert.equal(svg.filename, "distribution-shapes-from-deterministic-samples.svg");
    assert.equal(svg.types.at(-1), "image/svg+xml;charset=utf-8");
    const markup = svg.bytes.toString("utf8");
    assert.match(markup, /^<svg\b/);
    assert.match(markup, /width="1200"/);
    assert.match(markup, /height="800"/);
    assert.match(markup, /preserveAspectRatio="xMidYMid meet"/);
    assert.doesNotMatch(markup, /var\(|<script|(?:href|src)="https?:\/\//);

    const png = await download(page, "Download Distribution shapes from deterministic samples as PNG");
    assert.equal(png.filename, "distribution-shapes-from-deterministic-samples.png");
    assert.equal(png.types.at(-1), "image/png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: 1200, height: 800 });
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
  } finally {
    await page.close();
  }
});

for (const width of [390, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const mode of ["light", "dark"]) {
      test(`Violin ${width}px ${theme} ${mode} contains page and scales expanded SVG`, async () => {
        const page = await pageAt("/components/violin", { width, height: 900 });
        const errors = [];
        page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
        page.on("pageerror", (error) => errors.push(error.message));
        try {
          await page.evaluate(({ selected, dark }) => {
            document.documentElement.dataset.theme = selected;
            document.documentElement.classList.toggle("dark", dark);
          }, { selected: theme, dark: mode === "dark" });
          const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
          const pageWidth = await page.evaluate(() => ({ client: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth }));
          assert.equal(pageWidth.scroll, pageWidth.client);
          const initial = await wrapper.locator(".goshtoso-charts-violin__viewport svg").evaluate((svg) => {
            const rect = svg.getBoundingClientRect();
            return { width: rect.width, height: rect.height, ratio: rect.width / rect.height };
          });
          assert.ok(initial.width <= width);
          assert.ok(Math.abs(initial.ratio - 1.5) < 0.01);

          await openExpand(wrapper);
          const dialog = wrapper.getByRole("dialog", { name: "Distribution shapes from deterministic samples" });
          await dialog.waitFor({ state: "visible" });
          await page.waitForTimeout(350);
          const geometry = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
            const body = panel.children[1].getBoundingClientRect();
            const svg = panel.querySelector(".goshtoso-charts-violin__viewport svg").getBoundingClientRect();
            const panelRect = panel.getBoundingClientRect();
            const root = panel.querySelector(".goshtoso-charts-violin__viewport svg");
            return { body, svg, panel: panelRect.toJSON(), viewBoxRatio: root.viewBox.baseVal.width / root.viewBox.baseVal.height, preserveAspectRatio: root.getAttribute("preserveAspectRatio") };
          });
          assert.ok(Math.abs((geometry.panel.left + geometry.panel.right) / 2 - width / 2) < 4);
          assert.ok(geometry.svg.left >= geometry.body.left - 1 && geometry.svg.right <= geometry.body.right + 1);
          assert.ok(geometry.svg.top >= geometry.body.top - 1 && geometry.svg.bottom <= geometry.body.bottom + 1);
          assert.ok(Math.abs(geometry.viewBoxRatio - 1.5) < 0.01);
          assert.equal(geometry.preserveAspectRatio, "xMidYMid meet");
          assert.deepEqual(errors, []);
          await page.keyboard.press("Escape");
          await dialog.waitFor({ state: "hidden" });
        } finally {
          await page.close();
        }
      });
    }
  }
}

for (const mode of ["light", "dark"]) {
  test(`static SVG and PNG exports contain resolved ${mode} artifacts`, async () => {
    const page = await pageAt("/components/line");
    try {
      await page.evaluate((dark) => document.documentElement.classList.toggle("dark", dark), mode === "dark");
      const svg = await download(page, "Download HTTPS monitor latency in milliseconds as SVG");
      assert.equal(svg.filename, "https-monitor-latency.svg");
      assert.equal(svg.types.at(-1), "image/svg+xml;charset=utf-8");
      assert.ok(svg.bytes.length > 1000);
      const markup = svg.bytes.toString("utf8");
      assert.match(markup, /^<svg\b/);
      assert.match(markup, /width="720"/);
      assert.match(markup, /height="320"/);
      assert.doesNotMatch(markup, /var\(/);

      const png = await download(page, "Download HTTPS monitor latency in milliseconds as PNG");
      assert.equal(png.filename, "https-monitor-latency.png");
      assert.equal(png.types.at(-1), "image/png");
      assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
      const metadata = await sharp(png.bytes).metadata();
      assert.equal(metadata.width, 720);
      assert.equal(metadata.height, 320);
      const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
      for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
    } finally {
      await page.close();
    }
  });
}

test("interactive PNG export matches live chart dimensions and remains opaque", async () => {
  const page = await pageAt("/components/interactive/bar");
  try {
    const expected = await page.locator("[data-goshtoso-chart-wrapper]").first().evaluate((element) => {
      const instance = window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]"));
      return { width: instance.getWidth(), height: instance.getHeight() };
    });
    const png = await download(page, "Download Weekly deployments by environment as PNG");
    assert.equal(png.filename, "interactive-deployments.png");
    assert.equal(png.types.at(-1), "image/png");
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, expected);
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
  } finally {
    await page.close();
  }
});

test("Gauge thermal scale uses low-mid-high theme tokens and preserves instance on live theme switch", async () => {
  const page = await pageAt("/components/interactive/gauge");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    const initial = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      element.__gaugeInstance = instance;
      const series = instance.getOption().series[0];
      return { colors: series.axisLine.lineStyle.color, progressWidth: series.progress.width, detail: series.data[0].value };
    });
    assert.deepEqual(initial.colors.map((stop) => stop[0]), [0.34, 0.67, 1]);
    assert.equal(new Set(initial.colors.map((stop) => stop[1])).size, 3);
    assert.equal(initial.progressWidth, 6);
    assert.equal(initial.detail, 73);
    await page.evaluate(() => document.documentElement.classList.toggle("dark"));
    await page.waitForFunction((before) => {
      const host = document.querySelector("[_echarts_instance_]");
      const colors = window.echarts.getInstanceByDom(host).getOption().series[0].axisLine.lineStyle.color;
      return JSON.stringify(colors) !== JSON.stringify(before);
    }, initial.colors);
    const themed = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return { sameInstance: instance === element.__gaugeInstance, colors: instance.getOption().series[0].axisLine.lineStyle.color, value: instance.getOption().series[0].data[0].value };
    });
    assert.equal(themed.sameInstance, true);
    assert.equal(new Set(themed.colors.map((stop) => stop[1])).size, 3);
    assert.notDeepEqual(themed.colors, initial.colors);
    assert.equal(themed.value, 73);
	if (screenshotDirectory) await page.screenshot({ path: path.join(screenshotDirectory, "gauge-thermal-dark.png"), fullPage: true });
  } finally {
    await page.close();
  }
});

test("static Scatter exports SVG and opaque PNG with intrinsic dimensions", async () => {
  const page = await pageAt("/components/scatter");
  try {
	if (screenshotDirectory) await page.screenshot({ path: path.join(screenshotDirectory, "scatter-dense.png"), fullPage: true });
		const svg = await download(page, "Download Dense scatter data as SVG");
		assert.equal(svg.filename, "dense-scatter-data.svg");
    assert.equal(svg.types.at(-1), "image/svg+xml;charset=utf-8");
    assert.match(svg.bytes.toString("utf8"), /^<svg\b/);
    assert.doesNotMatch(svg.bytes.toString("utf8"), /var\(/);

		const png = await download(page, "Download Dense scatter data as PNG");
		assert.equal(png.filename, "dense-scatter-data.png");
    assert.equal(png.types.at(-1), "image/png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
		assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: 600, height: 400 });
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
  } finally {
    await page.close();
  }
});

test("static Heat Map preserves sequential colors and exports resolved 600x400 SVG and opaque PNG", async () => {
  const page = await pageAt("/components/heatmap");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    const presentation = await wrapper.evaluate((element) => {
      const cells = [...element.querySelectorAll("figure svg path[class]")];
      const viewport = element.querySelector(".goshtoso-charts-heatmap__viewport");
      return {
        count: cells.length,
        low: getComputedStyle(cells[15]).fill,
        high: getComputedStyle(cells[7]).fill,
        localOverflow: viewport.scrollWidth >= viewport.clientWidth,
        pageOverflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
      };
    });
    assert.equal(presentation.count, 25);
    assert.notEqual(presentation.low, presentation.high);
    assert.equal(presentation.localOverflow, true);
    assert.equal(presentation.pageOverflow, 0);

    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "Basic heat map" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForTimeout(350);
    const modalGeometry = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
      const panelRect = panel.getBoundingClientRect();
      const svgRect = panel.querySelector(".goshtoso-charts-heatmap__viewport > svg").getBoundingClientRect();
      return {
        panelContained: panelRect.left >= 0 && panelRect.right <= innerWidth + 1 && panelRect.top >= 0 && panelRect.bottom <= innerHeight + 1,
        centered: Math.abs((panelRect.left + panelRect.right) / 2 - innerWidth / 2) < 4,
        panelWidth: panelRect.width,
        chartContained: svgRect.left >= panelRect.left && svgRect.right <= panelRect.right + 1 && svgRect.top >= panelRect.top && svgRect.bottom <= panelRect.bottom + 1,
        aspect: svgRect.width / svgRect.height,
      };
    });
    assert.deepEqual({
      panelContained: modalGeometry.panelContained,
      centered: modalGeometry.centered,
      chartContained: modalGeometry.chartContained,
    }, { panelContained: true, centered: true, chartContained: true });
    assert.ok(modalGeometry.panelWidth >= 1000, `heat-map modal width ${modalGeometry.panelWidth}`);
    assert.ok(Math.abs(modalGeometry.aspect - 1.5) < 0.02, `heat-map modal aspect ${modalGeometry.aspect}`);
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    const { action: fullscreen } = await enterFullscreen(wrapper);
    await page.waitForFunction(() => Boolean(document.fullscreenElement));
    const fullscreenGeometry = await wrapper.evaluate((element) => {
      const viewport = element.querySelector(".goshtoso-charts-heatmap__viewport");
      const svg = viewport.querySelector("svg");
      return {
        pressed: element.querySelector('[id$="-fullscreen-action"]').getAttribute("aria-pressed"),
        wrapper: { width: element.clientWidth, height: element.clientHeight },
        svg: { width: Math.round(svg.getBoundingClientRect().width), height: Math.round(svg.getBoundingClientRect().height) },
        pageOverflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
      };
    });
    assert.equal(fullscreenGeometry.pressed, "true");
    assert.ok(fullscreenGeometry.wrapper.width >= 1200);
    assert.ok(fullscreenGeometry.svg.width >= 1200);
    assert.equal(fullscreenGeometry.pageOverflow, 0);
    await page.evaluate(() => document.exitFullscreen());
    await page.waitForFunction(() => !document.fullscreenElement);

    const svg = await download(page, "Download Basic heat map as SVG");
    assert.equal(svg.filename, "basic-heat-map.svg");
    assert.equal(svg.types.at(-1), "image/svg+xml;charset=utf-8");
    const markup = svg.bytes.toString("utf8");
    assert.match(markup, /^<svg\b/);
    assert.match(markup, /width="600"/);
    assert.match(markup, /height="400"/);
    assert.doesNotMatch(markup, /var\(/);
    assert.doesNotMatch(markup, /color-mix\(/);

    const png = await download(page, "Download Basic heat map as PNG");
    assert.equal(png.filename, "basic-heat-map.png");
    assert.equal(png.types.at(-1), "image/png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: 600, height: 400 });
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
  } finally {
    await page.close();
  }
});

test("static Radar controls preserve DOM and export resolved 600x400 SVG and opaque PNG", async () => {
  const page = await pageAt("/components/radar");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    await wrapper.evaluate((element) => { element.__radarContent = element.querySelector("[data-goshtoso-chart-content]"); });
    assert.equal(await wrapper.evaluate((element) => element.__radarContent === element.querySelector("[data-goshtoso-chart-content]")), true);

    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "Basic radar chart" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForFunction(
      (element) => element.classList.contains("goshtoso-charts-expanded"),
      await wrapper.elementHandle(),
    );
    assert.equal(await dialog.locator("[data-goshtoso-chart-content]").count(), 1);
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    await enterFullscreen(wrapper);
    await page.waitForFunction(() => document.fullscreenElement !== null);
    await page.evaluate(() => document.exitFullscreen());
    await page.waitForFunction(() => document.fullscreenElement === null);

    const svg = await download(page, "Download Basic radar chart as SVG");
    assert.equal(svg.filename, "basic-radar-chart.svg");
    assert.equal(svg.types.at(-1), "image/svg+xml;charset=utf-8");
    const markup = svg.bytes.toString("utf8");
    assert.match(markup, /^<svg\b/);
    assert.match(markup, /width="600"/);
    assert.match(markup, /height="400"/);
    assert.doesNotMatch(markup, /var\(/);

    const png = await download(page, "Download Basic radar chart as PNG");
    assert.equal(png.filename, "basic-radar-chart.png");
    assert.equal(png.types.at(-1), "image/png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: 600, height: 400 });
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
  } finally {
    await page.close();
  }
});

test("interactive Tree exports only opaque PNG at live dimensions", async () => {
  const page = await pageAt("/components/interactive/tree");
  try {
    assert.equal(await page.getByRole("button", { name: /as SVG/ }).count(), 0);
    const expected = await page.locator("[data-goshtoso-chart-wrapper]").first().evaluate((element) => {
      const instance = window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]"));
      return { width: instance.getWidth(), height: instance.getHeight() };
    });
    const png = await download(page, "Download Basic tree example as PNG");
    assert.equal(png.filename, "basic-tree-example.png");
    assert.equal(png.types.at(-1), "image/png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, expected);
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
  } finally {
    await page.close();
  }
});

test("Theme river stays responsive and theme-distinguishable across viewport and theme matrix", async () => {
  for (const width of [390, 1440]) {
    for (const theme of ["goshtoso", "araihu"]) {
      for (const mode of ["light", "dark"]) {
        const page = await pageAt("/components/interactive/theme-river", { width, height: 900 });
        const errors = [];
        page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
        page.on("pageerror", (error) => errors.push(error.message));
        try {
          await page.evaluate(({ selected, dark }) => {
            document.documentElement.dataset.theme = selected;
            document.documentElement.classList.toggle("dark", dark);
          }, { selected: theme, dark: mode === "dark" });
          await page.waitForTimeout(350);
          const result = await page.evaluate(() => {
            const wrapper = document.querySelector("[data-goshtoso-chart-wrapper]");
            const host = wrapper.querySelector("[_echarts_instance_]");
            const instance = window.echarts.getInstanceByDom(host);
            const colors = instance.getOption().color || [];
            return {
              client: document.documentElement.clientWidth,
              scroll: document.documentElement.scrollWidth,
              hostWidth: host.clientWidth,
              contentWidth: wrapper.querySelector("[data-goshtoso-chart-content]").clientWidth,
              chartWidth: instance.getWidth(),
              canvasWidth: Math.round(host.querySelector("canvas").getBoundingClientRect().width),
              hostHeight: host.clientHeight,
              distinct: new Set(colors.slice(0, 6)).size,
            };
          });
          assert.equal(result.scroll, result.client);
          assert.ok(result.hostWidth <= result.contentWidth + 1);
          assert.deepEqual({ chart: result.chartWidth, canvas: result.canvasWidth }, { chart: result.hostWidth, canvas: result.hostWidth });
          assert.ok(result.hostHeight >= 288 && result.hostHeight <= 500, `host height = ${result.hostHeight}`);
          assert.equal(result.distinct, 6);
          assert.deepEqual(errors, []);
        } finally {
          await page.close();
        }
      }
    }
  }
});

test("Theme river preserves one instance through host resize, modal, theme, and native fullscreen; PNG stays opaque", async () => {
  const page = await pageAt("/components/interactive/theme-river", { width: 1440, height: 900 });
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      element.__themeRiverInstance = window.echarts.getInstanceByDom(host);
    });
    const measure = () => wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return {
        sameInstance: instance === element.__themeRiverInstance,
        hostWidth: host.clientWidth,
        chartWidth: instance.getWidth(),
        canvasWidth: Math.round(host.querySelector("canvas").getBoundingClientRect().width),
        chartHeight: instance.getHeight(),
      };
    });

    await wrapper.evaluate((element) => { element.style.width = "607px"; });
    await page.waitForFunction(() => {
      const wrapper = document.querySelector("[data-goshtoso-chart-wrapper]");
      const host = wrapper.querySelector("[_echarts_instance_]");
      return host.clientWidth === 607 && window.echarts.getInstanceByDom(host).getWidth() === 607;
    });
    let state = await measure();
    assert.deepEqual({ same: state.sameInstance, host: state.hostWidth, chart: state.chartWidth, canvas: state.canvasWidth }, { same: true, host: 607, chart: 607, canvas: 607 });

    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "ThemeRiver-SingleAxis-Time" });
    await dialog.waitFor({ state: "visible" });
    await page.evaluate(() => {
      document.documentElement.dataset.theme = "araihu";
      document.documentElement.classList.add("dark");
    });
    await page.waitForTimeout(450);
    state = await measure();
    assert.equal(state.sameInstance, true);
    assert.deepEqual({ chart: state.chartWidth, canvas: state.canvasWidth }, { chart: state.hostWidth, canvas: state.hostWidth });
    const modalGeometry = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
      const rect = panel.getBoundingClientRect();
      return {
        centered: Math.abs((rect.left + rect.right) / 2 - innerWidth / 2) < 4,
        large: rect.width >= innerWidth * 0.9 && rect.height >= innerHeight * 0.8,
        contained: rect.left >= 0 && rect.right <= innerWidth + 1 && rect.top >= 0 && rect.bottom <= innerHeight + 1,
      };
    });
    assert.deepEqual(modalGeometry, { centered: true, large: true, contained: true });
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    await enterFullscreen(wrapper);
    await page.waitForFunction(() => document.fullscreenElement !== null);
    await page.waitForTimeout(350);
    state = await measure();
    assert.equal(state.sameInstance, true);
    assert.deepEqual({ chart: state.chartWidth, canvas: state.canvasWidth }, { chart: state.hostWidth, canvas: state.hostWidth });
    await page.evaluate(() => document.exitFullscreen());
    await page.waitForFunction(() => document.fullscreenElement === null);
    await page.waitForTimeout(350);

    const expected = await measure();
    assert.deepEqual({ width: expected.chartWidth, height: expected.chartHeight }, { width: 607, height: 337 });
    const png = await download(page, "Download ThemeRiver-SingleAxis-Time as PNG");
    assert.equal(png.filename, "theme-river-single-axis-time.png");
    assert.equal(png.types.at(-1), "image/png");
    assert.deepEqual([...png.bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const metadata = await sharp(png.bytes).metadata();
    assert.deepEqual({ width: metadata.width, height: metadata.height }, { width: expected.chartWidth, height: expected.chartHeight });
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    for (let index = 3; index < pixels.length; index += 4) assert.equal(pixels[index], 255);
  } finally {
    await page.close();
  }
});

test("static Export dropdown uses Goshtoso keyboard and Escape semantics", async () => {
  const page = await pageAt("/components/scatter");
  try {
		const trigger = page.getByRole("button", { name: "Export Dense scatter data" });
    await trigger.focus();
    await trigger.press("Enter");
    const menu = page.getByRole("menu");
    await menu.waitFor({ state: "visible" });
    assert.deepEqual(await menu.getByRole("menuitem").allTextContents(), ["SVG", "PNG"]);
		await page.keyboard.press("Tab");
    assert.equal(await menu.evaluate((element) => element.contains(document.activeElement)), true);
    await page.keyboard.press("Escape");
    await menu.waitFor({ state: "hidden" });
    assert.equal(await trigger.getAttribute("aria-expanded"), "false");
  } finally {
    await page.close();
  }
});

test("static transparent exports retain transparent pixels", async () => {
  const page = await pageAt("/components/pie");
  try {
    const svg = await download(page, "Download Observation states as SVG");
    const markup = svg.bytes.toString("utf8");
    assert.doesNotMatch(markup, /data-goshtoso-chart-export-surface/);
    assert.doesNotMatch(markup, /var\(/);

    const png = await download(page, "Download Observation states as PNG");
    assert.equal(png.types.at(-1), "image/png");
    const pixels = await sharp(png.bytes).ensureAlpha().raw().toBuffer();
    let transparentPixels = 0;
    for (let index = 3; index < pixels.length; index += 4) {
      if (pixels[index] < 255) transparentPixels += 1;
    }
    assert.ok(transparentPixels > 0, "transparent export retained no transparent pixels");
  } finally {
    await page.close();
  }
});

for (const width of [390, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const mode of ["light", "dark"]) {
      test(`${width}px ${theme} ${mode} keeps controls local, accessible, and page-owned layout intact`, async () => {
        const page = await pageAt("/components/line", { width, height: 900 });
        const errors = [];
        page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
        page.on("pageerror", (error) => errors.push(error.message));
        try {
          await page.evaluate(({ selected, dark }) => {
            document.documentElement.dataset.theme = selected;
            document.documentElement.classList.toggle("dark", dark);
          }, { selected: theme, dark: mode === "dark" });
          const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
          assert.equal(await wrapper.getByRole("group", { name: /chart controls/ }).count(), 1);
          assert.equal(await wrapper.locator("button:visible").count(), 2);
          const dimensions = await page.evaluate(() => ({
            client: document.documentElement.clientWidth,
            scroll: document.documentElement.scrollWidth,
          }));
          assert.equal(dimensions.scroll, dimensions.client);
          if (width === 390) {
            assert.deepEqual(await wrapper.locator("[data-goshtoso-chart-content]").evaluate((element) => ({
              localOverflow: getComputedStyle(element).overflowX,
              contained: element.scrollWidth >= element.clientWidth,
            })), { localOverflow: "auto", contained: true });
          }
          const focusTarget = width === 390
            ? wrapper.getByRole("button", { name: /More .* chart actions/ })
            : wrapper.locator('[id$="-stacked"] > button');
          await focusTarget.focus();
          assert.equal(await focusTarget.evaluate((element) => element === document.activeElement), true);
          assert.equal(await wrapper.getByRole("button", { name: /^Collapse / }).count(), 0);
          if (screenshotDirectory) {
            await page.screenshot({
              path: path.join(screenshotDirectory, `chart-controls-${width}-${theme}-${mode}.png`),
              fullPage: true,
            });
          }
          await openExpand(wrapper);
          const dialog = wrapper.getByRole("dialog", { name: "HTTPS monitor latency in milliseconds" });
          await dialog.waitFor({ state: "visible" });
          await page.waitForTimeout(350);
          const modalGeometry = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((panel) => {
            const panelRect = panel.getBoundingClientRect();
            const body = panel.children[1];
            const bodyRect = body.getBoundingClientRect();
            const svgRect = body.querySelector("svg").getBoundingClientRect();
            return {
              panelContained: panelRect.left >= 0 && panelRect.right <= innerWidth + 1 && panelRect.top >= 0 && panelRect.bottom <= innerHeight + 1,
              centered: Math.abs((panelRect.left + panelRect.right) / 2 - innerWidth / 2) < 4,
              chartContained: svgRect.left >= bodyRect.left && svgRect.right <= bodyRect.right + 1 && svgRect.top >= bodyRect.top && svgRect.bottom <= bodyRect.bottom + 1,
              localOverflow: body.scrollWidth <= body.clientWidth + 1,
            };
          });
          assert.deepEqual(modalGeometry, { panelContained: true, centered: true, chartContained: true, localOverflow: true });
          if (screenshotDirectory) {
            await page.screenshot({
              path: path.join(screenshotDirectory, `chart-expand-${width}-${theme}-${mode}.png`),
              fullPage: true,
            });
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

for (const width of [390, 1440]) {
  for (const theme of ["goshtoso", "araihu"]) {
    for (const mode of ["light", "dark"]) {
      test(`${width}px ${theme} ${mode} keeps Candlestick, Heat Map, Treemap, and Parallel contained and theme-legible`, async () => {
		for (const route of ["/components/candlestick", "/components/heatmap", "/components/interactive/treemap", "/components/interactive/parallel"]) {
          const page = await pageAt(route, { width, height: 900 });
          const errors = [];
          page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
          page.on("pageerror", (error) => errors.push(error.message));
          try {
            await page.evaluate(({ selected, dark }) => {
              document.documentElement.dataset.theme = selected;
              document.documentElement.classList.toggle("dark", dark);
            }, { selected: theme, dark: mode === "dark" });
            await page.waitForTimeout(350);
            assert.deepEqual(await page.evaluate(() => ({
              client: document.documentElement.clientWidth,
              scroll: document.documentElement.scrollWidth,
            })), { client: width, scroll: width });

            if (route.includes("candlestick")) {
              const result = await page.evaluate(() => {
				const colorCanvas = document.createElement("canvas");
				colorCanvas.width = 1;
				colorCanvas.height = 1;
				const colorContext = colorCanvas.getContext("2d", { willReadFrequently: true });
				const rgb = (value) => {
				  colorContext.clearRect(0, 0, 1, 1);
				  colorContext.fillStyle = value;
				  colorContext.fillRect(0, 0, 1, 1);
				  return [...colorContext.getImageData(0, 0, 1, 1).data.slice(0, 3)];
				};
                const luminance = (color) => rgb(color).map((channel) => {
                  const value = channel / 255;
                  return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
                }).reduce((sum, value, index) => sum + value * [0.2126, 0.7152, 0.0722][index], 0);
                const contrast = (foreground, background) => {
                  const values = [luminance(foreground), luminance(background)].sort((a, b) => b - a);
                  return (values[0] + 0.05) / (values[1] + 0.05);
                };
                const surface = getComputedStyle(document.body).backgroundColor;
                const increasing = getComputedStyle(document.querySelector(".goshtoso-charts-candlestick__direction--increasing")).color;
                const decreasing = getComputedStyle(document.querySelector(".goshtoso-charts-candlestick__direction--decreasing")).color;
                const viewport = document.querySelector(".goshtoso-charts-candlestick__viewport");
                return {
                  increasingContrast: contrast(increasing, surface),
                  decreasingContrast: contrast(decreasing, surface),
                  localOverflow: getComputedStyle(viewport).overflowX,
                  locallyContained: viewport.scrollWidth >= viewport.clientWidth,
                };
              });
              assert.ok(result.increasingContrast >= 3, `increasing contrast ${result.increasingContrast}`);
              assert.ok(result.decreasingContrast >= 3, `decreasing contrast ${result.decreasingContrast}`);
              assert.equal(result.localOverflow, "auto");
              assert.equal(result.locallyContained, true);
			} else if (route === "/components/heatmap") {
			  const result = await page.locator("[data-goshtoso-chart-wrapper]").first().evaluate((element) => {
				const cells = [...element.querySelectorAll("figure svg path[class]")];
				const viewport = element.querySelector(".goshtoso-charts-heatmap__viewport");
				return {
				  cells: cells.length,
				  low: getComputedStyle(cells[15]).fill,
				  high: getComputedStyle(cells[7]).fill,
				  localOverflow: getComputedStyle(viewport).overflowX,
				  locallyContained: viewport.scrollWidth >= viewport.clientWidth,
				};
			  });
			  assert.equal(result.cells, 25);
			  assert.notEqual(result.low, result.high);
			  assert.equal(result.localOverflow, "auto");
			  assert.equal(result.locallyContained, true);
			} else {
			  const geometry = await page.locator("[data-goshtoso-chart-wrapper]").first().evaluate((element) => {
                const host = element.querySelector("[_echarts_instance_]");
                const instance = window.echarts.getInstanceByDom(host);
                return {
                  host: host.clientWidth,
                  chart: instance.getWidth(),
                  canvas: Math.round(host.querySelector("canvas").getBoundingClientRect().width),
				  colors: instance.getOption().series.filter((series) => series.type === "parallel").map((series) => series.lineStyle.color),
                };
			  });
			  assert.deepEqual({ chart: geometry.chart, canvas: geometry.canvas }, { chart: geometry.host, canvas: geometry.host });
			  if (route.includes("parallel")) assert.equal(new Set(geometry.colors).size, 3);
            }
            assert.deepEqual(errors, []);
          } finally {
            await page.close();
          }
        }
      });
    }
  }
}
