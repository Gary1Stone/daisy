//Online.js

// reload this page passing the date as a query parameter
function loadOnlineDevices() {
    const date = document.getElementById("plotDate").value;
    const url = new URL(window.location.href);
    url.searchParams.set("date", date);
    window.location.href = url.toString();
}

function svgClicked(mid) {
    if (mid) {
        const popup = document.getElementById("popup");
        if (popup) {
            openModal(popup);
            loadMac(mid);
        }
    }
}

function doSave(){
    if (!validateForm()) return;
    saveMac();
    closeModal(document.getElementById("popup"));
}

// Fetch & Load the edit Mac dialog window values
async function loadMac(mid) {
    const formData = { Mid: txt2Int(mid) };
    try {
        const macInfo = await postForm("online/getmac", formData);

        const summary = document.getElementById('summary');
        const popupTitle = document.getElementById("popupTitle");
        const midInput = document.getElementById('mid');
        const macInput = document.getElementById('mac');
        const siteInput = document.getElementById('site');
        const nameInput = document.getElementById('name');
        const locationInput = document.getElementById('location');
        const noteInput = document.getElementById('note');

        if (summary) summary.innerHTML = macInfo.Summary;
        if (popupTitle) popupTitle.innerHTML = macInfo.Name;
        if (midInput) midInput.value = macInfo.Mid;
        if (macInput) macInput.value = macInfo.Mac;
        if (siteInput) siteInput.value = macInfo.Site;
        if (nameInput) nameInput.value = macInfo.Name;
        if (locationInput) locationInput.value = macInfo.Location;
        if (noteInput) noteInput.value = macInfo.Note;

        setDropdownValue("kind", macInfo.Kind);
        fillOfficeSelect(macInfo.Site, macInfo.Office); 
    } catch (error) {
        console.error("loadMac failed:", error);
    }
}

let oldSite = "";
function fillOfficeSelect(siteCode, defaultOffice) {
    const officeCtrl = document.getElementById('office');
    if (siteCode !== oldSite && typeof officeCache !== "undefined") {
        oldSite = siteCode;
        const filtered = officeCache.filter(i => i.p === siteCode);
        if (officeCtrl) {
            officeCtrl.innerHTML = "";
            officeCtrl.appendChild(new Option("", ""));
            filtered.forEach(i => {
                officeCtrl.appendChild(new Option(i.d, i.c));
            });
        }
    }
    if (officeCtrl) officeCtrl.value = defaultOffice;
}

function validateForm() {
    const nameInput = document.getElementById("name");
    const macForm = document.getElementById("macForm");
    if (nameInput && !nameInput.checkValidity()) return false;
    if (macForm && !macForm.checkValidity()) return false;
    return true;
}

async function saveMac() {   
    const midEl = document.getElementById("mid");
    const nameEl = document.getElementById("name");
    const kindEl = document.getElementById("kind");
    const officeEl = document.getElementById("office");
    const locationEl = document.getElementById("location");
    const noteEl = document.getElementById("note");

    const formData = {
        Mid: midEl ? txt2Int(midEl.value) : 0,
        Name: nameEl ? nameEl.value : "",
        Kind: kindEl ? kindEl.value : "",
        Office: officeEl ? officeEl.value : "",
        Location: locationEl ? locationEl.value : "",
        Note: noteEl ? noteEl.value : "",
        Intruder: 0,
    };
    
    try {
        const result = await postForm("online/setmac", formData);
        if (result !== "ok") {
            console.log(result || "Error occurred.");
        }
    } catch (error) {
        console.error("saveMac failed:", error);
    }
}

function showHelp() {
    openModal(document.getElementById("helpDialog"));
}

function doHistory() {
    const midInput = document.getElementById("mid");
    const mid = midInput ? midInput.value : "";
    window.location.href = encodeURI("history.html?mid=" + mid);
}
