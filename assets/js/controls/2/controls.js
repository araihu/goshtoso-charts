(function () {
  "use strict";

  if (window.__goshtosoChartsControls) return;

  var fullscreenTrigger = null;
  var fallbackWrapper = null;

  function wrapperFor(element) {
    return element && element.closest("[data-goshtoso-chart-wrapper]");
  }

  function chartInstance(wrapper) {
    if (!wrapper || !window.echarts) return null;
    var host = wrapper.querySelector("[_echarts_instance_]");
    return host ? window.echarts.getInstanceByDom(host) : null;
  }

  function resizeOnce(wrapper) {
    if (!wrapper) return;
    wrapper.dispatchEvent(new CustomEvent("goshtoso-charts:resize", { bubbles: true }));
  }

  function settledResize(wrapper) {
    if (!wrapper) return;
    requestAnimationFrame(function () {
      resizeOnce(wrapper);
      requestAnimationFrame(function () { resizeOnce(wrapper); });
    });
    clearTimeout(wrapper.__goshtosoChartResizeTimer);
    wrapper.__goshtosoChartResizeTimer = setTimeout(function () { resizeOnce(wrapper); }, 250);
  }

  function fullPageActive(wrapper) {
    return document.fullscreenElement === wrapper || fallbackWrapper === wrapper ||
      wrapper.classList.contains("goshtoso-charts-expanded");
  }

  function updateCollapseVisibility(wrapper) {
    var collapse = wrapper && wrapper.querySelector('[data-goshtoso-chart-control="collapse"]');
    if (collapse) collapse.hidden = fullPageActive(wrapper);
  }

  function setFullscreenState(wrapper, active) {
    var button = wrapper && wrapper.querySelector('[data-goshtoso-chart-control="fullscreen"]');
    if (button) {
      button.setAttribute("aria-pressed", active ? "true" : "false");
      button.setAttribute("aria-label", (active ? "Exit fullscreen for " : "Enter fullscreen for ") +
        wrapper.querySelector("[role=group]").getAttribute("aria-label").replace(/ chart controls$/, ""));
      var label = button.querySelector("[data-goshtoso-chart-control-label]");
      if (label) label.textContent = active ? "Exit fullscreen" : "Fullscreen";
    }
    updateCollapseVisibility(wrapper);
  }

  function leaveFallback() {
    if (!fallbackWrapper) return;
    var wrapper = fallbackWrapper;
    fallbackWrapper = null;
    wrapper.classList.remove("goshtoso-charts-fullscreen-fallback");
    setFullscreenState(wrapper, false);
    settledResize(wrapper);
    if (fullscreenTrigger && fullscreenTrigger.isConnected) fullscreenTrigger.focus();
    fullscreenTrigger = null;
  }

  function enterFallback(wrapper) {
    fallbackWrapper = wrapper;
    wrapper.classList.add("goshtoso-charts-fullscreen-fallback");
    setFullscreenState(wrapper, true);
    settledResize(wrapper);
  }

  function toggleFullscreen(button) {
    var wrapper = wrapperFor(button);
    if (!wrapper) return;
    if (document.fullscreenElement === wrapper) {
      document.exitFullscreen();
      return;
    }
    if (fallbackWrapper === wrapper) {
      leaveFallback();
      return;
    }
    fullscreenTrigger = button;
    if (wrapper.requestFullscreen) {
      wrapper.requestFullscreen().catch(function () {
        enterFallback(wrapper);
      });
      return;
    }
    enterFallback(wrapper);
  }

  function toggleCollapse(button) {
    var wrapper = wrapperFor(button);
    var content = wrapper && wrapper.querySelector("[data-goshtoso-chart-content]");
    if (!content) return;
    var expanded = button.getAttribute("aria-expanded") !== "false";
    content.hidden = expanded;
    button.setAttribute("aria-expanded", expanded ? "false" : "true");
    var label = button.querySelector("[data-goshtoso-chart-control-label]");
    if (label) label.textContent = expanded ? "Expand" : "Collapse";
    var chartLabel = wrapper.querySelector("[role=group]").getAttribute("aria-label").replace(/ chart controls$/, "");
    button.setAttribute("aria-label", (expanded ? "Expand " : "Collapse ") + chartLabel);
    if (!expanded) settledResize(wrapper);
  }

  function expandParts(wrapper) {
    var expand = wrapper && wrapper.querySelector("[data-goshtoso-chart-expand]");
    var modalRoot = expand && expand.firstElementChild;
    var dialog = modalRoot && modalRoot.querySelector('[role="dialog"]');
    var panel = dialog && dialog.querySelector(".goshtoso-charts-expand-panel");
    var body = panel && panel.children[1];
    var trigger = modalRoot && modalRoot.querySelector(":scope > button");
    return { expand: expand, dialog: dialog, body: body, trigger: trigger };
  }

  function closeExpand(wrapper) {
    if (!wrapper || !wrapper.classList.contains("goshtoso-charts-expanded")) return;
    var content = wrapper.querySelector("[data-goshtoso-chart-content]");
    var origin = content && content.__goshtosoChartOrigin;
    if (content && origin && origin.parentNode) origin.parentNode.insertBefore(content, origin.nextSibling);
    if (origin) origin.remove();
    if (content) {
      content.hidden = Boolean(content.__goshtosoChartWasHidden);
      delete content.__goshtosoChartWasHidden;
      delete content.__goshtosoChartOrigin;
    }
    wrapper.classList.remove("goshtoso-charts-expanded");
    updateCollapseVisibility(wrapper);
    settledResize(wrapper);
  }

  function openExpand(wrapper) {
    if (!wrapper || wrapper.classList.contains("goshtoso-charts-expanded")) return;
    var parts = expandParts(wrapper);
    var content = wrapper.querySelector("[data-goshtoso-chart-content]");
    if (!parts.dialog || !parts.body || !content || getComputedStyle(parts.dialog).display === "none") return;
    var origin = document.createComment("goshtoso-chart-content-origin");
    content.parentNode.insertBefore(origin, content);
    content.__goshtosoChartOrigin = origin;
    content.__goshtosoChartWasHidden = content.hidden;
    content.hidden = false;
    parts.body.appendChild(content);
    wrapper.classList.add("goshtoso-charts-expanded");
    updateCollapseVisibility(wrapper);
    settledResize(wrapper);
  }

  function prepareExpand(wrapper) {
    var parts = expandParts(wrapper);
    if (!parts.dialog || parts.dialog.__goshtosoChartObserved) return;
    parts.dialog.__goshtosoChartObserved = true;
    new MutationObserver(function () {
      if (getComputedStyle(parts.dialog).display === "none") closeExpand(wrapper);
    }).observe(parts.dialog, { attributes: true, attributeFilter: ["style"] });
  }

  function prepareExportLabel(wrapper) {
    var menu = wrapper && wrapper.querySelector("[data-goshtoso-chart-export-menu]");
    var trigger = menu && menu.querySelector(":scope > div > button");
    var group = wrapper && wrapper.querySelector("[role=group]");
    if (!trigger || !group) return;
    var label = group.getAttribute("aria-label").replace(/ chart controls$/, "");
    trigger.setAttribute("aria-label", "Export " + label);
  }

  function safeFilename(value) {
    var filename = String(value || "").trim().toLowerCase()
      .replace(/[^a-z0-9_-]+/g, "-").replace(/^[-_]+|[-_]+$/g, "").slice(0, 80)
      .replace(/[-_]+$/g, "");
    return filename || "goshtoso-chart";
  }

  function dimensions(svg) {
    var width = parseFloat(svg.getAttribute("width"));
    var height = parseFloat(svg.getAttribute("height"));
    if (!(width > 0)) width = svg.viewBox && svg.viewBox.baseVal.width;
    if (!(height > 0)) height = svg.viewBox && svg.viewBox.baseVal.height;
    if (!(width > 0) || !(height > 0)) {
      var bounds = svg.getBoundingClientRect();
      width = width > 0 ? width : bounds.width;
      height = height > 0 ? height : bounds.height;
    }
    if (!(width > 0) || !(height > 0)) throw new Error("Chart has no exportable dimensions.");
    return { width: Math.round(width), height: Math.round(height) };
  }

  function surfaceColor(wrapper) {
    var probe = document.createElement("span");
    probe.hidden = true;
    probe.style.color = "var(--color-chart-surface, #ffffff)";
    wrapper.appendChild(probe);
    var value = getComputedStyle(probe).color;
    probe.remove();
    if (value) return value;
    var figure = wrapper.querySelector("figure");
    var color = figure ? getComputedStyle(figure).backgroundColor : "";
    return color && color !== "rgba(0, 0, 0, 0)" ? color : "#ffffff";
  }

  function serializedSVG(wrapper) {
    var source = wrapper.querySelector("[data-goshtoso-chart-content] svg");
    if (!source) throw new Error("Chart has no SVG export source.");
    var size = dimensions(source);
    var clone = source.cloneNode(true);
    var sourceNodes = [source].concat(Array.from(source.querySelectorAll("*")));
    var cloneNodes = [clone].concat(Array.from(clone.querySelectorAll("*")));
    var properties = ["fill", "stroke", "color", "font-family", "font-size", "font-weight", "opacity"];
    sourceNodes.forEach(function (node, index) {
      var computed = getComputedStyle(node);
      cloneNodes[index].removeAttribute("style");
      Array.from(cloneNodes[index].attributes || []).forEach(function (attribute) {
        if (/var\(/.test(attribute.value)) cloneNodes[index].removeAttribute(attribute.name);
      });
      properties.forEach(function (property) {
        var value = computed.getPropertyValue(property);
        if (value) cloneNodes[index].style.setProperty(property, value);
      });
    });
    clone.setAttribute("xmlns", "http://www.w3.org/2000/svg");
    clone.setAttribute("width", String(size.width));
    clone.setAttribute("height", String(size.height));
    if (wrapper.dataset.goshtosoChartExportBackground === "transparent") {
      // go-analyze/charts SVG output starts with its full-canvas surface path.
      if (clone.firstElementChild) clone.firstElementChild.remove();
    } else {
      var rect = document.createElementNS("http://www.w3.org/2000/svg", "rect");
      rect.setAttribute("width", "100%");
      rect.setAttribute("height", "100%");
      rect.setAttribute("fill", surfaceColor(wrapper));
      rect.setAttribute("data-goshtoso-chart-export-surface", "");
      clone.insertBefore(rect, clone.firstChild);
    }
    var markup = new XMLSerializer().serializeToString(clone);
    if (/var\(/.test(markup)) throw new Error("Chart colors could not be resolved.");
    return { markup: markup, width: size.width, height: size.height };
  }

  function svgBlob(wrapper) {
    return new Blob([serializedSVG(wrapper).markup], { type: "image/svg+xml;charset=utf-8" });
  }

  async function staticPNGBlob(wrapper) {
    var serialized = serializedSVG(wrapper);
    var source = URL.createObjectURL(new Blob([serialized.markup], { type: "image/svg+xml;charset=utf-8" }));
    try {
      var image = new Image();
      image.src = source;
      await image.decode();
      var ratio = Number(wrapper.dataset.goshtosoChartExportPixelRatio) || 1;
      var canvas = document.createElement("canvas");
      canvas.width = Math.round(serialized.width * ratio);
      canvas.height = Math.round(serialized.height * ratio);
      var context = canvas.getContext("2d");
      if (!context) throw new Error("Canvas rendering is unavailable.");
      context.scale(ratio, ratio);
      context.drawImage(image, 0, 0, serialized.width, serialized.height);
      return await new Promise(function (resolve, reject) {
        canvas.toBlob(function (blob) {
          if (blob) resolve(blob); else reject(new Error("Browser could not encode PNG."));
        }, "image/png");
      });
    } finally {
      URL.revokeObjectURL(source);
    }
  }

  async function interactivePNGBlob(wrapper) {
    var chart = chartInstance(wrapper);
    if (!chart) throw new Error("Live chart instance is unavailable.");
    var transparent = wrapper.dataset.goshtosoChartExportBackground === "transparent";
    var dataURL = chart.getDataURL({
      type: "png",
      pixelRatio: Number(wrapper.dataset.goshtosoChartExportPixelRatio) || 1,
      backgroundColor: transparent ? "rgba(0,0,0,0)" : surfaceColor(wrapper)
    });
    var response = await fetch(dataURL);
    return response.blob();
  }

  function download(blob, filename) {
    var href = URL.createObjectURL(blob);
    var anchor = document.createElement("a");
    anchor.hidden = true;
    anchor.href = href;
    anchor.download = filename;
    document.body.append(anchor);
    anchor.click();
    anchor.remove();
    setTimeout(function () { URL.revokeObjectURL(href); }, 0);
  }

  async function exportChart(button, explicitFormat) {
    if (button.disabled) return;
    var wrapper = wrapperFor(button);
    if (!wrapper) return;
    var format = explicitFormat || button.dataset.goshtosoChartExport;
    var capability = wrapper.dataset.goshtosoChartCapability;
    var filename = safeFilename(wrapper.dataset.goshtosoChartExportFilename) + "." + format;
    var status = wrapper.querySelector("[data-goshtoso-chart-export-status]");
    button.disabled = true;
    button.setAttribute("aria-busy", "true");
    try {
      var blob;
      if (format === "svg" && capability === "static-svg") blob = svgBlob(wrapper);
      else if (format === "png" && capability === "static-svg") blob = await staticPNGBlob(wrapper);
      else if (format === "png" && capability === "interactive-raster") blob = await interactivePNGBlob(wrapper);
      else throw new Error("Unsupported chart export.");
      download(blob, filename);
      if (status) status.textContent = "Download ready: " + filename;
    } catch (error) {
      if (status) status.textContent = "Download failed: " + error.message;
      console.error("Chart export failed.", error);
    } finally {
      button.disabled = false;
      button.removeAttribute("aria-busy");
    }
  }

  document.addEventListener("click", function (event) {
    var control = event.target.closest("[data-goshtoso-chart-control]");
    if (control) {
      if (control.dataset.goshtosoChartControl === "collapse") toggleCollapse(control);
      if (control.dataset.goshtosoChartControl === "fullscreen") toggleFullscreen(control);
      return;
    }
    var exporter = event.target.closest("[data-goshtoso-chart-export]");
    if (exporter) exportChart(exporter);
    var expand = event.target.closest("[data-goshtoso-chart-expand]");
    if (expand) {
      var wrapper = wrapperFor(expand);
      var parts = expandParts(wrapper);
      if (event.target.closest("button") === parts.trigger) {
        prepareExpand(wrapper);
        requestAnimationFrame(function () { openExpand(wrapper); });
        setTimeout(function () { openExpand(wrapper); }, 50);
      }
    }
  });

  document.addEventListener("fullscreenchange", function () {
    var wrapper = document.fullscreenElement && wrapperFor(document.fullscreenElement);
    document.querySelectorAll("[data-goshtoso-chart-wrapper]").forEach(function (candidate) {
      setFullscreenState(candidate, candidate === wrapper);
    });
    if (wrapper) settledResize(wrapper);
    else if (fullscreenTrigger) {
      var previous = wrapperFor(fullscreenTrigger);
      settledResize(previous);
      if (fullscreenTrigger.isConnected) fullscreenTrigger.focus();
      fullscreenTrigger = null;
    }
  });

  document.addEventListener("keydown", function (event) {
    if (event.key === "Escape" && fallbackWrapper) leaveFallback();
  });

  window.__goshtosoChartsControls = {
    safeFilename: safeFilename,
    dimensions: dimensions,
    exportFromMenu: function (element, format) { exportChart(element, format); }
  };

  document.querySelectorAll("[data-goshtoso-chart-wrapper]").forEach(prepareExportLabel);
}());
