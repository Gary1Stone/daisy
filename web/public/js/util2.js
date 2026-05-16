// util2.js

const txt2Int = (value) => {
    const result = parseInt(value, 10);
    return isNaN(result) ? 0 : result;
};

const isDigits = (value) => typeof value === "string" && value.length > 0 ? /^\d+$/.test(value) : true;

const setDisplay = (el, show) => { if (el) el.style.display = show ? "block" : "none"; };

// Define a Button class to encapsulate button behavior
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
