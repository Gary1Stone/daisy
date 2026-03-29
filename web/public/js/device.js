/* global Metro */
// device.js

let gblOldColor = "";   //used to remember the icon's old color if the user toggels the Died status over and over
let oldImageName = "";  //remember what the previous picture was, in case user cancels


$(document).ready(function () {
    const cid = $("#cid").val();
    gblOldColor = $("#color").val();
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
    $("textarea").on("input", function () {
        btnSave.on();
        btnNew.off();
        btnDelete.off();
    });
    // if any of the 'input' elements are modified, change the save/add/delete states  
    $("input").on("input", function () {
        btnSave.on();
        btnNew.off();
        btnDelete.off();
    });
    // if any of the 'select' droplists are modified, change the save/add/delete states
    $("select").change(function () {
        btnSave.on();
        btnNew.off();
        btnDelete.off();
        // handle any interrelated or extra functionality droplists
        let sendData = getFormData();
        switch (this.id) {
            case "type": // device type changed, so change the asset id (name)
                showHideItemsByType();
                //let sendData = getFormData();
                sendData.task = "get_asset_id";
                $.post("device", sendData).then(response => {
                    reply = JSON.parse(response);
                    if (reply.success) {
                        $("#name").val(reply.msg);
                        $("#asset").val(reply.msg);
                    } else {
                        $("#msg").val(reply.msg);
                    }
                });
                break;
            case "status":
                if ($("#status").val() === "DIED") {
                    $("#color").val("fg-olive");
                } else {
                    if (gblOldColor === "fg-olive") {
                        gblOldColor = "fg-emerald";
                    }
                    $("#color").val(gblOldColor);
                } 
                break;
        }
    });
    //Turn off the div that relates to computers only, if it is another type of device 
    showHideItemsByType();
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
      cid: txt2Int($("#cid").val()),
      gid: txt2Int($("#gid").val()),
      uid: txt2Int($("#uid").val()),
      site: $("#site").val(),
      office: $("#office").val(),
      impact: "",
      trouble: "",
      wizard: "",
      type: $("#type").val(),
      inform_gid: 0,
      isReadonly: false,
    }
    return droplistRequest;
}
/**************************************************/


function changePic() {
    //Make sure device has a name - Only a-z, A-Z, 0-9 underscore or dash
    let imageName = $("#name").val().toLowerCase().replace(/[^a-zA-Z0-9_-]/g, '');
    if (imageName.length < 4) {
        imageName = imageName + "DEV01";
    }
    if (imageName.length > 20) {
        imageName = imageName.substr(0,20);
    }
    oldImageName = $("#image").val();
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
    $("#image").val(imageName);
    Metro.dialog.open("#uploadDialog");
}

function cancelUpload() {
    $("#image").val(oldImageName);
}

function showHideItemsByType() {
    $("#computerOnly, #softwareOnly, #ethernetFlag, #wifiFlag, #usbFlag, #cdFlag").hide();
    const deviceType = $("#type").val();
     switch (deviceType) {
        case "DESKTOP": 
        case "LAPTOP":
           $("#computerOnly, #softwareOnly, #ethernetFlag, #wifiFlag, #usbFlag, #cdFlag").show();
            break;
        case "PRINTER":
            $("#ethernetFlag, #wifiFlag, #usbFlag").show();            break;
            break;
        case "NETWORK":
        case "PHONE":
                $("#ethernetFlag, #wifiFlag").show();
                break;    
        }
}

function msg(str) {
    if (str.length === 0) {
        $("#msg").val("").removeClass("remark").removeClass("alert");
        return;
    }
    $("#msg").val(str).addClass("remark").addClass("alert");
    setTimeout(msg, 12345, "", true);
}

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

let btnNew = {
    id: "btnNew",
    state: "on",
    on: function () {
        if ($("#canNew").val() === "1") {
            $("#btnNew").show();
            this.state = "on";
        }
    },
    off: function () {
        $("#btnNew").hide();
        this.state = "off";
    }
};

let btnDelete = {
    id: "btnDelete",
    state: "on",
    on: function () {
        if ($("#canDelete").val() === "1") {
            $("#btnDelete").show();
            this.state = "on";
        } else {
            this.off();
        }
    },
    off: function () {
            $("#btnDelete").hide();
        this.state = "off";
    }
};

//settings={color:light, action:SIGHTING, label:Sighting, icon:mif-eye, active:0 aid:525, cid_ack:0, iid_ack:0, sid_ack:0, uid_ack:0 }
function pop(aid) {
    const $pop = $("#pop");
    const $actionID = $("#actionID");
    const $actionName = $("#actionName");
    const $cmd = $("#cmd");
    const $details = $("#details");
    const $aid = $("#aid" + aid);
    $pop.html("<p>" + $("#notes" + aid).html() + "</p>");
    $actionID.val(aid);
    const settings = JSON.parse($aid.val());
    $actionName.val(settings.action);
    $cmd.hide();
    $details.hide();
    if (settings.active && !settings.cid_ack) { 
        $("#cmd").show();
    }
    if (["BROKEN", "CARE", "DIED", "LOST", "REQUEST"].includes(settings.action)) {
        $details.show();
    }
    Metro.dialog.open("#NotesDialog");
}


