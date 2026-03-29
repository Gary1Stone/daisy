/* global Metro */
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
    ["SITE", "<span class='mif-map2'></span>&nbsp;Sites"],
    ["OFFICE", "<span class='mif-room'></span>&nbsp;Office"],
    ["GROUP", "<span class='mif-users'></span>&nbsp;Groups"],
    ["IMPACT", "<span class='mif-hammer'></span>&nbsp;Impact"],
    ["STATUS", "<span class='mif-question'></span>&nbsp;Status"],
    ["MAKE", "<span class='mif-location-city'></span>&nbsp;Make"],
    ["CORES", "<span class='mif-calculator2'></span>&nbsp;Cores"],
    ["DRIVETYPE", "<span class='mif-cabinet'></span>&nbsp;Drives"],
    ["OS", "<span class='mif-windows'></span>&nbsp;OS"],
    ["GEOFENCE", "<span class='mif-my-location'></span>&nbsp;GeoFence"],
    ["TROUBLE", "<span class='mif-news'></span>&nbsp;Trouble"],
    ["TYPE", "<span class='mif-star-half'></span>&nbsp;Asset ID Prefix"],
    ["KIND", "<span class='mif-display'></span>&nbsp;Kinds"]
]);

// Object that holds the parent (Device Type) for the trouble type => (DESKTOP, LAPTOP, PHONE...)
let troubleParent; 

$(document).ready(function () {
    btnSave.off();
    const jsonData = $("#deviceTypesJson").val();
    troubleParent = JSON.parse(jsonData);
    showSelection('SITE');
});

function showSelection(item) {
    Metro.dialog.close("#tableDialog");
    let sendData = {task: "", field: "", adminData: ""};
    sendData.task = "build_table";
    sendData.field = item;
    sendData.adminData = ""
    $.post("admin", sendData).then(response => {
        adminData = JSON.parse(response);
        if (item === "SITE") {
            siteData = JSON.parse(response);
        }
        if (adminData != null && adminData.length > 0) {
            buildTable();
        }
    });
    btnSave.off();
    $("#tableSelected").val(item);
}

function save() {
    if (adminData === null || adminData.length === 0) {
        return;
    }
    let sendData = {task: "", field: "", adminData: ""};
    sendData.task = "save_table";
    sendData.field = adminData[0].field;
    sendData.adminData = JSON.stringify(adminData);
    $.post("admin", sendData).then(response => {
        if (typeof response === 'string'  && response.startsWith("ERROR:")) {
            toast("Error saving Admin table. Refresh and retry.", "alert");
        } else {
            adminData = null; //Important to reset adminData after changes were saved, or they are repeated again and again.
            adminData = JSON.parse(response);
            if (sendData.field === "SITE") {
                siteData = JSON.parse(response);
            }
            buildTable();
            toast("Changes saved.", "success");
            btnSave.off()
        }
    });
}

