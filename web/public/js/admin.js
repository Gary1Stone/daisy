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

const menuLabels = new Map([
    ["SITE", "Sites"],
    ["OFFICE", "Office"],
    ["GROUP", "Groups"],
    ["IMPACT", "Impact"],
    ["STATUS", "Status"],
    ["MAKE", "Make"],
    ["CORES", "Cores"],
    ["DRIVETYPE", "Drives"],
    ["OS", "OS"],
    ["GEOFENCE", "GeoFence"],
    ["TROUBLE", "Trouble"],
    ["TYPE", "Asset ID Prefix"],
    ["KIND", "Kinds"]
]);

// Object that holds the parent (Device Type) for the trouble type => (DESKTOP, LAPTOP, PHONE...)
let troubleParent; 

// Icon Buttons
let btnTables;
let btnSave;
let btnNew;
let btnHelp;

document.addEventListener('DOMContentLoaded', function() {
  btnNew = new Button("btnNew");
  btnSave = new Button("btnSave");
  btnHelp = new Button("btnHelp");
  btnTables = new Button("btnTables");
  btnSave.off();
  
  const deviceTypesJson = document.getElementById("deviceTypesJson");
  if (deviceTypesJson) {
      troubleParent = JSON.parse(deviceTypesJson.value);
  }
  showSelection('SITE');
});

function showSelection(item) {
    const tblDialog = document.getElementById("tableDialog");
    if (tblDialog) {
        closeModal(tblDialog);
    }

//    Metro.dialog.close("#tableDialog");
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
            toast("Error saving Admin table. Refresh and retry.", "alert");
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

    let tbl = `
        <p class='text-center'><strong>${title}</strong></p>
        <table class='gtable'>
        <thead>
            <tr class='bg-black fg-white'>
                <th class='gcell' style='width: 15%;'>Code</th>
                <th class='gcell'>Description</th>
                ${headerExtraColumn}
                <th class='gcell'>Active</th>
                <th></th>
                <th></th>
                <th></th>
                <th></th>
            </tr>
        </thead>
        <tbody>
    `;

    adminData.forEach((item, index) => {
        if (item.delete) return;

        const rowClass = (cnt - index) % 2 === 0 ? " class='bg-lightGray'" : "";
        const descriptionControl = buildDescriptionControl(item.description, item.id);
        const activeControl = buildActiveControl(item.active, item.id);
        const extraColumn = setRowExtraColumn(tableSelected, item.parent, item.assetid, item.id);

        const moveUp = !isFirst
            ? `<a href="javascript:moveRowUp('${item.id}');">
                 <span class="mif-arrow-up mif-2x fg-blue"></span></a>`
            : "&nbsp;";
        isFirst = false;

        const moveDown = (cnt - index - 1) > 0
            ? `<a href="javascript:moveRowDown('${item.id}');">
                 <span class="mif-arrow-down mif-2x fg-blue"></span></a>`
            : "&nbsp;";

        const notInUse = item.inuse !== true
            ? `<a href="javascript:deleteRow('${item.id}');">
                 <span class="mif-cross mif-2x fg-red"></span></a>`
            : "&nbsp;";

        tbl += `
            <tr${rowClass}>
                <td class='gcell'>${item.code}</td>
                <td class='gcell'>${descriptionControl}</td>
                ${extraColumn}
                <td class='gcell'>${activeControl}</td>
                <td>${moveUp}</td>
                <td>${moveDown}</td>
                <td>${notInUse}</td>
            </tr>
        `;
    });

    // Final row for adding a new item, except for the TYPE
    if (tableSelected !== "TYPE") {
        const lastCodeControl = buildCodeControl("", 0);
        const lastDescriptionControl = buildDescriptionControl("", 0);
        const lastActiveControl = buildActiveControl(0, 0);
        const lastExtraColumn = setRowExtraColumn(tableSelected, "", "", 0);

        tbl += `
            <tr class='bg-black fg-white'>
                <td class='gcell'>${lastCodeControl}</td>
                <td class='gcell'>${lastDescriptionControl}</td>
                ${lastExtraColumn}
                <td class='gcell'>${lastActiveControl}</td>
                <td colspan='3'>
                    <a href="javascript:addRow();">
                        <span class='mif-plus mif-2x fg-green'></span></a>
                </td>
            </tr>
        `;
    }
    tbl += `
                    </a>
                </td>
            </tr>
            </tbody>
            </table>
        `;

    document.getElementById("adminTable").innerHTML = tbl;
    triggersBtnSave();
    setTimeout(() => { building = false; }, 2000);
}

