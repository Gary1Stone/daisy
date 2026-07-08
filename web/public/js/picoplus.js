//picoplus.js

const setDisplay = (el, show) => { if (el) el.style.display = show ? "block" : "none"; };
const isDigits = (value) => typeof value === "string" && value.length > 0 ? /^\d+$/.test(value) : true;
const txt2Int = (value) => {
    const result = parseInt(value, 10);
    return isNaN(result) ? 0 : result;
};

// Define an Icon Button class to encapsulate button behavior
// This allows for easy addition of more buttons in the future
class Button {
    constructor(btnId, forceOffIfNotAllowed = false) {
        this.btnId = btnId;
        this.forceOffIfNotAllowed = forceOffIfNotAllowed;
        this.state = "on";
    }
    on() {
        const btn = document.getElementById(this.btnId);
        if (btn && btn.dataset.allowed === "1") {
            setDisplay(btn, true);
            this.state = "on";
        } else if (this.forceOffIfNotAllowed) {
            this.off();
        }
    }
    off() {
        setDisplay(document.getElementById(this.btnId), false);
        this.state = "off";
    }
}

document.addEventListener("DOMContentLoaded", () => {
  themeSetter();
  setUpDropDowns();
});

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


/*
 * Theme Setter
 *
 * Pico.css - https://picocss.com
 * Copyright 2019-2024 - Licensed under MIT
 */

// Config
function themeSetter() {
  const themeQuery = window.matchMedia('(prefers-color-scheme: dark)');
  const applyTheme = () => {
    const stored = window.localStorage?.getItem("picoPreferredColorScheme") || 'auto';
    const theme = stored === 'auto' ? (themeQuery.matches ? 'dark' : 'light') : stored;
    document.documentElement.setAttribute('data-theme', theme);
  };
  applyTheme();
  // Automatically update the theme if system preferences change while the page is open
  themeQuery.addEventListener('change', applyTheme);
}


/*
 * Snackbar
 * An addon for Pico.css to have toast-like functionality
 *
 * Gary Stone
 * Copyright 2026 - Licensed under MIT
 *
 * Usage:
 *   Snackbar.push({
 *       message: "Saved successfully!", // mandatory: message
 *       type: "success",            // optional: "success", "error", "warning", or "info"
 *       duration: 6000              // optional: milliseconds, default is 3 seconds
 *       actionText: "Undo",         // optional: provide button for user to click, and onAction runs when they click it
 *       onAction: () => {
 *           console.log("Undo clicked");
 *       }
 *   });
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
    el.style.pointerEvents = "auto";
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


/*
 * Icon/Color DropDowns with Pico.css
 * 
 *
 * Gary Stone
 * Copyright 2026 - Licensed under MIT
 *
 */

function setUpDropDowns(root = document) {
    const scope = (typeof root === 'string') ? document.getElementById(root) : root;
    if (!scope) return;

    // Scope all logic within a loop for multiple drop-list instances
    scope.querySelectorAll('details[role="list"]:not([data-initialized])').forEach(dropdown => {
        dropdown.setAttribute('data-initialized', 'true');
        const container = dropdown.closest('.custom-select-container') || dropdown.parentElement;
        const hiddenInput = container.querySelector('.droplist-input');
        const summary = dropdown.querySelector('summary');
        if (!hiddenInput || !summary) return;
        const options = Array.from(dropdown.querySelectorAll('ul[role="listbox"] a'));
        let activeIndex = -1;

        const isDisabled = () => summary.getAttribute('aria-disabled') === 'true';

        // Handle Readonly (Disabled) state by blocking the toggle
        summary.addEventListener('click', (e) => { if (isDisabled()) e.preventDefault(); });

        // Handle Required state validation propagation
        hiddenInput.addEventListener('invalid', () => summary.setAttribute('aria-invalid', 'true'));

        // 1. Initialize Default Value for this specific instance (can have many dropdowns on same page, seperated by custom-select-container)
        function initDefault() {
            const defaultVal = hiddenInput.value;
            if (defaultVal) {
                const match = options.find(opt => opt.getAttribute('data-value') === defaultVal);
                if (match) selectOption(match);
            }
        }

        // 2. Selection Logic
        function selectOption(optionEl) {
            options.forEach(opt => opt.removeAttribute('aria-selected'));

            optionEl.setAttribute('aria-selected', 'true');
            summary.innerHTML = optionEl.innerHTML;
            hiddenInput.value = optionEl.getAttribute('data-value');
            dropdown.removeAttribute('open');
            summary.setAttribute('aria-invalid', 'false'); // Clear validation error on selection
            hiddenInput.dispatchEvent(new Event('change', { bubbles: true })); // Dispatch change event
            summary.focus();
        }

        // 3. Click Listeners
        options.forEach(option => {
            option.addEventListener('click', (e) => {
                e.preventDefault();
                selectOption(option);
            });
        });

        // 4. Keyboard Navigation (scoped strictly to this dropdown)
        dropdown.addEventListener('keydown', (e) => {
            const isOpen = dropdown.hasAttribute('open');

            if (!isOpen) {
                if ((e.key === 'ArrowDown' || e.key === 'Enter' || e.key === ' ') && !isDisabled()) {
                    e.preventDefault();
                    dropdown.setAttribute('open', '');
                    activeIndex = options.findIndex(opt => opt.getAttribute('aria-selected') === 'true');
                    if (activeIndex === -1) activeIndex = 0;
                    options[activeIndex].focus();
                }
                return;
            }

            if (e.key === 'Escape') {
                e.preventDefault();
                dropdown.removeAttribute('open');
                summary.focus();
            } else if (e.key === 'ArrowDown') {
                e.preventDefault();
                activeIndex = (activeIndex + 1) % options.length;
                options[activeIndex].focus();
            } else if (e.key === 'ArrowUp') {
                e.preventDefault();
                activeIndex = (activeIndex - 1 + options.length) % options.length;
                options[activeIndex].focus();
            } else if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                if (document.activeElement.tagName === 'A') {
                    selectOption(document.activeElement);
                }
            }
        });

        // Run initializer for each specific component
        initDefault();
    });
}

function checkDropdownValid(el) {
    const isValid = el.checkValidity(); // may have a blank entry, and is required
    el.setAttribute("aria-invalid", !isValid);
    const summary = el.closest('.custom-select-container')?.querySelector('summary');
    if (summary) summary.setAttribute("aria-invalid", !isValid);
    if (!isValid) {
      btnNew.off();
      btnDelete.off();
    }
}
