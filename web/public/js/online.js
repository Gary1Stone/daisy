//Online.js

// reload this page passing the date as a query parameter
function loadOnlineDevices() {
    const date = document.getElementById("plotDate").value;
    const url = new URL(window.location.href);
    url.searchParams.set("date", date);
    window.location.href = url.toString();
}

function svgClicked(mid) {
    Metro.dialog.open("#popup");
    loadMac(mid);
}

function doSave(){
    if (!validateForm()) return;
    saveMac();
    Metro.dialog.close("#popup");
}

// Fetch & Load the edit Mac dialog window values
function loadMac(mid) {
    const formData = { mid: mid };
    $.post("online/getmac", formData).then(response => {
        let macInfo;
        try {
            macInfo = JSON.parse(response);
        } catch (e) {
            console.log("Non-JSON response:", response);
            return;
        }
        $('#summary').html(macInfo.Summary);
        $("#popupTitle").html(macInfo.Name);
        $('#mid').val(macInfo.Mid);
        $('#mac').val(macInfo.Mac);
        $('#site').val(macInfo.Site);
        $('#name').val(macInfo.Name);
        $('#location').val(macInfo.Location);
        $('#note').val(macInfo.Note);
        // Damn Metro... 
        let plugin = Metro.getPlugin('#kind', 'select');
        if (plugin) {
            plugin.reset();
            if (macInfo.Kind) plugin.val(macInfo.Kind);
        }
        fillOfficeSelect(macInfo.Site, macInfo.Office); 
    });
}

// Damn Metro, you can't just update the value of the select element, 
// you have to reset the plugin first and then set the value through the plugin, 
// otherwise it won't update the UI.

let oldSite = "";
function fillOfficeSelect(siteCode, defaultOffice) {
    if (siteCode !== oldSite) {
        oldSite = siteCode;
        const filtered = officeCache.filter(i => i.p === siteCode);
        const officeCtrl = $('#office');
        officeCtrl.empty();
        officeCtrl.append(new Option("", ""));
        filtered.forEach(i => {
            officeCtrl.append(new Option(i.d, i.c));
        });
    }
    let plugin = Metro.getPlugin('#office', 'select');
    if (plugin) {
        plugin.reset();
        if (defaultOffice) plugin.val(defaultOffice);
    }
}

function validateForm() {
    if (!$("#name")[0].checkValidity()) return false;
    return $("#macForm")[0].checkValidity();
}

function saveMac() {   
    const formData = {
        Mid: txt2Int(document.getElementById("mid").value),
        Name: document.getElementById("name").value,
        Kind: document.getElementById("kind").value,
        Office: document.getElementById("office").value,
        Location: document.getElementById("location").value,
        Note: document.getElementById("note").value,
        Intruder: 0,
    };
    
    $.post("online/setmac", formData).then(response => {
        if (response != "ok") {
            console.log(response || "Error occurred.");
        }
    });
}

function showHelp() {
    Metro.dialog.open("#helpDialog");
}

function doHistory() {
    mid = $("#mid").val();
    window.location.href = encodeURI("history.html?mid=" + mid);
}
