// duplicates.js

$(document).ready(function () {
    $("#avoidSelect").change(function () {
        const optionSelected = $("#avoidSelect").val();
        const macsArray = optionSelected.split("_");
        const sendData = { 
            mac1: macsArray[0], 
            mac2: macsArray[1]
        }
        $.post("duplicates", sendData).then(response => {
            $("#chart").html(response);
        }); 
    });
});

function showHelp() {
    Metro.dialog.open("#helpDialog");
}

function showLink() {
    const avoid = document.getElementById("avoidSelect");
    const text = avoid.options[avoid.selectedIndex].text;
    $("#linkText").text(text);
    Metro.dialog.open("#linkDialog");
}

// Read the radio buttons form to see what is selected for linking devices
function recordLink() {
    let isSame = false;
    // Select the checked radio button in the 'device' group
    const selectedRadio = document.querySelector('input[name="device"]:checked');
    if (selectedRadio) {        // it might be null if none are checked
        const value = selectedRadio.value;
        if (value === "same") {
            isSame = true;
            isIgnore = false;
        } else if (value === "different") {
            isSame = false;
            isIgnore = false;
        } else { //  if Ignore
            isSame = false;
            isIgnore = true;
        }
    } else {
        return;      //  No radio button selected, do nothing
    }
    const optionSelected = $("#avoidSelect").val();
    const macsArray = optionSelected.split("_");
    const sendData = { 
        mac1: macsArray[0], 
        mac2: macsArray[1],
        isSame: isSame,
        isIgnore: isIgnore
    }
    $.post("duplicatesjoin", sendData).then(response => {
        if (response != "ok") {
            console.log(response);
        } else {
            window.location.reload();
        }
    }); 
}
