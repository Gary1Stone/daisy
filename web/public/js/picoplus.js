//picoplus.js
document.addEventListener("DOMContentLoaded", () => initCustomSelects());

const txt2Int = (value) => {
    const result = parseInt(value, 10);
    return isNaN(result) ? 0 : result;
};

const isDigits = (value) => typeof value === "string" && value.length > 0 ? /^\d+$/.test(value) : true;

const setDisplay = (el, show) => { if (el) el.style.display = show ? "block" : "none"; };

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

/******************************************************
 * Theme Setter
 ******************************************************/
document.addEventListener('DOMContentLoaded', () => {
  const themeQuery = window.matchMedia('(prefers-color-scheme: dark)');
  
  const applyTheme = () => {
    const stored = window.localStorage?.getItem("picoPreferredColorScheme") || 'auto';
    const theme = stored === 'auto' ? (themeQuery.matches ? 'dark' : 'light') : stored;
    document.documentElement.setAttribute('data-theme', theme);
  };

  applyTheme();
  // Automatically update the theme if system preferences change while the page is open
  themeQuery.addEventListener('change', applyTheme);
});

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

// *********************************************************************
// Custom select list with icons and pico formatting
// Mostly from https://www.w3schools.com/howto/howto_custom_select.asp
// *********************************************************************
function initCustomSelects(namedElementID = null) {
  let x, i, j, l, ll, selElmnt, a, b, c;
/* Look for any elements with the class "custom-select": */
if (namedElementID) {
  const container = document.getElementById(namedElementID);
  x = container ? container.getElementsByClassName("custom-select") : [];
} else {
  x = document.getElementsByClassName("custom-select");
} 
l = x.length;
for (i = 0; i < l; i++) {
  selElmnt = x[i].getElementsByTagName("select")[0];
  if (!selElmnt) continue;
  ll = selElmnt.length;
  
  // For each element, create a new DIV that will act as the selected item:
  a = document.createElement("DIV");
  a.classList.add("select-selected");
  a.style.alignItems = "center";
  a.style.gap = "0.5rem"; // Space between icon and text

  let selectedContent = "";
  let initialColor = "";

  if (selElmnt.selectedIndex !== -1) {
    selectedContent = selElmnt.options[selElmnt.selectedIndex].innerHTML;
    initialColor = selElmnt.options[selElmnt.selectedIndex].getAttribute("data-color");
  } else if (selElmnt.options.length > 0) {
    // If no item is selected, but there are options, display the first option's innerHTML
    selectedContent = selElmnt.options[0].innerHTML;
    initialColor = selElmnt.options[0].getAttribute("data-color");
  }

  if (selectedContent) {
    const tempDiv = document.createElement('div');
    tempDiv.innerHTML = selectedContent;

    const svgElement = tempDiv.querySelector('svg');
    let textNode = tempDiv.textContent.trim();

    // If an SVG was found, remove its outerHTML from the text content
    if (svgElement) {
        textNode = textNode.replace(svgElement.textContent, '').trim(); // Remove SVG's internal text content
        svgElement.style.flexShrink = "0"; // Prevent SVG from shrinking
        a.appendChild(svgElement);
    }

    const textSpan = document.createElement('span');
    textSpan.textContent = textNode;
    textSpan.style.whiteSpace = "nowrap";
    textSpan.style.overflow = "hidden";
    textSpan.style.textOverflow = "ellipsis";
    textSpan.style.flexGrow = "1"; // Allow text to grow and shrink
    textSpan.style.minWidth = "0"; // Important for ellipsis in flex containers
    a.appendChild(textSpan);

    if (initialColor) {
        // Apply color to the text span, not the whole div, to avoid coloring the SVG unless intended
        textSpan.style.color = initialColor;
    }
  }

  // Replicate aria-invalid behavior for Pico.css styling
  // We use a closure to capture the specific elements for this iteration
  const currentSelect = selElmnt;
  const currentDisplay = a;
  const syncInvalid = () => {
    const val = currentSelect.getAttribute("aria-invalid");
    if (val) currentDisplay.setAttribute("aria-invalid", val);
    else currentDisplay.removeAttribute("aria-invalid");
  };

  syncInvalid(); // Set initial state

  // Watch for dynamic validation changes (e.g. from checkValidity calls)
  new MutationObserver((mutations) => {
    mutations.forEach(m => m.attributeName === "aria-invalid" && syncInvalid());
  }).observe(currentSelect, { attributes: true });

  x[i].appendChild(a);

  /* For each element, create a new DIV that will contain the option list: */
  b = document.createElement("DIV");
  b.classList.add("select-items", "select-hide");
  for (j = 0; j < ll; j++) {
    /* For each option in the original select element, create a new DIV that will act as an option item: */
    c = document.createElement("DIV");
    c.style.display = "flex";
    c.style.alignItems = "center";
    c.style.gap = "0.5rem"; // Space between icon and text

    const optionContent = selElmnt.options[j].innerHTML;
    const itemColor = selElmnt.options[j].getAttribute("data-color");

    const tempDiv = document.createElement('div');
    tempDiv.innerHTML = optionContent;

    const svgElement = tempDiv.querySelector('svg');
    let textNode = tempDiv.textContent.trim();

    if (svgElement) {
        textNode = textNode.replace(svgElement.textContent, '').trim();
        svgElement.style.flexShrink = "0";
        c.appendChild(svgElement);
    }

    const textSpan = document.createElement('span');
    textSpan.textContent = textNode;
    textSpan.style.whiteSpace = "nowrap";
    textSpan.style.overflow = "hidden";
    textSpan.style.textOverflow = "ellipsis";
    textSpan.style.flexGrow = "1";
    textSpan.style.minWidth = "0";
    c.appendChild(textSpan);

    // Store the index to avoid brittle innerHTML comparison
    c.setAttribute("data-index", j);
    if (itemColor) {
        textSpan.style.color = itemColor;
    }

    c.addEventListener("click", function() {
        /* When an item is clicked, update the original select box, and the selected item: */
        let y, k, s, h, yl, idx;
        s = this.parentNode.parentNode.getElementsByTagName("select")[0];
        h = this.parentNode.previousSibling;

        idx = parseInt(this.getAttribute("data-index"));

        s.selectedIndex = idx;
        // Update the selected display with the new icon and text
        h.innerHTML = ''; // Clear existing content
        Array.from(this.children).forEach(child => h.appendChild(child.cloneNode(true)));
        h.querySelector('span').style.color = this.querySelector('span').style.color;

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
    /* When the select box is clicked, close any other select boxes, and open/close the current select box: */
    e.stopPropagation();
    closeAllSelect(this);
    this.nextSibling.classList.toggle("select-hide");
    this.classList.toggle("select-arrow-active");
  });
}
}


function closeAllSelect(elmnt) {
  /* A function that will close all select boxes in the document, except the current select box: */
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

/* If the user clicks anywhere outside the select box, then close all select boxes: */
document.addEventListener("click", closeAllSelect);
