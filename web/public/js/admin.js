// admin.js

let building = false; // If the table is being built, do not trigger any changes to the adminData

// Global variable to hold json administration table data object from the server
// Holds data for: SITE, OFFICE, GROUP,... but only one set at a time
let adminData = [
    {id: 0, description: "One", code: "ONE", parent: "", sequence: 1, field: "LIST", active: 1, assetid: "", permissions: "", update: false, add: false, delete: false, inuse: true, task: ""},
    {id: 1, description: "Two", code: "TWO", parent: "", sequence: 2, field: "LIST", active: 1, assetid: "", permissions: "", update: false, add: false, delete: false, inuse: true, task: ""},
    {id: 2, description: "Three", code: "THREE", parent: "", sequence: 3, field: "LIST", active: 1, assetid: "", permissions: "", update: false, add: false, delete: false, inuse: true, task: ""}
];

// Keep the site data for the site droplist in office configuration
// We need a copy of the site data to build the select droplists for OFFICE table (Parent/Child)
// The solution is that SITE is the first thing displayed, meaning that this data is saved in browser
// memory for when the user selects OFFICE
// Otherwise we would have to do a second seperate fetch when OFFICE is selected
let siteData = [
    {id: 0, description: "One", code: "ONE", parent: "", sequence: 1, field: "LIST", active: 1, assetid: "", permissions: "", update: false, add: false, delete: false, inuse: true, task: ""},
    {id: 1, description: "Two", code: "TWO", parent: "", sequence: 2, field: "LIST", active: 1, assetid: "", permissions: "", update: false, add: false, delete: false, inuse: true, task: ""},
    {id: 2, description: "Three", code: "THREE", parent: "", sequence: 3, field: "LIST", active: 1, assetid: "", permissions: "", update: false, add: false, delete: false, inuse: true, task: ""}
];

