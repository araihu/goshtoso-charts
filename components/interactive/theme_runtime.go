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
    var repeat = function (items, value) {
      return (items && items.length ? items : []).map(function () { return value; });
    };
    var resize = function (figure) {
      if (!figure.isConnected || !window.echarts) return;
      var host = figure.querySelector("[_echarts_instance_]");
      if (!host) return;
      var chart = window.echarts.getInstanceByDom(host);
      if (!chart) return;
      chart.resize();
    };
    var resizeObserver = window.ResizeObserver ? new ResizeObserver(function (entries) {
      entries.forEach(function (entry) { resize(entry.target); });
    }) : null;
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
        var themedItem = {
          id: series.id,
          name: series.name,
          animationDurationUpdate: 0,
          label: { color: text },
          endLabel: { color: text },
          emphasis: { label: { color: strong } },
          labelLine: { lineStyle: { color: outline } }
        };
        if (series.type === "boxplot" && managesSeriesItem(index)) {
          themedItem.itemStyle = { color: palette[index % palette.length], borderColor: palette[index % palette.length] };
        }
        if (series.type === "gauge") {
          themedItem.axisLine = { lineStyle: { color: [[1, surfaceAlt]] } };
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
      });
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
        radiusAxis: repeat(current.radiusAxis, axis),
        angleAxis: repeat(current.angleAxis, axis),
        radar: repeat(current.radar, radar),
        visualMap: themedVisualMaps,
        tooltip: repeat(current.tooltip, {
          backgroundColor: surfaceAlt,
          borderColor: outline,
          textStyle: { color: strong }
        }),
        series: themedSeries
      };
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
        if (resizeObserver) resizeObserver.observe(figure);
        resize(figure);
        apply(figure);
      },
      refresh: refresh
    };
  }
  runtime.register(document.currentScript.closest(".goshtoso-charts-interactive"));
}());
</script>`
