/*
 * Modal
 *
 * Pico.css - https://picocss.com
 * Copyright 2019-2024 - Licensed under MIT
 */

// Config
const isOpenClass = "modal-is-open";
const openingClass = "modal-is-opening";
const closingClass = "modal-is-closing";
const scrollbarWidthCssVar = "--pico-scrollbar-width";
const animationDuration = 400; // ms
let visibleModal = null;

// Toggle modal
const toggleModal = (event) => {
  event.preventDefault();
  const modal = document.getElementById(event.currentTarget.dataset.target);
  if (!modal) return;
  modal && (modal.open ? closeModal(modal) : openModal(modal));
};

// Open modal
const openModal = (modal) => {
  const { documentElement: html } = document;
  const scrollbarWidth = getScrollbarWidth();
  if (scrollbarWidth) {
    html.style.setProperty(scrollbarWidthCssVar, `${scrollbarWidth}px`);
  }
  html.classList.add(isOpenClass, openingClass);
  setTimeout(() => {
    visibleModal = modal;
    html.classList.remove(openingClass);
  }, animationDuration);
  modal.showModal();
};

// Close modal
const closeModal = (modal) => {
  visibleModal = null;
  const { documentElement: html } = document;
  html.classList.add(closingClass);
  setTimeout(() => {
    html.classList.remove(closingClass, isOpenClass);
    html.style.removeProperty(scrollbarWidthCssVar);
    modal.close();
  }, animationDuration);
};

// Close with a click outside
document.addEventListener("click", (event) => {
  if (visibleModal === null) return;
  const modalContent = visibleModal.querySelector("article");
  const isClickInside = modalContent.contains(event.target);
  !isClickInside && closeModal(visibleModal);
});

// Close with Esc key
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && visibleModal) {
    closeModal(visibleModal);
  }
});

// Get scrollbar width
const getScrollbarWidth = () => {
  const scrollbarWidth = window.innerWidth - document.documentElement.clientWidth;
  return scrollbarWidth;
};

// Is scrollbar visible
const isScrollbarVisible = () => {
  return document.body.scrollHeight > screen.height;
};


/*!
 * Minimal theme switcher
 *
 * Pico.css - https://picocss.com
 * Copyright 2019-2024 - Licensed under MIT
 */

const themeSwitcher = {
  // Config
  _scheme: "auto",
  menuTarget: "details.dropdown",
  buttonsTarget: "a[data-theme-switcher]",
  buttonAttribute: "data-theme-switcher",
  rootAttribute: "data-theme",
  localStorageKey: "picoPreferredColorScheme",

  // Init
  init() {
    this.scheme = this.schemeFromLocalStorage;
    this.initSwitchers();
  },

  // Get color scheme from local storage
  get schemeFromLocalStorage() {
    return window.localStorage?.getItem(this.localStorageKey) ?? this._scheme;
  },

  // Preferred color scheme
  get preferredColorScheme() {
    return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  },

  // Init switchers
  initSwitchers() {
    const buttons = document.querySelectorAll(this.buttonsTarget);
    buttons.forEach((button) => {
      button.addEventListener(
        "click",
        (event) => {
          event.preventDefault();
          // Set scheme
          this.scheme = button.getAttribute(this.buttonAttribute);
          // Close dropdown
          document.querySelector(this.menuTarget)?.removeAttribute("open");
        },
        false
      );
    });
  },

  // Set scheme
  set scheme(scheme) {
    if (scheme == "auto") {
      this._scheme = this.preferredColorScheme;
    } else if (scheme == "dark" || scheme == "light") {
      this._scheme = scheme;
    }
    this.applyScheme();
    this.schemeToLocalStorage();
  },

  // Get scheme
  get scheme() {
    return this._scheme;
  },

  // Apply scheme
  applyScheme() {
    document.querySelector("html")?.setAttribute(this.rootAttribute, this.scheme);
  },

  // Store scheme to local storage
  schemeToLocalStorage() {
    window.localStorage?.setItem(this.localStorageKey, this.scheme);
  },
};