//use innerHTML = 
const icons = {
    site:  `<svg height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M80-80v-481l280-119v80l200-80v120h320v480H80Zm80-80h640v-320H480v-82l-200 80v-78l-120 53v347Zm280-80h80v-160h-80v160Zm-160 0h80v-160h-80v160Zm320 0h80v-160h-80v160Zm280-320H680l40-320h120l40 320ZM160-160h640-640Z"></path></svg>`,
    office:  `<svg height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M120-120v-560h160v-160h400v320h160v400H520v-160h-80v160H120Zm80-80h80v-80h-80v80Zm0-160h80v-80h-80v80Zm0-160h80v-80h-80v80Zm160 160h80v-80h-80v80Zm0-160h80v-80h-80v80Zm0-160h80v-80h-80v80Zm160 320h80v-80h-80v80Zm0-160h80v-80h-80v80Zm0-160h80v-80h-80v80Zm160 480h80v-80h-80v80Zm0-160h80v-80h-80v80Z"></path></svg>`,
    group: `<svg height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M40-160v-112q0-34 17.5-62.5T104-378q62-31 126-46.5T360-440q66 0 130 15.5T616-378q29 15 46.5 43.5T680-272v112H40Zm720 0v-120q0-44-24.5-84.5T666-434q51 6 96 20.5t84 35.5q36 20 55 44.5t19 53.5v120H760ZM247-527q-47-47-47-113t47-113q47-47 113-47t113 47q47 47 47 113t-47 113q-47 47-113 47t-113-47Zm466 0q-47 47-113 47-11 0-28-2.5t-28-5.5q27-32 41.5-71t14.5-81q0-42-14.5-81T544-792q14-5 28-6.5t28-1.5q66 0 113 47t47 113q0 66-47 113ZM120-240h480v-32q0-11-5.5-20T580-306q-54-27-109-40.5T360-360q-56 0-111 13.5T140-306q-9 5-14.5 14t-5.5 20v32Zm296.5-343.5Q440-607 440-640t-23.5-56.5Q393-720 360-720t-56.5 23.5Q280-673 280-640t23.5 56.5Q327-560 360-560t56.5-23.5ZM360-240Zm0-400Z"></path></svg>`,
    impact:  `<svg height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M756-120 537-339l84-84 219 219-84 84Zm-552 0-84-84 276-276-68-68-28 28-51-51v82l-28 28-121-121 28-28h82l-50-50 142-142q20-20 43-29t47-9q24 0 47 9t43 29l-92 92 50 50-28 28 68 68 90-90q-4-11-6.5-23t-2.5-24q0-59 40.5-99.5T701-841q15 0 28.5 3t27.5 9l-99 99 72 72 99-99q7 14 9.5 27.5T841-701q0 59-40.5 99.5T701-561q-12 0-24-2t-23-7L204-120Z"></path></svg>`,
    status:  `<svg height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="m480-80-10-120h-10q-142 0-241-99t-99-241q0-142 99-241t241-99q71 0 132.5 26.5t108 73q46.5 46.5 73 108T800-540q0 75-24.5 144t-67 128q-42.5 59-101 107T480-80Zm80-146q71-60 115.5-140.5T720-540q0-109-75.5-184.5T460-800q-109 0-184.5 75.5T200-540q0 109 75.5 184.5T460-280h100v54Zm-72-107q12-12 12-29t-12-29q-12-12-29-12t-29 12q-12 12-12 29t12 29q12 12 29 12t29-12Zm-58-115h60q0-30 6-42t38-44q18-18 30-39t12-45q0-51-34.5-76.5T460-720q-44 0-74 24.5T344-636l56 22q5-17 19-33.5t41-16.5q27 0 40.5 15t13.5 33q0 17-10 30.5T480-558q-35 30-42.5 47.5T430-448Zm30-65Z"></path></svg>`,
    make:  `<svg height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M80-80v-481l280-119v80l200-80v120h320v480H80Zm80-80h640v-320H480v-82l-200 80v-78l-120 53v347Zm280-80h80v-160h-80v160Zm-160 0h80v-160h-80v160Zm320 0h80v-160h-80v160Zm280-320H680l40-320h120l40 320ZM160-160h640-640Z"></path></svg>`,
    cores:  `<svg height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M160-120q-33 0-56.5-23.5T80-200v-560q0-33 23.5-56.5T160-840h560q33 0 56.5 23.5T800-760v80h80v80h-80v80h80v80h-80v80h80v80h-80v80q0 33-23.5 56.5T720-120H160Zm0-80h560v-560H160v560Zm80-80h200v-160H240v160Zm240-280h160v-120H480v120Zm-240 80h200v-200H240v200Zm240 200h160v-240H480v240ZM160-760v560-560Z"></path></svg>`,
    drive:  `<svg height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M240-80q-33 0-56.5-23.5T160-160v-640q0-33 23.5-56.5T240-880h480q33 0 56.5 23.5T800-800v640q0 33-23.5 56.5T720-80H240Zm0-80h480v-640H240v640Zm80-80h320v-80H320v80Zm160-160q66 0 113-47t47-113q0-66-47-113t-113-47q-66 0-113 47t-47 113q0 66 47 113t113 47Zm0-120q-17 0-28.5-11.5T440-560q0-17 11.5-28.5T480-600q17 0 28.5 11.5T520-560q0 17-11.5 28.5T480-520Zm0-40Z"></path></svg>`,
    os:  `<svg height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M280-200h400q-4-49-30-90t-68-65l38-68q2-4 1-9t-6-7q-4-2-8.5-1t-6.5 5l-39 70q-20-8-40-12.5t-41-4.5q-21 0-41 4.5T399-365l-39-70q-2-5-6.5-5t-9.5 2l-4 15 38 68q-42 24-68 65t-30 90Zm96-66q-6-6-6-14t6-14q6-6 14-6t14 6q6 6 6 14t-6 14q-6 6-14 6t-14-6Zm180 0q-6-6-6-14t6-14q6-6 14-6t14 6q6 6 6 14t-6 14q-6 6-14 6t-14-6ZM240-80q-33 0-56.5-23.5T160-160v-640q0-33 23.5-56.5T240-880h320l240 240v480q0 33-23.5 56.5T720-80H240Zm280-520v-200H240v640h480v-440H520ZM240-800v200-200 640-640Z"></path></svg>`,
    geofence:  `<svg height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M536.5-503.5Q560-527 560-560t-23.5-56.5Q513-640 480-640t-56.5 23.5Q400-593 400-560t23.5 56.5Q447-480 480-480t56.5-23.5ZM480-186q122-112 181-203.5T720-552q0-109-69.5-178.5T480-800q-101 0-170.5 69.5T240-552q0 71 59 162.5T480-186Zm0 106Q319-217 239.5-334.5T160-552q0-150 96.5-239T480-880q127 0 223.5 89T800-552q0 100-79.5 217.5T480-80Zm0-480Z"></path></svg>`,
    trouble:  `<svg height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M160-120q-33 0-56.5-23.5T80-200v-560q0-33 23.5-56.5T160-840h640q33 0 56.5 23.5T880-760v560q0 33-23.5 56.5T800-120H160Zm0-80h640v-560H160v560Zm80-80h480v-80H240v80Zm0-160h160v-240H240v240Zm240 0h240v-80H480v80Zm0-160h240v-80H480v80ZM160-200v-560 560Z"></path></svg>`,
    type:  `<svg height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M508.5-291.5Q520-303 520-320t-11.5-28.5Q497-360 480-360t-28.5 11.5Q440-337 440-320t11.5 28.5Q463-280 480-280t28.5-11.5ZM440-440h80v-240h-80v240Zm40 360q-83 0-156-31.5T197-197q-54-54-85.5-127T80-480q0-83 31.5-156T197-763q54-54 127-85.5T480-880q83 0 156 31.5T763-763q54 54 85.5 127T880-480q0 83-31.5 156T763-197q-54 54-127 85.5T480-80Zm0-80q134 0 227-93t93-227q0-134-93-227t-227-93q-134 0-227 93t-93 227q0 134 93 227t227 93Zm0-320Z"></path></svg>`,
    kinds:  `<svg height="24px" viewBox="0 -960 960 960" width="24px" fill="currentColor"><path d="M400-560ZM160-160q-33 0-56.5-23.5T80-240v-480q0-33 23.5-56.5T160-800h640q33 0 56.5 23.5T880-720H160v480h80v80h-80Zm640-80v-320H640v320h160Zm-180 80q-25 0-42.5-17.5T560-220v-360q0-25 17.5-42.5T620-640h200q25 0 42.5 17.5T880-580v360q0 25-17.5 42.5T820-160H620Zm100-300q13 0 21.5-9t8.5-21q0-13-8.5-21.5T720-520q-12 0-21 8.5t-9 21.5q0 12 9 21t21 9ZM340-160l-20-70q-19-17-29.5-40T280-320q0-27 10.5-50t29.5-40l20-70h120l20 70q19 17 29.5 40t10.5 50q0 27-10.5 50T480-230l-20 70H340Zm60-100q26 0 43-17.5t17-42.5q0-25-18-42.5T400-380q-24 0-42 17t-18 43q0 26 17 43t43 17Zm320-140Z"></path></svg>`
};

