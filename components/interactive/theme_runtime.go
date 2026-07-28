package interactive

import (
	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func explicitColors(style charttheme.Style) string {
	if len(style.Colors) > 0 {
		return "true"
	}
	return "false"
}

// themeRuntime bridges Goshtoso's computed semantic CSS tokens into the canvas
// renderer. The script is constant library-owned markup; no chart or
// application values are interpolated into it.
func themeRuntime() templ.Component { return templ.Raw(themeRuntimeMarkup) }

const themeRuntimeMarkup = `<script data-goshtoso-charts-theme-runtime>
(function () {
  "use strict";
  var key = "__goshtosoChartsThemeRuntime";
	var runtime = window[key];
	if (!runtime) {
		var figures = new Set();
		var observedHosts = new WeakSet();
    var scheduled = false;
    var colorCanvas = document.createElement("canvas");
    colorCanvas.width = 1;
    colorCanvas.height = 1;
    var colorContext = colorCanvas.getContext("2d", { willReadFrequently: true });
    var rendererColor = function (value, fallback) {
      if (!colorContext) return value || fallback;
      colorContext.clearRect(0, 0, 1, 1);
      colorContext.fillStyle = value || fallback;
      colorContext.fillRect(0, 0, 1, 1);
      var pixel = colorContext.getImageData(0, 0, 1, 1).data;
      return "rgba(" + pixel[0] + ", " + pixel[1] + ", " + pixel[2] + ", " + (pixel[3] / 255) + ")";
    };
		var cssColor = function (figure, name, fallback) {
      var probe = document.createElement("span");
      probe.hidden = true;
      probe.style.color = "var(" + name + ", " + fallback + ")";
      figure.appendChild(probe);
      var value = getComputedStyle(probe).color;
      probe.remove();
      return rendererColor(value, fallback);
		};
		var classColor = function (figure, className, fallback) {
			var probe = document.createElement("span");
			probe.hidden = true;
			probe.className = className;
			figure.appendChild(probe);
			var value = getComputedStyle(probe).color;
			probe.className = "";
			var inherited = getComputedStyle(probe).color;
			probe.remove();
			if (value === inherited) return rendererColor(fallback, fallback);
			return rendererColor(value, fallback);
		};
		var classColorOrFallback = function (figure, className, fallback) {
			var inherited = getComputedStyle(figure).color;
			var resolved = classColor(figure, className, fallback);
			return resolved === rendererColor(inherited, fallback) ? fallback : resolved;
		};
    var repeat = function (items, value) {
      return (items && items.length ? items : []).map(function () { return value; });
    };
    var resize = function (host) {
      if (!host.isConnected || !window.echarts) return;
      var chart = window.echarts.getInstanceByDom(host);
      if (!chart) return;
      chart.resize();
    };
    var pendingResizeHosts = new WeakSet();
    var scheduleResize = function (host) {
      if (!host || pendingResizeHosts.has(host)) return;
      pendingResizeHosts.add(host);
      requestAnimationFrame(function () {
        pendingResizeHosts.delete(host);
        resize(host);
      });
    };
    var resizeObserver = window.ResizeObserver ? new ResizeObserver(function (entries) {
      entries.forEach(function (entry) { scheduleResize(entry.target); });
    }) : null;
    var apply = function (figure) {
      if (!figure.isConnected || !window.echarts) return;
      var host = figure.querySelector("[_echarts_instance_]");
      if (!host) return;
			var chart = window.echarts.getInstanceByDom(host);
			if (!chart) return;
			if (!observedHosts.has(host) && resizeObserver) {
				observedHosts.add(host);
				resizeObserver.observe(host);
			}

      var surface = cssColor(figure, "--color-chart-surface", "#ffffff");
      var surfaceAlt = cssColor(figure, "--color-chart-surface-alt", surface);
      var outline = cssColor(figure, "--color-chart-outline", "#64748b");
      var grid = cssColor(figure, "--color-chart-grid", outline);
      var text = cssColor(figure, "--color-chart-text", "#1f2937");
      var strong = cssColor(figure, "--color-chart-text-strong", text);
      var muted = cssColor(figure, "--color-chart-text-muted", text);
      var scaleLow = cssColor(figure, "--color-chart-scale-low", surfaceAlt);
      var scaleMid = cssColor(figure, "--color-chart-scale-mid", "#3b82f6");
      var scaleHigh = cssColor(figure, "--color-chart-scale-high", "#1e3a8a");
      var seriesColors = [];
      for (var colorIndex = 1; colorIndex <= 8; colorIndex += 1) {
        seriesColors.push(cssColor(figure, "--color-chart-series-" + colorIndex, "#2563eb"));
      }
      var current = chart.getOption();
      var explicitColors = figure.getAttribute("data-goshtoso-charts-explicit-colors") === "true";
      var explicitAnimation = figure.getAttribute("data-goshtoso-charts-explicit-animation") || "default";
      var palette = explicitColors && current.color && current.color.length ? current.color : seriesColors;
			var themeSeriesItems = (figure.getAttribute("data-goshtoso-charts-theme-series-items") || "")
        .split(",").filter(Boolean).map(function (value) { return Number(value); });
			var managesSeriesItem = function (index) { return themeSeriesItems.indexOf(index) !== -1; };
			var gaugeScale = null;
			try { gaugeScale = JSON.parse(figure.getAttribute("data-goshtoso-charts-gauge-scale") || "null"); } catch (_) {}
			var candlestickStyles = [];
			try { candlestickStyles = JSON.parse(figure.getAttribute("data-goshtoso-charts-candlestick-styles") || "[]"); } catch (_) {}
			var liquidGauge = null;
			try { liquidGauge = JSON.parse(figure.getAttribute("data-goshtoso-charts-liquid") || "null"); } catch (_) {}
			var gaugeColors = gaugeScale && (gaugeScale.stops || []).map(function (stop) {
				if (stop.token === "low") return scaleLow;
				if (stop.token === "mid") return scaleMid;
				if (stop.token === "high") return scaleHigh;
				if (stop.class) return classColor(figure, stop.class, text);
				return rendererColor(stop.color, text);
			});
			if (gaugeScale && gaugeScale.reverse) gaugeColors.reverse();
      var axis = {
        axisLabel: { color: muted },
        axisLine: { lineStyle: { color: outline } },
        axisTick: { lineStyle: { color: outline } },
        splitLine: { lineStyle: { color: grid } },
        nameTextStyle: { color: text }
      };
      var radar = {
        axisName: { color: text },
        axisLine: { lineStyle: { color: outline } },
        splitLine: { lineStyle: { color: grid } },
        splitArea: { areaStyle: { color: [surface, surfaceAlt] } }
      };
			var themedSeries = (current.series || []).map(function (series, index) {
				// Tree disclosure, Sunburst view-root, and Treemap view-root/layout state live inside their series
				// models. Presentation-only series updates can rebuild those models.
				// Global palette and text/background tokens still theme all three. Never
				// re-supply partial Treemap levels while native navigation state is active.
				if (series.type === "tree" || series.type === "sunburst" || series.type === "treemap") return null;
        var themedItem = {
          id: series.id,
          name: series.name,
          animationDurationUpdate: 0,
          label: { color: text },
          endLabel: { color: text },
          emphasis: { label: { color: strong } },
          labelLine: { lineStyle: { color: outline } }
        };
		if (series.type === "parallel") {
			var seriesPathColor = series.className
				? classColor(figure, series.className, palette[index % palette.length])
				: (series.lineStyle && series.lineStyle.color
					? rendererColor(series.lineStyle.color, palette[index % palette.length])
					: palette[index % palette.length]);
			themedItem.lineStyle = { color: seriesPathColor };
			themedItem.data = (series.data || []).map(function (item) {
				if (!item || Array.isArray(item) || (!item.className && !(item.lineStyle && item.lineStyle.color))) return item;
				var itemColor = item.className
					? classColor(figure, item.className, seriesPathColor)
					: rendererColor(item.lineStyle.color, seriesPathColor);
				return Object.assign({}, item, { lineStyle: Object.assign({}, item.lineStyle || {}, { color: itemColor }) });
			});
		}
		if (series.type === "wordCloud") {
			themedItem.data = (series.data || []).map(function (item, itemIndex) {
				if (!item || typeof item !== "object") return item;
				var wordColor = item.sourceColor
					? rendererColor(item.sourceColor, palette[itemIndex % palette.length])
					: (item.className ? classColorOrFallback(figure, item.className, palette[itemIndex % palette.length]) : palette[itemIndex % palette.length]);
				return Object.assign({}, item, { textStyle: Object.assign({}, item.textStyle || {}, { color: wordColor }) });
			});
		}
		if (series.type === "liquidFill") {
			var liquidPaint = function (paint, fallback) {
				if (!paint) return fallback;
				if (paint.class) return classColorOrFallback(figure, paint.class, fallback);
				if (paint.color) return rendererColor(paint.color, fallback);
				return fallback;
			};
			var wavePaint = liquidGauge && liquidGauge.style;
			var waveColor = liquidPaint(wavePaint, palette[index % palette.length]);
			themedItem.color = (series.data || []).map(function (_, waveIndex) {
				return wavePaint && (wavePaint.color || wavePaint.class) ? waveColor : palette[waveIndex % palette.length];
			});
			if (wavePaint && wavePaint.opacity !== undefined) themedItem.itemStyle = { opacity: wavePaint.opacity };
			var liquidOutline = liquidGauge && liquidGauge.outline;
			themedItem.outline = { itemStyle: {
				borderColor: liquidPaint(liquidOutline, outline),
				borderWidth: liquidOutline && liquidOutline.width !== undefined ? liquidOutline.width : 8
			} };
			var liquidBackground = liquidGauge && liquidGauge.background;
			themedItem.backgroundStyle = {
				color: liquidPaint(liquidBackground, surfaceAlt),
				borderColor: liquidBackground && liquidBackground.borderClass
					? classColorOrFallback(figure, liquidBackground.borderClass, outline)
					: rendererColor(liquidBackground && liquidBackground.borderColor, outline),
				borderWidth: liquidBackground && liquidBackground.borderWidth !== undefined ? liquidBackground.borderWidth : 0
			};
			var liquidLabel = liquidGauge && liquidGauge.label;
			themedItem.label = { color: liquidPaint(liquidLabel, strong) };
		}
        if (series.type === "boxplot" && managesSeriesItem(index)) {
          themedItem.itemStyle = { color: palette[index % palette.length], borderColor: palette[index % palette.length] };
        }
			if (series.type === "candlestick") {
				var candleStyle = candlestickStyles[index] || {};
				var riseBase = candleStyle.riseColor
					? rendererColor(candleStyle.riseColor, "#15803d")
					: cssColor(figure, "--color-chart-increasing", "#15803d");
				var fallBase = candleStyle.fallColor
					? rendererColor(candleStyle.fallColor, "#b91c1c")
					: cssColor(figure, "--color-chart-decreasing", "#b91c1c");
				var rise = candleStyle.riseClass
					? classColor(figure, candleStyle.riseClass, riseBase)
					: riseBase;
				var fall = candleStyle.fallClass
					? classColor(figure, candleStyle.fallClass, fallBase)
					: fallBase;
				var riseBorder = candleStyle.riseBorderColor ? rendererColor(candleStyle.riseBorderColor, rise) : rise;
				var fallBorder = candleStyle.fallBorderColor ? rendererColor(candleStyle.fallBorderColor, fall) : fall;
				themedItem.itemStyle = {
					color: rise, color0: fall, borderColor: riseBorder, borderColor0: fallBorder,
					borderWidth: candleStyle.borderWidth || 0
				};
			}
			if (series.type === "gauge") {
				themedItem.axisLine = { lineStyle: { width: 20, color: gaugeScale ? gaugeScale.stops.map(function (stop, stopIndex) { return [stop.position, gaugeColors[stopIndex]]; }) : [[1, surfaceAlt]] } };
          themedItem.axisLabel = { color: muted };
          themedItem.axisTick = { lineStyle: { color: outline } };
          themedItem.splitLine = { lineStyle: { color: outline } };
          themedItem.detail = { color: strong };
          themedItem.title = { color: text };
          if (managesSeriesItem(index)) {
            themedItem.itemStyle = { color: palette[index % palette.length] };
            themedItem.progress = { itemStyle: { color: palette[index % palette.length] } };
            themedItem.pointer = { itemStyle: { color: palette[index % palette.length] } };
          }
        }
        return themedItem;
			}).filter(Boolean);
      var themedVisualMaps = (current.visualMap || []).map(function () {
        var visualMap = { textStyle: { color: text } };
        if (!explicitColors) visualMap.inRange = { color: [scaleLow, scaleMid, scaleHigh] };
        return visualMap;
      });
			var themed = {
        backgroundColor: surface,
        textStyle: { color: text },
        title: repeat(current.title, { textStyle: { color: strong }, subtextStyle: { color: muted } }),
        legend: repeat(current.legend, { textStyle: { color: text } }),
        xAxis: repeat(current.xAxis, axis),
        yAxis: repeat(current.yAxis, axis),
        singleAxis: repeat(current.singleAxis, axis),
        radiusAxis: repeat(current.radiusAxis, axis),
        angleAxis: repeat(current.angleAxis, axis),
		parallelAxis: repeat(current.parallelAxis, axis),
        radar: repeat(current.radar, radar),
        visualMap: themedVisualMaps,
        tooltip: repeat(current.tooltip, {
          backgroundColor: surfaceAlt,
          borderColor: outline,
          textStyle: { color: strong }
        }),
        series: themedSeries
			};
			var treeExpansion = [];
			chart.getModel().eachSeriesByType("tree", function (seriesModel) {
				var data = seriesModel.getData();
				for (var dataIndex = 0; dataIndex < data.count(); dataIndex += 1) {
					var node = data.tree.getNodeByDataIndex(dataIndex);
					if (node && node.children && node.children.length) treeExpansion.push({ seriesIndex: seriesModel.seriesIndex, dataIndex: dataIndex, expanded: node.isExpand });
				}
			});
			var sunburstViewRoots = [];
			chart.getModel().eachSeriesByType("sunburst", function (seriesModel) {
				var viewRoot = seriesModel.getViewRoot && seriesModel.getViewRoot();
				if (viewRoot) sunburstViewRoots.push({ seriesIndex: seriesModel.seriesIndex, dataIndex: viewRoot.dataIndex });
			});
			var treemapViewRoots = [];
			chart.getModel().eachSeriesByType("treemap", function (seriesModel) {
				var viewRoot = seriesModel.getViewRoot && seriesModel.getViewRoot();
				if (viewRoot) treemapViewRoots.push({ seriesIndex: seriesModel.seriesIndex, dataIndex: viewRoot.dataIndex });
			});
      if (!explicitColors) {
        themed.color = seriesColors;
      }
      var reduceMotion = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
      if (reduceMotion && explicitAnimation === "default") {
        themed.animation = false;
        themed.animationDuration = 0;
        themed.animationDurationUpdate = 0;
      }
			chart.setOption(themed, { notMerge: false, lazyUpdate: false, silent: true });
			treeExpansion.forEach(function (state) {
				var seriesModel = chart.getModel().getSeriesByIndex(state.seriesIndex);
				var node = seriesModel && seriesModel.getData().tree.getNodeByDataIndex(state.dataIndex);
				if (node && node.isExpand !== state.expanded) chart.dispatchAction({ type: "treeExpandAndCollapse", seriesIndex: state.seriesIndex, dataIndex: state.dataIndex });
			});
			sunburstViewRoots.forEach(function (state) {
				var seriesModel = chart.getModel().getSeriesByIndex(state.seriesIndex);
				var targetNode = seriesModel && seriesModel.getData().tree.getNodeByDataIndex(state.dataIndex);
				if (targetNode) chart.dispatchAction({ type: "sunburstRootToNode", seriesIndex: state.seriesIndex, targetNode: targetNode });
			});
			treemapViewRoots.forEach(function (state) {
				var seriesModel = chart.getModel().getSeriesByIndex(state.seriesIndex);
				var targetNode = seriesModel && seriesModel.getData().tree.getNodeByDataIndex(state.dataIndex);
				if (targetNode) chart.dispatchAction({ type: "treemapRootToNode", seriesIndex: state.seriesIndex, targetNode: targetNode });
			});
    };
    var refresh = function () {
      if (scheduled) return;
      scheduled = true;
      requestAnimationFrame(function () {
        scheduled = false;
        figures.forEach(function (figure) {
          if (figure.isConnected) apply(figure); else figures.delete(figure);
        });
      });
    };
    var observer = new MutationObserver(refresh);
    observer.observe(document.documentElement, {
      attributes: true,
      subtree: false,
      attributeFilter: ["class", "data-theme"]
    });
    var preferredDark = window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)");
    if (preferredDark) {
      if (preferredDark.addEventListener) preferredDark.addEventListener("change", refresh);
      else if (preferredDark.addListener) preferredDark.addListener(refresh);
    }
    var preferredMotion = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)");
    if (preferredMotion) {
      if (preferredMotion.addEventListener) preferredMotion.addEventListener("change", refresh);
      else if (preferredMotion.addListener) preferredMotion.addListener(refresh);
    }
    runtime = window[key] = {
      register: function (figure) {
        figures.add(figure);
        var host = figure.querySelector("[_echarts_instance_]");
        if (resizeObserver && host) resizeObserver.observe(host);
        scheduleResize(host);
        apply(figure);
      },
      refresh: refresh
    };
  }
  runtime.register(document.currentScript.closest(".goshtoso-charts-interactive"));
}());
</script>`
