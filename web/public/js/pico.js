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


/*!
 * Snackbar
 * An addon for Pico.css to have toast-like functionality
 *
 * Gary Stone
 * Copyright 2026 - Licensed under MIT
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

// Custom Sidebar menu
function toggleMenu() {
  const sidebar = document.getElementById('sidebar');
  const body = document.body;
  sidebar.classList.toggle('is-active');
  body.classList.toggle('menu-open');
}

document.addEventListener("click", (event) => {
    const sidebar = document.getElementById("sidebar");

    if (sidebar && sidebar.classList.contains("is-active")) {
        const isClickInsideSidebar = sidebar.contains(event.target);
        const isToggleClick = event.target.closest(".sidebar-toggle");

        if (!isClickInsideSidebar && !isToggleClick) {
            sidebar.classList.remove("is-active");
            document.body.classList.remove("menu-open");
        }
    }
});

document.addEventListener("keydown", (event) => {
    const sidebar = document.getElementById("sidebar");
    if (event.key === "Escape" && sidebar && sidebar.classList.contains("is-active")) {
        sidebar.classList.remove("is-active");
        document.body.classList.remove("menu-open");
    }
});
