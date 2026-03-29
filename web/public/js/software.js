/* global Metro */

$(document).ready(function () {
    const sid = $("#sid").val();
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
    $("input").on("input", function(){
        btnSave.on();
        btnNew.off();
        btnDelete.off();
    });
    
    //if any of the 'select' droplists are modified, change the save/add/delete states
    $("select").change(function () {
        btnSave.on();
        btnNew.off();
        btnDelete.off();
    });

    //when user changes the name of the software, check if the name is not already in use
    $("#name").blur(function () {
        let sendData = getFormData();
        sendData.task = "unique";
        $.post("software", sendData).then(response => {
            reply = JSON.parse(response);
            if (reply.success === true) {
                $("#nameError").hide(); 
            } else {
                $("#nameError").val(reply.msg);
                $("#nameError").show();
            } 
        });
    });
});

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


//pop up a dialog for displaying the message details
//settings={"color":"light ","action":"INSTALL","label":"Install Soft...","icon":"mif-apps","active":0,"aid":176,"cid_ack":1,"iid_ack":0,"sid_ack":1,"uid_ack":0}
function pop(aid) {
    $("#pop").html("<p>" + $("#notes" + aid).html() + "</p>");
    $("#actionID").val(aid);
    const settings = JSON.parse($("#aid" + aid).val());
    $("#actionName").val(settings.action);
    $("#cmd").hide();
    if (settings.active && !settings.sid_ack) {
        $("#cmd").show();
    }
    Metro.dialog.open("#NotesDialog");
}

function acceptAction() {
    const aid = txt2Int($("#actionID").val());
    fetchLog(aid);
}


function fetchLog(aid = 0) {
    sendData = getFormData();
    sendData.task = "getactionlog";
    sendData.aid = aid;
    $.post("software", sendData).then(response => {
        $("#actionLogDiv").html(response);
    });
}

// Deleting a record is simply deeing it if not used anywhere
// else setting the active flag = 0 and moving name to old name
// its still in the database but not used again.
function deleteRecord() {
    if (btnDelete.state !== "on") return;
    Metro.dialog.create({
        title: "Delete this software record?",
        content: "<div><p>Deleting a record is permanent.</p><p>Are you sure you want to delete the " + $("#software").val() + " record?</p></div>",
        actions: [{
                caption: "Delete",
                cls: "js-dialog-close alert",
                onclick: function () {
                    let sendData = getFormData();
                    sendData.task = "delete";
                    $.post("software", sendData).then(response => {
                        reply = JSON.parse(response);
                        if (reply.success) {
                            addRecord();  //clears the displayed record
                        } else {
                            toast(reply.msg, "alert");
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


function isDigits(value) {
    if (typeof value === "string" && value.length > 0) {
        digitsOnly = /^\d+$/;  // d=[0-9] 
        return digitsOnly.test(value);
    }
    return true;
}

function validateForm(sendData) {
    if (!isDigits(sendData.sid)) return false;    
    if (!$("#name")[0].checkValidity()) return false;
    //Check if the name is unique (onBlur sets if error message visible or not)
    if ($("#nameError").is(':visible')) {
        $("#name").focus();
        return false;
    }
    if (sendData.link.length > 0) {
        const link = document.getElementById("link");
        if (!link.checkValidity()) {
            console.log(link.validationMessage);
            return false;
        }
    }    
    return $("#theForm")[0].checkValidity();
}


function saveRecord() {
    if (btnSave.state !== "on") return false;
    let sendData = getFormData();
    if (!validateForm(sendData)) return false;
    if (sendData.sid === 0) {
        sendData.task = "add";
    }

    $.post("software", sendData).then(response => {
        reply = JSON.parse(response);
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
    });
    return false;
}

function getFormData() {
    let sendData = {task: "", sid: 0, name: "", licenses: 0, 
        license_key: "", product: "", source: "", link: "",
        notes: "", active: 0, reuseable: 0, aid: 0, 
        showhistory: 0, purchased: "", inv_name: "", 
        pre_installed: 0, free: 0 };
    sendData.task = "save";
    sendData.sid = txt2Int($("#sid").val());
    sendData.name = $("#name").val().trim();
    sendData.licenses = txt2Int($("#licenses").val());
    sendData.license_key = $("#license_key").val().trim();
    sendData.product = $("#product").val().trim();
    sendData.source = $("#source").val().trim();
    sendData.link = $("#link").val().trim();
    sendData.notes = $("#notes").val().trim();
    sendData.active = $("#active").val();
    sendData.reuseable = ($("#reuseable").prop("checked")) ? 1 : 0;
    sendData.showhistory = ($("#filter").prop("checked")) ? 1 : 0;
    sendData.purchased = $("#purchased").val();
    sendData.inv_name = $("#inv_name").val();
    sendData.pre_installed = txt2Int($("#pre_installed").val());
    sendData.free = ($("#free").prop("checked")) ? 1 : 0;
    return sendData;
}

//The list of inv_names that are already used
let usedNames = {"a": "a", "b":"b"};

function popDialog() {
    //Get the inventory name and populate the search field
    const inv_name =$("#inv_name").val();
    $("#search").val(inv_name);
    //If already have the list, just open popup
    if ($("#inv_list").children().length > 0) {
      Metro.dialog.open("#matchDialog");
      filterList();
      return;
    }
    let sendData = { task: "get_software_inventory" };
    $.post("inventory", sendData).then((response) => {
        Metro.dialog.open("#matchDialog");
        reply = JSON.parse(response);
        if (!reply.success) {
            console.log(reply.msg);
        } else {
            usedNames = reply.used_names.filter(item => item !== inv_name);
            $("#inv_list").html(reply.inv_table);
            filterList();
        }
    });
  }
  
//Search the used names list for any matches
function closeDialog() {
    newName = $("#search").val();
    if (newName.length > 0 && isMatch(newName)) {
        $("#errorMsg").html("Sorry, already in use on another software record.");
        return;
    }
    $("#inv_name").val(newName); 
    Metro.dialog.close("#matchDialog");
    btnSave.on();
    btnNew.off();
    btnDelete.off();
}

function isMatch(name) {
    for (let i = 0; i < usedNames.length; i++) {
        if (name.startsWith(usedNames[i]) || usedNames[i].startsWith(name)) {
            return true;
        }
    }
    return false;
}

function fillSearch(newName) {
    $("#search").val(newName);
    filterList();
}
  
function filterList() {
    errorMsg = "";
    newName = $("#search").val();
    if (newName.length > 0 && isMatch(newName)){
        errorMsg = "Sorry, already in use."
    }
    $("#errorMsg").html(errorMsg);
    let matchCount = 0;
    let input = document.getElementById("search");
    let filter = input.value;
    let table = document.getElementById("inv_table");
    let tr = table.getElementsByTagName("tr");
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
    $("#matchCount").html(matchCount + " matches")
}