// Helper functions
function setHeaderExtraColumn(tableSelected) {
    switch (tableSelected) {
        case "OFFICE": return "<th class='gcell'>Site</th>";
        case "TROUBLE": return "<th class='gcell'>Device Type</th>";
        case "GROUP": return "<th class='gcell'>Permissions</th>";
        case "TYPE": return "<th class='gcell'>Asset ID Prefix</th>";
        default: return "";
    }
}

function setRowExtraColumn(tableSelected, parent, assetid, id) {
    switch (tableSelected) {
        case "OFFICE": return `<td class='gcell'>${buildParentControl(parent, id)}</td>`;
        case "TROUBLE": return `<td class='gcell'>${buildTroubleParentControl(parent, id)}</td>`;
        case "GROUP": return `<td class='gcell'>${buildGroupPermissionsControl(id)}</td>`;
        case "TYPE": return `<td class='gcell'>${buildAssetIdControl(assetid, id)}</td>`;
        default: return "<td class='gcell'></td>";
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
    return `
        <input type='text' 
            id='${ctlId}'  name='${ctlId}' 
            value="${txt}" placeholder='Description'
            title='Description' required
            minlength='1' maxlength='${maxLength}'
            data-role='input'
            data-validate='required minlength=1 maxlength=${maxLength}'
            onblur="doTextbox('${rowId}')" >
        <span class='invalid_feedback' id='${errorId}'>Mandatory with only A-Z or 0-9, ${maxLength} Characters</span>
    `;
}

function buildActiveControl(isChecked, rowId) {
    const ctlId = `active${rowId}`;
    const checked = isChecked === 1 ? "checked" : "";
    return `
        <input type='checkbox' 
            id='${ctlId}' name='${ctlId}' 
            data-role='checkbox' data-caption='Active'
            ${checked}
            onclick="doCheckbox('${rowId}');" >
    `;
}

function buildParentControl(txt, rowId) {
    const ctlId = `parent${rowId}`;
    let options = siteData.map(item => {
        const selected = item.code === txt ? "selected" : "";
        return `<option value="${item.code}" ${selected}>${item.description}</option>`;
    }).join('');
    return `
        <select 
            id='${ctlId}'  name='${ctlId}' 
            data-role='select'  data-filter='false' 
            onchange="doParentSelect('${rowId}')" >
        ${options}
        </select>
    `;
}

function buildTroubleParentControl(txt, rowId) {
    const ctlId = `parent${rowId}`;
    let options = troubleParent.map(item => {
        const selected = item.name === txt ? "selected" : "";
        return `
            <option value="${item.name}"  ${selected}
                data-template="<span class='${item.icon} icon'></span> $1" >
                &nbsp;${item.description}
            </option>
        `;
    }).join('');
    return `
        <select id='${ctlId}' name='${ctlId}' 
            data-role='select' data-filter='false' 
            onchange="doParentSelect('${rowId}')" >
        ${options}
        </select>
    `;
}

function buildGroupPermissionsControl(rowId) {
    return `
        <button type='button' class='button info' onclick="popPermissions('${rowId}');">Edit...</button>
    `;
}

function buildAssetIdControl(txt, rowId) {
    const ctlId = `asset${rowId}`;
    const errorId = `asseterror${rowId}`;
    const maxLength = 6;
    return `
        <input type='text' 
            id='${ctlId}'  name='${ctlId}' 
            value="${txt}" placeholder='Asset ID Prefix'
            title='Asset ID Prefix' required
            minlength='1' maxlength='${maxLength}'
            data-role='input'
            data-validate='required minlength=1 maxlength=${maxLength}'
            onblur="doAssetId('${rowId}')" >
        <span class='invalid_feedback' id='${errorId}'>Mandatory with only A-Z or 0-9, ${maxLength} Characters</span>
    `;
}

function buildCodeControl(txt, rowId) {
    const tableSelected = document.getElementById("tableSelected").value;
    const isGroupOrImpact = tableSelected === "GROUP" || tableSelected === "IMPACT";
    const isGeoFence = tableSelected === "GEOFENCE";
    const codeErrorId = `codeerror${rowId}`;
    let validationRules, errorMessage;
    if (isGroupOrImpact) {
        validationRules = "required minlength=1 maxlength=2 integer";
        errorMessage = "Mandatory, unique and only numbers (1-99)";
    } else if (isGeoFence) {
        validationRules = "required minlength=1 maxlength=30";
        errorMessage = "Must be lat,lon such as: 43.865050,-79.849630";
    } else {
        validationRules = "required minlength=1 maxlength=20";
        errorMessage = "Mandatory and unique with only: A to Z, 0 to 9";
    }
    return `
        <input type='text' 
            id='code${rowId}' name='code${rowId}' value="${txt || ""}" 
            placeholder='Code' title='Code' required 
            minlength='1' maxlength='20' 
            style='text-transform:uppercase' 
            data-role='input' data-validate='${validationRules}' >
        <span class='invalid_feedback' id='${codeErrorId}'>${errorMessage}</span>
    `;
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
    const errorId = "descrerror" + id.toString();
    const descrEl = document.getElementById(ctlId);
    const errorEl = document.getElementById(errorId);
    const descr = descrEl.value;
    
    if (errorEl) errorEl.style.display = 'none';
    
    // Validate the input text (txt)
    if (descr === undefined || descr.length < 1 || descr !== descr.replace(/[^a-zA-Z0-9.() ]/g, '')) {
        if (errorEl) errorEl.style.display = 'block';
        return;
    }

    const idx = adminData.findIndex((item => item.id === id));
    if (idx >= 0 && idx < adminData.length) {
        if (adminData[idx].description !== descr) {
            adminData[idx].description = descr;
            adminData[idx].update = true; //Flag server to update this record
        }
    }
}




function doAssetId(rowId) {
    const id = txt2Int(rowId);
    const ctlId = "asset" + id.toString();
    const errorId = "asseterror" + id.toString();
    const inputEl = document.getElementById(ctlId);
    const errorEl = document.getElementById(errorId);
    const txt = inputEl.value;
    if (errorEl) errorEl.style.display = 'none';
    
    // Validate the input text (txt)
    if (txt === undefined || txt.length < 1 || txt !== txt.replace(/[^a-zA-Z0-9]/g, '')) {
        if (errorEl) errorEl.style.display = 'block';
        return;
    }

    let idx = adminData.findIndex((item => item.id === id));
    if (idx >= 0 && idx < adminData.length) {
        if (adminData[idx].assetid !== txt) {
            adminData[idx].assetid = txt;
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
    buildTable();
}

function deleteRow(rowId) {
    const id = txt2Int(rowId);
    const idx = adminData.findIndex((item => item.id === id));
    if (idx >= 0 && idx < adminData.length) {
        adminData[idx].delete = true;
    }
    buildTable();
}

function addRow() {
    const code = document.getElementById("code0").value.toString().toUpperCase();
    const descr = document.getElementById("descr0").value.toString();
    let parent = "";
    let permissions = "";
    const tableSelected = document.getElementById("tableSelected").value;
    
    if (tableSelected === "OFFICE") {
        parent = document.getElementById("parent0").value.toString();
        const err = document.getElementById("parenterror0");
        if (err) err.style.display = 'none';
    } else if (tableSelected === "PERMISSIONS") {
        permissions = readPermissions();
    } else if (tableSelected === "TROUBLE") {
        parent = document.getElementById("parent0").value.toString();
        const err = document.getElementById("parenterror0");
        if (err) err.style.display = 'none';
    }
    const active = document.getElementById("active0").checked ? 1 : 0;
    
    const codeErr = document.getElementById("codeerror0");
    const descrErr = document.getElementById("descrerror0");
    if (codeErr) codeErr.style.display = 'none';
    if (descrErr) descrErr.style.display = 'none';
    
    let idx = adminData.findIndex((item => item.code === code));
    
    if (tableSelected === "GROUP") {
        if (!isDigits(code) || idx > -1 || code.length < 1 || txt2Int(code) < 1 ) {
            if (codeErr) codeErr.style.display = 'block';
            return;
        }
    } else if (tableSelected === "IMPACT") {
        if (!isDigits(code) || idx > -1 || code.length < 1 || txt2Int(code) < 1 ) {
            if (codeErr) codeErr.style.display = 'block';
            return;
        }
    } else if (tableSelected === "GEOFENCE") {
        if (!isLatLon(code) || idx > -1 || code.length < 1) {
            if (codeErr) codeErr.style.display = 'block';
            return;
        }
    } else {
        if (code !== code.replace(/[+-]?[^A-Za-z0-9 ]/g, '') || idx > -1 || code.length < 1) {
            if (codeErr) codeErr.style.display = 'block';
            return;
        }
    }

    if (descr !== descr.replace(/[+-]?[^A-Za-z0-9 ]/g, '') || descr.length < 1) {
        if (descrErr) descrErr.style.display = 'block';
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

//checks for 40.838383,-79.44848  format (in one field)
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
//    Metro.dialog.open("#helpDialog");
}

function showTableSelect() {
    const tableDialog = document.getElementById("tableDialog");
    if (tableDialog) {
        openModal(tableDialog);
    }
//    Metro.dialog.open("#tableDialog");
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
    
//    Metro.dialog.open("#permissionsDialog");
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
}
