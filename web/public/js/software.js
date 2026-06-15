// software.js

// Declare iconbar button variables at top level so they are available to checkValid, saveRecord, etc.
let btnSave, btnNew, btnDelete, usedNames = [];

// Cache DOM elements using getters to ensure they are available when needed
const UI = {
    form: () => document.getElementById('theForm'),
    sid: () => document.getElementById("sid"),
    name: () => document.getElementById("name"),
    licenses: () => document.getElementById("licenses"),
    license_key: () => document.getElementById("license_key"),
    product: () => document.getElementById("product"),
    source: () => document.getElementById("source"),
    link: () => document.getElementById("link"),
    notes: () => document.getElementById("notes"),
    active: () => document.getElementById("active"),
    reuseable: () => document.getElementById("reuseable"),
    filter: () =>  document.getElementById("filter"),
    purchased: () => document.getElementById("purchased"),
    inv_name: () => document.getElementById("inv_name"),
    pre_installed: () => document.getElementById("pre_installed"),
    free: () => document.getElementById("free"),
    pop: () => document.getElementById("pop"),
    actionID: () => document.getElementById("actionID"),
    actionName: () => document.getElementById("actionName"),
    cmd: () => document.getElementById("cmd"),
    notesDialog: () => document.getElementById("NotesDialog"),
    actionLogDiv: () => document.getElementById("actionLogDiv"),
    deleteDialog: () => document.getElementById("deleteDialog"),
    softwareName: () => document.getElementById("softwareName"),
    nameError: () => document.getElementById("nameError"),
    matchDialog: () => document.getElementById("matchDialog"),
    search: () => document.getElementById("search"),
    invList: () => document.getElementById("inv_list"),
    invTable: () => document.getElementById("inv_table"),
    errorMsg: () => document.getElementById("errorMsg"),
    matchCount: () => document.getElementById("matchCount")
};

// Page loaded event
document.addEventListener('DOMContentLoaded', function() {
    const sid = UI.sid().value;
    btnSave = new Button("btnSave");
    btnNew = new Button("btnNew");
    btnDelete = new Button("btnDelete", true);

    const isNew = isDigits(sid) && txt2Int(sid) === 0;
    if (isNew) {
        btnSave.on();
        btnNew.off();
        btnDelete.off();
    } else {
        btnSave.off();
        btnNew.on();
        btnDelete.on();
    }

    const form = UI.form();
    if (form) {
        form.addEventListener("input", (e) => {
            if (e.target.id !== "filter") updateButtonStates();
        });
        form.addEventListener("change", (e) => {
            if (e.target.matches("select")) updateButtonStates();
            if (e.target.id === "name") onNameChange();
        });
    }
});

function updateButtonStates() {
    btnSave.on();
    btnNew.off();
    btnDelete.off();
}

async function onNameChange() {
    const sendData = getFormData();
    sendData.task = "unique";
    try {
        const reply = await postForm("software", sendData);
        const data = typeof reply === "string" ? JSON.parse(reply) : reply;
        const nameInput = UI.name();
        const nameError = UI.nameError();
        if (nameInput && nameError) {
            if (data.success) {
                nameInput.setAttribute("aria-invalid") === "true"
                nameError.style.display = "none";
            } else {
                nameInput.setAttribute("aria-invalid") === "false"
                nameError.value = data.msg;
                nameError.style.display = "block";
            }
        }
    } catch (e) {
        console.error("Uniqueness check failed:", e);
    }
}

function pop(aid) {
    const notesEl = document.getElementById("notes" + aid);
    const aidInput = document.getElementById("aid" + aid);

    if (UI.pop() && notesEl) {
        UI.pop().innerHTML = `<p>${notesEl.innerHTML}</p>`;
    }
    if (UI.actionID()) UI.actionID().value = aid;
    
    if (aidInput) {
        const settings = JSON.parse(aidInput.value);
        if (UI.actionName()) UI.actionName().value = settings.action;
        if (UI.cmd()) {
            setDisplay(UI.cmd(), !!(settings.active && !settings.sid_ack));
        }
    }
    openModal(UI.notesDialog());
}

function acceptAction() {
    const aidEl = UI.actionID();
    fetchLog(aidEl ? txt2Int(aidEl.value) : 0);
}

async function fetchLog(aid = 0) {
    const sendData = getFormData();
    if (sendData.sid === 0) return;
    sendData.task = "getactionlog";
    sendData.aid = aid;
    try {
        const html = await postForm("software", sendData);
        if (UI.actionLogDiv()) UI.actionLogDiv().innerHTML = html;
        buildTable("actionlog");
    } catch (e) {
        toast("fetchLog failed: " + e, "error");
    }
}

function deleteRecord() {
    if (btnDelete.state !== "on") return;
    openModal(UI.deleteDialog());
    const name = UI.name()?.value;
    if (name) UI.softwareName().innerHTML = name;
}