const menuLabels = new Map([
    ["SITE", icons.site + " Sites"],
    ["OFFICE", icons.office +" Office"],
    ["GROUP", icons.group + " Groups"],
    ["IMPACT", icons.impact +" Impact"],
    ["STATUS", icons.status + " Status"],
    ["MAKE", icons.make + " Make"],
    ["CORES", icons.cores + " Cores"],
    ["DRIVETYPE", icons.drive + " Drives"],
    ["OS", icons.os + " OS"],
    ["GEOFENCE", icons.geofence + " GeoFence"],
    ["TROUBLE", icons.trouble + " Trouble"],
    ["TYPE", icons.type + " Asset ID Prefix "],
    ["KIND", icons.kinds + " Kinds"]
]);

// Icon Buttons
let btnTables;
let btnSave;
let btnNew;
let btnHelp;

// Object that holds the parent (Device Type) for the trouble type => (DESKTOP, LAPTOP, PHONE...)
let troubleParent; 

document.addEventListener('DOMContentLoaded', function() {
  btnNew = new Button("btnNew");
  btnSave = new Button("btnSave");
  btnTables = new Button("btnTables");
  btnHelp = new Button("btnHelp");
  btnSave.off();
  btnNew.off();
  btnTables.on();
  
  const deviceTypesJson = document.getElementById("deviceTypesJson");
  if (deviceTypesJson) {
      troubleParent = JSON.parse(deviceTypesJson.value);
  }
  showSelection('SITE');
});

