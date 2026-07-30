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
func themeRuntime() templ.Component { return templ.Raw(ThemeRuntimeMarkup) }

const ThemeRuntimeMarkup = `<script data-goshtoso-charts-theme-runtime>
(function () {
  "use strict";
  var key = "__goshtosoChartsThemeRuntime";
		var runtime = window[key];
		if (!runtime) {
			var figures = new Set();
			var responsiveFigures = new Map();
			var targetFigures = new WeakMap();
			var pieAutoEmphasisStates = new WeakMap();
			var settleFrames = 2;
			var maxSettleFrames = 10;
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
    var scatter3DColdToWarmFallbacks = [
      "#313695", "#4575b4", "#74add1", "#abd9e9", "#e0f3f8",
      "#fee090", "#fdae61", "#f46d43", "#d73027", "#a50026"
    ];
		var stopPieAutoEmphasis = function (figure) {
			var state = pieAutoEmphasisStates.get(figure);
			if (!state) return;
			if (state.timer) clearInterval(state.timer);
			if (state.chart && state.index >= 0 && (!state.chart.isDisposed || !state.chart.isDisposed())) {
				state.chart.dispatchAction({ type: "downplay", seriesIndex: state.seriesIndex, dataIndex: state.index });
				state.chart.dispatchAction({ type: "hideTip" });
			}
			pieAutoEmphasisStates.delete(figure);
		};
	    var unregister = function (figure) {
			stopPieAutoEmphasis(figure);
	      var state = responsiveFigures.get(figure);
	      if (!state) return;
	      state.targets.forEach(function (target) {
	        var owners = targetFigures.get(target);
	        if (!owners) return;
	        owners.delete(figure);
	        if (!owners.size) {
	          if (resizeObserver) resizeObserver.unobserve(target);
	          targetFigures.delete(target);
	        }
	      });
	      responsiveFigures.delete(figure);
	      figures.delete(figure);
	    };
	    var resize = function (figure, state) {
	      if (!figure.isConnected || !window.echarts) {
	        if (!figure.isConnected) unregister(figure);
	        return null;
	      }
	      var host = figure.querySelector("[_echarts_instance_]");
	      if (!host) return null;
	      var bounds = host.getBoundingClientRect();
	      var width = Math.round(bounds.width);
	      var height = Math.round(bounds.height);
	      if (!(width > 0) || !(height > 0)) return null;
	      var chart = window.echarts.getInstanceByDom(host);
	      if (!chart) return null;
	      if (chart.getWidth() !== width || chart.getHeight() !== height) {
	        chart.resize({ width: width, height: height, animation: { duration: 0 } });
	      }
	      var geometry = width + "x" + height;
	      state.stableFrames = geometry === state.geometry ? state.stableFrames + 1 : 0;
	      state.geometry = geometry;
	      return geometry;
	    };
	    var scheduleResize = function (figure) {
	      var state = responsiveFigures.get(figure);
	      if (!state || state.scheduled) return;
	      state.scheduled = true;
	      state.stableFrames = 0;
	      state.frameCount = 0;
	      var step = function () {
	        if (!responsiveFigures.has(figure)) return;
	        state.frameCount += 1;
	        resize(figure, state);
	        if (state.stableFrames < settleFrames && state.frameCount < maxSettleFrames) {
	          requestAnimationFrame(step);
	          return;
	        }
	        state.scheduled = false;
	      };
	      requestAnimationFrame(step);
	    };
	    var resizeObserver = window.ResizeObserver ? new ResizeObserver(function (entries) {
	      entries.forEach(function (entry) {
	        var owners = targetFigures.get(entry.target);
	        if (owners) owners.forEach(scheduleResize);
	      });
	    }) : null;
	    var observeTarget = function (figure, target) {
	      if (!resizeObserver || !target) return;
	      var state = responsiveFigures.get(figure);
	      if (!state || state.targets.has(target)) return;
	      state.targets.add(target);
	      var owners = targetFigures.get(target);
	      if (!owners) {
	        owners = new Set();
	        targetFigures.set(target, owners);
	        resizeObserver.observe(target);
	      }
	      owners.add(figure);
	    };
	    var registerResponsive = function (figure) {
	      if (responsiveFigures.has(figure)) {
	        scheduleResize(figure);
	        return;
	      }
	      responsiveFigures.set(figure, {
	        targets: new Set(), scheduled: false, stableFrames: 0, frameCount: 0, geometry: ""
	      });
	      var host = figure.querySelector("[_echarts_instance_]");
	      observeTarget(figure, host);
	      observeTarget(figure, host && host.parentElement);
	      scheduleResize(figure);
	    };
		var syncPieAutoEmphasis = function (figure, chart) {
			var raw = figure.getAttribute("data-goshtoso-charts-pie-auto-emphasis") || "";
			var config = null;
			try { config = raw ? JSON.parse(raw) : null; } catch (_) {}
			var reduced = Boolean(preferredMotion && preferredMotion.matches);
			var existing = pieAutoEmphasisStates.get(figure);
			var signature = raw + ":" + reduced;
			if (existing && existing.chart === chart && existing.signature === signature) return;
			stopPieAutoEmphasis(figure);
			if (!config || reduced) return;
			var option = chart.getOption();
			var series = option.series && option.series[config.seriesIndex];
			var count = series && series.data ? series.data.length : 0;
			if (!count) return;
			var state = {
				chart: chart, signature: signature, seriesIndex: config.seriesIndex,
				index: -1, timer: 0
			};
			var tick = function () {
				if (!figure.isConnected || (chart.isDisposed && chart.isDisposed())) {
					unregister(figure);
					return;
				}
				if (state.index >= 0) chart.dispatchAction({ type: "downplay", seriesIndex: state.seriesIndex, dataIndex: state.index });
				state.index = (state.index + 1) % count;
				chart.dispatchAction({ type: "highlight", seriesIndex: state.seriesIndex, dataIndex: state.index });
				if (config.showTooltip) chart.dispatchAction({ type: "showTip", seriesIndex: state.seriesIndex, dataIndex: state.index });
			};
			state.timer = setInterval(tick, config.interval);
			pieAutoEmphasisStates.set(figure, state);
		};
	    var scheduleAll = function () {
	      responsiveFigures.forEach(function (_, figure) { scheduleResize(figure); });
	    };
	    window.addEventListener("resize", scheduleAll);
	    document.addEventListener("goshtoso-charts:resize", function (event) {
	      var wrapper = event.target.closest && event.target.closest("[data-goshtoso-chart-wrapper]");
	      if (!wrapper) return;
	      wrapper.querySelectorAll(".goshtoso-charts-interactive").forEach(scheduleResize);
	    });
	    document.addEventListener("goshtoso-charts:export-request", function (event) {
	      var detail = event.detail;
	      if (!detail || detail.format !== "png" || detail.dataURL) return;
	      var wrapper = event.target.closest && event.target.closest("[data-goshtoso-chart-wrapper]");
	      var figure = wrapper && wrapper.querySelector(".goshtoso-charts-interactive");
	      var host = figure && figure.querySelector("[_echarts_instance_]");
	      var chart = host && window.echarts && window.echarts.getInstanceByDom(host);
	      if (!chart) return;
	      detail.dataURL = chart.getDataURL({
	        type: "png",
	        pixelRatio: detail.pixelRatio,
	        backgroundColor: detail.backgroundColor
	      });
	    });
	    new MutationObserver(function () {
	      responsiveFigures.forEach(function (_, figure) {
	        if (!figure.isConnected) unregister(figure);
	      });
	    }).observe(document.documentElement, { childList: true, subtree: true });
	    var apply = function (figure) {
	      if (!figure.isConnected || !window.echarts) return;
	      var host = figure.querySelector("[_echarts_instance_]");
	      if (!host) return;
			var chart = window.echarts.getInstanceByDom(host);
			if (!chart) return;

      var surface = cssColor(figure, "--color-chart-surface", "#ffffff");
      var surfaceAlt = cssColor(figure, "--color-chart-surface-alt", surface);
      var outline = cssColor(figure, "--color-chart-outline", "#64748b");
      var grid = cssColor(figure, "--color-chart-grid", outline);
      var axisColor = cssColor(figure, "--color-chart-axis", outline);
      var text = cssColor(figure, "--color-chart-text", "#1f2937");
      var strong = cssColor(figure, "--color-chart-text-strong", text);
      var muted = cssColor(figure, "--color-chart-text-muted", text);
      var scaleLow = cssColor(figure, "--color-chart-scale-low", surfaceAlt);
      var scaleMid = cssColor(figure, "--color-chart-scale-mid", "#3b82f6");
      var scaleHigh = cssColor(figure, "--color-chart-scale-high", "#1e3a8a");
      var seriesColors = [];
      for (var colorIndex = 1; colorIndex <= 12; colorIndex += 1) {
        seriesColors.push(cssColor(figure, "--color-chart-series-" + colorIndex, "#2563eb"));
      }
      var divergingColors = [];
      for (var divergingIndex = 1; divergingIndex <= 5; divergingIndex += 1) {
        divergingColors.push(cssColor(figure, "--color-chart-diverging-" + divergingIndex, [scaleLow, scaleLow, scaleMid, scaleHigh, scaleHigh][divergingIndex - 1]));
      }
      var current = chart.getOption();
      var explicitColors = figure.getAttribute("data-goshtoso-charts-explicit-colors") === "true";
      var explicitVisualMapColors = figure.getAttribute("data-goshtoso-charts-explicit-visual-map-colors") === "true";
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
			var geoGeometryPaint = null;
			try { geoGeometryPaint = JSON.parse(figure.getAttribute("data-goshtoso-charts-geo-geometry-paint") || "null"); } catch (_) {}
			var geoSeriesPaints = [];
			try { geoSeriesPaints = JSON.parse(figure.getAttribute("data-goshtoso-charts-geo-series-paints") || "[]"); } catch (_) {}
			var scatter3DPaints = [];
			try { scatter3DPaints = JSON.parse(figure.getAttribute("data-goshtoso-charts-scatter3d-paints") || "[]"); } catch (_) {}
			var scatter3DColdToWarm = figure.getAttribute("data-goshtoso-charts-scatter3d-cold-to-warm") === "true";
			var bar3DPaints = [];
			try { bar3DPaints = JSON.parse(figure.getAttribute("data-goshtoso-charts-bar3d-paints") || "[]"); } catch (_) {}
			var bar3DColdToWarm = figure.getAttribute("data-goshtoso-charts-bar3d-cold-to-warm") === "true";
			var surface3DPaints = [];
			try { surface3DPaints = JSON.parse(figure.getAttribute("data-goshtoso-charts-surface3d-paints") || "[]"); } catch (_) {}
			var surface3DColdToWarm = figure.getAttribute("data-goshtoso-charts-surface3d-cold-to-warm") === "true";
			var line3DPaints = [];
			try { line3DPaints = JSON.parse(figure.getAttribute("data-goshtoso-charts-line3d-paints") || "[]"); } catch (_) {}
			var line3DColdToWarm = figure.getAttribute("data-goshtoso-charts-line3d-cold-to-warm") === "true";
			var line3DAutoRotate = figure.getAttribute("data-goshtoso-charts-line3d-auto-rotate") === "true";
			var reduceMotion = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
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
        axisLine: { lineStyle: { color: axisColor } },
        axisTick: { lineStyle: { color: axisColor } },
        splitLine: { lineStyle: { color: grid } },
        nameTextStyle: { color: text }
      };
      var radar = {
        axisName: { color: text },
        axisLine: { lineStyle: { color: axisColor } },
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
		if (series.type === "map") {
			themedItem.showLegendSymbol = false;
			themedItem.itemStyle = Object.assign({}, series.itemStyle || {}, { areaColor: surfaceAlt, borderColor: outline });
			themedItem.data = (series.data || []).map(function (item, itemIndex) {
				if (!item || typeof item !== "object") return item;
				var regionColor = item.sourceColor
					? rendererColor(item.sourceColor, palette[itemIndex % palette.length])
					: (item.className ? classColorOrFallback(figure, item.className, palette[itemIndex % palette.length]) : "");
				if (!regionColor) return item;
				return Object.assign({}, item, { itemStyle: Object.assign({}, item.itemStyle || {}, { color: regionColor }) });
			});
		}
		if ((series.type === "scatter" || series.type === "effectScatter") && series.coordinateSystem === "geo") {
			var geoSeriesPaint = geoSeriesPaints[index] || {};
			var geoSeriesColor = geoSeriesPaint.class
				? classColorOrFallback(figure, geoSeriesPaint.class, palette[index % palette.length])
				: (geoSeriesPaint.color
					? rendererColor(geoSeriesPaint.color, palette[index % palette.length])
					: (series.itemStyle && series.itemStyle.color
						? rendererColor(series.itemStyle.color, palette[index % palette.length])
						: palette[index % palette.length]));
			themedItem.itemStyle = Object.assign({}, series.itemStyle || {}, { color: geoSeriesColor });
			themedItem.data = (series.data || []).map(function (item) {
				if (!item || typeof item !== "object") return item;
				var pointColor = item.sourceColor
					? rendererColor(item.sourceColor, geoSeriesColor)
					: (item.className ? classColorOrFallback(figure, item.className, geoSeriesColor) : "");
				if (!pointColor) return item;
				return Object.assign({}, item, { itemStyle: Object.assign({}, item.itemStyle || {}, { color: pointColor }) });
			});
		}
		if (series.type === "scatter3D") {
			var scatter3DPaint = scatter3DPaints[index] || {};
			var scatter3DSeriesColor = scatter3DPaint.class
				? classColorOrFallback(figure, scatter3DPaint.class, palette[index % palette.length])
				: (scatter3DPaint.color
					? rendererColor(scatter3DPaint.color, palette[index % palette.length])
					: palette[index % palette.length]);
			themedItem.itemStyle = Object.assign({}, series.itemStyle || {}, { color: scatter3DSeriesColor });
			themedItem.data = (series.data || []).map(function (item) {
				if (!item || typeof item !== "object") return item;
				var pointColor = item.sourceColor
					? rendererColor(item.sourceColor, scatter3DSeriesColor)
					: (item.className ? classColorOrFallback(figure, item.className, scatter3DSeriesColor) : "");
				if (!pointColor) return item;
				return Object.assign({}, item, { itemStyle: Object.assign({}, item.itemStyle || {}, { color: pointColor }) });
			});
		}
		if (series.type === "bar3D") {
			var bar3DPaint = bar3DPaints[index] || {};
			var bar3DSeriesColor = bar3DPaint.class
				? classColorOrFallback(figure, bar3DPaint.class, palette[index % palette.length])
				: (bar3DPaint.color
					? rendererColor(bar3DPaint.color, palette[index % palette.length])
					: palette[index % palette.length]);
			themedItem.itemStyle = Object.assign({}, series.itemStyle || {}, { color: bar3DSeriesColor });
			themedItem.data = (series.data || []).map(function (item) {
				if (!item || typeof item !== "object") return item;
				var cellColor = item.sourceColor
					? rendererColor(item.sourceColor, bar3DSeriesColor)
					: (item.className ? classColorOrFallback(figure, item.className, bar3DSeriesColor) : "");
				if (!cellColor) return item;
				return Object.assign({}, item, { itemStyle: Object.assign({}, item.itemStyle || {}, { color: cellColor }) });
			});
		}
		if (series.type === "surface") {
			var surface3DPaint = surface3DPaints[index] || {};
			var surface3DSeriesColor = surface3DPaint.class
				? classColorOrFallback(figure, surface3DPaint.class, palette[index % palette.length])
				: (surface3DPaint.color
					? rendererColor(surface3DPaint.color, palette[index % palette.length])
					: palette[index % palette.length]);
			themedItem.itemStyle = Object.assign({}, series.itemStyle || {}, { color: surface3DSeriesColor });
			themedItem.data = (series.data || []).map(function (item) {
				if (!item || typeof item !== "object") return item;
				var pointColor = item.sourceColor
					? rendererColor(item.sourceColor, surface3DSeriesColor)
					: (item.className ? classColorOrFallback(figure, item.className, surface3DSeriesColor) : "");
				if (!pointColor) return item;
				return Object.assign({}, item, { itemStyle: Object.assign({}, item.itemStyle || {}, { color: pointColor }) });
			});
		}
		if (series.type === "line3D") {
			var line3DPaint = line3DPaints[index] || {};
			var line3DSeriesColor = line3DPaint.class
				? classColorOrFallback(figure, line3DPaint.class, palette[index % palette.length])
				: (line3DPaint.color
					? rendererColor(line3DPaint.color, palette[index % palette.length])
					: palette[index % palette.length]);
			themedItem.lineStyle = Object.assign({}, series.lineStyle || {}, { color: line3DSeriesColor });
		}
		if (series.type === "line") {
			var lineSeriesColor = palette[index % palette.length];
			if (series.markPoint) themedItem.markPoint = Object.assign({}, series.markPoint, {
				label: Object.assign({}, series.markPoint.label || {}, { color: strong }),
				itemStyle: Object.assign({}, series.markPoint.itemStyle || {}, { color: lineSeriesColor })
			});
			if (series.markLine) themedItem.markLine = Object.assign({}, series.markLine, {
				label: Object.assign({}, series.markLine.label || {}, { color: strong }),
				lineStyle: Object.assign({}, series.markLine.lineStyle || {}, { color: outline })
			});
			if (series.markArea) themedItem.markArea = Object.assign({}, series.markArea, {
				label: Object.assign({}, series.markArea.label || {}, { color: strong }),
				itemStyle: Object.assign({}, series.markArea.itemStyle || {}, { color: scaleHigh, opacity: 0.18 })
			});
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
          themedItem.axisTick = { lineStyle: { color: axisColor } };
          themedItem.splitLine = { lineStyle: { color: axisColor } };
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
        if (scatter3DColdToWarm || bar3DColdToWarm || surface3DColdToWarm || line3DColdToWarm) {
          visualMap.inRange = { color: Array.from({ length: 10 }, function (_, index) {
            return cssColor(figure, "--color-chart-scatter3d-" + (index + 1), scatter3DColdToWarmFallbacks[index]);
          }) };
        } else if (!explicitVisualMapColors) visualMap.inRange = { color: divergingColors };
        return visualMap;
      });
			var themedGeo = (current.geo || []).map(function (geo) {
				var areaColor = geoGeometryPaint && geoGeometryPaint.class
					? classColorOrFallback(figure, geoGeometryPaint.class, surfaceAlt)
					: (geoGeometryPaint && geoGeometryPaint.color
						? rendererColor(geoGeometryPaint.color, surfaceAlt)
						: surfaceAlt);
				return { itemStyle: Object.assign({}, geo.itemStyle || {}, { areaColor: areaColor, borderColor: axisColor }) };
			});
			var themedCalendars = (current.calendar || []).map(function (calendar) {
				var calendarItemStyle = calendar.itemStyle || {};
				return {
					itemStyle: Object.assign({}, calendarItemStyle, {
						color: calendarItemStyle.color ? rendererColor(calendarItemStyle.color, surfaceAlt) : surfaceAlt,
						borderColor: calendarItemStyle.borderColor ? rendererColor(calendarItemStyle.borderColor, axisColor) : axisColor
					}),
					splitLine: { lineStyle: { color: axisColor } },
					dayLabel: Object.assign({}, calendar.dayLabel || {}, { color: muted }),
					monthLabel: Object.assign({}, calendar.monthLabel || {}, { color: muted }),
					yearLabel: Object.assign({}, calendar.yearLabel || {}, { color: text })
				};
			});
			var themed = {
        backgroundColor: surface,
        textStyle: { color: text },
        title: repeat(current.title, { textStyle: { color: strong }, subtextStyle: { color: muted } }),
        legend: repeat(current.legend, { textStyle: { color: text } }),
        xAxis: repeat(current.xAxis, axis),
        yAxis: repeat(current.yAxis, axis),
        xAxis3D: repeat(current.xAxis3D, axis),
        yAxis3D: repeat(current.yAxis3D, axis),
        zAxis3D: repeat(current.zAxis3D, axis),
        singleAxis: repeat(current.singleAxis, axis),
        radiusAxis: repeat(current.radiusAxis, axis),
        angleAxis: repeat(current.angleAxis, axis),
		parallelAxis: repeat(current.parallelAxis, axis),
        radar: repeat(current.radar, radar),
        geo: themedGeo,
				calendar: themedCalendars,
        visualMap: themedVisualMaps,
        tooltip: repeat(current.tooltip, {
          backgroundColor: surfaceAlt,
          borderColor: outline,
          textStyle: { color: strong }
        }),
        series: themedSeries
			};
			if (line3DAutoRotate) {
				themed.grid3D = (current.grid3D || []).map(function (grid3D) {
					return { viewControl: Object.assign({}, grid3D.viewControl || {}, { autoRotate: !reduceMotion }) };
				});
			}
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
			syncPieAutoEmphasis(figure, chart);
    };
    var refresh = function () {
      if (scheduled) return;
      scheduled = true;
      requestAnimationFrame(function () {
        scheduled = false;
	        figures.forEach(function (figure) {
	          if (figure.isConnected) apply(figure); else unregister(figure);
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
	        if (!figure) return;
	        figures.add(figure);
	        registerResponsive(figure);
	        apply(figure);
	      },
      refresh: refresh
    };
  }
  runtime.register(document.currentScript.closest(".goshtoso-charts-interactive"));
}());
</script>`
