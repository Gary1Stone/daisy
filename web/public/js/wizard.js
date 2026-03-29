/* global Metro */

let page = 0;
let pageLoading = true;
let endReached = false;

const wizards = {
    SIGHTING: { icon: "mif-eye", title: "Sighting", steps: ["DeviceSelect", "SiteSelect", "LocationText", "UserSelect", "NotesText"] },
    USING: { icon: "mif-checkmark", title: "Using", steps: ["DeviceSelect", "NotesText"] },
    CLAIMING: { icon: "mif-tag", title: "Claiming", steps: ["DeviceSelect", "SiteSelect", "LocationText", "NotesText"] },
    GIVING: { icon: "mif-tag", title: "Giving", steps: ["DeviceSelect", "SiteSelect", "LocationText", "UserSelect", "NotesText"] },
    BROKEN: { icon: "mif-heart-broken", title: "Broken", steps: ["DeviceSelect", "SiteSelect", "LocationText", "UserSelect", "TroubleSelect", "NotesText", "ImpactSelect"] },
    LOST: { icon: "mif-steps", title: "Lost", steps: ["DeviceSelect", "UserSelect", "TroubleSelect", "NotesText", "ImpactSelect"] },
    DIED: { icon: "mif-stethoscope", title: "Died", steps: ["DeviceSelect", "UserSelect", "TroubleSelect", "NotesText", "ImpactSelect"] },
    CARE: { icon: "mif-wrench", title: "Care", steps: ["DeviceSelect", "SiteSelect", "LocationText", "UserSelect", "TroubleSelect", "NotesText", "ImpactSelect"] },
    BACKUP: { icon: "mif-copy", title: "Backup", steps: ["DeviceSelect", "NotesText"] },
    INSTALL: { icon: "mif-apps", title: "Install", steps: ["DeviceSelect", "SoftwareSelect", "DateChoose", "NotesText"] },
    REMOVE: { icon: "mif-apps", title: "Remove", steps: ["DeviceSelect", "SoftwareSelect", "DateChoose", "NotesText"] },
    REQUEST: { icon: "mif-file-binary", title: "Request", steps: ["DeviceSelect", "SoftwareSelect", "UserSelect", "TroubleSelect", "NotesText"] }
};

$(document).ready(function () {
    const wizKey = $("#wizKey").val();
    const cid = $("#cid").val(); // If have cid, skip to next step
    nav.init(wizKey, cid);
    const now = new Date(); // Calendar control
    let calendar = Metro.getPlugin($("#cal"), "calendar");
    const maxDate = now.toISOString().split("T")[0]; // get YYYY-MM-DD
    calendar.setMinDate("2001-01-01");
    calendar.setMaxDate(maxDate);
    calendar.setShow(maxDate);
    let secs = now.getTime() / 1000;
    secs = secs.toFixed(0);
    $("#installed").val(secs); //default to today in milliseconds since epoch
    pageLoading = false;
});

function selectDevice(cid) {
    $("#cid").val(cid);
    const cidJson = $("#cid" + cid).html();
    const dev = JSON.parse(cidJson);
    $("#searchResults").html("");
    $("#pic").attr("src", "images/" + dev.small_image);
    const model = dev.model.substring(0,25);
    const icon = "<span class='" + dev.icon +" icon'></span>&nbsp;"
    $("#cidDetails").html("<p>" + icon + dev.name + "</p><p>" + dev.make + " (" + dev.year + ")</p><p>" + model + "</p>");
    getSiteCtrl(); // Also Triggers getOfficeCtrl() Changes on computer selected
    getGroupCtrl(); // Also Triggers getPersonControl() User changes on computer selected, thus their group
    getLocationCtrl(); // Changes on computer selected
    nav.next();
}

function popFilters() {
    nav.goto("DeviceSelect");
    Metro.dialog.open("#searchDialog");
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
        const response = await $.post("wizard",sendData);
        $("#" + target).html(response);
    } catch (error) {
        console.error("Error while posting data:", error);
    }
}

function getOfficeSearchCtrl() {
    updateCtrl("get_office_search_control", "selectSearchOffice")
}

function getUserSearchCtrl() {
    updateCtrl("get_user_search_control", "selectSearchUser")
}

function search() {
    if (pageLoading) return;
    page = 0;
    sendData = getSearchFormData("search_for_devices");
    $.post("devices", sendData).then((response) => {
        $("#cards").html(response);
        let msg = "";
        if (response.length < 10) {
            msg = "No device matches found!";
            Metro.dialog.open("#searchDialog");
        }
        $("#searchError").html(msg);
    });
}

function getSearchFormData(task) {
    let sendData = {
        task: task,
        page: page,
        cid: txt2Int($("#cid").val()),
        devtype: $("#type").val(),
        site: $("#siteSearch").val(),
        office: $("#officeSearch").val(),
        gid: txt2Int($("#groupSearch").val()),
        uid: txt2Int($("#userSearch").val()),
        searchtxt: $("#txtSearch").val(),
        islate: 0,
        ismissing: 0
    }
    return sendData;
}