const isValid = (ctlId) => {
    const el = document.getElementById(ctlId);
    if (el) {
        if (!el.checkValidity()) {
            el.setAttribute("aria-invalid", "true");
            return { valid: false, value: el.value };
        } else {
            el.setAttribute("aria-invalid", "false");
            return { valid: true, value: el.value };
        }
    }
    return { valid: false, value: "" };
};

function showSelection(item) {
    const tblDialog = document.getElementById("tableDialog");
    if (tblDialog) {
        closeModal(tblDialog);
    }
    const sendData = {task: "build_table", field: item, adminData: ""};
    
    fetch("admin", {
        method: "POST",
        body: new URLSearchParams(sendData)
    })
    .then(response => response.text())
    .then(response => {
        adminData = JSON.parse(response);
        if (item === "SITE") {
            siteData = JSON.parse(response);
        }
        if (adminData !== null && adminData.length > 0) {
            buildTable();
        }
    });
    
    btnSave.off();
    const tableSelected = document.getElementById("tableSelected");
    if (tableSelected) {
        tableSelected.value = item;
    }
}

function saveRecord() {
    if (adminData === null || adminData.length === 0) {
        return;
    }

    // search all input elements for aria-invalid='true'
    const form = document.querySelector('#theForm');
    //    const allControls = form.elements; // Returns an HTMLFormControlsCollection
    // If you specifically want only 'input' tags:
    const inputsOnly = Array.from(form.elements).filter(el => el.tagName === 'INPUT');
    for (let i = 0; i < inputsOnly.length; i++) {
        const el = inputsOnly[i];
        if (el.getAttribute("aria-invalid") === "true") {
            toast("There are invalid inputs. Please Correct.", "error");
            return;
        }
    }


    let sendData = {task: "", field: "", adminData: ""};
    sendData.task = "save_table";
    sendData.field = adminData[0].field;
    sendData.adminData = JSON.stringify(adminData);
    
    fetch("admin", {
        method: "POST",
        body: new URLSearchParams(sendData)
    })
    .then(response => response.text())
    .then(response => {
        if (typeof response === 'string'  && response.startsWith("ERROR:")) {
            toast("Error saving Admin table. Refresh and retry.", "error");
        } else {
            adminData = null;
            adminData = JSON.parse(response);
            if (sendData.field === "SITE") {
                siteData = JSON.parse(response);
            }
            buildTable();
            toast("Changes saved.", "success");
            btnSave.off();
        }
    });
}

function buildTable() {
    building = true;
    adminData.sort(getSortOrder("sequence"));
    const tableSelected = document.getElementById("tableSelected").value;
    const title = menuLabels.get(adminData[0].field);
    const headerExtraColumn = setHeaderExtraColumn(tableSelected);
    const cnt = adminData.length;
    let isFirst = true;

    let tbl = `<p><strong>${title}</strong></p>
        <table id="settingsTable"><thead><tr>
        <th style="width: 15%; white-space: nowrap;">Code</th><th>Description</th>${headerExtraColumn}
        <th style="width: 1%; white-space: nowrap;">Active</th><th style="width: 1%; white-space: nowrap;"></th>
        <th style="width: 1%; white-space: nowrap;"></th><th style="width: 1%; white-space: nowrap;"'></th>
        </tr></thead><tbody>`;

    adminData.forEach((item, index) => {
        if (item.delete) return;
        const descriptionControl = buildDescriptionControl(item.description, item.id);
        const activeControl = buildActiveControl(item.active, item.id);
        const extraColumn = setRowExtraColumn(tableSelected, item.parent, item.assetid, item.id);
        const moveUp = !isFirst ? `<a href="javascript:moveRowUp('${item.id}');">⬆️</a>` : "&nbsp;";
        isFirst = false;
        const moveDown = (cnt - index - 1) > 0 ? `<a href="javascript:moveRowDown('${item.id}');">⬇️</a>` : "&nbsp;";
        const notInUse = item.inuse !== true ? `<a href="javascript:deleteRow('${item.id}');">❌</a>` : "&nbsp;";
        tbl += `<tr><td>${item.code}</td><td>${descriptionControl}</td>${extraColumn}<td>${activeControl}</td><td>${moveUp}</td><td>${moveDown}</td><td>${notInUse}</td></tr>`;
    });

    // Final row for adding a new item, except for the TYPE
    if (tableSelected !== "TYPE") {
        const lastCodeControl = buildCodeControl("", 0);
        const lastDescriptionControl = buildDescriptionControl("", 0);
        const lastActiveControl = buildActiveControl(0, 0);
        const lastExtraColumn = setRowExtraColumn(tableSelected, "", "", 0);
        tbl += `<tr><td>${lastCodeControl}</td><td>${lastDescriptionControl}</td>${lastExtraColumn}<td>${lastActiveControl}</td><td colspan='3'><a href="javascript:addRow();">➕</a></td></tr>`;
    }
    tbl += `</tbody></table>`;
    document.getElementById("adminTable").innerHTML = tbl;
    triggersBtnSave();
    setTimeout(() => { building = false; }, 2000);
}

