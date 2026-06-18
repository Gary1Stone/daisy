// device.js

// Declare iconbar button variables at top level so they are available to checkValid, saveRecord, etc.
let btnSave, btnNew, btnDelete;
let gblOldColor = "";   //used to remember the icon's old color if the user toggles the Died status over and over
let oldImageName = "";  //remember what the previous picture was, in case user cancels

// Cache DOM elements using getters to ensure they are available when needed
const UI = {
    form: () => document.getElementById('theForm'),
    name: () => document.getElementById("name"),
    active: () => document.getElementById("active"),
    speed: () => document.getElementById("speed"),
    cid: () => document.getElementById("cid"),
    color: () => document.getElementById("color"),
    curUid: () => document.getElementById("curUid"),
    image: () => document.getElementById("image"),
    type: () => document.getElementById("type"),
    status: () => document.getElementById("status"),
    site: () => document.getElementById("site"),
    office: () => document.getElementById("office"),
    location: () => document.getElementById("location"),
    year: () => document.getElementById("year"),
    make: () => document.getElementById("make"),
    model: () => document.getElementById("model"),
    cpu: () => document.getElementById("cpu"),
    cores: () => document.getElementById("cores"),
    ram: () => document.getElementById("ram"),
    drivetype: () => document.getElementById("drivetype"),
    drivesize: () => document.getElementById("drivesize"),
    cd: () => document.getElementById("cd"),
    wifi: () => document.getElementById("wifi"),
    ethernet: () => document.getElementById("ethernet"),
    usb: () => document.getElementById("usb"),
    gpu: () => document.getElementById("gpu"),
    notes: () => document.getElementById("notes"),
    gid: () => document.getElementById("gid"),
    uid: () => document.getElementById("uid"),
    os: () => document.getElementById("os"),
    serial_number: () => document.getElementById("serial_number"),
    asset: () => document.getElementById("asset"),
    actionID: () => document.getElementById("actionID"),
    actionName: () => document.getElementById("actionName"),
    cmd: () => document.getElementById("cmd"),
    details: () => document.getElementById("details"),
    filter: () => document.getElementById("filter"),
    actionLogDiv: () => document.getElementById("actionLogDiv"),
    deviceName: () => document.getElementById("deviceName"),
    typeErr: () => document.getElementById("typeErr"),
    uploadDialog: () => document.getElementById("uploadDialog"),
    ajaxfile: () => document.getElementById("ajaxfile"),
    notesDialog: () => document.getElementById("NotesDialog"),
    deleteDialog: () => document.getElementById("deleteDialog"),
    actionID: () => document.getElementById("actionID"),
    actionName: () => document.getElementById("actionName")
};

// Page loaded event
document.addEventListener('DOMContentLoaded', function() {
    
    // Initialize the iconbar button instances once the scripts and DOM are ready
    btnSave = new Button("btnSave");
    btnNew = new Button("btnNew");
    btnDelete = new Button("btnDelete", true);
    btnSave.on(); 

    const form = UI.form();
    if (form) {
        form.addEventListener('submit', (event) => {
            event.preventDefault();
        });
    }

    const cid = UI.cid().value;
    gblOldColor = UI.color().value;
    //Set initial button state depending if a record is displayed or not
    if (isDigits(cid) && txt2Int(cid) === 0) {
        btnNew.off(); btnDelete.off();
    } else {
        btnNew.on(); btnDelete.on();
    }

    if (form) {
        // Use event delegation for input/textarea changes
        form.addEventListener("input", (e) => {
            if (e.target.matches("input[type='text'], input[type='email'], input[type='number'], textarea")) {
                checkValid(e.target);
            }
        });
        // Handle select changes and specific logic for 'type' and dropdowns
        form.addEventListener("change", (e) => {
            if (e.target.matches("select, input[type='checkbox'], .droplist-input")) {
                if (e.target.id === "type") onTypeChange();
                if (e.target.classList.contains("droplist-input")) {
                    checkDropdownValid(e.target); // in picoplus.js
                } else {
                    checkValid(e.target);
                }
            }
        });
    }
    showHideItemsByType();
});

// If fails validation, set to off/invalid
function checkValid(el) {
    btnNew.off();
    btnDelete.off();
    const isValid = el.checkValidity();
    el.setAttribute("aria-invalid", !isValid);
}

/**************************************************/
/*         Droplist handling                      */
/**************************************************/
function getPersonCtrl() {
    getCtrl("selectPerson", ctrlData("USER"));
}

function getOfficeCtrl() {
    getCtrl("selectOffice", ctrlData("OFFICE"));
}

function ctrlData(task) {
    let droplistRequest = { 
      task: task, 
      isTicket: false,
      isWizard: false,
      cid: txt2Int(UI.cid().value),
      gid: txt2Int(UI.gid().value),
      uid: txt2Int(UI.uid().value),
      site: UI.site().value,
      office: UI.office().value,
      impact: "",
      trouble: "",
      wizard: "",
      type: UI.type().value,
      inform_gid: 0,
      isReadonly: false,
    }
    return droplistRequest;
}
/**************************************************/