function getWizFormData(task) {
    let sendData = {
        task: task,
        cid: txt2Int($("#cid").val()), // Computer ID
        sid: txt2Int($("#sid").val()), // Software ID
        gid: txt2Int($("#gid").val()), // Group ID for the user
        uid: txt2Int($("#uid").val()), // User ID (person the action is assigned to, not to be informed
        site: $("#site").val(),
        office: $("#office").val(),
        location: $("#location").val(),
        notes: $("#notes").val(),
        installed: txt2Int($("#installed").val()),
        impact: txt2Int($("#impact").val()),
        trouble: txt2Int($("#trouble").val()),
        wizard: $("#wizKey").val(),
        type: $("#type").val(), // Device Type DESKTOP, LAPTOP...
    }
    return sendData
}

// Check if the user has scrolled to the bottom of the page
$(window).scroll(function() {
    if ($(window).scrollTop() + $(window).height() >= $(document).height() - 200) {
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
        const response = await $.post("devices", sendData);
        if (response.length === 0) {
            endReached = true;
        } else {
            $("#cards").append(response);
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
        this.labels = JSON.parse($("#labels").html());
        $("#cidLabel").html(this.labels[this.curWizard].DeviceSelect);
        $("#siteLabel").html(this.labels[this.curWizard].SiteSelect);
        $("#locationLabel").html(this.labels[this.curWizard].LocationText);
        $("#uidLabel").html(this.labels[this.curWizard].UserSelect);
        $("#notesLabel").html(this.labels[this.curWizard].NotesText);
        $("#impactLabel").html(this.labels[this.curWizard].ImpactSelect);
        $("#sidLabel").html(this.labels[this.curWizard].SoftwareSelect);
        $("#dateLabel").html(this.labels[this.curWizard].DateChoose);
        $("#troubleLabel").html(this.labels[this.curWizard].TroubleSelect); //new
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
            $("#SelectedDevice").show();
        } else {
            $("#SelectedDevice").hide();
        }
    },
    previous: function () {
        this.curStep--;
        if (this.curStep < 0) {
            this.curStep = 0;
        }
        this.show();
        if (this.curStep < 1) {
            $("#SelectedDevice").hide();
        } else {
            $("#SelectedDevice").show();
        }
    },
    goto: function (panelName) {
        let idx = wizards[this.curWizard].steps.indexOf(panelName);
        if (idx >= 0 && idx <= this.numSteps) {
            this.curStep = idx;
        }
        this.show();
        if (this.curStep > 0) {
            $("#SelectedDevice").show();
        } else {
            $("#SelectedDevice").hide();
        }
    },
    hide: function () {
        this.panels.forEach((panel) => {
            $("#" + panel).hide();
        });
    },
    show: function () {
        this.hide();
        $("#" + wizards[this.curWizard].steps[this.curStep]).show();
        this.setButtons();
    },
    setButtons() {
        $("#btnPrev").removeClass("disabled"); //enable back button
        $("#btnNext").removeClass("disabled"); //enable next button
        $("#btnFinish").addClass("disabled"); //disable finish button
        if (this.curStep === 0) {
            $("#btnPrev").addClass("disabled");
            $("#btnNext").addClass("disabled");
            Metro.dialog.open("#searchDialog");
        } else if (this.curStep === this.numSteps) {
            $("#btnNext").addClass("disabled");
            $("#btnFinish").removeClass("disabled");
        } else {
            $("#btnNext").removeClass("disabled"); //enable next button
        }
    },
    getHelpMessage() {
        return this.labels[this.curWizard].Help;
    },
};

function saveDate(sel, day, el) {
    let secs = sel[0] / 1000;
    secs = secs.toFixed(0);
    $("#installed").val(secs);
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
    $("#helpMsg").html(nav.getHelpMessage());
    Metro.dialog.open("#helpDialog");
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
        $("#notesError").show();
        return false;
    }
    return true;
}

function isNoteValid(noteVal) {
    if (noteVal.length === 0) {
        nav.goto("NotesText");
        $("#notesError").show();
        return false;
    }
    return true;
}

//can assign to a user/group or just a group
function isUidValid(uidVal, gidVal) {
    if (uidVal === 0 && gidVal === 0) {
        nav.goto("UserSelect");
        $("#uidLabel").addClass("fg-red");
        return false;
    }
    return true;
}

function isUidGidOfficeValid(uidVal, gidVal, officeVal) {
    if ((uidVal === 0 && gidVal === 0) || (officeVal.length === 0)) {
        $("#uidLabel").addClass("fg-red");
        nav.goto("SiteSelect");
        $("#siteLabel").addClass("fg-red");
        return false;
    }
    return true;
}

//Note, not sure we want the user to select no software for install/remove.
//We should only track the packages we purchased.
//But if the software was installed on someone elses computer..., we still need to track it
function isSidValid(sidVal) {
    if (sidVal === 0 && $("#notes").val().length === 0) {
        nav.goto("NotesText");
        $("#notesError").show();
        return false;
    }
    return true;
}

function isOtherComputerValid(cidVal) {
    if (cidVal === 0 && $("#notes").val().length === 0) {
        nav.goto("NotesText");
        $("#notesError").show();
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
        $("#dateError").show();
        return false;
    }
    return true;
}

function onFinish() {
    const sendData = getWizFormData(nav.curWizard.toUpperCase());
    if (!isMandatory(sendData)) {
        return;
    }
    $("#btnFinish").addClass("disabled"); //disable finish button so no second submission
    $.post("wizard", sendData).then((response) => {
        if (response === "Okay") {
            nav.hide();
            $("#thankYou").show();
            setTimeout(() => { window.location.href = encodeURI("home.html"); }, 3000); 
        } else {
            console.log(response);
        }
    });
}
