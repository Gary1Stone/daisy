// Control.js


//const popup = document.getElementById('popup');
//As user changes something turn on the Save button
$(document).ready(function () {
    // If User clicks on off-popup background, close the dialog
    popup.addEventListener("click", e => {
        const popupDimensions = popup.getBoundingClientRect()
        if (
            e.clientX < popupDimensions.left ||
            e.clientX > popupDimensions.right ||
            e.clientY < popupDimensions.top ||
            e.clientY > popupDimensions.bottom
        ) {
            $("#popupContent").html("");
            $("#popupTitle").html("");
            Metro.dialog.close("#popup");
        }
    });
});

function showActiveUsers() {
    getActiveUsers();
    Metro.dialog.open("#popup");
}

function doPopupOkay(){
    $("#popupContent").html("");
    $("#popupTitle").html("");
//    Metro.dialog.close("#popup");
}

function endSession(id) {
    alert("end session");
    $.post("control", getFormData("end_session", id)).then(response => {
        $("#popupContent").html(response);
        $("#popupTitle").html("Active Users");
    });
}

function getActiveUsers() {
   const sendData = getFormData("get_active_users");
    $.post("control", sendData).then(response => {
        $("#popupContent").html(response);
        $("#popupTitle").html("Active Users");
    });
}

function getAttacks(duration) {
    const sendData = getFormData("get_attacks", duration);
    let title = "Attacks ";
    if (duration === 1) {
        title += "(Day)";
    } else if (duration === 7) {
        title += "(Week)";
    } else if (duration === 30) {
        title += "(Month)";
    }
    $.post("control", sendData).then(response => {
        $("#popupContent").html(response);
        $("#popupTitle").html(title);
        Metro.dialog.open("#popup");
    });
}


function getFormData(task, id=0) {
    return {task: task, id: id};
}