async function onTypeChange() {
    showHideItemsByType();
    const sendData = getFormData();
    if (sendData.type !== "") {
        sendData.task = "get_asset_id";
        try {
            const reply = await postForm("device", sendData);
            if (reply.success) {
                UI.name().value = reply.msg;
                const assetEl = UI.asset();
                if (assetEl) assetEl.innerHTML = reply.msg;
            } else {
                toast(reply.msg || "Failed to retrieve Asset ID", "error");
            }
        } catch (e) {
            toast(e, "error");
        }
    } else {
        UI.name().value = "";
        const assetEl = UI.asset();
        if (assetEl) assetEl.innerHTML = "";
    }
}

function changePic() {
    //Make sure device has a name - Only a-z, A-Z, 0-9 underscore or dash
    let imageName = UI.name().value.toLowerCase().replace(/[^a-zA-Z0-9_-]/g, '');
    if (imageName.length < 4) {
        imageName = imageName + "DEV01";
    }
    if (imageName.length > 20) imageName = imageName.substring(0, 20);
    oldImageName = UI.image().value || "";
    let version = 1;
    const versionMatch = oldImageName.match(/-v(\d+)\.jpg$/);
    if (versionMatch) {
        version = parseInt(versionMatch[1], 10) + 1;
    } else if (oldImageName.length > 0) {
        version = 2; // Assume v1 if named but no version string found
    }
    UI.image().value = `${imageName}-v${version}.jpg`;
    openModal(UI.uploadDialog());
}

function cancelUpload() {
    UI.image().value = oldImageName;
}

function showHideItemsByType() {
    const allItems = ["computerOnly", "softwareOnly", "ethernetFlag", "wifiFlag", "usbFlag", "cdFlag", "backups"];
    const typeMap = {
        "DESKTOP": allItems,
        "LAPTOP": allItems,
        "PRINTER": ["ethernetFlag", "wifiFlag", "usbFlag"],
        "NETWORK": ["ethernetFlag", "wifiFlag"],
        "PHONE": ["ethernetFlag", "wifiFlag"]
    };
    const activeItems = typeMap[UI.type().value] || [];
    allItems.forEach(id => setDisplay(document.getElementById(id), activeItems.includes(id)));
}

// TODO: WORK NEEDED HERE
function pop(aid) {
    const popEl = document.getElementById("pop");
    const notesEl = document.getElementById("notes" + aid);
    const actionIDEl = UI.actionID();
    const actionNameEl = UI.actionName();
    const cmdEl = UI.cmd();
    const detailsEl = UI.details();
    const aidInput = document.getElementById("aid" + aid);
    if (popEl && notesEl) {
        popEl.innerHTML = "<p>" + notesEl.innerHTML + "</p>";
    }
    if (actionIDEl) actionIDEl.value = aid;
    const settings = JSON.parse(aidInput.value);
    if (actionNameEl) actionNameEl.value = settings.action;
    setDisplay(cmdEl, false);
    setDisplay(detailsEl, false);
    if (settings.active && !settings.cid_ack) { 
        setDisplay(cmdEl, true);
    }
    if (["BROKEN", "CARE", "DIED", "LOST", "REQUEST"].includes(settings.action)) {
        setDisplay(detailsEl, true);
    }
    openModal(UI.notesDialog());
}

function goTicket() {
    const aid = txt2Int(UI.actionID().value);
    window.location.href = encodeURI("ticket.html?aid=" + aid);
}

function acceptAction() {
    const aid = txt2Int(UI.actionID().value);
    fetchLog(aid);
}

async function fetchLog(aid = 0) {
    const sendData = getFormData();
    if (sendData.cid === 0) return; //Stop getting log when there is no record
    sendData.task = "getactionlog";
    sendData.aid = aid;
    const filter = UI.filter();
    if (filter && filter.checked) {
        sendData.showHistory = 1
    }
    try {
        const html = await postForm("device", sendData);
        if (html) UI.actionLogDiv().innerHTML = html;
        buildTable("actionlog");
    } catch (e) { 
        console.error(e);
        toast("Failed to fetch action log", "error");
    }
}

function deleteRecord() {
    if (btnDelete.state !== "on") return;
    openModal(UI.deleteDialog());
    const name = UI.name().value;
    if (name.length > 0) {
        UI.deviceName().innerHTML = name;
    }
}

async function confirmDelete() {
    if (btnDelete.state !== "on") return;
    let sendData = getFormData();
    sendData.task = "delete";
    try {
        const reply = await postForm("device", sendData);
        if (reply.success) {
            addRecord();  //clears the displayed record
        } else {
            closeModal(UI.deleteDialog());
            toast(reply.msg, "alert");
        }
    } catch (e) {
        console.error(e);
        toast(e, "error");
    }
}