function goTicket() {
    const aid = txt2Int($("#actionID").val());
    window.location.href = encodeURI("ticket.html?aid=" + aid);
}

function acceptAction() {
    const aid = txt2Int($("#actionID").val());
    fetchLog(aid);
}

function fetchLog(aid = 0) {
    let sendData = getFormData();
    if (sendData.cid === 0) return; //Stop getting log when there is no record
    sendData.task = "getactionlog";
    sendData.aid = aid;
    if ($("#filter").is(":checked")) {
        sendData.showHistory = 1
    }
    $.post("device", sendData).then(response => {
        $("#actionLogDiv").html(response);
    });
}

// Deleting a record is simply marking the deleted flag (active) = 0;
// its still in the database but not used again.
// issue: what if someone decides to have the same name, then we delete it
// and all of its sub records (action_log...),
function deleteRecord() {
    if (btnDelete.state !== "on") return;
    Metro.dialog.create({
        title: "Delete this device record?",
        content: "<div>Deleting a record cannot be undone.<br> Are you sure you want to delete the " + $("#name").val() + " record?</div>",
        actions: [{
                caption: "Delete",
                cls: "js-dialog-close alert",
                onclick: function () {
                    let sendData = getFormData();
                    sendData.task = "delete";
                    $.post("device", sendData).then(response => {
                        reply = JSON.parse(response);
                        if (reply.success) {
                            addRecord();  //clears the displayed record
                        } else {
                            msg(reply.msg);
                        }
                    });
                }
            },
            {
                caption: "Cancel",
                cls: "js-dialog-close",
                onclick: function () {}
            }]
    });
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
    if (!$("#name")[0].checkValidity()) return false;
    //TYPE, MAKE, STATUS, WIFI, CD, YEAR, 
    if (!$("#serial_number")[0].checkValidity()) return false;
    //GID, UID, OS, SITE, OFFICE, 
    if (!$("#location")[0].checkValidity()) return false;
    if (!$("#ram")[0].checkValidity()) return false;
    if (!$("#cpu")[0].checkValidity()) return false;
    //CORES, DRIVETYPE, 
    if (!$("#gpu")[0].checkValidity()) return false;
    if (!$("#notes")[0].checkValidity()) return false;
    if (!isDigits(sendData.cid)) return false;
    if (!isDigits(sendData.year)) return false;
    if (!isDigits(sendData.cores)) return false;
    if (!isDigits(sendData.cd)) return false;
    if (!isDigits(sendData.drivesize)) return false;
    if (!isDigits(sendData.ram)) return false;
    if (!isDigits(sendData.wifi)) return false;
    if (sendData.name.length < 6) return false;
    if ($("#nameError").is(':visible')) {
        $("#name").focus();
        return false;
    }
    return $("#theForm")[0].checkValidity();
}

function saveRecord() {
    if (btnSave.state !== "on") return false;
    let sendData = getFormData();
    if (!validateForm(sendData)) return false;
    if (sendData.cid === 0) {
        sendData.task = "add";
    } else {
        savePreInstalled();
    }
    $.post("device", sendData).then(response => {
        reply = JSON.parse(response);
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
    });
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
    sendData.cid = txt2Int($("#cid").val());
    sendData.name = $("#name").val();
    sendData.name = sendData.name.toUpperCase();
    sendData.type = $("#type").val();
    sendData.site = $("#site").val();
    sendData.office = $("#office").val();
    sendData.location = $("#location").val();
    sendData.year = txt2Int($("#year").val());
    sendData.make = $("#make").val();
    sendData.model = $("#model").val();
    sendData.cpu = $("#cpu").val();
    sendData.cores = txt2Int($("#cores").val());
    sendData.ram = txt2Int($("#ram").val());
    sendData.drivetype = $("#drivetype").val();
    sendData.drivesize = txt2Int($("#drivesize").val());
    sendData.notes = $("#notes").val();
    sendData.gpu = $("#gpu").val();
    sendData.cd = ($("#cd").prop("checked")) ? 1 : 0;
    sendData.wifi = ($("#wifi").prop("checked")) ? 1 : 0;
    sendData.ethernet = ($("#ethernet").prop("checked")) ? 1 : 0;
    sendData.usb = ($("#usb").prop("checked")) ? 1 : 0;
    sendData.active = txt2Int($("#active").val());
    sendData.image = $("#image").val();
    sendData.color = $("#color").val();
    sendData.speed = txt2Int($("#speed").val());
    sendData.uid = txt2Int($("#uid").val());
    sendData.status = $("#status").val();
    sendData.os = $("#os").val();
    sendData.serial_number = $("#serial_number").val();
    sendData.gid = txt2Int($("#gid").val());
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
    formData.append("uploadfile", ajaxfile.files[0], $("#image").val());
  
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
    $.post("preinstalled", payload).then(response => {
        if (response !== "okay") {
            msg(response);
        }
    });
}