// Init
themeSwitcher.init();


/******************************************************
 * Snackbar
 * An addon for Pico.css to have toast-like functionality
 *
 * Gary Stone
 * Copyright 2026 - Licensed under MIT
 *
 *
 *
Snackbar Usage:
  Snackbar.push({
      message: "Saved successfully!", // mandatory: message
      type: "success",            // optional: "success", "error", "warning", or "info"
      duration: 6000              // optional: milliseconds, default is 3 seconds
      actionText: "Undo",         // optional: provide button for user to click, and onAction runs when they click it
      onAction: () => {
          console.log("Undo clicked");
      }
  });
 *
 *
 */

const Snackbar = (() => {
  let container = null;
  const queue = [];
  let active = false;

  //<div id="snackbar-container" aria-live="polite" aria-atomic="true"></div>
  function getContainer() {
    if (!container) {
      container = document.getElementById("snackbar-container");
      if (!container) {
        container = document.createElement("div");
        container.id = "snackbar-container";
        container.setAttribute("aria-live", "polite");
        container.setAttribute("aria-atomic", "true");
        document.body.appendChild(container); 
      }
    }
    return container;
  }

  function showNext() {
    if (queue.length === 0) {
      active = false;
      return;
    }
    active = true;

    const { message, duration, actionText, onAction, type: rawType } = queue.shift();
    const validTypes = ["success", "error", "warning", "info"];
    const type = validTypes.includes(rawType) ? rawType : "info";
    const el = document.createElement("article");
    el.className = `snackbar ${type}`;
    el.setAttribute("role", "status");
    
    const icons = {
        success: "✔",
        error: "✖",
        warning: "⚠",
        info: "ℹ"
    };

    const icon = document.createElement("span");
    icon.textContent = icons[type] || "";
    icon.style.opacity = "0.7";
    icon.style.fontSize = "0.9rem";

    const text = document.createElement("span");
    text.textContent = message;

    // Wrap icon + text together
    const content = document.createElement("div");
    content.style.display = "flex";
    content.style.alignItems = "center";
    content.style.gap = "0.5rem";

    content.appendChild(icon);
    content.appendChild(text);

    // Add to snackbar
    el.appendChild(content);

    // Action button
    if (actionText && onAction) {
      const btn = document.createElement("button");
      btn.textContent = actionText;
      btn.onclick = () => {
        onAction();
        remove();
      };
      el.appendChild(btn);
    }

    // Close button
    const close = document.createElement("span");
    close.textContent = "✕";
    close.className = "close";
    close.onclick = remove;
    el.appendChild(close);

    getContainer().appendChild(el);

    // Animate in
    requestAnimationFrame(() => {
      el.classList.add("show");
    });

    let timeout = setTimeout(remove, duration);
    let remaining = duration;
    let start = Date.now();

    // Pause on hover
    el.addEventListener("mouseenter", () => {
      clearTimeout(timeout);
      remaining -= Date.now() - start;
    });

    el.addEventListener("mouseleave", () => {
      start = Date.now();
      timeout = setTimeout(remove, remaining);
    });

    function remove() {
      el.classList.remove("show");
      setTimeout(() => {
        el.remove();
        showNext();
      }, 300);
    }
  }

  function push(options) {
    queue.push({
      duration: 3000,
      ...options
    });

    if (!active) showNext();
  }

  return { push };
})();

// Simplified Snackbar into toast
function toast(msg, type = "info") {
    Snackbar.push({
      message: msg,
      type: type
    });
}

// // Custom Sidebar menu
// function toggleMenu() {
//   const sidebar = document.getElementById('sidebar');
//   const body = document.body;
//   sidebar.classList.toggle('is-active');
//   body.classList.toggle('menu-open');
// }

// document.addEventListener("click", (event) => {
//     const sidebar = document.getElementById("sidebar");

