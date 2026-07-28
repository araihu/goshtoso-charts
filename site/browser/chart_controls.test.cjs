const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

const baseURL = process.env.BASE_URL || "http://127.0.0.1:8096";
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
    server = spawn("go", ["run", "./cmd/server", "-port", "8096"], {
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
		const menu = page.getByRole("menu");
		if (!(await menu.isVisible())) await trigger.click();
		await menu.getByRole("menuitem", { name: match[2], exact: true }).first().click();
  }
  const artifact = await pending;
	const artifactPath = await artifact.path();
	assert.ok(artifactPath);
	const openMenu = page.getByRole("menu");
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

test("collapse preserves static DOM and interactive instance, then resizes on expand", async () => {
  for (const route of ["/components/line", "/components/scatter", "/components/interactive/bar", "/components/interactive/tree"]) {
    const page = await pageAt(route);
    try {
      const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
      await wrapper.evaluate((element) => {
        element.__contentIdentity = element.querySelector("[data-goshtoso-chart-content]").firstElementChild;
        const host = element.querySelector("[_echarts_instance_]");
        element.__instanceIdentity = host && window.echarts.getInstanceByDom(host);
        element.__resizeEvents = 0;
        element.addEventListener("goshtoso-charts:resize", () => { element.__resizeEvents += 1; });
      });
      const collapse = wrapper.locator('[data-goshtoso-chart-control="collapse"]');
      await collapse.press("Enter");
      assert.equal(await collapse.getAttribute("aria-expanded"), "false");
      assert.equal(await wrapper.locator("[data-goshtoso-chart-content]").getAttribute("hidden"), "");
      await collapse.press(" ");
      assert.equal(await collapse.getAttribute("aria-expanded"), "true");
      await page.waitForTimeout(100);
      const result = await wrapper.evaluate((element) => ({
        sameContent: element.__contentIdentity === element.querySelector("[data-goshtoso-chart-content]").firstElementChild,
        sameInstance: !element.__instanceIdentity || element.__instanceIdentity === window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")),
        resizeEvents: element.__resizeEvents,
      }));
      assert.equal(result.sameContent, true);
      assert.equal(result.sameInstance, true);
      assert.ok(result.resizeEvents >= 2, `resize events = ${result.resizeEvents}`);
    } finally {
      await page.close();
    }
  }
});

test("tree expansion state and instance survive collapse and fullscreen", async () => {
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

    const collapse = wrapper.locator('[data-goshtoso-chart-control="collapse"]');
    await collapse.click();
    await collapse.click();
    await page.waitForTimeout(100);

    const fullscreen = wrapper.locator('[data-goshtoso-chart-control="fullscreen"]');
    await fullscreen.click();
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

test("shared ResizeObserver converges a 847px interactive canvas into a 607px flex host without re-init", async () => {
  const page = await pageAt("/components/interactive/tree");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      const data = instance.getModel().getSeriesByIndex(0).getData();
      let nodeIndex = -1;
      for (let index = 0; index < data.count(); index += 1) if (data.getName(index) === "Node3") nodeIndex = index;
      instance.dispatchAction({ type: "treeExpandAndCollapse", seriesIndex: 0, dataIndex: nodeIndex });
      element.__flexInstance = instance;
      element.__flexNodeIndex = nodeIndex;
      host.style.width = "847px";
      instance.resize();
    });
    await page.waitForFunction(() => document.querySelector("[_echarts_instance_]").clientWidth === 847);
    await wrapper.evaluate((element) => { element.querySelector("[_echarts_instance_]").style.width = "607px"; });
    await page.waitForFunction(() => {
      const host = document.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth === 607 && instance.getWidth() === 607 && Math.round(host.querySelector("canvas").getBoundingClientRect().width) === 607;
    });
    assert.deepEqual(await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      const node = instance.getModel().getSeriesByIndex(0).getData().tree.getNodeByDataIndex(element.__flexNodeIndex);
      return { sameInstance: instance === element.__flexInstance, expanded: node.isExpand, host: host.clientWidth, chart: instance.getWidth(), canvas: Math.round(host.querySelector("canvas").getBoundingClientRect().width) };
    }), { sameInstance: true, expanded: true, host: 607, chart: 607, canvas: 607 });
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

test("Expand uses Goshtoso Modal, scales static Scatter, preserves collapse, and restores focus", async () => {
  const page = await pageAt("/components/scatter");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    const collapse = wrapper.locator('[data-goshtoso-chart-control="collapse"]');
    const trigger = wrapper.locator("[data-goshtoso-chart-expand] > div > button").first();
    await wrapper.evaluate((element) => {
      element.__expandedContent = element.querySelector("[data-goshtoso-chart-content]");
    });
    await collapse.click();
    assert.equal(await collapse.getAttribute("aria-expanded"), "false");
    await trigger.click();
		const dialog = wrapper.getByRole("dialog", { name: "Dense scatter data" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForTimeout(350);
    assert.equal(await collapse.isHidden(), true);
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
      collapsed: element.querySelector("[data-goshtoso-chart-content]").hidden,
      collapseVisible: !element.querySelector('[data-goshtoso-chart-control="collapse"]').hidden,
      focusReturned: document.activeElement === element.querySelector("[data-goshtoso-chart-expand] > div > button"),
    })), { sameContent: true, collapsed: true, collapseVisible: true, focusReturned: true });
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
    const trigger = wrapper.locator("[data-goshtoso-chart-expand] > div > button").first();
    await trigger.click();
    const dialog = wrapper.getByRole("dialog", { name: "Basic tree example" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForTimeout(350);
    assert.equal(await wrapper.locator('[data-goshtoso-chart-control="collapse"]').isHidden(), true);
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
      collapseVisible: !element.querySelector('[data-goshtoso-chart-control="collapse"]').hidden,
      focusReturned: document.activeElement === element.querySelector("[data-goshtoso-chart-expand] > div > button"),
    })), { sameInstance: true, collapseVisible: true, focusReturned: true });
  } finally {
    await page.close();
  }
});

