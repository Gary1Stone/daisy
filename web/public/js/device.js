/* global Metro */
// device.js

let gblOldColor = "";   //used to remember the icon's old color if the user toggels the Died status over and over
let oldImageName = "";  //remember what the previous picture was, in case user cancels


// Declare iconbar button variables at top level so they are available to checkValid, saveRecord, etc.
let btnSave, btnNew, btnDelete;

// Page loaded event
document.addEventListener('DOMContentLoaded', function() {
    // Initialize the iconbar button instances once the scripts and DOM are ready
    btnSave = new Button("btnSave");
    btnNew = new Button("btnNew");
    btnDelete = new Button("btnDelete", true); // true for forceOffIfNotAllowed
    const cidEl = document.getElementById("cid");
    const cid = cidEl ? cidEl.value : "0";
    const colorEl = document.getElementById("color");
    gblOldColor = colorEl ? colorEl.value : "";
    if (isDigits(cid) && txt2Int(cid) === 0) {
        btnSave.on();
        btnNew.off(); 
        btnDelete.off();
    } else {
        btnSave.off();
        btnNew.on();
        btnDelete.on(); 
    } 
    // if any of the 'textArea' elements are modified, change the save/add/delete states  
    document.querySelectorAll("textarea").forEach(el => {
        el.addEventListener("input", function () {
            btnSave.on();
            btnNew.off();
            btnDelete.off();
        });
    });
    // if any of the 'input' elements are modified, change the save/add/delete states  
    document.querySelectorAll("input").forEach(el => {
        el.addEventListener("input", function () {
            btnSave.on();
            btnNew.off();
            btnDelete.off();
        });
    });
    // if any of the 'select' droplists are modified, change the save/add/delete states
    document.querySelectorAll("select").forEach(el => {
        el.addEventListener("change", async function () {
        btnSave.on();
        btnNew.off();
        btnDelete.off();
        // handle any interrelated or extra functionality droplists
        let sendData = getFormData();
        switch (this.id) {
            case "type": // device type changed, so change the asset id (name)
                showHideItemsByType();
                sendData.task = "get_asset_id";
                try {
                    const reply = await postForm("device", sendData);
                    if (reply.success) {
                        document.getElementById("name").value = reply.msg;
                        const assetEl = document.getElementById("asset");
                        if (assetEl) assetEl.value = reply.msg;
                    } else {
                        msg(reply.msg);
                    }
                } catch (e) { console.error(e); }
                break;
            case "status":
                const statusEl = document.getElementById("status");
                if (statusEl && statusEl.value === "DIED") {
                    if (colorEl) colorEl.value = "fg-olive";
                } else {
                    if (gblOldColor === "fg-olive") {
                        gblOldColor = "fg-emerald";
                    }
                    if (colorEl) colorEl.value = gblOldColor;
                } 
                break;
        }
    })});
    //Turn off the div that relates to computers only, if it is another type of device 
    showHideItemsByType();
    // Get device name if #name blank
    const nameEl = document.getElementById("name");
    if (nameEl && nameEl.value.length < 1) {
        // trigger onchange on type select list to get a name
        document.getElementById("type").dispatchEvent(new Event('change', { bubbles: true }));
    }  
});


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
      cid: txt2Int(document.getElementById("cid").value),
      gid: txt2Int(document.getElementById("gid").value),
      uid: txt2Int(document.getElementById("uid").value),
      site: document.getElementById("site").value,
      office: document.getElementById("office").value,
      impact: "",
      trouble: "",
      wizard: "",
      type: document.getElementById("type").value,
      inform_gid: 0,
      isReadonly: false,
    }
    return droplistRequest;
}
/**************************************************/


function changePic() {
    //Make sure device has a name - Only a-z, A-Z, 0-9 underscore or dash
    let imageName = document.getElementById("name").value.toLowerCase().replace(/[^a-zA-Z0-9_-]/g, '');
    if (imageName.length < 4) {
        imageName = imageName + "DEV01";
    }
    if (imageName.length > 20) {
        imageName = imageName.substr(0,20);
    }
    oldImageName = document.getElementById("image").value;
    if (typeof oldImageName !== "string" ) {
        oldImageName = "";
    }
    //if there was no previous image
    if (oldImageName.length < 1) {
        imageName = imageName + "-v1.jpg";
    } else {
        //find the current counter in the name string aaaaaaaa-vxxxxx.jpg
        let toEnd = imageName.length;
        if (oldImageName.endsWith(".jpg")) {
            toEnd = oldImageName.length - 4; //remove .jpg
        }
        let toStart = oldImageName.lastIndexOf("-v");
        if (toStart === -1) {
            toStart = toEnd;
        } else {        //May not have -v in the name due to legacy
            toStart = toStart + 2; //offset for the two characters (-v)
            if (toStart > toEnd) {
                toStart = toEnd;
            }
        }
        const cnt = txt2Int(oldImageName.slice(toStart, toEnd)) + 1;
        imageName = imageName + "-v" + cnt.toString() + ".jpg";
    }    
    document.getElementById("image").value = imageName;
    openModal(document.getElementById("uploadDialog"));
}

