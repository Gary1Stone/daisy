/* global Metro */
const isEmailValid = /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
let pageLoading = true;

$(document).ready(function () {
    const uid = $("#uid").val();
    //Set initial button state depending if a record is displayed or not
    if (isDigits(uid) && txt2Int(uid) === 0) {
        btnSave.on();  //No record, cannot save, but if they fill it out, we want to save
        btnNew.off(); //No point in showing the same screen again
        btnDelete.off(); //No record, nothing to delete
    } else {
        btnSave.off(); //Record just displayed, nothing to save, user has to make changes first
        btnNew.on(); //Record displayed, can create new
        btnDelete.on(); //record displayed, can delete it
    }
    //if any of the 'input' elements are modified, change the save/add/delete states   
    $("input").on("input", function () {
        btnSave.on(); //As user changes something turn on the Save button
        btnNew.off(); //As user changes something turn off the New button
        btnDelete.off(); //As user changes something turn off the delete button
    });

    //if any of the 'select' droplists are modified, change the save/add/delete states
    $("select").change(function () {
        btnSave.on(); //As user changes something turn on the Save button
        btnNew.off(); //As user changes something turn off the New button
        btnDelete.off(); //As user changes something turn off the delete button
    });
    
    //when user changes the email of the user, check if the email is not already in use
    $("#user").blur(function () {
        let sendData = getFormData();
        sendData.task = "unique";
        if (!isEmailValid.test(sendData.user)) {
            $("#userError").val("ERROR: User ID must be an email address");
            $("#userError").show();
            return;
        }
        $.post("profile", sendData).then(response => {
            reply = JSON.parse(response);
            if (reply.success) {
                $("#userError").hide(); 
            } else {
                $("#userError").val(reply.msg);
                $("#userError").show();
 //               $("#purge").html(sendData.user)
            }
        });
    });
   pageLoading = false; 
});

function getPersonCtrl() {
    return;
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

// Deleting a record is simply setting the active flag = 0 and move its user ID to another column
// its still in the database but never used again.
function deleteRecord() {
    if (btnDelete.state !== "on") return;
    Metro.dialog.create({
        title: "Deleting profile record!",
        content: "<div><p>Deleting a profile is permanent.</p><p>Are you sure you want to delete the " + $("#user").val() + " profile?</p></div>",
        actions: [{
                caption: "Delete",
                cls: "js-dialog-close alert",
                onclick: function () {
                    let sendData = getFormData();
                    sendData.task = "delete";
                    $.post("profile", sendData).then(response => {
                        reply = JSON.parse(response);
                        if (reply.success) {
                            addRecord();  //clears the displayed record
                        } else {
                            console.log(reply.msg);
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
//Display this screen with a uid=0 (user ID = UID)
//when user presses save, in the servlet, detect if record id (UID) is 0, then insert record.
//then send the uid to be used inside this form
function addRecord() {
    if (btnNew.state !== "on") return;
    let url = window.location.href;
    const i = url.indexOf("?");
    if (i < 0) {
        url = url + "?uid=" + encodeURIComponent("0");
    } else {
        url = url.substring(0, i) + "?uid=" + encodeURIComponent("0");
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
    if (!isDigits(sendData.uid)) return false;
    if (!$("#user")[0].checkValidity()) return false;
    if (!$("#first")[0].checkValidity()) return false;
    if (!$("#last")[0].checkValidity()) return false;
    //Check if the user id is unique (onBlur sets if error message visible or not)
    if ($("#userError").is(':visible')) {
        $("#user").focus();
        return false;
    }
    return $("#theForm")[0].checkValidity();
}

function saveRecord() {
    if (btnSave.state !== "on") return false;
    let sendData = getFormData();
    if (!validateForm(sendData)) return false;
    if (sendData.uid === 0) {
        sendData.task = "add";
    }

    //If the user changed their own name, update the Menubar label
    if ($("#curUid").val() === sendData.uid) {
        $("#userName").val(sendData.first + " " + sendData.last);
    }    
    
    $.post("profile", sendData).then(response => {
        reply = JSON.parse(response);
        if (!reply.success) {
            console.log(reply.msg);
        } else {    //Refresh the page
            let url = window.location.href;
            const i = url.indexOf("?");
            if (i < 0) {
                url = url + "?uid=" + encodeURIComponent(reply.uid);
            } else {
                url = url.substring(0, i) + "?uid=" + encodeURIComponent(reply.uid);
            }
            window.location.href =  encodeURI(url);
        }
    });
    return false;
}

function getFormData() {
    let sendData = {task: "", uid: 0, user: "", first: "", last: "", 
        gid: 0, active: 0, geo_fence: "", geo_radius: 10, pwd_reset: 0, 
        color: "", notify: 0 };
    sendData.task = "save";
    sendData.uid = txt2Int($("#uid").val());
    sendData.user = $("#user").val();
    sendData.first = $("#first").val();
    sendData.last = $("#last").val();
    sendData.gid = txt2Int($("#gid").val());
    sendData.geo_fence = $("#geo_fence").val();
    sendData.geo_radius = txt2Int($("#geo_radius").val());
    sendData.pwd_reset = $("#pwd_reset").val();
    sendData.active = ($("#active").prop("checked")) ? 1 : 0;
    sendData.color = $("#color").val();
    sendData.notify = ($("#notify").prop("checked")) ? 1 : 0;
    return sendData;
}

function resetBanned(UID) {
    let sendData = getFormData();
    sendData.task = "unban";
    sendData.uid = UID;
    $.post("profile", sendData).then(response => {
        reply = JSON.parse(response);
        if (!reply.success) {
            toast(reply.msg);
        } else {
            $("#bttn").html("");
        }
    });
}

function ackAlert(aid = 0) {
    let sendData = {
        task: "get_alerts", 
        aid: aid, 
        uid:  $("#uid").val()
    };
    $.post("home", sendData).then(response => {
        $("#alerts").html(response);
    });
}
