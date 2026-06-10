// software.js

// Declare iconbar button variables at top level so they are available to checkValid, saveRecord, etc.
let btnSave, btnNew, btnDelete;

// Page loaded event

document.addEventListener('DOMContentLoaded', function() {
    const sidEl = document.getElementById("sid");
    const sid = sidEl ? sidEl.value : "";

    // Initialize the iconbar button instances once the scripts and DOM are ready
    btnSave = new Button("btnSave");
    btnNew = new Button("btnNew");
    btnDelete = new Button("btnDelete", true); // true for forceOffIfNotAllowed

    //Set initial button state depending if a record is displayed or not
    if (isDigits(sid) && txt2Int(sid) === 0) {
        btnSave.on();
        btnNew.off();
        btnDelete.off();
    } else {
        btnSave.off();
        btnNew.on();
        btnDelete.on();
    } 

    //if any of the 'input' elements are modified, change the save/add/delete states  
    document.querySelectorAll("input").forEach(el => {
        el.addEventListener("input", function(){
            btnSave.on();
            btnNew.off();
            btnDelete.off();
        });
    });
    
    //if any of the 'select' droplists are modified, change the save/add/delete states
    document.querySelectorAll("select").forEach(el => {
        el.addEventListener("change", function () {
            btnSave.on();
            btnNew.off();
            btnDelete.off();
        });
    });

    //when user changes the name of the software, check if the name is not already in use
    const nameInput = document.getElementById("name");
    if (nameInput) {
        nameInput.addEventListener("blur", async function () {
        let sendData = getFormData();
        sendData.task = "unique";
            try {
                await postForm("software", sendData, (reply) => {
                    if (typeof reply === "string") reply = JSON.parse(reply);
                    const nameError = document.getElementById("nameError");
                    if (reply.success === true) {
                        nameError.style.display = "none";
                    } else {
                        nameError.value = reply.msg;
                        nameError.style.display = "block";
                    }
                });
            } catch (e) {
                console.error("Uniqueness check failed:", e);
            }
        });
    }
});


//pop up a dialog for displaying the message details
//settings={"color":"light ","action":"INSTALL","label":"Install Soft...","icon":"mif-apps","active":0,"aid":176,"cid_ack":1,"iid_ack":0,"sid_ack":1,"uid_ack":0}
function pop(aid) {
    const popEl = document.getElementById("pop");
    const notesEl = document.getElementById("notes" + aid);
    const aidInput = document.getElementById("aid" + aid);
    const actionIDInput = document.getElementById("actionID");
    const actionNameInput = document.getElementById("actionName");
    const cmdEl = document.getElementById("cmd");

    if (popEl && notesEl) {
        popEl.innerHTML = "<p>" + notesEl.innerHTML + "</p>";
    }
    if (actionIDInput) actionIDInput.value = aid;
    
    if (aidInput) {
        const settings = JSON.parse(aidInput.value);
        if (actionNameInput) actionNameInput.value = settings.action;
        if (cmdEl) {
            cmdEl.style.display = (settings.active && !settings.sid_ack) ? "block" : "none";
        }
    }
    openModal(document.getElementById("NotesDialog"));
}

function acceptAction() {
    const actionIDInput = document.getElementById("actionID");
    const aid = actionIDInput ? txt2Int(actionIDInput.value) : 0;
    fetchLog(aid);
}

async function fetchLog(aid = 0) {
    const sendData = getFormData();
    sendData.task = "getactionlog";
    sendData.aid = aid;
    try {
        await postForm("software", sendData, (html) => {
            const div = document.getElementById("actionLogDiv");
            if (div) div.innerHTML = html;
        });
    } catch (e) {
        console.error("fetchLog failed:", e);
    }
}

// Deleting a record is simply deeing it if not used anywhere
// else setting the active flag = 0 and moving name to old name
// its still in the database but not used again.
function deleteRecord() {
    if (btnDelete.state !== "on") return;
    openModal(document.getElementById("deleteDialog"));
    const name = document.getElementById("name").value;
    if (name.length > 0) {
        document.getElementById("softwareName").innerHTML = name;
    }
}