async function confirmDelete() {
    if (btnDelete.state !== "on") return;
    const sendData = getFormData();
    sendData.task = "delete";
    try {
        const reply = await postForm("software", sendData);
        const data = typeof reply === "string" ? JSON.parse(reply) : reply;
        if (data.success) {
            addRecord();
        } else {
            closeModal(UI.deleteDialog());
            toast(data.msg, "alert");
        }
    } catch (e) {
        console.error(e);
        toast(e, "error");
    }
}


function addRecord() {
    if (btnNew.state !== "on" && txt2Int(UI.sid().value) !== 0) return;
    const url = new URL(window.location.href);
    url.searchParams.set("sid", "0");
    window.location.href = encodeURI(url.toString());
}

function validateForm(data) {
    if (!isDigits(data.sid.toString())) return false;    
    const nameInput = UI.name();
    // Check aria-invalid of nameInput, it may have been set by function onNameChange()
    if (nameInput && nameInput.getAttribute("aria-invalid") === "true") {
        return false;
    }
    // Check validity of nameInput
    if (nameInput && !nameInput.checkValidity()) {
        nameInput.setAttribute("aria-invalid", "true");
        return false;
    } else {
        nameInput.setAttribute("aria-invalid", "false");
    }

    if (data.link.length > 0) {
        const link = UI.link();
        if (link && !link.checkValidity()) return false;
    }
    
    // Check all the input fields if valid and set aria-invalid for each
    const form = UI.form();
    if (form) {
        Array.from(form.elements).forEach(el => {
            if (el.willValidate) {
                el.setAttribute("aria-invalid", el.checkValidity() ? "false" : "true");
            }
        });
    }

    return form ? form.checkValidity() : false;
}

async function saveRecord() {
    if (btnSave.state !== "on") return false;
    const sendData = getFormData();
    if (!validateForm(sendData)) return false;
    if (sendData.sid === 0) sendData.task = "add";

    try {
        const reply = await postForm("software", sendData);
        const data = typeof reply === "string" ? JSON.parse(reply) : reply;
        if (!data.success) {
            toast(data.msg, "alert");
        } else {
            const url = new URL(window.location.href);
            url.searchParams.set("sid", data.sid);
            window.location.href = encodeURI(url.toString());
        }
    } catch (e) {
        console.error("Save failed:", e);
        toast("Save failed", "error");
    }
    return false;
}

function getFormData() {
    return {
        task: "save",
        sid: txt2Int(UI.sid()?.value),
        name: UI.name()?.value.trim() || "",
        licenses: txt2Int(UI.licenses()?.value),
        license_key: UI.license_key()?.value.trim() || "",
        product: UI.product()?.value.trim() || "",
        source: UI.source()?.value.trim() || "",
        link: UI.link()?.value.trim() || "",
        notes: UI.notes()?.value.trim() || "",
        active: UI.active()?.value || "0",
        reuseable: UI.reuseable()?.checked ? 1 : 0,
        showhistory: UI.filter()?.checked ? 1 : 0,
        purchased: UI.purchased()?.value || "",
        inv_name: UI.inv_name()?.value || "",
        pre_installed: txt2Int(UI.pre_installed()?.value),
        free: UI.free()?.checked ? 1 : 0
    };
}

async function popDialog() {
    const invName = UI.inv_name()?.value || "";
    if (UI.search()) UI.search().value = invName;

    //If already have the list, just open popup
    const invList = UI.invList();
    if (invList && invList.children.length > 0) {
      openModal(UI.matchDialog());
      filterList();
      return;
    }
    
    try {
        const reply = await postForm("inventory", { task: "get_software_inventory" });
        const data = typeof reply === "string" ? JSON.parse(reply) : reply;
        if (data.success) {
            usedNames = data.used_names.filter(item => item !== invName);
            if (invList) invList.innerHTML = data.inv_table;
            openModal(UI.matchDialog());
            filterList();
        }
    } catch (e) {
        console.error("Inventory fetch failed:", e);
    }
}
  
//Search the used names list for any matches
function closeDialog() {
        const newName = UI.search()?.value || "";
    if (newName && isMatch(newName)) {
        if (UI.errorMsg()) UI.errorMsg().innerHTML = "Sorry, already in use on another software record.";
        return;
    }
    if (UI.inv_name()) UI.inv_name().value = newName; 
    closeModal(UI.matchDialog());
    updateButtonStates();
}

function isMatch(name) {
    return usedNames.some(used => name.startsWith(used) || used.startsWith(name));
}

function fillSearch(newName) {
    if (UI.search()) UI.search().value = newName;
    filterList();
}
  
function filterList() {
    const filter = UI.search()?.value || "";
    if (UI.errorMsg()) UI.errorMsg().innerHTML = isMatch(filter) ? "Sorry, already in use." : "";
    let matchCount = 0;
        const invTable = UI.invTable();
    if (!invTable) return;
       Array.from(invTable.getElementsByTagName("tr")).forEach(row => {
        const td = row.cells[0];
        if (td) {
            const txtValue = td.textContent || td.innerText;
            const matches = txtValue.trim().indexOf(filter) === 0;
            row.style.display = matches ? "table-row" : "none";
            if (matches) matchCount++;
        }
    });
    if (UI.matchCount()) UI.matchCount().innerHTML = `${matchCount} matches`;
}