test("native fullscreen enters and exits, preserves instance, resizes, and returns focus", async () => {
  const page = await pageAt("/components/interactive/line");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    const button = wrapper.locator('[data-goshtoso-chart-control="fullscreen"]');
    const initial = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      element.__instanceIdentity = instance;
      element.__resizeEvents = 0;
      element.addEventListener("goshtoso-charts:resize", () => { element.__resizeEvents += 1; });
      return { width: instance.getWidth(), height: instance.getHeight() };
    });
    await button.click();
    await page.waitForFunction(() => Boolean(document.fullscreenElement));
    await page.waitForTimeout(350);
    assert.equal(await button.getAttribute("aria-pressed"), "true");
    assert.equal(await wrapper.locator('[data-goshtoso-chart-control="collapse"]').isHidden(), true);
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
      focusReturned: document.activeElement === element.querySelector('[data-goshtoso-chart-control="fullscreen"]'),
      collapseVisible: !element.querySelector('[data-goshtoso-chart-control="collapse"]').hidden,
    }));
    assert.equal(restored.sameInstance, true);
    assert.equal(restored.focusReturned, true);
    assert.equal(restored.collapseVisible, true);
    assert.ok(restored.resizeEvents >= 4, `resize events = ${restored.resizeEvents}`);
  } finally {
    await page.close();
  }
});

test("Escape exits the safe fullscreen fallback and returns focus", async () => {
  const page = await pageAt("/components/interactive/bar");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    const button = wrapper.locator('[data-goshtoso-chart-control="fullscreen"]');
    const initial = await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      element.__fallbackInstance = instance;
      return { width: instance.getWidth(), height: instance.getHeight() };
    });
    await wrapper.evaluate((element) => { element.requestFullscreen = undefined; });
    await button.click();
    await page.waitForTimeout(350);
    assert.equal(await wrapper.evaluate((element) => element.classList.contains("goshtoso-charts-fullscreen-fallback")), true);
    assert.equal(await button.getAttribute("aria-pressed"), "true");
    assert.equal(await wrapper.locator('[data-goshtoso-chart-control="collapse"]').isHidden(), true);
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
      focusReturned: document.activeElement === element.querySelector('[data-goshtoso-chart-control="fullscreen"]'),
      sameInstance: element.__fallbackInstance === window.echarts.getInstanceByDom(element.querySelector("[_echarts_instance_]")),
      collapseVisible: !element.querySelector('[data-goshtoso-chart-control="collapse"]').hidden,
    })), { active: false, focusReturned: true, sameInstance: true, collapseVisible: true });
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
    await wrapper.locator("[data-goshtoso-chart-expand] > div > button").first().click();
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

    await wrapper.locator("[data-goshtoso-chart-expand] > div > button").first().click();
    const dialog = wrapper.getByRole("dialog", { name: "Basic sunburst example" });
    await dialog.waitFor({ state: "visible" });
    await page.evaluate(() => document.documentElement.classList.add("dark"));
    await page.waitForTimeout(400);
    state = await measure();
    assert.equal(state.sameInstance, true);
    assert.equal(state.viewRoot, initial.target);
    assert.equal(state.canvasWidth, state.hostWidth);
    assert.equal(await wrapper.getByRole("button", { name: /^Collapse / }).isHidden(), true);
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    const fullscreen = wrapper.getByRole("button", { name: /^Enter fullscreen / });
    await fullscreen.click();
    await page.waitForFunction(() => document.fullscreenElement !== null);
    await page.waitForTimeout(350);
    state = await measure();
    assert.equal(state.sameInstance, true);
    assert.equal(state.viewRoot, initial.target);
    assert.equal(state.canvasWidth, state.hostWidth);
    assert.equal(await wrapper.getByRole("button", { name: /^Collapse / }).isHidden(), true);
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

