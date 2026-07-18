// wizard.js

let page = 0;
let endReached = false;

const wizards = {
    SIGHTING: { icon: "eye", title: "Sighting", steps: ["DeviceSelect", "SiteSelect", "LocationText", "UserSelect", "NotesText"] },
    USING: { icon: "check", title: "Using", steps: ["DeviceSelect", "NotesText"] },
    CLAIMING: { icon: "tag", title: "Claiming", steps: ["DeviceSelect", "SiteSelect", "LocationText", "NotesText"] },
    GIVING: { icon: "tag", title: "Giving", steps: ["DeviceSelect", "SiteSelect", "LocationText", "UserSelect", "NotesText"] },
    BROKEN: { icon: "broken", title: "Broken", steps: ["DeviceSelect", "SiteSelect", "LocationText", "UserSelect", "TroubleSelect", "NotesText", "ImpactSelect"] },
    LOST: { icon: "steps", title: "Lost", steps: ["DeviceSelect", "UserSelect", "TroubleSelect", "NotesText", "ImpactSelect"] },
    DIED: { icon: "stethoscope", title: "Died", steps: ["DeviceSelect", "UserSelect", "TroubleSelect", "NotesText", "ImpactSelect"] },
    CARE: { icon: "wrench", title: "Care", steps: ["DeviceSelect", "SiteSelect", "LocationText", "UserSelect", "TroubleSelect", "NotesText", "ImpactSelect"] },
    BACKUP: { icon: "copy", title: "Backup", steps: ["DeviceSelect", "NotesText"] },
    INSTALL: { icon: "software", title: "Install", steps: ["DeviceSelect", "SoftwareSelect", "DateChoose", "NotesText"] },
    REMOVE: { icon: "software", title: "Remove", steps: ["DeviceSelect", "SoftwareSelect", "DateChoose", "NotesText"] },
    REQUEST: { icon: "install", title: "Request", steps: ["DeviceSelect", "SoftwareSelect", "UserSelect", "TroubleSelect", "NotesText"] }
};

const UI = {
    thanksDialog: () => document.getElementById("thanksDialog"),
    helpDialog: () => document.getElementById("helpDialog"),
    installed:() => document.getElementById("installed"),
    filterDialog: () => document.getElementById("filterDialog"),
    cid: () => document.getElementById("cid"),
    type: () => document.getElementById("type"),
    siteSearch: () => document.getElementById("siteSearch"),
    officeSearch: () => document.getElementById("officeSearch"),
    groupSearch: () => document.getElementById("groupSearch"),
    userSearch: () => document.getElementById("userSearch"),
    textSearch: () => document.getElementById("textSearch"),
    isLate: () => document.getElementById("isLate"),
    isMissing: () => document.getElementById("isMissing")
}

function el(id) { return document.getElementById(id); }
function val(id) { const e = el(id); return e ? e.value : ""; }
function html(id) { const e = el(id); return e ? e.innerHTML : ""; }
function setHtml(id, content) { const e = el(id); if (e) e.innerHTML = content; }
function setVal(id, v) { const e = el(id); if (e) e.value = v; }
function showEl(id) { const e = el(id); if (e) e.style.display = ""; }
function hideEl(id) { const e = el(id); if (e) e.style.display = "none"; }
function addClass(id, cls) { const e = el(id); if (e) e.classList.add(cls); }
function removeClass(id, cls) { const e = el(id); if (e) e.classList.remove(cls); }

document.addEventListener('DOMContentLoaded', function() {
    const wizKey = val("wizKey");
    const cid = val("cid"); // If have cid, skip past device select/filter
    nav.init(wizKey, cid);

    const today = new Date().toISOString().split('T')[0]; // get YYYY-MM-DD
    const cal = UI.installed();
    if (cal) {
        cal.value = today;
        cal.setAttribute('max', today);
    }
});

function selectDevice(cid) {
    setVal("cid", cid);
    const cidJson = html("cid" + cid);
    const dev = JSON.parse(cidJson);
    setHtml("searchResults", "");
    const pic = el("pic"); if (pic) pic.src = "images/" + dev.small_image;
    const model = dev.model.substring(0,25);
    const icon = "<span class='" + dev.icon +" icon'></span>&nbsp;"
    setHtml("cidDetails", "<p>" + icon + dev.name + "</p><p>" + dev.make + " (" + dev.year + ")</p><p>" + model + "</p>");
    getSiteCtrl(); // Also Triggers getOfficeCtrl() Changes on computer selected
    getGroupCtrl(); // Also Triggers getPersonControl() User changes on computer selected, thus their group
    getLocationCtrl(); // Changes on computer selected
    nav.next();
}

