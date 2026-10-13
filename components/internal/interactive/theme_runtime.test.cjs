const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const test = require('node:test');

const source = fs.readFileSync(path.join(__dirname, 'theme_runtime.go'), 'utf8');
const script = source.match(/const ThemeRuntimeMarkup = `<script[^>]*>([\s\S]*?)<\/script>`/)[1];

test('detached chart subtrees dispose once; synchronous modal moves preserve instances', () => {
  const observers = [];
  const observed = new Set();
  const instances = new Map();
  const frames = [];
  const context = {
    window: {
      addEventListener() {},
      echarts: { getInstanceByDom: host => instances.get(host) },
    },
    document: {
      documentElement: {},
      currentScript: { closest: () => null },
      addEventListener() {},
      createElement: () => ({ style: {}, getContext: () => null, remove() {} }),
    },
    getComputedStyle: () => ({ color: '#123456' }),
    requestAnimationFrame: fn => frames.push(fn),
    ResizeObserver: class {
      observe(target) { observed.add(target); }
      unobserve(target) { observed.delete(target); }
    },
    MutationObserver: class {
      constructor(callback) { this.callback = callback; observers.push(this); }
      observe(target, options) { this.options = options; }
    },
  };
  context.window.ResizeObserver = context.ResizeObserver;
  vm.runInNewContext(script, context);
  const runtime = context.window.__goshtosoChartsThemeRuntime;
  const removedSubtree = observers.find(observer => observer.options.childList);
  const makeFigure = () => {
    const figure = {
      isConnected: true,
      getAttribute: () => '',
      appendChild() {},
      querySelector: () => host,
    };
    const host = { parentElement: figure, getBoundingClientRect: () => ({ width: 600, height: 400 }) };
    const chart = {
      disposals: 0, themes: 0,
      getOption: () => ({}),
      getModel: () => ({ eachSeriesByType() {} }),
      getWidth: () => 600, getHeight: () => 400,
      isDisposed() { return this.disposals > 0; },
      setOption() { this.themes++; },
      dispose() { this.disposals++; instances.delete(host); },
    };
    instances.set(host, chart);
    runtime.register(figure);
    return { figure, host, chart };
  };

  const moving = makeFigure();
  assert.equal(moving.chart.themes, 1);
  moving.figure.isConnected = false;
  moving.figure.isConnected = true;
  removedSubtree.callback();
  assert.equal(moving.chart.disposals, 0, 'moving a chart into its modal must preserve it');
  assert(observed.has(moving.host));

  const nested = makeFigure();
  moving.figure.isConnected = false;
  nested.figure.isConnected = false;
  removedSubtree.callback();
  removedSubtree.callback();
  assert.equal(moving.chart.disposals, 1);
  assert.equal(nested.chart.disposals, 1);
  assert.equal(observed.size, 0, 'detached subtree resize targets must be released');

  for (let i = 0; i < 30; i++) {
    const replacement = makeFigure();
    replacement.figure.isConnected = false;
    removedSubtree.callback();
    assert.equal(replacement.chart.disposals, 1);
    assert.equal(instances.size, 0, 'repeated HTML snapshots must not retain instances');
  }
  while (frames.length) frames.shift()();
  assert.equal(observed.size, 0);
});

test('bundled renderer retains line gaps and excludes them from averages', () => {
  const echarts = require('../../../assets/js/runtime/echarts/5.6.0/echarts.min.js');
  const chart = echarts.init(null, null, { renderer: 'svg', ssr: true, width: 600, height: 400 });
  try {
    chart.setOption({
      animation: false,
      xAxis: { type: 'category', data: ['A', 'B', 'C', 'D', 'E'] }, yAxis: {},
      series: [{ type: 'line', data: [{ value: null }, { value: 0 }, { value: null }, { value: 42 }, { value: null }], markLine: { data: [{ type: 'average' }] } }],
    });
    const series = chart.getModel().getSeriesByIndex(0);
    const data = series.getData();
    assert(Number.isNaN(data.get('y', 0)));
    assert.equal(data.get('y', 1), 0);
    assert(Number.isNaN(data.get('y', 2)));
    assert.equal(data.get('y', 3), 42);
    assert.equal(series.get('connectNulls'), false);
    const svg = chart.renderToSVGString();
    assert(svg.includes('>21</text>'), 'average should be 21 excluding nulls');
    assert(!svg.includes('NaN'), 'SVG must not contain invalid coordinates');
  } finally {
    chart.dispose();
  }
});
