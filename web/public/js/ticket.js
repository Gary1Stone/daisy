// Ticket.js - JavaScript for ticket.html

let btnSave = {
    id: "btnSave",
    state: "on",
    on: function () {
        if ($("#canSave").val() === "1") {
            $("#btnSave").show(); 
            this.state = "on";
        }
    },
    off: function () {
        $("#btnSave").hide();
        this.state = "off";
    }
};


//As user changes something turn on the Save button
$(document).ready(function () {
    btnSave.off();
    const active = $("#active").val();
    if (active === "1") {
        $("#addButton").removeAttr("disabled"); //Add comment button
        $("input").on("input", function () {
            btnSave.on(); 
        });
        $("select").change(function () {
            btnSave.on();
        });
        $("#minusButton").removeAttr("disabled");
        $("#minusButton").click(function () {
            btnSave.on();
        });
        $("#plusButton").removeAttr("disabled");
        $("#plusButton").click(function () {
            btnSave.on();
        });
        setInterval(updateDuration, 60000);
    }
});


//Duration is the elapsed time the ticket has been open (in seconds)
//Convert into days or hrs or mins
function updateDuration() {
    let retVal = "";
    const start = txt2Int($("#openedGMT").val());
    const end = txt2Int($("#closedGMT").val());
    const now = Math.floor(Date.now() / 1000);
    let duration = now - start;
    if (end > 1000) {
        duration = end - start;
    }
    const days = Math.floor(duration / 86400);
    const hours = Math.floor((duration % 86400) / 3600);
    const minutes = Math.floor(((duration % 86400) % 3600) / 60);
    if (days > 0) {
        retVal += `${days} days `;
    } else if (hours > 0) {
        retVal += `${hours} hrs `;
    } else {
        retVal += `${minutes} mins`;
    }
    $("#duration").html(retVal);
}

async function updateCtrl(task, target) {
    if (pageLoading) return;
    try {
        const response = await $.post("ticket", getFormData(task));
        $("#" + target).html(response);
    } catch (error) {
        console.error("Error while posting data:", error);
    }
}


function addComment() {
    let log = $("#log").val();
    if (log.length === 0) {
        $("#log").focus();
        return false;
    }
    const cmd = $("#cmd").val();
    updateCtrl("add_log", "workLog");
    $("#log").val(""); 
    if (cmd === "CLOSED") {
        setTimeout(reloadTicket, 200);
    }
    return false;
}

function showRouteDialog() {
    Metro.dialog.open("#routeDialog");
}

function route() {
    updateCtrl("route_ticket", "workLog");
    setTimeout(reloadTicket, 200);
}

const reloadTicket = () => {
    const aid = txt2Int($("#aid").val());
    if (!isNaN(aid)) {
        window.location.href = `ticket.html?aid=${encodeURIComponent(aid)}`;
    }
}

function plusInform() {
    Metro.dialog.open('#addInform');
}

function minusInform() {
    const informs = document.getElementById('informs');
    const selectedOption = informs.options[informs.selectedIndex];
    if (selectedOption) {
        informs.remove(informs.selectedIndex);
    }
}

function addInform() {
    const selectElement = document.getElementById('inform');
    const selectedOption = selectElement.options[selectElement.selectedIndex];
    if (selectedOption) {
        const uid = selectedOption.value;
        const name = selectedOption.text;
        if (txt2Int(uid) > 0) {
            const informlist = document.getElementById('informs');
            let newOption = document.createElement("option");
            newOption.text = name;
            newOption.value = uid;
            if (!isItInTheInformsListAlready(uid)) {
                informlist.add(newOption);
            }
        }
    }
}

function isItInTheInformsListAlready(uid) {
    const select = document.getElementById("informs");
    let optionExists = false;
    for (let i = 0; i < select.options.length; i++) {
        if (select.options[i].value === uid) {
            optionExists = true;
            break;
        }
    }
    return optionExists
}

function getFormData(task) {
    let informsList = [];
    const select = document.getElementById("informs");
    for (let i = 0; i < select.options.length; i++) {
        informsList.push(txt2Int(select.options[i].value));
    }
    let informsCSV = informsList.join(',');    
    let sendData = { 
        task: task,
        aid: txt2Int($("#aid").val()), 
        cid: txt2Int($("#cid").val()), 
        cid_ack: $("#cid_ack").prop("checked") ? 1 : 0, 
        sid: txt2Int($("#sid").val()), 
        sid_ack: $("#sid_ack").prop("checked") ? 1 : 0,
        trouble: txt2Int($("#trouble").val()), 
        report: $("#report").val(), 
        impact: txt2Int($("#impact").val()), 
        gid: txt2Int($("#gid").val()), 
        uid: txt2Int($("#uid").val()), 
        uid_ack: $("#uid_ack").prop("checked") ? 1 : 0, 
        inform_gid: txt2Int($("#inform_gid").val()), 
        inform: txt2Int($("#inform").val()), 
        inform_ack: $("#inform_ack").prop("checked") ? 1 : 0,
        cmd:  $("#cmd").val(), 
        log: $("#log").val(),
        oldgid: txt2Int($("#oldgid").val()),
        oldgroup: $("#oldgroup").val(),
        olduid: txt2Int($("#olduid").val()),
        olduser: $("#olduser").val(),
        informs: informsCSV
    };
    return sendData
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
function getInformPersonCtrl() {
    getCtrl("selectInformUser", ctrlData("USERINFORM"));
}
function ctrlData(task) {
    let droplistRequest = { 
        task: task, 
        isTicket: false,
        isWizard: false,
        cid: txt2Int($("#cid").val()),
        gid: txt2Int($("#gid").val()),
        uid: txt2Int($("#uid").val()),
        site: "",
        office: "",
        impact: txt2Int($("#impact").val()),
        trouble: txt2Int($("#trouble").val()),
        wizard: "",
        type: $("#type").val(),
        inform_gid: txt2Int($("#inform_gid").val()),
        isReadonly: false,
    }
    return droplistRequest;
}
/**************************************************/