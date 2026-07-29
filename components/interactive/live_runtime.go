package interactive

import "github.com/a-h/templ"

func liveRuntime() templ.Component { return templ.Raw(liveRuntimeMarkup) }

const liveRuntimeMarkup = `<script data-goshtoso-charts-live-runtime>
(function () {
  "use strict";
  var key = "__goshtosoChartsLiveRuntime";
  var runtime = window[key];
  if (!runtime) {
    var streams = new Map();
    var finiteValues = function (values, count) {
      return Array.isArray(values) && values.length === count && values.every(function (value) {
        return typeof value === "number" && Number.isFinite(value);
      });
    };
    var cartesianSnapshot = function (chart, payload) {
      if (!payload || !Array.isArray(payload.categories) || !Array.isArray(payload.series)) return;
      if (!payload.categories.every(function (value) { return typeof value === "string"; })) return;
      var valid = payload.series.length > 0 && payload.series.every(function (series) {
        return series && typeof series.name === "string" && series.name.length > 0 &&
          finiteValues(series.values, payload.categories.length);
      });
      if (!valid) return;

      var configured = (chart.getOption().series || []);
      if (configured.length !== payload.series.length) return;
      var configuredByName = new Map();
      for (var index = 0; index < configured.length; index += 1) {
        var configuredSeries = configured[index];
        if (!configuredSeries.name || configuredByName.has(configuredSeries.name)) return;
        configuredByName.set(configuredSeries.name, configuredSeries);
      }
      var seen = new Set();
      var updates = [];
      for (var seriesIndex = 0; seriesIndex < payload.series.length; seriesIndex += 1) {
        var snapshotSeries = payload.series[seriesIndex];
        var existing = configuredByName.get(snapshotSeries.name);
        if (!existing || seen.has(snapshotSeries.name)) return;
        seen.add(snapshotSeries.name);
        updates.push({
          id: existing.id,
          name: existing.name,
          data: snapshotSeries.values,
          animationDurationUpdate: 0
        });
      }
      chart.setOption({
        xAxis: [{ data: payload.categories }],
        series: updates
      });
    };
    var closeDetached = function () {
      streams.forEach(function (source, figure) {
        if (!figure.isConnected) {
          source.close();
          streams.delete(figure);
        }
      });
    };
    var observer = new MutationObserver(closeDetached);
    observer.observe(document.documentElement, { childList: true, subtree: true });
    runtime = window[key] = {
      register: function (figure) {
        if (!figure || streams.has(figure) || !window.EventSource || !window.echarts) return;
        var url = figure.getAttribute("data-goshtoso-charts-live-url");
        if (!url) return;
        var host = figure.querySelector("[_echarts_instance_]");
        var chart = host && window.echarts.getInstanceByDom(host);
        if (!chart) return;
        var source = new EventSource(url);
        var eventName = figure.getAttribute("data-goshtoso-charts-live-event") || "message";
        source.addEventListener(eventName, function (event) {
          var payload;
          try { payload = JSON.parse(event.data); } catch (_) { return; }
          if (figure.getAttribute("data-goshtoso-charts-live-shape") === "cartesian") {
            cartesianSnapshot(chart, payload);
          }
        });
        streams.set(figure, source);
      }
    };
  }
  runtime.register(document.currentScript.closest(".goshtoso-charts-interactive"));
}());
</script>`