// Helper functions
function setHeaderExtraColumn(tableSelected) {
    switch (tableSelected) {
        case "OFFICE": return "<th>Site</th>";
        case "TROUBLE": return "<th>Device Type</th>";
        case "GROUP": return "<th>Permissions</th>";
        case "TYPE": return "<th>Asset ID Prefix</th>";
        default: return "";
    }
}

function setRowExtraColumn(tableSelected, parent, assetid, id) {
    switch (tableSelected) {
        case "OFFICE": return `<td>${buildParentControl(parent, id)}</td>`;
        case "TROUBLE": return `<td>${buildTroubleParentControl(parent, id)}</td>`;
        case "GROUP": return `<td>${buildGroupPermissionsControl(id)}</td>`;
        case "TYPE": return `<td>${buildAssetIdControl(assetid, id)}</td>`;
        default: return "";
    }
}

//Comparer Function for sorting JSON object array  
function getSortOrder(prop) {
    return function (a, b) {
        if (a[prop] > b[prop]) {
            return 1;
        } else if (a[prop] < b[prop]) {
            return -1;
        }
        return 0;
    };
}

function buildDescriptionControl(txt, rowId) {
    const ctlId = `descr${rowId}`;
    const errorId = `descrerror${rowId}`;
    const maxLength = 30;
    return `<input type='text' id='${ctlId}'  name='${ctlId}' value="${txt}" placeholder='Description'
            title='Description' required minlength='1' maxlength='${maxLength}' pattern="[a-zA-Z0-9.\\s\\(\\)\\-]+"
            aria-invalid="false" aria-describedby='${errorId}'
            onchange="doTextbox('${rowId}')" >
        <small class="err" id='${errorId}'>Mandatory with only A-Z or 0-9, ${maxLength} Characters</small>`;
}

function buildActiveControl(isChecked, rowId) {
    const ctlId = `active${rowId}`;
    const checked = isChecked === 1 ? "checked" : "";
    return `<input type='checkbox' id='${ctlId}' name='${ctlId}' ${checked} onclick="doCheckbox('${rowId}');" >`;
}

function buildParentControl(txt, rowId) {
    const ctlId = `parent${rowId}`;
    let options = siteData.map(item => {
        const selected = item.code === txt ? "selected" : "";
        return `<option value="${item.code}" ${selected}>${item.description}</option>`;
    }).join('');
    return `<select id='${ctlId}' name='${ctlId}' onchange="doParentSelect('${rowId}')" >${options}</select>`;
}

function buildTroubleParentControl(txt, rowId) {
    const ctlId = `parent${rowId}`;
    let options = troubleParent.map(item => {
        const selected = item.name === txt ? "selected" : "";
        return `<option value="${item.name}"  ${selected} >${item.description}</option>`;
    }).join('');
    return `<select id='${ctlId}' name='${ctlId}' onchange="doParentSelect('${rowId}')" >${options}</select>`;
}