function cancelUpload() {
    document.getElementById("image").value = oldImageName;
}

function showHideItemsByType() {
    const items = ["computerOnly", "softwareOnly", "ethernetFlag", "wifiFlag", "usbFlag", "cdFlag"];
    items.forEach(id => setDisplay(document.getElementById(id), false));
    const deviceType = document.getElementById("type").value;
     switch (deviceType) {
        case "DESKTOP": 
        case "LAPTOP":
           items.forEach(id => setDisplay(document.getElementById(id), true));
            break;
        case "PRINTER":
            ["ethernetFlag", "wifiFlag", "usbFlag"].forEach(id => setDisplay(document.getElementById(id), true));
            break;
        case "NETWORK":
        case "PHONE":
                ["ethernetFlag", "wifiFlag"].forEach(id => setDisplay(document.getElementById(id), true));
                break;    
        }
}

function msg(str) {
    const msgEl = document.getElementById("msg");
    if (!msgEl) return;
    if (str.length === 0) {
        msgEl.value = "";
        msgEl.classList.remove("remark", "alert");
        return;
    }
    msgEl.value = str;
    msgEl.classList.add("remark", "alert");
    setTimeout(msg, 12345, "", true);
}


//settings={color:light, action:SIGHTING, label:Sighting, icon:mif-eye, active:0 aid:525, cid_ack:0, iid_ack:0, sid_ack:0, uid_ack:0 }
function pop(aid) {
    const popEl = document.getElementById("pop");
    const notesEl = document.getElementById("notes" + aid);
    const actionIDEl = document.getElementById("actionID");
    const actionNameEl = document.getElementById("actionName");
    const cmdEl = document.getElementById("cmd");
    const detailsEl = document.getElementById("details");
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
    
    openModal(document.getElementById("NotesDialog"));
}


function goTicket() {
    const aid = txt2Int(document.getElementById("actionID").value);
    window.location.href = encodeURI("ticket.html?aid=" + aid);
}

function acceptAction() {
    const aid = txt2Int(document.getElementById("actionID").value);
    fetchLog(aid);
}

async function fetchLog(aid = 0) {
    let sendData = getFormData();
    if (sendData.cid === 0) return; //Stop getting log when there is no record
    sendData.task = "getactionlog";
    sendData.aid = aid;
    const filter = document.getElementById("filter");
    if (filter && filter.checked) {
        sendData.showHistory = 1
    }
    try {
        const html = await postForm("device", sendData);
        if (html) document.getElementById("actionLogDiv").innerHTML = html;
    } catch (e) { console.error(e); }
}

// Deleting a record is simply marking the deleted flag (active) = 0;
// its still in the database but not used again.
// issue: what if someone decides to have the same name, then we delete it
// and all of its sub records (action_log...),
function deleteRecord() {
    if (btnDelete.state !== "on") return;
    openModal(document.getElementById("deleteDialog"));
    const name = document.getElementById("name").value;
    if (name.length > 0) {
        document.getElementById("deviceName").innerHTML = name;
    }
}

