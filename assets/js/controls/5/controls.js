(function () {
  "use strict";

  if (window.__goshtosoChartsControls) return;

  var fullscreenTrigger = null;
  var fallbackWrapper = null;
  var lifecycleObserver = null;

  function wrapperFor(element) {
    return element && typeof element.closest === "function"
      ? element.closest("[data-goshtoso-chart-wrapper]") : null;
  }

  function chartLabel(wrapper) {
    var group = wrapper && wrapper.querySelector("[role=group]");
    var label = group && group.getAttribute("aria-label");
    return label ? label.replace(/ chart controls$/, "") : "chart";
  }

  function normalizeWrapperMode(mode) {
    if (mode === "" || mode === "enabled") return "enabled";
    if (mode === "disabled" || mode === "hidden") return mode;
    return null;
  }

  function wrapperMode(wrapper) {
    var mode = wrapper && wrapper.dataset.goshtosoChartWrapperMode;
    return normalizeWrapperMode(mode) || "enabled";
  }

  function actionFieldset(wrapper) {
    return wrapper && wrapper.querySelector("[data-goshtoso-chart-actions-fieldset]");
  }

  function setActionsDisabled(wrapper, disabled) {
    var fieldset = actionFieldset(wrapper);
    if (!fieldset) return;
    fieldset.disabled = disabled;
    if (disabled) fieldset.setAttribute("aria-disabled", "true");
    else fieldset.removeAttribute("aria-disabled");
    fieldset.querySelectorAll("button, [role=menuitem], a").forEach(function (action) {
      if (disabled) {
        if (action.dataset.goshtosoChartWrapperDisabled !== "true") {
          action.dataset.goshtosoChartWrapperTabindex = action.hasAttribute("tabindex")
            ? action.getAttribute("tabindex") : "__absent__";
          action.dataset.goshtosoChartWrapperAriaDisabled = action.hasAttribute("aria-disabled")
            ? action.getAttribute("aria-disabled") : "__absent__";
        }
        action.setAttribute("aria-disabled", "true");
        action.setAttribute("tabindex", "-1");
        action.dataset.goshtosoChartWrapperDisabled = "true";
      } else if (action.dataset.goshtosoChartWrapperDisabled === "true") {
        if (action.dataset.goshtosoChartWrapperAriaDisabled === "__absent__") action.removeAttribute("aria-disabled");
        else if (action.dataset.goshtosoChartWrapperAriaDisabled !== undefined) {
          action.setAttribute("aria-disabled", action.dataset.goshtosoChartWrapperAriaDisabled);
        }
        if (action.dataset.goshtosoChartWrapperTabindex === "__absent__") action.removeAttribute("tabindex");
        else if (action.dataset.goshtosoChartWrapperTabindex !== undefined) {
          action.setAttribute("tabindex", action.dataset.goshtosoChartWrapperTabindex);
        }
        delete action.dataset.goshtosoChartWrapperAriaDisabled;
        delete action.dataset.goshtosoChartWrapperTabindex;
        delete action.dataset.goshtosoChartWrapperDisabled;
      }
    });
  }

  function setWrapperDOMState(wrapper, mode) {
    wrapper.dataset.goshtosoChartWrapperMode = mode;
    setActionsDisabled(wrapper, mode === "disabled");
    if (mode === "hidden") {
      wrapper.hidden = true;
      wrapper.setAttribute("inert", "");
      wrapper.setAttribute("aria-hidden", "true");
    } else {
      wrapper.hidden = false;
      wrapper.removeAttribute("inert");
      wrapper.removeAttribute("aria-hidden");
    }
  }

  function actionsFor(wrapper, action) {
    if (!wrapper) return [];
    return Array.from(wrapper.querySelectorAll(
      '[data-goshtoso-chart-control="' + action + '"], [id*="-' + action + '-action"], ' +
      '[id*="-' + action + '-primary-action"]'
    ));
  }

  function actionTrigger(action) {
    var menu = action && action.closest('[role="menu"]');
    var dropdown = menu && menu.parentElement;
    return (dropdown && dropdown.querySelector(":scope > button")) || action;
  }

  function setActionLabel(action, text) {
    var label = action.querySelector("[data-goshtoso-chart-control-label]");
    if (label) {
      label.textContent = text;
      return;
    }
    var replaced = Array.from(action.childNodes).some(function (node) {
      if (node.nodeType !== Node.TEXT_NODE || !node.textContent.trim()) return false;
      node.textContent = text;
      return true;
    });
    if (replaced) return;
    var nested = action.querySelector("span > span:last-child");
    if (nested) nested.textContent = text;
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

  function setFullscreenState(wrapper, active) {
    actionsFor(wrapper, "fullscreen").forEach(function (button) {
      button.setAttribute("aria-pressed", active ? "true" : "false");
      button.setAttribute("aria-label", (active ? "Exit fullscreen for " : "Enter fullscreen for ") +
        chartLabel(wrapper));
      setActionLabel(button, active ? "Exit fullscreen" : "Fullscreen");
    });
  }

  function leaveFallback(restoreFocus) {
    if (!fallbackWrapper) return;
    var wrapper = fallbackWrapper;
    fallbackWrapper = null;
    wrapper.classList.remove("goshtoso-charts-fullscreen-fallback");
    setFullscreenState(wrapper, false);
    settledResize(wrapper);
    if (restoreFocus !== false && fullscreenTrigger && fullscreenTrigger.isConnected) fullscreenTrigger.focus();
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
    if (!wrapper || wrapperMode(wrapper) !== "enabled") return;
    if (document.fullscreenElement === wrapper) {
      var exit = document.exitFullscreen();
      if (exit && typeof exit.catch === "function") exit.catch(function () {});
      return;
    }
    if (fallbackWrapper === wrapper) {
      leaveFallback(true);
      return;
    }
    fullscreenTrigger = actionTrigger(button);
    if (wrapper.requestFullscreen) {
      wrapper.requestFullscreen().catch(function () {
        enterFallback(wrapper);
      });
      return;
    }
    enterFallback(wrapper);
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

  function closeExpand(wrapper, settle, restoreFocus) {
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
    if (settle !== false) settledResize(wrapper);
    var returnFocus = wrapper.__goshtosoChartExpandReturnFocus;
    delete wrapper.__goshtosoChartExpandReturnFocus;
    if (restoreFocus !== false && returnFocus && returnFocus.isConnected) {
      setTimeout(function () { returnFocus.focus(); }, 0);
    }
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
    settledResize(wrapper);
  }

  function prepareExpand(wrapper) {
    var parts = expandParts(wrapper);
    if (!parts.dialog || parts.dialog.__goshtosoChartObserved) return;
    parts.dialog.__goshtosoChartObserved = true;
    parts.dialog.__goshtosoChartObserver = new MutationObserver(function () {
      if (getComputedStyle(parts.dialog).display === "none") closeExpand(wrapper);
    });
    parts.dialog.__goshtosoChartObserver.observe(parts.dialog, { attributes: true, attributeFilter: ["style"] });
  }

  function expandFromMenu(element) {
    var wrapper = wrapperFor(element);
    var parts = expandParts(wrapper);
    if (!wrapper || wrapperMode(wrapper) !== "enabled" || !parts.trigger) return;
    wrapper.__goshtosoChartExpandReturnFocus = actionTrigger(element);
    prepareExpand(wrapper);
    parts.trigger.click();
  }

  function prepareActions(wrapper) {
    setFullscreenState(wrapper, document.fullscreenElement === wrapper || fallbackWrapper === wrapper);
    actionsFor(wrapper, "expand").forEach(function (action) {
      action.setAttribute("aria-label", "Expand " + chartLabel(wrapper));
    });
    var stackedTrigger = wrapper && wrapper.querySelector('[id$="-stacked"] > button');
    if (stackedTrigger) stackedTrigger.setAttribute("aria-label", "Expand " + chartLabel(wrapper));
    var exportTrigger = wrapper && wrapper.querySelector('[id$="-export"] > button');
    if (exportTrigger) exportTrigger.setAttribute("aria-label", "Export " + chartLabel(wrapper));
    wrapper.querySelectorAll('[id*="-export-"][title]').forEach(function (action) {
      action.setAttribute("aria-label", action.getAttribute("title"));
    });
    var parts = expandParts(wrapper);
    if (parts.trigger) {
      parts.trigger.setAttribute("aria-hidden", "true");
      parts.trigger.setAttribute("tabindex", "-1");
    }
  }

  function closeTransientUI(wrapper) {
    if (!wrapper) return;
    wrapper.querySelectorAll('button[aria-expanded="true"]').forEach(function (trigger) {
      trigger.click();
    });
    var parts = expandParts(wrapper);
    delete wrapper.__goshtosoChartExpandReturnFocus;
    closeExpand(wrapper, true, false);
    if (parts.dialog && getComputedStyle(parts.dialog).display !== "none") {
      var closeButton = parts.dialog.querySelector('button[aria-label="close modal"]');
      if (closeButton) closeButton.click();
    }
    if (fallbackWrapper === wrapper) leaveFallback(false);
    if (document.fullscreenElement === wrapper) {
      fullscreenTrigger = null;
      var exit = document.exitFullscreen();
      if (exit && typeof exit.catch === "function") exit.catch(function () {});
    }
  }

  function applyWrapperMode(wrapper, mode, focusReturn, emitChange) {
    mode = normalizeWrapperMode(mode);
    if (!wrapper || !mode) return false;
    var previousMode = wrapperMode(wrapper);
    if (previousMode === mode) {
      setWrapperDOMState(wrapper, mode);
      return true;
    }
    if (mode === "disabled" || mode === "hidden") closeTransientUI(wrapper);
    setWrapperDOMState(wrapper, mode);
    if ((previousMode === "hidden" && mode !== "hidden") ||
        (previousMode === "disabled" && mode === "enabled")) {
      settledResize(wrapper);
    }
    var active = document.activeElement;
    if ((mode === "hidden" || (mode === "disabled" && active && actionFieldset(wrapper) && actionFieldset(wrapper).contains(active))) &&
        active && wrapper.contains(active)) {
      if (focusReturn && typeof focusReturn.focus === "function" && focusReturn.isConnected && !wrapper.contains(focusReturn)) {
        focusReturn.focus();
      } else if (typeof active.blur === "function") {
        active.blur();
      }
    }
    if (emitChange !== false) {
      wrapper.dispatchEvent(new CustomEvent("goshtoso-charts:wrapper-mode-change", {
        bubbles: true,
        detail: { previousMode: previousMode, mode: mode }
      }));
    }
    return true;
  }

  function prepareWrapper(wrapper) {
    if (!wrapper || !wrapper.isConnected) return;
    var mode = wrapperMode(wrapper);
    setWrapperDOMState(wrapper, mode);
    prepareActions(wrapper);
    prepareExpand(wrapper);
    wrapper.dataset.goshtosoChartWrapperInitialized = "true";
  }

  function prepareWithin(root) {
    if (!root) return;
    if (root.matches && root.matches("[data-goshtoso-chart-wrapper]")) prepareWrapper(root);
    else if (root.closest) prepareWrapper(root.closest("[data-goshtoso-chart-wrapper]"));
    if (root.querySelectorAll) {
      root.querySelectorAll("[data-goshtoso-chart-wrapper]").forEach(prepareWrapper);
    }
  }

  function cleanupWrapper(wrapper) {
    if (!wrapper || wrapper.isConnected) return;
    closeExpand(wrapper, false, false);
    clearTimeout(wrapper.__goshtosoChartResizeTimer);
    delete wrapper.__goshtosoChartResizeTimer;
    var parts = expandParts(wrapper);
    if (parts.dialog && parts.dialog.__goshtosoChartObserver) {
      parts.dialog.__goshtosoChartObserver.disconnect();
      delete parts.dialog.__goshtosoChartObserver;
      delete parts.dialog.__goshtosoChartObserved;
    }
    if (fallbackWrapper === wrapper) {
      fallbackWrapper = null;
      fullscreenTrigger = null;
      wrapper.classList.remove("goshtoso-charts-fullscreen-fallback");
    }
    if (document.fullscreenElement === wrapper && document.exitFullscreen) {
      var exit = document.exitFullscreen();
      if (exit && typeof exit.catch === "function") exit.catch(function () {});
    }
    if (fullscreenTrigger && wrapper.contains(fullscreenTrigger)) fullscreenTrigger = null;
    delete wrapper.__goshtosoChartExpandReturnFocus;
    delete wrapper.dataset.goshtosoChartWrapperInitialized;
  }

  function observeWrappers() {
    if (lifecycleObserver || !document.documentElement) return;
    lifecycleObserver = new MutationObserver(function (records) {
      records.forEach(function (record) {
        record.addedNodes.forEach(function (node) {
          if (node.nodeType === Node.ELEMENT_NODE) prepareWithin(node);
        });
        record.removedNodes.forEach(function (node) {
          if (node.nodeType !== Node.ELEMENT_NODE) return;
          var wrappers = [];
          if (node.matches && node.matches("[data-goshtoso-chart-wrapper]")) wrappers.push(node);
          if (node.querySelectorAll) {
            wrappers = wrappers.concat(Array.from(node.querySelectorAll("[data-goshtoso-chart-wrapper]")));
          }
          wrappers.forEach(cleanupWrapper);
        });
      });
    });
    lifecycleObserver.observe(document.documentElement, { childList: true, subtree: true });
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
    var transparent = wrapper.dataset.goshtosoChartExportBackground === "transparent";
    var request = {
      format: "png",
      pixelRatio: Number(wrapper.dataset.goshtosoChartExportPixelRatio) || 1,
      backgroundColor: transparent ? "rgba(0,0,0,0)" : surfaceColor(wrapper),
      dataURL: ""
    };
    wrapper.dispatchEvent(new CustomEvent("goshtoso-charts:export-request", {
      bubbles: true,
      detail: request
    }));
    if (!request.dataURL) throw new Error("Live chart export is unavailable.");
    var response = await fetch(request.dataURL);
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
    if (!wrapper || wrapperMode(wrapper) !== "enabled") return;
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
    var target = event.target;
    if (!target || typeof target.closest !== "function") return;
    var control = target.closest("[data-goshtoso-chart-control]");
    if (control) {
      if (control.dataset.goshtosoChartControl === "fullscreen") toggleFullscreen(control);
      return;
    }
    var exporter = target.closest("[data-goshtoso-chart-export]");
    if (exporter) exportChart(exporter);
    var expand = target.closest("[data-goshtoso-chart-expand]");
    if (expand) {
      var wrapper = wrapperFor(expand);
      var parts = expandParts(wrapper);
      if (target.closest("button") === parts.trigger) {
        prepareExpand(wrapper);
        requestAnimationFrame(function () { openExpand(wrapper); });
        setTimeout(function () { openExpand(wrapper); }, 50);
      }
    }
  });

  document.addEventListener("goshtoso-charts:set-wrapper-mode", function (event) {
    var wrapper = wrapperFor(event.target);
    var detail = event.detail || {};
    if (!wrapper) return;
    applyWrapperMode(wrapper, detail.mode, detail.focusReturn, true);
  });

  document.addEventListener("htmx:load", function (event) {
    prepareWithin(event.detail && event.detail.elt ? event.detail.elt : event.target);
  });

  document.addEventListener("htmx:afterSwap", function (event) {
    prepareWithin(event.detail && event.detail.target ? event.detail.target : event.target);
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
    if (event.key !== "Escape") return;
    if (event.target && typeof event.target.closest === "function") {
      var menu = event.target.closest('[role="menu"]');
      var trigger = menu && menu.parentElement && menu.parentElement.querySelector(":scope > button");
      if (trigger) setTimeout(function () { trigger.focus(); }, 50);
    }
    if (fallbackWrapper) leaveFallback(true);
  });

  window.__goshtosoChartsControls = {
    safeFilename: safeFilename,
    dimensions: dimensions,
    exportFromMenu: function (element, format) { exportChart(element, format); },
    expandFromMenu: expandFromMenu,
    toggleFullscreen: toggleFullscreen,
    setWrapperMode: function (wrapper, mode, focusReturn) {
      return applyWrapperMode(wrapper, mode, focusReturn, true);
    }
  };

  prepareWithin(document);
  observeWrappers();
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", function () { prepareWithin(document); }, { once: true });
  }
}());