function buildGroupPermissionsControl(rowId) {
    return `<button type='button' class='secondary' onclick="popPermissions('${rowId}');">Edit...</button>`;
}

function buildAssetIdControl(txt, rowId) {
    const ctlId = `asset${rowId}`;
    const errorId = `asseterror${rowId}`;
    const maxLength = 6;
    return `<input type='text' id='${ctlId}' name='${ctlId}' value="${txt}" placeholder='Asset ID Prefix'
            title='Asset ID Prefix' required minlength='1' maxlength='${maxLength}' pattern='[a-zA-Z0-9]+'
            aria-invalid="false" aria-describedby='${errorId}'
            onchange="doAssetId('${rowId}')" >
        <small class="err" id='${errorId}'>Mandatory with only A-Z or 0-9, ${maxLength} Characters</small>`;
}

function buildCodeControl(txt, rowId) {
    const tableSelected = document.getElementById("tableSelected").value;
    const isGroupOrImpact = tableSelected === "GROUP" || tableSelected === "IMPACT";
    const isGeoFence = tableSelected === "GEOFENCE";
    const codeErrorId = `codeerror${rowId}`;
    let validationRules, errorMessage;
    if (isGroupOrImpact) {
        validationRules = "required minlength='1' maxlength='2' integer";
        errorMessage = "Mandatory, unique and only numbers (1-99)";
    } else if (isGeoFence) {
        validationRules = "required minlength='1' maxlength='30'";
        errorMessage = "Must be lat,lon such as: 43.865050,-79.849630";
    } else {
        validationRules = "required minlength='1' maxlength='20' pattern='[A-Z0-9]+' ";
        errorMessage = "Mandatory and unique with only: A to Z, 0 to 9";
    }
    return `<input type='text' id='code${rowId}' name='code${rowId}' value="${txt || ""}" 
            placeholder='Code' title='Code' ${validationRules}
            style='text-transform:uppercase'
            aria-invalid="false" aria-describedby='${codeErrorId}'
            onchange="doCodeValidate('${rowId}')" >
        <small class="err" id='${codeErrorId}'>${errorMessage}</small>`;
} // <-----<<<<<<<<  check for further code control checking

function doCodeValidate(rowId) {
    const id = txt2Int(rowId);
    const ctlId = "code" + id.toString();
    const reply = isValid(ctlId);
}

function doCheckbox(rowId) {
    const id = txt2Int(rowId);
    const ctlId = "active" + id.toString();
    const idx = adminData.findIndex((item => item.id === id));
    if (idx >= 0 && idx < adminData.length) {
        adminData[idx].active = (document.getElementById(ctlId).checked) ? 1 : 0;
        adminData[idx].update = true; //Flag server to update this record
    }
}

function doTextbox(rowId) {
    const id = txt2Int(rowId);
    const ctlId = "descr" + id.toString();
    const reply = isValid(ctlId);
    if (!reply.valid) return;
    const idx = adminData.findIndex((item => item.id === id));
    if (idx >= 0 && idx < adminData.length) {
        if (adminData[idx].description !== reply.value) {
            adminData[idx].description = reply.value;
            adminData[idx].update = true; //Flag server to update this record
        }
    }
}

function doAssetId(rowId) {
    const id = txt2Int(rowId);
    const ctlId = "asset" + id.toString();
    const reply = isValid(ctlId);
    if (!reply.valid) return;
    let idx = adminData.findIndex((item => item.id === id));
    if (idx >= 0 && idx < adminData.length) {
        if (adminData[idx].assetid !== reply.value) {
            adminData[idx].assetid = reply.value;
            adminData[idx].update = true; // Flag server to update this record
        }
    }
}

function doParentSelect(rowId) {
    if (building) return;
    const id = txt2Int(rowId);
    const ctlId = "parent" + id.toString();
    const idx = adminData.findIndex((item => item.id === id));
    if (idx >= 0 && idx < adminData.length) {
        adminData[idx].parent = document.getElementById(ctlId).value;
        adminData[idx].update = true; //Flag server to update this record
    }
}