//Adding a record is a two step process
//Display this screen with a cid=0 (ComputerID = CID)
//when user presses save, in the servlet, detect if record id (CID) is 0, then insert record.
//then send the cid to be used inside this form
function addRecord() {
    if (btnNew.state !== "on") return;
    let url = window.location.href;
    const i = url.indexOf("?");
    if (i < 0) {
        url = url + "?cid=" + encodeURIComponent("0");
    } else {
        url = url.substring(0, i) + "?cid=" + encodeURIComponent("0");
    }
    window.location.href =  encodeURI(url);
}

function validateForm(data) {
    const form = UI.form();
    const isBaseValid = form.checkValidity();
    // Set aria-invalid for all form controls based on standard validation
    Array.from(form.elements).forEach(el => {
        if (el.willValidate) {
            el.setAttribute("aria-invalid", el.checkValidity() ? "false" : "true");
        }
    });
    // Secondary logic checks
    let isTypeValid = true;
    if (UI.type().value === "") {
        isTypeValid = false;
        toast("The device type is required.", "error");
    }
    setDisplay(UI.typeErr(), !isTypeValid);
    const isRamValid = data.ram >= 0;
    UI.ram().setAttribute("aria-invalid", isRamValid ? "false" : "true");
    const isNameValid = data.name.trim().length >= 6;
    UI.name().setAttribute("aria-invalid", isNameValid ? "false" : "true");
    return isBaseValid && isTypeValid && isRamValid && isNameValid;
}

async function saveRecord() {
    btnSave.off();
    let sendData = getFormData();
    if (!validateForm(sendData)) return false;
    if (sendData.cid === 0) {
        sendData.task = "add";
    } else {
        savePreInstalled();
    }
    try {
        const reply = await postForm("device", sendData);
        if (!reply.success) {
            toast(reply.msg, "alert")
            console.log(reply.msg);  //display error message
        } else {
            let url = window.location.href;
            const i = url.indexOf("?");
            if (i < 0) {
                url = url + "?cid=" + encodeURIComponent(reply.cid);
            } else {
                url = url.substring(0, i) + "?cid=" + encodeURIComponent(reply.cid);
            }
            //need to wait some millisecnds for the server to complete creating/resizing the photo and updating the database
            setTimeout(function() {
                window.location.href = url;
            }, 200);
        }
    } catch (e) { 
        console.error(e);
        toast("Failed to save record", "error");
    }
    return false;
}

function getFormData() {
    return {
        task: "save",
        cid: txt2Int(UI.cid().value),
        name: UI.name().value.trim(),
        type: UI.type().value,
        site: UI.site().value,
        office: UI.office().value,
        location: UI.location().value.trim(),
        year: txt2Int(UI.year().value),
        make: UI.make().value,
        model: UI.model().value.trim(),
        cpu: UI.cpu().value.trim(),
        cores: txt2Int(UI.cores().value),
        ram: txt2Int(UI.ram().value),
        drivetype: UI.drivetype().value,
        drivesize: txt2Int(UI.drivesize().value),
        notes: UI.notes().value.trim(),
        gpu: UI.gpu().value.trim(),
        cd: UI.cd().checked ? 1 : 0,
        wifi: UI.wifi().checked ? 1 : 0,
        ethernet: UI.ethernet().checked ? 1 : 0,
        usb: UI.usb().checked ? 1 : 0,
        active: txt2Int(UI.active().value),
        image: UI.image().value,
        color: UI.color().value,
        speed: txt2Int(UI.speed().value),
        uid: txt2Int(UI.uid().value),
        status: UI.status().value,
        os: UI.os().value,
        serial_number: UI.serial_number().value.trim(),
        gid: txt2Int(UI.gid().value),
        aid: 0,
        showHistory: 0
    };
}

async function uploadFile() {
    let formData = new FormData();
    const fileInput = UI.ajaxfile();
    
    // Check if a file is selected
    if (!fileInput || fileInput.files.length === 0) {
      return;
    }
    if (typeof fileInput.files[0] === "undefined") {
      return;
    }
  
    // Append the file to the FormData with its new name
    formData.append("uploadfile", fileInput.files[0], UI.image().value);
  
    try {
      // Perform the file upload using fetch
      const response = await fetch("upload", {
        method: "POST",
        body: formData,
      });
  
      // Check if the upload was successful (status code 2xx)
      if (response.ok) {
        // Show the new image
        saveRecord()
      } else {
        // Handle the case where the upload was not successful
        toast("File upload failed: " + response.statusText, "alert");
      }
    } catch (error) {
      // Handle fetch errors
      toast("Error during file upload: " + error.message, "alert");
    }
  }

// Use MAP because the user can toggle the same item on and off many times
const preInstalled = new Map([]);
function setPreInstalled(id, checked) {
    const rowid = txt2Int(id);
    if (rowid > 0) {
        preInstalled.set(rowid, checked);
    }
}

function savePreInstalled() {
    if (preInstalled.size === 0) {
        return;
    }
    const items = Array.from(preInstalled).map(([id, checked]) => ({
        id,
        chk: checked ? 1 : 0
    }));

    // transmit array of objects to server
    postJSON("preinstalled", items, (reply) => {
        if (reply !== "okay") {
           console.log(reply);
        }
    });
}
