(function () {
  "use strict";

  function searchRoot(input) {
    return input.closest("[data-docs-search]");
  }

  function visibleItems(root) {
    return Array.from(root.querySelectorAll("[data-docs-search-item]"))
      .filter(function (item) { return !item.hidden; });
  }

  function close(root) {
    var input = root.querySelector("input");
    root.querySelector("[data-docs-search-results]").hidden = true;
    input.setAttribute("aria-expanded", "false");
    visibleItems(root).forEach(function (item) {
      item.setAttribute("aria-selected", "false");
    });
  }

  function update(input) {
    var root = searchRoot(input);
    var query = input.value.trim().toLocaleLowerCase();
    var terms = query.split(/\s+/).filter(Boolean);
    var results = root.querySelector("[data-docs-search-results]");
    var matches = 0;

    root.querySelectorAll("[data-docs-search-item]").forEach(function (item) {
      var matched = terms.length !== 0 && terms.every(function (term) {
        return item.dataset.search.includes(term);
      });
      item.hidden = !matched;
      item.setAttribute("aria-selected", "false");
      if (matched) matches += 1;
    });

    root.querySelector("[data-docs-search-empty]").hidden = query === "" || matches !== 0;
    results.hidden = query === "";
    input.setAttribute("aria-expanded", query === "" ? "false" : "true");
  }

  function select(items, index) {
    items.forEach(function (item, itemIndex) {
      item.setAttribute("aria-selected", itemIndex === index ? "true" : "false");
    });
    items[index].focus();
  }

  document.addEventListener("input", function (event) {
    if (event.target.matches("[data-docs-search] input")) update(event.target);
  });

  document.addEventListener("keydown", function (event) {
    var root = event.target.closest("[data-docs-search]");
    if (!root) return;

    var input = root.querySelector("input");
    var items = visibleItems(root);
    var current = items.indexOf(document.activeElement);

    if (event.key === "Escape") {
      close(root);
      input.focus();
      return;
    }
    if (items.length === 0) return;
    if (event.key === "ArrowDown") {
      event.preventDefault();
      select(items, current < items.length - 1 ? current + 1 : 0);
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      select(items, current > 0 ? current - 1 : items.length - 1);
    } else if (event.key === "Enter" && event.target === input) {
      event.preventDefault();
      items[0].click();
    }
  });

  document.addEventListener("click", function (event) {
    document.querySelectorAll("[data-docs-search]").forEach(function (root) {
      if (!root.contains(event.target)) close(root);
    });
  });
})();