function moveRowUp(rowId) {
    const id = txt2Int(rowId);
    const idx = adminData.findIndex((item => item.id === id));
    const seq = adminData[idx].sequence;
    let prevRow = idx - 1;

    while (adminData[prevRow].delete === true) {    //skip past deleted rows
        prevRow--;
        if (prevRow < 0)
            break;
    }

    if (prevRow >= 0) {
        adminData[idx].sequence = adminData[prevRow].sequence;
        adminData[idx].update = true;
        adminData[prevRow].sequence = seq;
        adminData[prevRow].update = true;
    }
    btnSave.on();
    buildTable();
}

function moveRowDown(rowId) {
    const id = txt2Int(rowId);
    const idx = adminData.findIndex((item => item.id === id));
    const seq = adminData[idx].sequence;
    let nextRow = idx + 1;

    while (adminData[nextRow].delete === true) {    //skip past deleted rows
        nextRow++;
        if (nextRow > adminData.length)
            break;
    }

    if (nextRow < adminData.length) {
        adminData[idx].sequence = adminData[nextRow].sequence;
        adminData[idx].update = true;
        adminData[nextRow].sequence = seq;
        adminData[nextRow].update = true;
    }
    btnSave.on();
    buildTable();
}

function deleteRow(rowId) {
    const id = txt2Int(rowId);
    const idx = adminData.findIndex((item => item.id === id));
    if (idx >= 0 && idx < adminData.length) {
        adminData[idx].delete = true;
    }
    btnSave.on();
    buildTable();
}

function addRow() {
    let parent = "";
    let permissions = "";
    const tableSelected = document.getElementById("tableSelected").value;
    const active = document.getElementById("active0").checked ? 1 : 0;
    
    //Get the new "code" value as uppercase string
    let reply = isValid("code0");
    if (!reply.valid) return;
    const code = reply.value.toString().toUpperCase();

    //This is a new addition, so we should not find an index for it
    let idx = adminData.findIndex((item => item.code === code));
    if (idx > -1) {
        toast("Code already exists:" + code, "error");
        console.log("Code already exists:" + code);
        return;
    }

    // Get the new description value as string
    reply = isValid("descr0");
    if (!reply.valid) return;
    const descr = reply.value.toString();

    switch (tableSelected) {
        case "OFFICE", "TROUBLE":
            reply = isValid("parent0");
            if (!reply.valid) return;
            parent = reply.value.toString();
            break;
        case "PERMISSIONS":
            permissions = readPermissions();
            break;
        case "GROUP", "IMPACT":
            if (!isDigits(code) || code.length < 1 || txt2Int(code) < 1 ) {
                document.getElementById("code0").setAttribute("aria-invalid", "true");
                return;
            }
            break;
        case "GEOFENCE":
            if (!isLatLon(code) || code.length < 1) {
                document.getElementById("code0").setAttribute("aria-invalid", "true");
                return;
            }
            break;
        default:
             if (code !== code.replace(/[+-]?[^A-Za-z0-9 ]/g, '') || code.length < 1) {
                document.getElementById("code0").setAttribute("aria-invalid", "true");
                return;
             }
            break;
        }

    if (descr !== descr.replace(/[+-]?[^A-Za-z0-9 ]/g, '') || descr.length < 1) {
        document.getElementById("descr0").setAttribute("aria-invalid", "true");
        return;
    }

    // Find the item with the maximum id in the array
    // Its just a placeholder. Database will assign id when insert occurs
    const maxId = adminData.reduce((prev, current) => (prev.id > current.id) ? prev : current);
    // Find the item with the maximum sequence in the array
    const maxSeq = adminData.reduce((prev, current) => (prev.sequence > current.sequence) ? prev : current);
    // Create the item to be added
    let newItem = {id: 0, description: "", code: "", parent: "", sequence: 1, field: "", active: 1, update: false, add: true, delete: false, inuse: false, task: ""};
    newItem.id = maxId.id + 1;
    newItem.description = descr;
    newItem.code = code;
    
    if (tableSelected === "OFFICE") {
        newItem.parent = parent;
    } else if (tableSelected === "GROUP") {
        newItem.permissions = permissions;
    }
    newItem.sequence = maxSeq.sequence + 1;
    newItem.field = adminData[0].field;
    newItem.active = active;
    adminData.push(newItem);
    buildTable();
}

