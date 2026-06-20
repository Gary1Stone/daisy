/* Profile.js */

// Declare iconbar button variables at top level so they are available to checkValid, saveRecord, etc.
let btnSave, btnNew, btnDelete;

// Cache DOM elements using getters to ensure they are available when needed
const UI = {
    form: () => document.getElementById('theForm'),
    uid: () => document.getElementById("uid"),
    user: () => document.getElementById("user"),
    first: () => document.getElementById("first"),
    last: () => document.getElementById("last"),
    gid: () => document.getElementById("gid"),
    geo_fence: () => document.getElementById("geo_fence"),
    geo_radius: () => document.getElementById("geo_radius"),
    pwd_reset: () => document.getElementById("pwd_reset"),
    color: () => document.getElementById("color"),
    active: () => document.getElementById("active"),
    notify: () => document.getElementById("notify"),
    curUid: () => document.getElementById("curUid"),
    userName: () => document.getElementById("userName"),
    alerts: () => document.getElementById("alerts")
};


// Page loaded event
document.addEventListener('DOMContentLoaded', function() {
    
    //Set initial button state depending if a record is displayed or not
    btnSave = new Button("btnSave");
    btnNew = new Button("btnNew");
    btnDelete = new Button("btnDelete", true); // true for forceOffIfNotAllowed
    btnSave.on(); 
    const uid = UI.uid().value;
    if (isDigits(uid) && txt2Int(uid) === 0) {
        btnNew.off(); btnDelete.off();
    } else {
        btnNew.on(); btnDelete.on();
    }

    // Use event delegation for all form inputs
    const form = UI.form();
    if (form) {
        form.addEventListener('submit', (event) => {
            event.preventDefault();
        });
        form.addEventListener("input", (e) => {
            if (e.target.matches("input[type='text'], input[type='email'], input[type='number']")) {
                checkValid(e.target);
            }
        });
        form.addEventListener("change", (e) => {
            if (e.target.matches("select, input[type='checkbox'], .droplist-input")) {
                if (e.target.classList.contains("droplist-input")) {
                    checkDropdownValid(e.target); // in picoplus.js
                } else {
                    checkValid(e.target);
                }
            }
        });
    }
});

function checkValid(el) {
    btnNew.off(); 
    btnDelete.off();
    // If user has spaces before or after value, reject
    if (el.value !== el.value.trim()) {
        el.setAttribute("aria-invalid", "true");
        el.setCustomValidity("Leading and trailing spaces are not allowed");
        toast("Please remove leading and trailing spaces", "warning");
        return;
    }

    // Clear any previous custom errors
    if (el.validity.customError) {
        el.setCustomValidity("");
    } 
    
    // If fails validation, set to off/invalid
    const isValid = el.checkValidity();
    el.setAttribute("aria-invalid", !isValid);
    if (!isValid) return;

    // Now have user input, check if it changed.
    if (el.id === "user" && el.value !== el.defaultValue) {
        checkUnique(el);
    }
}

async function checkUnique(el) {
    const sendData = getFormData();
    sendData.task = "unique";
    try {
        await postJSON("profile", sendData, (reply) => {
            if (reply.success) {
                el.setAttribute("aria-invalid", "false");
            } else {
                el.setAttribute("aria-invalid", "true");
                el.setCustomValidity("User ID must unique"); // This is how to set the input field to invalid
            }
            el.defaultValue = el.value;
        });
    } catch (error) {
        toast(error, "error");
        console.error("Uniqueness check failed:", error);
    }
}

function getPersonCtrl() { return; }

function deleteRecord(event) {
    if (btnDelete.state !== "on") return;
    toggleModal(event);
    const userValue = UI.user().value;
    document.getElementById("displayName").value = userValue;
}

async function deleteProfile() {
    const sendData = getFormData();
    sendData.task = "delete";
    try {
        await postJSON("profile", sendData, (reply) => {
            if (reply.success) {
                location.href='profile.html?uid=0'; // clears the displayed record
            } else {
                toast(reply.msg, "error");
                console.error(reply.msg);
            }
        });
    } catch (error) {
        toast("Delete failed:" + error, "error");
        console.error("Delete failed:", error);
    }
}

// Adding a record is a two step process
// Display this screen with a uid=0 (user ID = UID)
// when user presses save, in the servlet, detect if record id (UID) is 0, then insert record.
// then send the uid to be used inside this form
function addRecord(event) {
  location.href='profile.html?uid=0';
}

function validateForm(sendData) {
    if (!isDigits(sendData.uid)) return false;

    const userField = UI.user();
    if (!userField.checkValidity()) {
        userField.setAttribute("aria-invalid", "true");
        return false;
    }

    if (userField.getAttribute("aria-invalid") === "true") {
        userField.focus();
        return false;
    }

    const firstField = UI.first();
    if (!firstField.checkValidity())  {
        firstField.setAttribute("aria-invalid", "true");
        return false;
    }

    const lastField = UI.last();
    if (!lastField.checkValidity())  {
        lastField.setAttribute("aria-invalid", "true");
        return false;
    }
    
    return UI.form().checkValidity();
}

function getFormData() {
    return {
        task: "save", 
        uid: txt2Int(UI.uid().value), 
        user: UI.user().value.trim(), 
        first: UI.first().value.trim(), 
        last: UI.last().value.trim(),
        gid: txt2Int(UI.gid().value),
        geo_fence: UI.geo_fence().value, 
        geo_radius: txt2Int(UI.geo_radius().value),
        pwd_reset: txt2Int(UI.pwd_reset().value), 
        color: UI.color().value, 
        active: UI.active().checked ? 1 : 0,
        notify: UI.notify().checked ? 1 : 0
    };
}

async function resetBanned(UID) {
    const sendData = getFormData();
    sendData.task = "unban";
    sendData.uid = UID;
    try {
        await postJSON("profile", sendData, (reply) => {
        if (!reply.success) {
            toast(reply.msg);
        } else {
            document.getElementById("bttn").innerHTML = "";
        }
        });
    } catch (error) {
        toast("Unban failed:" + error, "error");
        console.error("Unban failed:", error);
    }
}

async function ackAlert(aid = 0) {
    const sendData = {
        task: "get_alerts", 
        aid: txt2Int(aid), 
        uid: txt2Int(UI.uid().value)
    };
    try {
        await htmx("home", sendData, "alerts");
    } catch (error) {
        toast("Alert acknowledgment failed:" + error, "error");
        console.error("Alert acknowledgment failed:", error);
    }
}

async function saveRecord(event) {
    btnSave.off();
    const sendData = getFormData();
    if (!validateForm(sendData)) return;
    if (sendData.uid === 0) { sendData.task = "add"; }
    try {
        await postJSON("profile", sendData, (reply) => {
            if (!reply.success) {
                toast(reply.msg, "error");
                console.error(reply.msg);
            } else {    // Refresh the page
                let url = window.location.href;
                const i = url.indexOf("?");
                if (i < 0) {
                    url = url + "?uid=" + encodeURIComponent(reply.uid);
                } else {
                    url = url.substring(0, i) + "?uid=" + encodeURIComponent(reply.uid);
                }
                window.location.href =  encodeURI(url);
            }
        });
    } catch (error) {
        toast("Save failed:" + error, "error");
        console.error("Save failed:", error);
    }
}
