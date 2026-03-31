// snackbar.js

const Snackbar = (() => {
  const container = document.getElementById("snackbar-container");
  const queue = [];
  let active = false;

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

    container.appendChild(el);

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




$(document).ready(function () {

    Snackbar.push({
    message: "Saved successfully!",
    type: "success"
    });

    Snackbar.push({
    message: "Item deleted",
    actionText: "Undo",
    onAction: () => {
        console.log("Undo clicked");
    }
    });

    Snackbar.push({
    message: "Failed to save",
    type: "error"
    });

    Snackbar.push({
    message: "Low disk space",
    type: "warning"
    });

    Snackbar.push({
    message: "Item deleted",
    actionText: "Undo",
    onAction: () => {
        console.log("Undo clicked");
    }
    });

    Snackbar.push({
    message: "Uploading file...",
    duration: 6000
    });
});