//checks for 40.838383,-79.44848 format (in one field)
function isLatLon(value) {    
    if (typeof value !== "string" || value.length < 3) return false;
    const latLonOnly = /^[-+]?([1-8]?\d(\.\d+)?|90(\.0+)?),\s*[-+]?(180(\.0+)?|((1[0-7]\d)|([1-9]?\d))(\.\d+)?)$/;
    return latLonOnly.test(value);   
}

function showHelp() {
    const helpDialog = document.getElementById("helpDialog");
    if (helpDialog) {
        openModal(helpDialog);
    }
}

function showTableSelect() {
    const tableDialog = document.getElementById("tableDialog");
    if (tableDialog) {
        openModal(tableDialog);
    }
}

//PCRUD:ACRUD:DCRUD:SCRUD:TCRUD
function popPermissions(rowId) {
    // MAP of the names of the Sliders
    const sliders = new Map([
        ["P", "#profileSlider"],
        ["D", "#deviceSlider"],
        ["S", "#softwareSlider"],
        ["T", "#ticketSlider"],
        ["A", "#adminSlider"]
      ]);
    const lookup = {"C": 1, "CR": 2, "CRU": 3, "CRUD": 4};
    const id = txt2Int(rowId);
    document.getElementById("permRowId").value = id;
    const idx = adminData.findIndex((item => item.id === id));
    let perms = "P:D:S:T:A";
    let dialogTitle = "Group";
    if (idx >-1 ) {
        dialogTitle = adminData[idx].description;
        perms = adminData[idx].permissions;
    }
    document.getElementById("GroupName").value = dialogTitle;

    const permissionsDialog = document.getElementById("permissionsDialog");
    if (permissionsDialog) {
        openModal(permissionsDialog);
    }
    
    const crud = perms.split(":");
    for (let i = 0; i < crud.length; i++) {
        const val = lookup[crud[i].slice(1)] || 0;
        const slider = sliders.get(crud[i].charAt(0));
        document.querySelector(slider).value = val;
        setSlider(slider, val);
    }
}

function updatePermissions() {
    const permissions = readPermissions();
    const id = txt2Int(document.getElementById("permRowId").value);
    const idx = adminData.findIndex(item => item.id === id);
    if (idx < 0 && idx < adminData.length) return;
    if (adminData[idx].permissions !== permissions) {
        adminData[idx].permissions = permissions;
        adminData[idx].update = true; // Flag server to update this record
    }
    btnSave.on()
    buildTable();
}

function readPermissions() {
    const sliders = [ "#profileSlider", "#deviceSlider", "#softwareSlider", "#ticketSlider", "#adminSlider" ];
    const starters = ["P", "D", "S", "T", "A"];
    const lookup = [":", "C:", "CR:", "CRU:", "CRUD:"];
    let permissions = sliders.map((selector, i) => {
        const val = parseInt(document.querySelector(selector).value, 10) || 0;
        return starters[i] + lookup[val];
    }).join("").toUpperCase().replace(/[^CRUD:PDSAT]/g, '');
    // remove trailing : from permissions string
    if (permissions.endsWith(":")) {
        permissions = permissions.slice(0, -1);
    }
    return permissions;
}

function setSlider(ctrl, val) {
    const str = ["None", "Create", "Create and Read", "Create, Read and Update", "Create, Read, Update and Delete"];
    const i = txt2Int(val);    
    const slider = document.querySelector(ctrl);
    if (slider) slider.value = i;
    const label = document.querySelector(ctrl + "Label");
    if (label) label.textContent = str[i];
}

function triggersBtnSave() {
    document.querySelectorAll("input").forEach(el => {
        el.addEventListener("input", () => { btnSave.on(); });
    });
    document.querySelectorAll("select").forEach(el => {
        el.addEventListener("change", () => { btnSave.on(); });
    });
    document.querySelectorAll("select").forEach(el => {
        el.addEventListener("change", () => { btnSave.on(); });
    });
    document.querySelectorAll("checkbox").forEach(el => {
        el.addEventListener("change", () => { btnSave.on(); });
    });
}