function popFilters() {
    nav.goto("DeviceSelect");
    const modal = UI.filterDialog();
    if (modal) modal.showModal();
}

function getOfficeSearchCtrl() {
    updateCtrl("get_office_search_control", "selectSearchOffice")
}

function getUserSearchCtrl() {
    updateCtrl("get_user_search_control", "selectSearchUser")
}

// Asynchronously posts data to the server and updates the target element
async function updateCtrl(task, target) {
    if (pageLoading) return;
    let sendData = {};
    if (task.includes("search")) {
        sendData = getSearchFormData(task)
    } else {
        sendData = getWizFormData(task)
    }
    try {
        const response = await postForm("wizard", sendData);
        document.getElementById(target).innerHTML = response;
    } catch (error) {
        console.error("Error while posting data:", error);
    }
}

function search() {
    if (pageLoading) return;
    page = 0;
    sendData = getSearchFormData("search_for_devices");
    postForm("devices", sendData, (response) => {
        document.getElementById("cards").innerHTML = response;
        let msg = "";
        if (response.length < 10) {
            msg = "No device matches found!";
            openModal(UI.filterDialog());
        } else {
            closeModal(UI.filterDialog());
        }
        setHtml("searchError", msg);
    });
}

function getSearchFormData(task) {
    let sendData = {
        task: task,
        page: page,
        cid: UI.cid().value,
        devtype:  UI.type().value,
        gid: txt2Int(UI.groupSearch().value),
        uid: txt2Int(UI.userSearch().value),
        site: UI.siteSearch().value,
        office: UI.officeSearch().value,
        searchtxt: UI.textSearch().value,
        islate: UI.isLate().checked,
        ismissing: UI.isMissing().checked
    }
    return sendData;
}

function getWizFormData(task) {
    //date to UTC seconds
    const now = new Date();
    let unixTimestampSeconds = Math.floor(now.getTime() / 1000);
    const dateControl = UI.installed();
    const dateString = dateControl.value; // e.g., "2026-06-22"
    if (dateString) {
        const localDate = new Date(`${dateString}T00:00`); // 1. Force local time interpretation by appending T00:00
        unixTimestampSeconds = Math.floor(localDate.getTime() / 1000);  // 2. Convert to Unix Epoch in seconds (This value is implicitly UTC)
    }

    let sendData = {
        task: task,
        cid: txt2Int(UI.cid().value), // Computer ID
        sid: txt2Int(val("sid")), // Software ID
        gid: txt2Int(val("gid")), // Group ID for the user
        uid: txt2Int(val("uid")), // User ID (person the action is assigned to, not to be informed
        site: val("site"),
        office: val("office"),
        location: val("location"),
        notes: val("notes"),
        installed: unixTimestampSeconds,
        impact: txt2Int(val("impact")),
        trouble: txt2Int(val("trouble")),
        wizard: val("wizKey"),
        type: val("type"), // Device Type DESKTOP, LAPTOP...
    }
    return sendData
}

// Check if the user has scrolled to the bottom of the page
window.addEventListener('scroll', function() {
    const scrollTop = window.scrollY || window.pageYOffset;
    const windowHeight = window.innerHeight;
    const documentHeight = document.documentElement.scrollHeight;

    if (scrollTop + windowHeight >= documentHeight - 200) {
        if (!endReached && !pageLoading) {
            pageLoading = true;
            page++;
            getNextBlock();
        }
    }
});
  
// Fetches the next block of devices and appends it to the existing content
async function getNextBlock() {
    try {
        const sendData = getSearchFormData("search_for_devices");
        const response = await postForm("devices", sendData);
        if (response.length === 0) {
            endReached = true;
        } else {
            document.getElementById("cards").insertAdjacentHTML('beforeend', response);
        } 
        // Wait 1/2 second before getting next block
        setTimeout(() => { pageLoading = false; }, 500);
    } catch (error) {
        console.error("Error while posting data:", error);
        pageLoading = false;
    }
}