function confirmDelete() {
    if (btnDelete.state !== "on") return;
    let sendData = getFormData();
    sendData.task = "delete";
    try {
        await postForm("software", sendData, (reply) => {
            if (typeof reply === "string") reply = JSON.parse(reply);
        if (reply.success) {
            addRecord();  //clears the displayed record
        } else {
            closeModal(document.getElementById("deleteDialog"));
            toast(reply.msg, "alert");
        }
        });
    } catch (e) {
        console.error(e);
        toast(e, "error");
    }
}


//Adding a record is a two step process
//Display this screen with a sid=0 (Software ID = SID)
//when user presses save, in the servlet, detect if record id (SID) is 0, then insert record.
//then send the uid to be used inside this form
function addRecord() {
    if (btnNew.state !== "on") return;
    let url = window.location.href;
    const i = url.indexOf("?");
    if (i < 0) {
        url = url + "?sid=0"; // + encodeURIComponent("0");
    } else {
        url = url.substring(0, i) + "?sid=0"; // + encodeURIComponent("0");
    }
    window.location.href =  encodeURI(url);
}


// function isDigits(value) {
//     if (typeof value === "string" && value.length > 0) {
//         const digitsOnly = /^\d+$/;  // d=[0-9] 
//         return digitsOnly.test(value);
//     }
//     return true;
// }

function validateForm(sendData) {
    if (!isDigits(sendData.sid)) return false;    
    const nameInput = document.getElementById("name");
    const nameError = document.getElementById("nameError");
    const theForm = document.getElementById("theForm");

    if (nameInput && !nameInput.checkValidity()) return false;
    //Check if the name is unique (onBlur sets if error message visible or not)
    if (nameError && nameError.style.display !== "none") {
        if (nameInput) nameInput.focus();
        return false;
    }
    if (sendData.link.length > 0) {
        const link = document.getElementById("link");
        if (!link.checkValidity()) {
            console.warn(link.validationMessage);
            return false;
        }
    }    
    return theForm ? theForm.checkValidity() : false;
}


async function saveRecord() {
    if (btnSave.state !== "on") return false;
    let sendData = getFormData();
    if (!validateForm(sendData)) return false;
    if (sendData.sid === 0) {
        sendData.task = "add";
    }

    try {
        const reply = await postForm("software", sendData);
        if (typeof reply === "string") reply = JSON.parse(reply);
        
        if (!reply.success) {
            console.log(reply.msg);  //display error message
        } else {             //Refresh the page
            let url = window.location.href;
            const i = url.indexOf("?");
            if (i < 0) {
                url = url + "?sid=" + reply.sid;
            } else {
                url = url.substring(0, i) + "?sid=" + reply.sid;
            }
            window.location.href =  encodeURI(url);
        }
    } catch (e) {
        console.error("Save failed:", e);
    }
    return false;
}

function getFormData() {
    const sidEl = document.getElementById("sid");
    const nameEl = document.getElementById("name");
    const licensesEl = document.getElementById("licenses");
    const licenseKeyEl = document.getElementById("license_key");
    const productEl = document.getElementById("product");
    const sourceEl = document.getElementById("source");
    const linkEl = document.getElementById("link");
    const notesEl = document.getElementById("notes");
    const activeEl = document.getElementById("active");
    const reuseableEl = document.getElementById("reuseable");
    const filterEl = document.getElementById("filter");
    const purchasedEl = document.getElementById("purchased");
    const invNameEl = document.getElementById("inv_name");
    const preInstalledEl = document.getElementById("pre_installed");
    const freeEl = document.getElementById("free");

    return {
        task: "save",
        sid: sidEl ? txt2Int(sidEl.value) : 0,
        name: nameEl ? nameEl.value.trim() : "",
        licenses: licensesEl ? txt2Int(licensesEl.value) : 0,
        license_key: licenseKeyEl ? licenseKeyEl.value.trim() : "",
        product: productEl ? productEl.value.trim() : "",
        source: sourceEl ? sourceEl.value.trim() : "",
        link: linkEl ? linkEl.value.trim() : "",
        notes: notesEl ? notesEl.value.trim() : "",
        active: activeEl ? activeEl.value : "0",
        reuseable: (reuseableEl && reuseableEl.checked) ? 1 : 0,
        showhistory: (filterEl && filterEl.checked) ? 1 : 0,
        purchased: purchasedEl ? purchasedEl.value : "",
        inv_name: invNameEl ? invNameEl.value : "",
        pre_installed: preInstalledEl ? txt2Int(preInstalledEl.value) : 0,
        free: (freeEl && freeEl.checked) ? 1 : 0
    };
}

