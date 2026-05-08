/* global Metro, txt2Int, toast, postJSON, htmx */
const isEmailValid = /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;

// Cache DOM elements using getters to ensure they are retrieved when needed
const UI = {
    form: () => document.getElementById('theForm'),
    uid: () => document.getElementById("uid"),
    user: () => document.getElementById("user"),
    userError: () => document.getElementById("userError"),
    canSave: () => document.getElementById("canSave"),
    canNew: () => document.getElementById("canNew"),
    canDelete: () => document.getElementById("canDelete"),
    btnSave: () => document.getElementById("btnSave"),
    btnNew: () => document.getElementById("btnNew"),
    btnDelete: () => document.getElementById("btnDelete"),
    curUid: () => document.getElementById("curUid"),
    userName: () => document.getElementById("userName"),
    alerts: () => document.getElementById("alerts")
};

const setDisplay = (el, show) => { if (el) el.style.display = show ? "block" : "none"; };
const isDigits = (value) => typeof value === "string" && value.length > 0 ? /^\d+$/.test(value) : true;

// Page loaded event
document.addEventListener('DOMContentLoaded', function() {
    const form = UI.form();
    if (form) {
        form.addEventListener('submit', (event) => {
        event.preventDefault();
        saveRecord(event);
    });
    }

    const uid = UI.uid().value;
    //Set initial button state depending if a record is displayed or not
    if (isDigits(uid) && txt2Int(uid) === 0) {
        btnSave.on();  //No record, cannot save, but if they fill it out, we want to save
        btnNew.off(); //No point in showing the same screen again
        btnDelete.off(); //No record, nothing to delete
    } else {
        btnSave.off(); //Record just displayed, nothing to save, user has to make changes first
        btnNew.on(); //Record displayed, can create new
        btnDelete.on(); //record displayed, can delete it
    }

    //if any of the 'input' elements are modified, change the save/add/delete states   
    document.querySelectorAll("input").forEach(el => {
        el.addEventListener("input", () => { btnSave.on(); btnNew.off(); btnDelete.off(); });
    });

    // if any of the 'select' droplists are modified, change the save/add/delete states
    document.querySelectorAll("select").forEach(el => {
        el.addEventListener("change", () => { btnSave.on(); btnNew.off(); btnDelete.off(); });
    });
    
    // when user changes the email of the user, check if the email is not already in use
    const userEl = UI.user();
    if (userEl) userEl.addEventListener("blur", handleUserBlur);
});

async function handleUserBlur() {
    const sendData = getFormData();
    sendData.task = "unique";
    const userEl = UI.user();
    const errorEl = UI.userError();

    if (!userEl.checkValidity() || !isEmailValid.test(sendData.user)) {
        errorEl.value = "ERROR: User ID must be an email address";
        setDisplay(errorEl, true);
        return;
    }

    try {
        await postJSON("profile", sendData, (reply) => {
            if (reply.success) {
                setDisplay(errorEl, false);
            } else {
                errorEl.value = reply.msg;
                setDisplay(errorEl, true);
            }
        });
    } catch (error) {
        console.error("Uniqueness check failed:", error);
    }
}

function getPersonCtrl() { return; }

const btnSave = {
    state: "on",
    on() {
        if (UI.canSave().value === "1") { setDisplay(UI.btnSave(), true); this.state = "on"; }
    },
    off() { setDisplay(UI.btnSave(), false); this.state = "off"; }
};

const btnNew = {
    state: "on",
    on() {
        if (UI.canNew().value === "1") { setDisplay(UI.btnNew(), true); this.state = "on"; }
    },
    off() { setDisplay(UI.btnNew(), false); this.state = "off"; }
};

const btnDelete = {
    state: "on",
    on() {
        if (UI.canDelete().value === "1") { setDisplay(UI.btnDelete(), true); this.state = "on"; }
        else { this.off(); }
    },
    off() { setDisplay(UI.btnDelete(), false); this.state = "off"; }
};

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
                addRecord(); // clears the displayed record
            } else {
                console.error(reply.msg);
                toast(reply.msg, "alert");
            }
        });
    } catch (error) {
        console.error("Delete failed:", error);
    }
}

// Adding a record is a two step process
// Display this screen with a uid=0 (user ID = UID)
// when user presses save, in the servlet, detect if record id (UID) is 0, then insert record.
// then send the uid to be used inside this form
function addRecord() {
    if (btnNew.state !== "on") return;
    let url = window.location.href;
    const i = url.indexOf("?");
    if (i < 0) {
        url = url + "?uid=" + encodeURIComponent("0");
    } else {
        url = url.substring(0, i) + "?uid=" + encodeURIComponent("0");
    }
    window.location.href =  encodeURI(url);
}

function validateForm(sendData) {
    if (!isDigits(sendData.uid)) return false;
    if (!UI.user().checkValidity()) return false;
    if (!document.getElementById("first").checkValidity()) return false;
    if (!document.getElementById("last").checkValidity()) return false;
    // Check if the user id is unique (onBlur sets if error message visible or not)
    if (UI.userError().style.display !== "none") {
        UI.user().focus();
        return false;
    }
    return UI.form().checkValidity();
}

function getFormData() {
    return {
        task: "save", 
        uid: txt2Int(UI.uid().value), 
        user: UI.user().value, 
        first: document.getElementById("first").value, 
        last: document.getElementById("last").value, 
        gid: txt2Int(document.getElementById("gid").value),
        geo_fence: document.getElementById("geo_fence").value, 
        geo_radius: txt2Int(document.getElementById("geo_radius").value),
        pwd_reset: document.getElementById("pwd_reset").value, 
        color: document.getElementById("color").value, 
        active: document.getElementById("active").checked ? 1 : 0,
        notify: document.getElementById("notify").checked ? 1 : 0
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
        console.error("Unban failed:", error);
    }
}

async function ackAlert(aid = 0) {
    const sendData = {
        task: "get_alerts", 
        aid: aid, 
        uid: UI.uid().value
    };
    try {
        await htmx("home", sendData, "alerts");
    } catch (error) {
        console.error("Alert acknowledgment failed:", error);
    }
}

async function saveRecord(event) {
    if (btnSave.state !== "on") return;
    const sendData = getFormData();
    if (!validateForm(sendData)) return;

    if (sendData.uid === 0) { sendData.task = "add"; }

    // Get the button that triggered the submit event.
    const submitButton = event.submitter;

    // --- Show loading state on the submit button ---
    submitButton.setAttribute('aria-busy', 'true');
    submitButton.disabled = true;
    
    // If the user changed their own name, update the Menubar label
    if (UI.curUid().value === String(sendData.uid)) {
        UI.userName().value = `${sendData.first} ${sendData.last}`;
    }    

    try {
        await postJSON("profile", sendData, (reply) => {
            submitButton.setAttribute('aria-busy', 'false');
            submitButton.disabled = false;

        if (!reply.success) {
                console.error(reply.msg);
            toast(reply.msg);
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
        console.error("Save failed:", error);
        submitButton.setAttribute('aria-busy', 'false');
        submitButton.disabled = false;
    }
}