//     if (sidebar && sidebar.classList.contains("is-active")) {
//         const isClickInsideSidebar = sidebar.contains(event.target);
//         const isToggleClick = event.target.closest(".sidebar-toggle");

//         if (!isClickInsideSidebar && !isToggleClick) {
//             sidebar.classList.remove("is-active");
//             document.body.classList.remove("menu-open");
//         }
//     }
// });

// document.addEventListener("keydown", (event) => {
//     const sidebar = document.getElementById("sidebar");
//     if (event.key === "Escape" && sidebar && sidebar.classList.contains("is-active")) {
//         sidebar.classList.remove("is-active");
//         document.body.classList.remove("menu-open");
//     }
// });

// *********************************************************************
// Custom select list with icons and pico formatting
// Mostly from https://www.w3schools.com/howto/howto_custom_select.asp
// *********************************************************************
function initCustomSelects() {
  let x, i, j, l, ll, selElmnt, a, b, c;
/* Look for any elements with the class "custom-select": */
x = document.getElementsByClassName("custom-select");
l = x.length;
for (i = 0; i < l; i++) {
  selElmnt = x[i].getElementsByTagName("select")[0];
  if (!selElmnt) continue;
  ll = selElmnt.length;
  /* For each element, create a new DIV that will act as the selected item: */
  a = document.createElement("DIV");
  a.classList.add("select-selected");
  if (selElmnt.selectedIndex !== -1) {
    a.innerHTML = selElmnt.options[selElmnt.selectedIndex].innerHTML;
  }
  x[i].appendChild(a);
  /* For each element, create a new DIV that will contain the option list: */
  b = document.createElement("DIV");
  b.classList.add("select-items", "select-hide");
  for (j = 0; j < ll; j++) {
    /* For each option in the original select element,
    create a new DIV that will act as an option item: */
    c = document.createElement("DIV");
    c.innerHTML = selElmnt.options[j].innerHTML;
    // Store the index to avoid brittle innerHTML comparison
    c.setAttribute("data-index", j);
    c.addEventListener("click", function() {
        /* When an item is clicked, update the original select box,
        and the selected item: */
        let y, k, s, h, yl, idx;
        s = this.parentNode.parentNode.getElementsByTagName("select")[0];
        h = this.parentNode.previousSibling;
        idx = parseInt(this.getAttribute("data-index"));

        s.selectedIndex = idx;
        h.innerHTML = this.innerHTML;
        y = this.parentNode.getElementsByClassName("same-as-selected");
        yl = y.length;
        for (k = 0; k < yl; k++) {
            y[k].classList.remove("same-as-selected");
        }
        this.classList.add("same-as-selected");
        
        // Dispatch a change event so other scripts know the value changed
        s.dispatchEvent(new Event('change', { bubbles: true }));
        
        h.click();
    });
    b.appendChild(c);
  }
  x[i].appendChild(b);
  a.addEventListener("click", function(e) {
    /* When the select box is clicked, close any other select boxes,
    and open/close the current select box: */
    e.stopPropagation();
    closeAllSelect(this);
    this.nextSibling.classList.toggle("select-hide");
    this.classList.toggle("select-arrow-active");
  });
}
}

document.addEventListener("DOMContentLoaded", initCustomSelects);

function closeAllSelect(elmnt) {
  /* A function that will close all select boxes in the document,
  except the current select box: */
  var x, y, i, xl, yl, arrNo = [];
  x = document.getElementsByClassName("select-items");
  y = document.getElementsByClassName("select-selected");
  xl = x.length;
  yl = y.length;
  for (i = 0; i < yl; i++) {
    if (elmnt == y[i]) {
      arrNo.push(i)
    } else {
      y[i].classList.remove("select-arrow-active");
    }
  }
  for (i = 0; i < xl; i++) {
    if (arrNo.indexOf(i) === -1) {
      x[i].classList.add("select-hide");
    }
  }
}

/* If the user clicks anywhere outside the select box,
then close all select boxes: */
document.addEventListener("click", closeAllSelect);