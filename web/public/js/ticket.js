// Ticket.js - JavaScript for ticket.html

let btnSave = {
    id: "btnSave",
    state: "on",
    on: function () {
        const canSave = document.getElementById("canSave");
        const btn = document.getElementById(this.id);
        if (canSave && canSave.value === "1") {
            if (btn) btn.style.display = "block";
            this.state = "on";
        }
    },
    off: function () {
        const btn = document.getElementById(this.id);
        if (btn) btn.style.display = "none";
        this.state = "off";
    }
};


//As user changes something turn on the Save button
document.addEventListener('DOMContentLoaded', function() {
    btnSave.off();
const activeEl = document.getElementById("active");
    if (activeEl && activeEl.value === "1") {
        const addButton = document.getElementById("addButton");
        if (addButton) addButton.removeAttribute("disabled"); //Add comment button
        document.querySelectorAll("input").forEach(el => {
            el.addEventListener("input", () => btnSave.on());
        });
        document.querySelectorAll("select").forEach(el => {
            el.addEventListener("change", () => btnSave.on());
        });
        const minusButton = document.getElementById("minusButton");
        if (minusButton) {
            minusButton.removeAttribute("disabled");
            minusButton.addEventListener("click", () => btnSave.on());
        }
        const plusButton = document.getElementById("plusButton");
        if (plusButton) {
            plusButton.removeAttribute("disabled");
            plusButton.addEventListener("click", () => btnSave.on());
        }
        setInterval(updateDuration, 60000);
    }
});


//Duration is the elapsed time the ticket has been open (in seconds)
//Convert into days or hrs or mins
function updateDuration() {
    let retVal = "";
    const openedGMT = document.getElementById("openedGMT");
    const closedGMT = document.getElementById("closedGMT");
    const start = openedGMT ? txt2Int(openedGMT.value) : 0;
    const end = closedGMT ? txt2Int(closedGMT.value) : 0;
    const now = Math.floor(Date.now() / 1000);

    if (start === 0) return;

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
    const durationEl = document.getElementById("duration");
    if (durationEl) durationEl.innerHTML = retVal;
}

async function updateCtrl(task, target) {
    if (pageLoading) return;
    try {
        const response = await fetch("ticket", {
            method: "POST",
            body: new URLSearchParams(getFormData(task))
        });
        const html = await response.text();
        const targetEl = document.getElementById(target);
        if (targetEl) targetEl.innerHTML = html;
    } catch (error) {
        console.error("Error while posting data:", error);
    }
}


function addComment() {
    const logEl = document.getElementById("log");
    let log = logEl ? logEl.value : "";
    if (log.length === 0) {
        if (logEl) logEl.focus();
        return false;
    }
    const cmdEl = document.getElementById("cmd");
    const cmd = cmdEl ? cmdEl.value : "";
    updateCtrl("add_log", "workLog");
    if (logEl) logEl.value = ""; 
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
    const aidEl = document.getElementById("aid");
    const aid = aidEl ? txt2Int(aidEl.value) : NaN;
    if (!isNaN(aid)) {
        window.location.href = `ticket.html?aid=${encodeURIComponent(aid)}`;
    }
}

function plusInform() {
    Metro.dialog.open('#addInform');
}

function minusInform() {
    const informs = document.getElementById('informs');
    if (!informs) return;
    const selectedOption = informs.options[informs.selectedIndex];
    if (selectedOption) {
        informs.remove(informs.selectedIndex);
    }
}

function addInform() {
    const selectElement = document.getElementById('inform');
    if (!selectElement) return;
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
    if (!select) return false;
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
    if (select) {
    for (let i = 0; i < select.options.length; i++) {
        informsList.push(txt2Int(select.options[i].value));
    }
    }
    let informsCSV = informsList.join(',');    
    
    const val = (id) => {
        const el = document.getElementById(id);
        return el ? el.value : "";
    };
    const checked = (id) => {
        const el = document.getElementById(id);
        return (el && el.checked) ? 1 : 0;
    };

    let sendData = { 
        task: task,
        aid: txt2Int(val("aid")), 
        cid: txt2Int(val("cid")), 
        cid_ack: checked("cid_ack"), 
        sid: txt2Int(val("sid")), 
        sid_ack: checked("sid_ack"),
        trouble: txt2Int(val("trouble")), 
        report: val("report"), 
        impact: txt2Int(val("impact")), 
        gid: txt2Int(val("gid")), 
        uid: txt2Int(val("uid")), 
        uid_ack: checked("uid_ack"), 
        inform_gid: txt2Int(val("inform_gid")), 
        inform: txt2Int(val("inform")), 
        inform_ack: checked("inform_ack"),
        cmd:  val("cmd"), 
        log: val("log"),
        oldgid: txt2Int(val("oldgid")),
        oldgroup: val("oldgroup"),
        olduid: txt2Int(val("olduid")),
        olduser: val("olduser"),
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
        cid: txt2Int(document.getElementById("cid")?.value),
        gid: txt2Int(document.getElementById("gid")?.value),
        uid: txt2Int(document.getElementById("uid")?.value),
        site: "",
        office: "",
        impact: txt2Int(document.getElementById("impact")?.value),
        trouble: txt2Int(document.getElementById("trouble")?.value),
        wizard: "",
        type: document.getElementById("type")?.value,
        inform_gid: txt2Int(document.getElementById("inform_gid")?.value),
        isReadonly: false,
    }
    return droplistRequest;
}
/**************************************************/