// correlation.js

// if user changes the sliders, show the values in the readout
$(document).ready(function () {
    $("#jaccard").change(function () {
        $("#jaccardValue").html($("#jaccard").val());
    });
    $("#pearson").change(function () {
        $("#pearsonValue").html($("#pearson").val());
    });
});

function showHelp() {
    Metro.dialog.open("#helpDialog");
}

function setHighHigh() {
    jsign("&gt;");
    $("#jaccard").val("0.70").trigger("change");
    psign("&gt;");
    $("#pearson").val("0.85").trigger("change");
    $('#fixed').prop('checked', true);
    $('input[name="hostnames"]').prop('checked', false);
}

function setLowHigh() {
    jsign("&lt;");
    $("#jaccard").val("0.10").trigger("change");
    psign("&gt;");
    $("#pearson").val("0.95").trigger("change");
    $('#random').prop('checked', true);
    $('input[name="hostnames"]').prop('checked', false);
}

// Get the value of the radio selections
function reloadTable() {
    const sendData = { 
        Jaccard: txt2Int($("#jaccard").val()*100),
        Pearson: txt2Int($("#pearson").val()*100),
        Fixed: $('input[name="macfilter"]:checked').val() === "1",
        Random: $('input[name="macfilter"]:checked').val() === "2",
        Hostnames: $('input[name="hostnames"]:checked').val() === "1",
        Jsign: $("#jsign").html() === "&gt;",
        Psign: $("#psign").html() === "&gt;"
    }
    $.post("correlation", sendData).then(response => {
        if (response === "CRITICAL SERVER ERROR!") {
            console.log(response);
            toast(response, "alert");
            response = "";
        }
        $("#tableData").html(response);
    }); 
}

// Search Filters show
function popFilters() {
  Metro.dialog.open("#searchDialog");
}

function jsign(newsign) {
    //test if newsign exists
    if (newsign) {
        $("#jsign").html(newsign);
        return;
    }
    const currentSign = $("#jsign").html();
    if (currentSign === "&gt;") {
        $("#jsign").html("&lt;");
    } else {
        $("#jsign").html("&gt;");
    }
    return false;
}

function psign(newsign) {
    if (newsign) {
        $("#psign").html(newsign);
        return;
    }
    const currentSign = $("#psign").html();
    if (currentSign === "&gt;") {
        $("#psign").html("&lt;");
    } else {
        $("#psign").html("&gt;");
    }
    return false;
}

function showLink(mac1, name1, host1, mac2, name2, host2) {
    document.getElementById("mac1").value = mac1;
    document.getElementById("mac2").value = mac2;
    $("#linkText").html(`${name1} (${host1})<br>and<br>${name2}(${host2})`);
    Metro.dialog.open("#linkDialog");
}

// Read the radio buttons form to see what is selected for linking devices
function recordLink() {
    let isSame = false;
    let isIgnore = false;
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
    const sendData = { 
        mac1:document.getElementById("mac1").value, 
        mac2:document.getElementById("mac2").value,
        isSame: isSame,
        isIgnore: isIgnore
    }
    $.post("duplicatesjoin", sendData).then(response => {
        if (response !== "ok") {
            console.log(response);
        } else {
            reloadTable();
        }
    }); 
}