function buildTable() {
    building = true;
    adminData.sort(getSortOrder("sequence"));
    const tableSelected = $("#tableSelected").val();
    const title = menuLabels.get(adminData[0].field);
    const headerExtraColumn = setHeaderExtraColumn(tableSelected);
    let cnt = adminData.length;
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
                        <span class='mif-plus mif-2x fg-green'></span>
                    </a>
                </td>
            </tr>
            </tbody>
            </table>
        `;
    }
    tbl += "</tbody></table>";

    $("#adminTable").html(tbl);
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
        <button class='button info' onclick="popPermissions('${rowId}');">Edit...</button>
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
    const tableSelected = $("#tableSelected").val();
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
    const ctlId = "#active" + id.toString();
    let idx = adminData.findIndex((item => item.id === id));
    if (idx >= 0 && idx < adminData.length) {
        adminData[idx].active = ($(ctlId).prop("checked")) ? 1 : 0;
        adminData[idx].update = true; //Flag server to update this record
    }
}

function doTextbox(rowId) {
    const id = txt2Int(rowId);
    const ctlId = "#descr" + id.toString();
    const errorId = "#descrerror" + id.toString();
    const descr = $(ctlId).val();
    $(errorId).hide();
    
    // Validate the input text (txt)
    if (descr === undefined || descr.length < 1 || descr !== descr.replace(/[^a-zA-Z0-9.() ]/g, '')) {
        $(errorId).show();
        return;
    }

    let idx = adminData.findIndex((item => item.id === id));
    if (idx >= 0 && idx < adminData.length) {
        if (adminData[idx].description !== descr) {
            adminData[idx].description = descr;
            adminData[idx].update = true; //Flag server to update this record
        }
    }
}




function doAssetId(rowId) {
    const id = txt2Int(rowId);
    const ctlId = "#asset" + id.toString();
    const errorId = "#asseterror" + id.toString();
    const txt = $(ctlId).val();
    $(errorId).hide();
    
    // Validate the input text (txt)
    if (txt === undefined || txt.length < 1 || txt !== txt.replace(/[^a-zA-Z0-9]/g, '')) {
        $(errorId).show();
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
    const ctlId = "#parent" + id.toString();
    let idx = adminData.findIndex((item => item.id === id));
    if (idx >= 0 && idx < adminData.length) {
        adminData[idx].parent = $(ctlId).val();
        adminData[idx].update = true; //Flag server to update this record
    }
}

function moveRowUp(rowId) {
    const id = txt2Int(rowId);
    let idx = adminData.findIndex((item => item.id === id));
    let seq = adminData[idx].sequence;
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
    let idx = adminData.findIndex((item => item.id === id));
    let seq = adminData[idx].sequence;
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
    let idx = adminData.findIndex((item => item.id === id));
    if (idx >= 0 && idx < adminData.length) {
        adminData[idx].delete = true;
    }
    buildTable();
}

function addRow() {
    const code = $("#code0").val().toString().toUpperCase();
    const descr = $("#descr0").val().toString();
    let parent = "";
    let permissions = "";
    const tableSelected = $("#tableSelected").val();
    
    if (tableSelected === "OFFICE") {
        parent = $("#parent0").val().toString();
        $("#parenterror0").hide();
    } else if (tableSelected === "PERMISSIONS") {
        permissions = readPermissions();
    } else if (tableSelected === "TROUBLE") {
        parent = $("#parent0").val().toString();
        $("#parenterror0").hide();
    }
    const active = $("#active0").prop("checked") ? 1 : 0;
    $("#codeerror0").hide();
    $("#descrerror0").hide();
    
    let idx = adminData.findIndex((item => item.code === code));
    
    if (tableSelected === "GROUP") {
        if (!isDigits(code) || idx > -1 || code.length < 1 || txt2Int(code) < 1 ) {
            $("#codeerror0").show();
            return;
        }
    } else if (tableSelected === "IMPACT") {
        if (!isDigits(code) || idx > -1 || code.length < 1 || txt2Int(code) < 1 ) {
            $("#codeerror0").show();
            return;
        }
    } else if (tableSelected === "GEOFENCE") {
        if (!isLatLon(code) || idx > -1 || code.length < 1) {
            $("#codeerror0").show();
            return;
        }
    } else {
        if (code !== code.replace(/[+-]?[^A-Za-z0-9 ]/g, '') || idx > -1 || code.length < 1) {
            $("#codeerror0").show();
            return;
        }
    }

    if (descr !== descr.replace(/[+-]?[^A-Za-z0-9 ]/g, '') || descr.length < 1) {
        $("#descrerror0").show();
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
    latLonOnly = /^[-+]?([1-8]?\d(\.\d+)?|90(\.0+)?),\s*[-+]?(180(\.0+)?|((1[0-7]\d)|([1-9]?\d))(\.\d+)?)$/;
    return latLonOnly.test(value);   
}

function isDigits(value) {
    if (typeof value !== "string" || value.length < 1) return false; //true????
    digitsOnly = /^\d+$/;
    return digitsOnly.test(value);
}


function showHelp() {
    Metro.dialog.open("#helpDialog");
}

function showTableSelect() {
    Metro.dialog.open("#tableDialog");
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
    $("#permRowId").val(id);
    let idx = adminData.findIndex((item => item.id === id));
    let perms = "P:D:S:T:A";
    let dialogTitle = "Group";
    if (idx >-1 ) {
        dialogTitle = adminData[idx].description;
        perms = adminData[idx].permissions;
    }
    $("#GroupName").val(dialogTitle);
    Metro.dialog.open("#permissionsDialog")
    let crud = perms.split(":");
    for (let i = 0; i < crud.length; i++) {
        const val = lookup[crud[i].slice(1)] || 0;
        let slider = sliders.get(crud[i].charAt(0));
        $(slider).val(val);
        setSlider(slider, val);
    }
}

function updatePermissions() {
    let permissions = readPermissions();
    const id = txt2Int($("#permRowId").val());
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
        const val = parseInt($(selector).val(), 10) || 0;
        return starters[i] + lookup[val];
    }).join("").toUpperCase().replace(/[^CRUD:PDSAT]/g, '');
    // remove trailing : from permissions string
    if (permissions.endsWith(":")) {
        permissions = permissions.slice(0, -1);
    }
    return permissions;
}

function setSlider(ctrl, val) {
    let str = ["None", "Create", "Create and Read", "Create, Read and Update", "Create, Read, Update and Delete"];
    const i = txt2Int(val);    
    $(ctrl).val(i);
    $(ctrl+"Label").text(str[i]);
}

const btnSave = {
    id: "btnSave",
    state: "on",
    on() {
        if ($("#canSave").val() === "1" && this.state === "off") {
            $("#" + this.id).show();
            this.state = "on";
        }
    },
    off() {
        if (this.state === "on") {
            $("#" + this.id).hide();
            this.state = "off";
        }
    }
};

function triggersBtnSave() {
    $("input").on("input", function () {
        btnSave.on()
    });
    $("select").change(function () {
        btnSave.on();
    });
}