let nav = {
    curWizard: "SIGHTING",
    curStep: -1,
    numSteps: 0,
    wizKeys: ["SIGHTING", "USING", "CLAIMING", "GIVING", "BROKEN", "LOST", "DIED", "CARE", "BACKUP", "INSTALL", "REMOVE", "REQUEST"],
    panels: [],
    labels: {},
    init: function (wizKey, cid) {
        this.curWizard = this.wizKeys.find((key) => {
            return key === wizKey;
        });
        if (typeof this.curWizard === "undefined" || this.curWizard === null) {
            this.curWizard = "SIGHTING";
        }
        this.curStep = 0;
        this.wizKeys.forEach((wizKey) => {
            wizards[wizKey].steps.forEach((panel) => {
                if (!this.panels.includes(panel)) {
                    //Build list of all the panel names/IDs
                    this.panels.push(panel); //to hide all in a simple loop for navigation
                }
            });
        });
        this.numSteps = wizards[this.curWizard].steps.length - 1;
        this.labels = JSON.parse(html("labels"));
        setHtml("cidLabel", this.labels[this.curWizard].DeviceSelect);
        setHtml("siteLabel", this.labels[this.curWizard].SiteSelect);
        setHtml("locationLabel", this.labels[this.curWizard].LocationText);
        setHtml("uidLabel", this.labels[this.curWizard].UserSelect);
        setHtml("notesLabel", this.labels[this.curWizard].NotesText);
        setHtml("impactLabel", this.labels[this.curWizard].ImpactSelect);
        setHtml("sidLabel", this.labels[this.curWizard].SoftwareSelect);
        setHtml("dateLabel", this.labels[this.curWizard].DateChoose);
        setHtml("troubleLabel", this.labels[this.curWizard].TroubleSelect); //new
        if (cid > 0) {
            nav.next();
        }
        this.show();
    },
    next: function () {
        this.curStep++;
        if (this.curStep > this.numSteps) {
            this.curStep--;
        }
        this.show();
        if (this.curStep > 0) {
            showEl("SelectedDevice");
        } else {
            hideEl("SelectedDevice");
        }
    },
    previous: function () {
        this.curStep--;
        if (this.curStep < 0) {
            this.curStep = 0;
        }
        this.show();
        if (this.curStep < 1) {
            hideEl("SelectedDevice");
        } else {
            showEl("SelectedDevice");
        }
    },
    goto: function (panelName) {
        let idx = wizards[this.curWizard].steps.indexOf(panelName);
        if (idx >= 0 && idx <= this.numSteps) {
            this.curStep = idx;
        }
        this.show();
        if (this.curStep > 0) {
            showEl("SelectedDevice");
        } else {
            hideEl("SelectedDevice");
        }
    },
    hide: function () {
        this.panels.forEach((panel) => {
            hideEl(panel);
        });
    },
    show: function () {
        this.hide();
        showEl(wizards[this.curWizard].steps[this.curStep]);
        this.setButtons();
    },
    setButtons() {
        removeClass("btnPrev", "disabled"); //enable back button
        removeClass("btnNext", "disabled"); //enable next button
        addClass("btnFinish", "disabled"); //disable finish button
        if (this.curStep === 0) {
            addClass("btnPrev", "disabled");
            addClass("btnNext", "disabled");
            openModal(UI.filterDialog());
        } else if (this.curStep === this.numSteps) {
            addClass("btnNext", "disabled");
            removeClass("btnFinish", "disabled");
        } else {
            removeClass("btnNext", "disabled"); //enable next button
        }
    },
    getHelpMessage() {
        return this.labels[this.curWizard].Help;
    },
};

function saveDate(sel, day, el) {
    let secs = sel[0] / 1000;
    secs = secs.toFixed(0);
    setVal("installed", secs);
}

function getSiteCtrl() {
    updateCtrl("get_site_control","selectSite");
}

function getOfficeCtrl() {
    updateCtrl("get_office_control","selectOffice");
}

function getGroupCtrl() {
    updateCtrl("get_group_control","selectGroup");
}

function getPersonCtrl() {
    updateCtrl("get_person_control","selectPerson");
}

function getImpactCtrl() {
    updateCtrl("get_impact_control","selectImpact");
}

