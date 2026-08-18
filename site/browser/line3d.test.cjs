const assert = require("node:assert/strict");
const { after, before, test } = require("node:test");
const fs = require("node:fs/promises");
const net = require("node:net");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { chromium } = require("playwright");
const sharp = require("sharp");

let baseURL = process.env.BASE_URL;
let browser;
let server;

async function freePort() {
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
  for (let attempt = 0; attempt < 200; attempt += 1) {
    try {
      if ((await fetch(`${baseURL}/components/interactive/line-3d`)).ok) return;
    } catch {
      // Test-owned random-port server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Line3D verification server did not start at ${baseURL}`);
}

before(async () => {
  if (!baseURL) {
    const port = await freePort();
    baseURL = `http://127.0.0.1:${port}`;
    server = spawn("go", ["run", "./cmd/server", "-port", String(port)], {
      cwd: path.resolve(__dirname, ".."), detached: true, stdio: "pipe",
    });
  }
  await ready();
  browser = await chromium.launch({ headless: true });
});

after(async () => {
  await browser?.close();
  if (server?.pid) {
    try { process.kill(-server.pid, "SIGTERM"); } catch { /* already stopped */ }
  }
});

async function pageAt(viewport, reducedMotion = "no-preference") {
  const page = await browser.newPage({ viewport, acceptDownloads: true });
  await page.emulateMedia({ reducedMotion });
  await page.addInitScript(() => {
    const createObjectURL = URL.createObjectURL.bind(URL);
    globalThis.__chartBlobTypes = [];
    URL.createObjectURL = (blob) => {
      globalThis.__chartBlobTypes.push(blob.type);
      return createObjectURL(blob);
    };
  });
  await page.goto(`${baseURL}/components/interactive/line-3d`, { waitUntil: "domcontentloaded" });
  await page.locator("[data-line3d-variant]").first().waitFor();
  await page.waitForFunction(() => {
    const hosts = [...document.querySelectorAll("[data-line3d-variant] [_echarts_instance_]")];
    return hosts.length === 2 && hosts.every((host) => Boolean(window.echarts.getInstanceByDom(host)));
  });
  await page.waitForFunction(() => Boolean(document.documentElement._x_dataStack));
  return page;
}

function wrapperFor(page, variant = "base") {
  return page.locator(`[data-line3d-variant="${variant}"] [data-goshtoso-chart-wrapper]`).first();
}

async function openExpand(wrapper) {
  const trigger = wrapper.locator('[id$="-primary-action"]:visible').first();
  await trigger.waitFor({ state: "visible" });
  await trigger.click();
  return trigger;
}

async function measure(wrapper) {
  return wrapper.evaluate(async (element) => {
    const host = element.querySelector("[_echarts_instance_]");
    const instance = window.echarts.getInstanceByDom(host);
    const option = instance.getOption();
    const values = option.series[0].data.map((point) => Array.isArray(point) ? point.slice(0, 3) : point.value.slice(0, 3));
    const bytes = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(JSON.stringify(values)));
    const summary = element.parentElement.querySelector("[data-line3d-exact-data]");
    return {
      sameInstance: !element.__line3DInstance || instance === element.__line3DInstance,
      hostWidth: host.clientWidth,
      hostHeight: host.clientHeight,
      chartWidth: instance.getWidth(),
      chartHeight: instance.getHeight(),
      canvasCount: host.querySelectorAll("canvas").length,
      type: option.series[0].type,
      name: option.series[0].name,
      count: values.length,
      hash: [...new Uint8Array(bytes)].map((value) => value.toString(16).padStart(2, "0")).join(""),
      values,
      range: [option.visualMap[0].min, option.visualMap[0].max],
      calculable: option.visualMap[0].calculable,
      scaleColors: option.visualMap[0].inRange.color,
      autoRotate: option.grid3D[0].viewControl.autoRotate,
      background: option.backgroundColor,
      axisText: option.xAxis3D[0].axisLabel.color,
      grid: option.xAxis3D[0].splitLine.lineStyle.color,
      formula: summary.querySelector("[data-line3d-formula]").textContent,
      summary: summary.querySelector("[data-line3d-summary]").textContent.replace(/\s+/g, " ").trim(),
      motion: summary.querySelector("[data-line3d-motion]").textContent.replace(/\s+/g, " ").trim(),
      tableRows: summary.querySelectorAll("tr").length,
    };
  });
}

test("both treatments preserve exact plotted order, CSV, local assets, route, and search", async () => {
  const page = await pageAt({ width: 1440, height: 900 });
  try {
    const resources = await page.evaluate(() => [...document.scripts].map((script) => script.src).filter(Boolean));
    const names = ["echarts/5.6.0", "word-cloud/2.1.0", "liquid/3.1.0", "three-d/2.0.9", "maps/ibge-mmd-2025"];
    const indexes = names.map((name) => resources.findIndex((url) => url.includes(name)));
    assert.ok(indexes.every((index) => index >= 0), JSON.stringify({ resources, indexes }));
    assert.ok(indexes.every((value, index) => index === 0 || indexes[index - 1] < value), JSON.stringify(indexes));

    const base = await measure(wrapperFor(page));
    assert.equal(base.type, "line3D");
    assert.equal(base.name, "line3D");
    assert.ok(base.canvasCount >= 1);
    assert.equal(base.count, 25000);
    assert.equal(base.hash, "63f356cee0db8603edec10d54e8aec4f5eba291e8c99b961119c9543fd6c63a4");
    assert.deepEqual(base.values[0], [1.25, 0, 0]);
    assert.deepEqual(base.range, [0, 30]);
    assert.equal(base.calculable, true);
    assert.equal(base.scaleColors.length, 10);
    assert.equal(base.autoRotate, false);
    assert.match(base.formula, /i \/ 1000.*cos\(75.*sin\(75/);
    assert.match(base.summary, /25000 ordered points.*t domain \[0, 24\.999\].*X domain \[-1\.2489045114273476, 1\.25\].*Y domain \[-1\.2497258288146875, 1\.2497201051428937\].*Z domain \[-1\.9368409642920035, 26\.985899643193864\]/);
    assert.match(base.motion, /remains stationary/);
    assert.equal(base.tableRows, 0);

    const rotating = await measure(wrapperFor(page, "auto-rotate"));
    assert.equal(rotating.type, "line3D");
    assert.equal(rotating.count, 25000);
    assert.equal(rotating.hash, base.hash);
    assert.deepEqual(rotating.values, base.values);
    assert.deepEqual(rotating.range, base.range);
    assert.deepEqual(rotating.scaleColors, base.scaleColors);
    assert.equal(rotating.autoRotate, true);
    assert.match(rotating.motion, /rotates automatically.*initial drawing may animate.*Reduced-motion/);
    assert.equal(rotating.tableRows, 0);

    const expectedByVariant = { base, "auto-rotate": rotating };
    for (const variant of ["base", "auto-rotate"]) {
      const wrapper = wrapperFor(page, variant);
      const pending = page.waitForEvent("download");
      await wrapper.locator("xpath=..").getByRole("link", { name: "Download all exact points as CSV" }).click();
      const artifact = await pending;
      const csv = await fs.readFile(await artifact.path(), "utf8");
      const lines = csv.trimEnd().split("\n");
      assert.equal(lines[0], "series,index,x,y,z");
      assert.equal(lines.length - 1, 25000);
      const csvValues = lines.slice(1).map((line, index) => {
        const columns = line.slice('"line3D",'.length).split(",");
        assert.equal(Number(columns[0]), index);
        return columns.slice(1).map(Number);
      });
      assert.deepEqual(csvValues, expectedByVariant[variant].values);
    }

    const search = await page.request.get(`${baseURL}/components/line`);
    assert.match(await search.text(), /data-search="line 3d interactive \/ 3d interactive-line-3d"/);
    for (const route of ["/components/interactive/scatter-3d", "/components/interactive/bar-3d", "/components/interactive/surface-3d"]) {
      assert.equal((await page.request.get(`${baseURL}${route}`)).status(), 200);
    }
    for (const asset of ["/charts/assets/js/controls/6/controls.js", "/charts/assets/js/runtime/three-d/2.0.9/runtime.min.js", "/charts/assets/js/runtime/echarts/5.6.0/echarts.min.js"]) {
      assert.equal((await page.request.get(`${baseURL}${asset}`)).status(), 200);
    }
  } finally {
    await page.close();
  }
});

test("390 and 1440 widths retain light/dark theme contrast and responsive convergence", async () => {
  const themed = new Set();
  for (const width of [390, 1440]) {
    for (const dark of [false, true]) {
      const page = await pageAt({ width, height: 900 });
      try {
        await page.evaluate((darkMode) => document.documentElement.classList.toggle("dark", darkMode), dark);
        await page.waitForTimeout(350);
        const wrapper = wrapperFor(page);
        const state = await measure(wrapper);
        assert.deepEqual({ chartWidth: state.chartWidth, chartHeight: state.chartHeight }, { chartWidth: state.hostWidth, chartHeight: state.hostHeight });
        assert.deepEqual(await page.evaluate(() => ({ client: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth })), { client: width, scroll: width });
        assert.equal(await wrapper.evaluate((element) => {
          const host = element.querySelector("[_echarts_instance_]").getBoundingClientRect();
          const content = element.querySelector("[data-goshtoso-chart-content]").getBoundingClientRect();
          return Math.abs((host.left + host.right) / 2 - (content.left + content.right) / 2) < 2;
        }), true);
        const contrast = await page.evaluate(({ background, foreground }) => {
          const parse = (value) => (value.match(/[\d.]+/g) || []).slice(0, 3).map(Number);
          const luminance = (value) => {
            const rgb = parse(value).map((channel) => channel / 255).map((channel) => channel <= 0.04045 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4);
            return 0.2126 * rgb[0] + 0.7152 * rgb[1] + 0.0722 * rgb[2];
          };
          const first = luminance(background);
          const second = luminance(foreground);
          return (Math.max(first, second) + .05) / (Math.min(first, second) + .05);
        }, { background: state.background, foreground: state.axisText });
        assert.ok(contrast >= 4.5, `contrast ${contrast}: ${JSON.stringify(state)}`);
        assert.notEqual(state.grid, state.background);
        themed.add(`${state.background}|${state.axisText}|${state.scaleColors.join("|")}`);
      } finally {
        await page.close();
      }
    }
  }
  assert.equal(themed.size, 2);
});

test("auto-rotation obeys reduced motion and remains stable", async () => {
  const page = await pageAt({ width: 1440, height: 900 }, "reduce");
  try {
    const rotating = await measure(wrapperFor(page, "auto-rotate"));
    assert.equal(rotating.autoRotate, false);
    assert.match(rotating.motion, /Reduced-motion preference disables animation and keeps the same chart stationary/);
    await page.emulateMedia({ reducedMotion: "no-preference" });
    await page.waitForFunction(() => {
      const wrapper = document.querySelector('[data-line3d-variant="auto-rotate"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      return window.echarts.getInstanceByDom(host).getOption().grid3D[0].viewControl.autoRotate === true;
    });
    assert.equal((await measure(wrapperFor(page, "auto-rotate"))).autoRotate, true);
  } finally {
    await page.close();
  }
});

test("large centered modal preserves instance and exports opaque PNG", async () => {
  const page = await pageAt({ width: 1440, height: 900 });
  const errors = [];
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  page.on("pageerror", (error) => errors.push(error.message));
  try {
    const wrapper = wrapperFor(page);
    await wrapper.evaluate((element) => {
      const host = element.querySelector("[_echarts_instance_]");
      element.__line3DInstance = window.echarts.getInstanceByDom(host);
    });
    const initial = await measure(wrapper);
    await wrapper.evaluate((element) => {
      element.querySelector(".goshtoso-charts-interactive > .container").style.width = "60%";
    });
    await page.waitForFunction((initialWidth) => {
      const wrapper = document.querySelector('[data-line3d-variant="base"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth < initialWidth && instance.getWidth() === host.clientWidth;
    }, initial.hostWidth);
    assert.equal((await measure(wrapper)).sameInstance, true);
    await wrapper.evaluate((element) => {
      element.querySelector(".goshtoso-charts-interactive > .container").style.width = "";
    });
    await page.waitForFunction((initialWidth) => {
      const wrapper = document.querySelector('[data-line3d-variant="base"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth === initialWidth && instance.getWidth() === host.clientWidth;
    }, initial.hostWidth);
    await openExpand(wrapper);
    const dialog = wrapper.getByRole("dialog", { name: "basic line3d example" });
    await dialog.waitFor({ state: "visible" });
    await page.waitForFunction((initialWidth) => {
      const wrapper = document.querySelector('[data-line3d-variant="base"] [data-goshtoso-chart-wrapper]');
      const host = wrapper.querySelector("[_echarts_instance_]");
      const instance = window.echarts.getInstanceByDom(host);
      return host.clientWidth > initialWidth && instance.getWidth() === host.clientWidth;
    }, initial.hostWidth);
    assert.equal((await measure(wrapper)).sameInstance, true);
    const panel = await dialog.locator(".goshtoso-charts-expand-panel").evaluate((element) => {
      const rect = element.getBoundingClientRect();
      return {
        centered: Math.abs((rect.left + rect.right) / 2 - innerWidth / 2) < 4,
        widthRatio: rect.width / innerWidth,
        heightRatio: rect.height / innerHeight,
        contained: rect.left >= 0 && rect.right <= innerWidth + 1 && rect.top >= 0 && rect.bottom <= innerHeight + 1,
      };
    });
    assert.equal(panel.centered, true);
    assert.equal(panel.contained, true);
    assert.ok(panel.widthRatio >= .75 && panel.heightRatio >= .55, JSON.stringify(panel));
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });

    const pending = page.waitForEvent("download");
    await wrapper.getByRole("button", { name: "Export basic line3d example" }).click();
    await wrapper.locator('[id$="-export-png-action"]:visible').first().click();
    const artifact = await pending;
    const bytes = await fs.readFile(await artifact.path());
    assert.equal(artifact.suggestedFilename(), "basic-line3d-example.png");
    assert.equal(await page.evaluate(() => globalThis.__chartBlobTypes.at(-1)), "image/png");
    assert.deepEqual([...bytes.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    const image = sharp(bytes).ensureAlpha().raw();
    const { data, info } = await image.toBuffer({ resolveWithObject: true });
    let minimumAlpha = 255;
    for (let index = 3; index < data.length; index += info.channels) minimumAlpha = Math.min(minimumAlpha, data[index]);
    assert.equal(minimumAlpha, 255);
    assert.equal((await measure(wrapper)).sameInstance, true);
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});

test("pinned CDN core and 3D extension use SRI and register Line3D in order", async () => {
  const page = await browser.newPage();
  const errors = [];
  page.on("pageerror", (error) => errors.push(error.message));
  try {
    await page.setContent(`<!doctype html><html><head>
      <script src="https://cdn.jsdelivr.net/npm/echarts@5.6.0/dist/echarts.min.js" integrity="sha384-pPi0zxBAoDu6+JXW/C68UZLvBUUtU+7zonhif43rqj7pxsGyqyqzcian2Rj37Rss" crossorigin="anonymous"></script>
      <script src="https://cdn.jsdelivr.net/npm/echarts-gl@2.0.9/dist/echarts-gl.min.js" integrity="sha384-f4gAUkb5Y6LE9n50CbiH1hCBCw7021OeJu0ZrgRpgW6G1CZjPR8cu33e8rCFLqCl" crossorigin="anonymous"></script>
    </head><body><div id="chart" style="width:400px;height:300px"></div>
    <script>const chart=echarts.init(document.getElementById("chart"));chart.setOption({xAxis3D:{},yAxis3D:{},zAxis3D:{},grid3D:{},series:[{type:"line3D",data:[[0,0,0],[1,0,1]]}]});</script>
    </body></html>`, { waitUntil: "networkidle" });
    assert.deepEqual(await page.evaluate(() => {
      const scripts = [...document.scripts].filter((item) => item.src);
      const instance = window.echarts.getInstanceByDom(document.getElementById("chart"));
      return {
        type: instance.getOption().series[0].type,
        canvases: document.querySelectorAll("#chart canvas").length,
        order: scripts.map((item) => item.src),
        integrity: scripts.map((item) => item.integrity),
      };
    }), {
      type: "line3D",
      canvases: 1,
      order: [
        "https://cdn.jsdelivr.net/npm/echarts@5.6.0/dist/echarts.min.js",
        "https://cdn.jsdelivr.net/npm/echarts-gl@2.0.9/dist/echarts-gl.min.js",
      ],
      integrity: [
        "sha384-pPi0zxBAoDu6+JXW/C68UZLvBUUtU+7zonhif43rqj7pxsGyqyqzcian2Rj37Rss",
        "sha384-f4gAUkb5Y6LE9n50CbiH1hCBCw7021OeJu0ZrgRpgW6G1CZjPR8cu33e8rCFLqCl",
      ],
    });
    assert.deepEqual(errors, []);
  } finally {
    await page.close();
  }
});