//The list of inv_names that are already used
let usedNames = {"a": "a", "b":"b"};

async function popDialog() {
    const invNameEl = document.getElementById("inv_name");
    const searchEl = document.getElementById("search");
    const invListEl = document.getElementById("inv_list");
    
    const inv_name = invNameEl ? invNameEl.value : "";
    if (searchEl) searchEl.value = inv_name;

    //If already have the list, just open popup
    if (invListEl && invListEl.children.length > 0) {
      openModal(document.getElementById("matchDialog"));
      filterList();
      return;
    }
    
    let sendData = { task: "get_software_inventory" };
    try {
        const reply = await postForm("inventory", sendData);
        if (typeof reply === "string") reply = JSON.parse(reply);
        
        if (!reply.success) {
            console.log(reply.msg);
        } else {
            usedNames = reply.used_names.filter(item => item !== inv_name);
            if (invListEl) invListEl.innerHTML = reply.inv_table;
            openModal(document.getElementById("matchDialog"));
            filterList();
        }
    } catch (e) {
        console.error("Inventory fetch failed:", e);
    }
  }
  
//Search the used names list for any matches
function closeDialog() {
    const searchEl = document.getElementById("search");
    const invNameEl = document.getElementById("inv_name");
    const errorMsgEl = document.getElementById("errorMsg");
    const newName = searchEl ? searchEl.value : "";

    if (newName.length > 0 && isMatch(newName)) {
        if (errorMsgEl) errorMsgEl.innerHTML = "Sorry, already in use on another software record.";
        return;
    }
    if (invNameEl) invNameEl.value = newName; 
    closeModal(document.getElementById("matchDialog"));
    btnSave.on();
    btnNew.off();
    btnDelete.off();
}

function isMatch(name) {
    for (let i = 0; i < usedNames.length; i++) {
        if (typeof usedNames[i] === 'string' && (name.startsWith(usedNames[i]) || usedNames[i].startsWith(name))) {
            return true;
        }
    }
    return false;
}

function fillSearch(newName) {
    const searchEl = document.getElementById("search");
    if (searchEl) searchEl.value = newName;
    filterList();
}
  
function filterList() {
    let errorMsg = "";
    const searchInput = document.getElementById("search");
    const errorMsgEl = document.getElementById("errorMsg");
    const matchCountEl = document.getElementById("matchCount");
    const invTable = document.getElementById("inv_table");

    const newName = searchInput ? searchInput.value : "";
    if (newName.length > 0 && isMatch(newName)){
        errorMsg = "Sorry, already in use."
    }
    if (errorMsgEl) errorMsgEl.innerHTML = errorMsg;

    let matchCount = 0;
    const filter = newName;
    if (!invTable) return;

    const tr = invTable.getElementsByTagName("tr");
    let td, a, txtValue;
    for (let i = 0; i < tr.length; i++) {
        td = tr[i].getElementsByTagName("td")[0];
        if (td) {
            a = td.getElementsByTagName("a")[0];
            if (a) {
                txtValue = a.textContent || a.innerText;
            } else {
                txtValue = td.textContent || td.innerText;
            }
            if (txtValue.indexOf(filter) == 0) {
                tr[i].style.display = "table-row";
                matchCount++;
            } else {
                tr[i].style.display = "none";
            }
        }
    }
    if (matchCountEl) matchCountEl.innerHTML = matchCount + " matches";
}