function getLocationCtrl() {
    updateCtrl("get_location_from_device","locationDiv");
}

function getTroubleCtrl() {
    updateCtrl("get_trouble_control","selectTrouble");
}

function showHelp() {
    setHtml("helpMsg", nav.getHelpMessage());
    openModal(UI.helpDialog());
}

//check Mandatory fields for action
function isMandatory(sendData) {
    let isGood = true;
    switch (nav.curWizard) {
        case "SIGHTING": {
            isGood = true;
            break;
        }
        case "USING": {
            isGood = true;
            break;
        }
        case "CLAIMING": {
            isGood = true;
            break;
        }
        case "GIVING": {
            isGood = isUidGidOfficeValid(sendData.uid, sendData.gid, sendData.office);
            break;
        }
        case "BROKEN": {
            isGood = isNoteValid(sendData.notes);
            break;
        }
        case "LOST": {
            isGood = isNoteValid(sendData.notes);
            break;
        }
        case "DIED": {
            isGood = isNoteValid(sendData.notes);
            break;
        }
        case "CARE": {
            isGood = isUidValid(sendData.uid, sendData.gid) && isNoteValid(sendData.notes);
            break;
        }
        case "BACKUP": {
            isGood = isNoteValid(sendData.notes);
            break;
        }
        case "INSTALL": {
            isGood = isSidValid(sendData.sid) && isInstallValid(sendData.installed) && isOtherComputerValid(sendData.cid);
            break;
        }
        case "REMOVE": {
            isGood = isSidValid(sendData.sid) && isInstallValid(sendData.installed) && isOtherComputerValid(sendData.cid);
            break;
        }
        case "REQUEST": {
            isGood = isSoftwareRequestValid(sendData.sid, sendData.notes);
            break;
        }
    }
    return isGood;
}

function isSoftwareRequestValid(sidVal, noteVal) {
    if (sidVal === 0 && noteVal.length === 0) {
        nav.goto("NotesText");
        showEl("notesError");
        return false;
    }
    return true;
}

function isNoteValid(noteVal) {
    if (noteVal.length === 0) {
        nav.goto("NotesText");
        showEl("notesError");
        return false;
    }
    return true;
}

//can assign to a user/group or just a group
function isUidValid(uidVal, gidVal) {
    if (uidVal === 0 && gidVal === 0) {
        nav.goto("UserSelect");
        addClass("uidLabel", "fg-red");
        return false;
    }
    return true;
}

function isUidGidOfficeValid(uidVal, gidVal, officeVal) {
    if ((uidVal === 0 && gidVal === 0) || (officeVal.length === 0)) {
        addClass("uidLabel", "fg-red");
        nav.goto("SiteSelect");
        addClass("siteLabel", "fg-red");
        return false;
    }
    return true;
}

//Note, not sure we want the user to select no software for install/remove.
//We should only track the packages we purchased.
//But if the software was installed on someone elses computer..., we still need to track it
function isSidValid(sidVal) {
    if (sidVal === 0 && val("notes").length === 0) {
        nav.goto("NotesText");
        showEl("notesError");
        return false;
    }
    return true;
}

function isOtherComputerValid(cidVal) {
    if (cidVal === 0 && val("notes").length === 0) {
        nav.goto("NotesText");
        showEl("notesError");
        return false;
    }
    return true;
}

function isInstallValid(secondsSinceEpoch) {
    const now = new Date();
    let nowTimestamp = now.getTime() / 1000;
    nowTimestamp = nowTimestamp.toFixed(0); 
    // 978307201 = jan 1, 2001 GMT
    if (secondsSinceEpoch < 978307201 || secondsSinceEpoch > nowTimestamp) {
        nav.goto("DateChoose");
        showEl("dateError");
        return false;
    }
    return true;
}

function onFinish() {
    const sendData = getWizFormData(nav.curWizard.toUpperCase());
    if (!isMandatory(sendData)) {
        return;
    }
    addClass("btnFinish", "disabled"); //disable finish button so no second submission
    postForm("wizard", sendData, (response) => {
        if (response === "Okay") {
            nav.hide();
            showEl("thankYou");
            setTimeout(() => { window.location.href = encodeURI("home.html"); }, 3000); 
        } else {
            console.log(response);
        }
    });
}
