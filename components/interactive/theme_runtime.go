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
    var cssColor = function (figure, name, fallback) {
      var probe = document.createElement("span");
      probe.hidden = true;
      probe.style.color = "var(" + name + ", " + fallback + ")";
      figure.appendChild(probe);
      var value = getComputedStyle(probe).color;
      probe.remove();
      return value || fallback;
    };
    var repeat = function (items, value) {
      return (items && items.length ? items : []).map(function () { return value; });
    };
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
      var seriesColors = [];
      for (var colorIndex = 1; colorIndex <= 8; colorIndex += 1) {
        seriesColors.push(cssColor(figure, "--color-chart-series-" + colorIndex, "#2563eb"));
      }
      var current = chart.getOption();
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
      var themedSeries = (current.series || []).map(function () {
        return {
          label: { color: text },
          endLabel: { color: text },
          emphasis: { label: { color: strong } },
          labelLine: { lineStyle: { color: outline } }
        };
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
        visualMap: repeat(current.visualMap, { textStyle: { color: text } }),
        tooltip: repeat(current.tooltip, {
          backgroundColor: surfaceAlt,
          borderColor: outline,
          textStyle: { color: strong }
        }),
        series: themedSeries
      };
      if (figure.getAttribute("data-goshtoso-charts-explicit-colors") !== "true") {
        themed.color = seriesColors;
      }
      chart.setOption(themed, false, true);
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
      subtree: true,
      attributeFilter: ["class", "data-theme"]
    });
    var preferredDark = window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)");
    if (preferredDark) {
      if (preferredDark.addEventListener) preferredDark.addEventListener("change", refresh);
      else if (preferredDark.addListener) preferredDark.addListener(refresh);
    }
    runtime = window[key] = {
      register: function (figure) {
        figures.add(figure);
        apply(figure);
      },
      refresh: refresh
    };
  }
  runtime.register(document.currentScript.closest(".goshtoso-charts-interactive"));
}());
</script>`