test("Treemap focus and native breadcrumb back survive collapse, modal, theme, resize, and fullscreen on one instance", async () => {
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

	const collapse = wrapper.locator('[data-goshtoso-chart-control="collapse"]');
	await collapse.click();
	await collapse.click();
	await page.waitForTimeout(350);
	let state = await measure();
	assert.equal(state.sameInstance, true);
	assert.equal(state.viewRoot, "d1");
	assert.deepEqual({ chart: state.chartWidth, canvas: state.canvasWidth }, { chart: state.hostWidth, canvas: state.hostWidth });

    await wrapper.locator("[data-goshtoso-chart-expand] > div > button").first().click();
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

    await wrapper.getByRole("button", { name: /^Enter fullscreen / }).click();
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

test("static Radar controls preserve DOM and export resolved 600x400 SVG and opaque PNG", async () => {
  const page = await pageAt("/components/radar");
  try {
    const wrapper = page.locator("[data-goshtoso-chart-wrapper]").first();
    await wrapper.evaluate((element) => { element.__radarContent = element.querySelector("[data-goshtoso-chart-content]"); });
    const collapse = wrapper.getByRole("button", { name: /^Collapse / });
    await collapse.click();
    assert.equal(await wrapper.locator("[data-goshtoso-chart-content]").getAttribute("hidden"), "");
    await wrapper.getByRole("button", { name: /^Expand Basic radar chart$/ }).click();
    assert.equal(await wrapper.evaluate((element) => element.__radarContent === element.querySelector("[data-goshtoso-chart-content]")), true);

    await wrapper.locator("[data-goshtoso-chart-expand] > div > button").first().click();
    const dialog = wrapper.getByRole("dialog", { name: "Basic radar chart" });
    await dialog.waitFor({ state: "visible" });
    assert.equal(await collapse.isHidden(), true);
    assert.equal(await dialog.locator("[data-goshtoso-chart-content]").count(), 1);
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    const fullscreen = wrapper.getByRole("button", { name: /^Enter fullscreen / });
    await fullscreen.click();
    await page.waitForFunction(() => document.fullscreenElement !== null);
    assert.equal(await collapse.isHidden(), true);
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
          assert.equal(await wrapper.getByRole("button").count(), 4);
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
          const collapse = wrapper.getByRole("button", { name: /^Collapse / });
          await collapse.focus();
          assert.equal(await collapse.evaluate((element) => element === document.activeElement), true);
          if (screenshotDirectory) {
            await page.screenshot({
              path: path.join(screenshotDirectory, `chart-controls-${width}-${theme}-${mode}.png`),
              fullPage: true,
            });
          }
          const expand = wrapper.locator("[data-goshtoso-chart-expand] > div > button").first();
          await expand.click();
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
      test(`${width}px ${theme} ${mode} keeps Candlestick and Treemap contained and theme-legible`, async () => {
        for (const route of ["/components/candlestick", "/components/interactive/treemap"]) {
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
            } else {
			  const geometry = await page.locator("[data-goshtoso-chart-wrapper]").first().evaluate((element) => {
                const host = element.querySelector("[_echarts_instance_]");
                const instance = window.echarts.getInstanceByDom(host);
                return {
                  host: host.clientWidth,
                  chart: instance.getWidth(),
                  canvas: Math.round(host.querySelector("canvas").getBoundingClientRect().width),
                };
			  });
			  assert.deepEqual({ chart: geometry.chart, canvas: geometry.canvas }, { chart: geometry.host, canvas: geometry.host });
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