function confirmDelete() {
    if (btnDelete.state !== "on") return;
    let sendData = getFormData();
    sendData.task = "delete";
    try {
        const reply = await postForm("device", sendData);
        if (typeof reply === "string") reply = JSON.parse(reply);
        
        if (reply.success) {
            addRecord();  //clears the displayed record
        } else {
            closeModal(document.getElementById("deleteDialog"));
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

function isDigits(value) {
    if (typeof value === "string" && value.length > 0) {
        digitsOnly = /^\d+$/;  // d=[0-9] 
        return digitsOnly.test(value);
    }
    return true;
}

function validateForm(sendData) {
     if (!document.getElementById("name").checkValidity()) return false;
    //TYPE, MAKE, STATUS, WIFI, CD, YEAR, 
    if (!document.getElementById("serial_number").checkValidity()) return false;
    //GID, UID, OS, SITE, OFFICE, 
    if (!document.getElementById("location").checkValidity()) return false;
    if (!document.getElementById("ram").checkValidity()) return false;
    if (!document.getElementById("cpu").checkValidity()) return false;
    //CORES, DRIVETYPE, 
    if (!document.getElementById("gpu").checkValidity()) return false;
    if (!document.getElementById("notes").checkValidity()) return false;

    if (!isDigits(sendData.cid)) return false;
    if (!isDigits(sendData.year)) return false;
    if (!isDigits(sendData.ram)) return false;
    if (!isDigits(sendData.wifi)) return false;
    if (sendData.name.length < 6) return false;


    // if ($("#nameError").is(':visible')) {
    //     $("#name").focus();
    //     return false;
    // }
    return document.getElementById("theForm").checkValidity();
}

async function saveRecord() {
    if (btnSave.state !== "on") return false;
    let sendData = getFormData();
    if (!validateForm(sendData)) return false;
    if (sendData.cid === 0) {
        sendData.task = "add";
    } else {
        savePreInstalled();
    }
    try {
        const reply = await postForm("device", sendData);
        if (typeof reply === "string") reply = JSON.parse(reply);
        
        if (!reply.success) {
            msg(reply.msg);  //display error message
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
    } catch (e) { console.error(e); }
    return false;
}

function getFormData() {
    let sendData = {task: "", cid: 0, name: "", type: "", site: "",
        office: "", location: "", year: 0, make: "", model: "",
        cpu: "", cores: 0, ram: 0, drivetype: "", drivesize: 0,
        cd: 0, wifi: 0, ethernet: 0, usb: 0, gpu: "", notes: "", active: 0, last_updated_by: 0,
        image: "", color: "", speed: 0, uid: 0, status: "", os: "", 
        serial_number: "", gid: 0, aid: 0, showHistory: 0};
    sendData.task = "save";
    sendData.type = document.getElementById("type").value;
    sendData.site = document.getElementById("site").value;
    sendData.office = document.getElementById("office").value;
    sendData.location = document.getElementById("location").value;
    sendData.year = txt2Int(document.getElementById("year").value);
    sendData.make = document.getElementById("make").value;
    sendData.model = document.getElementById("model").value;
    sendData.cpu = document.getElementById("cpu").value;
    sendData.cores = txt2Int(document.getElementById("cores").value);
    sendData.ram = txt2Int(document.getElementById("ram").value);
    sendData.drivetype = document.getElementById("drivetype").value;
    sendData.drivesize = txt2Int(document.getElementById("drivesize").value);
    sendData.notes = document.getElementById("notes").value;
    sendData.gpu = document.getElementById("gpu").value;
    sendData.cd = (document.getElementById("cd").checked) ? 1 : 0;
    sendData.wifi = (document.getElementById("wifi").checked) ? 1 : 0;
    sendData.ethernet = (document.getElementById("ethernet").checked) ? 1 : 0;
    sendData.usb = (document.getElementById("usb").checked) ? 1 : 0;
    sendData.active = txt2Int(document.getElementById("active").value);
    sendData.image = document.getElementById("image").value;
    sendData.color = document.getElementById("color").value;
    sendData.speed = txt2Int(document.getElementById("speed").value);
    sendData.uid = txt2Int(document.getElementById("uid").value);
    sendData.status = document.getElementById("status").value;
    sendData.os = document.getElementById("os").value;
    sendData.serial_number = document.getElementById("serial_number").value;
    sendData.gid = txt2Int(document.getElementById("gid").value);
    return sendData;
}

async function uploadFile() {
    let formData = new FormData();
    
    // Check if a file is selected
    if (ajaxfile.files.length === 0) {
      return;
    }
    if (typeof ajaxfile.files[0] === "undefined") {
      return;
    }
  
    // Append the file to the FormData with its new name
    formData.append("uploadfile", ajaxfile.files[0], document.getElementById("image").value);
  
    try {
      // Perform the file upload using fetch
      const response = await fetch("upload", {
        method: "POST",
        body: formData,
      });
  
      // Check if the upload was successful (status code 2xx)
      if (response.ok) {
        // Show the new image
        btnSave.on();
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
    rowid = txt2Int(id);
    if (rowid > 0) {
        preInstalled.set(rowid, checked);
    }
}

function savePreInstalled() {
    if (preInstalled.size === 0) {
        return;
    }
    let items = [];
    for (const [key, value] of preInstalled) {
        let item = {id: key, chk: 0};
        if (value) {
            item.chk = 1;
        } else {
            item.chk = 0;
        }
        items.push(item);
    }
    // transmit array of objects to server
    payload = JSON.stringify(items);
    postJSON("preinstalled", items, (reply) => {
        if (reply !== "okay") {
            msg(reply);
        }
    });
}
