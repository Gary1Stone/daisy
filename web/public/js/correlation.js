// correlation.js

// if user changes the sliders, show the values in the readout
document.addEventListener('DOMContentLoaded', function() {
    const jaccard = document.getElementById("jaccard");
    if (jaccard) {
        jaccard.addEventListener("change", function () {
            const val = document.getElementById("jaccardValue");
            if (val) val.innerHTML = jaccard.value;
        });
    }
    const pearson = document.getElementById("pearson");
    if (pearson) {
        pearson.addEventListener("change", function () {
            const val = document.getElementById("pearsonValue");
            if (val) val.innerHTML = pearson.value;
        });
    }
});

function showHelp() {
    openModal(document.getElementById("helpDialog"));
}

function setHighHigh() {
    jsign("&gt;");
    const jaccard = document.getElementById("jaccard");
    if (jaccard) {
        jaccard.value = "0.70";
        jaccard.dispatchEvent(new Event('change'));
    }
    psign("&gt;");
    const pearson = document.getElementById("pearson");
    if (pearson) {
        pearson.value = "0.85";
        pearson.dispatchEvent(new Event('change'));
    }
    const fixed = document.getElementById('fixed');
    if (fixed) fixed.checked = true;
    document.querySelectorAll('input[name="hostnames"]').forEach(el => el.checked = false);
}

function setLowHigh() {
    jsign("&lt;");
    const jaccard = document.getElementById("jaccard");
    if (jaccard) {
        jaccard.value = "0.10";
        jaccard.dispatchEvent(new Event('change'));
    }
    psign("&gt;");
    const pearson = document.getElementById("pearson");
    if (pearson) {
        pearson.value = "0.95";
        pearson.dispatchEvent(new Event('change'));
    }
    const random = document.getElementById('random');
    if (random) random.checked = true;
    document.querySelectorAll('input[name="hostnames"]').forEach(el => el.checked = false);
}

// Get the value of the radio selections
async function reloadTable() {
    const jaccard = document.getElementById("jaccard");
    const pearson = document.getElementById("pearson");
    const macFilterChecked = document.querySelector('input[name="macfilter"]:checked');
    const hostnamesChecked = document.querySelector('input[name="hostnames"]:checked');
    const jsignEl = document.getElementById("jsign");
    const psignEl = document.getElementById("psign");

    const sendData = { 
        Jaccard: txt2Int((jaccard ? jaccard.value : 0) * 100),
        Pearson: txt2Int((pearson ? pearson.value : 0) * 100),
        Fixed: macFilterChecked ? macFilterChecked.value === "1" : false,
        Random: macFilterChecked ? macFilterChecked.value === "2" : false,
        Hostnames: hostnamesChecked ? hostnamesChecked.value === "1" : false,
        Jsign: jsignEl ? (jsignEl.innerHTML === ">" || jsignEl.innerHTML === "&gt;") : false,
        Psign: psignEl ? (psignEl.innerHTML === ">" || psignEl.innerHTML === "&gt;") : false
    }
    
    try {
        const result = await postForm("correlation", sendData);
        if (result === "CRITICAL SERVER ERROR!") {
            console.error(result);
            toast(result, "alert");
        } else {
            const tableData = document.getElementById("tableData");
            if (tableData) tableData.innerHTML = result;
        }
    } catch (error) {
        console.error("Reload table failed:", error);
    }
}

// Search Filters show
function popFilters() {
    openModal(document.getElementById("searchDialog"));
}

function jsign(newsign) {
    const el = document.getElementById("jsign");
    if (!el) return false;
    //test if newsign exists
    if (newsign) {
        el.innerHTML = newsign;
        return;
    }
    const currentSign = el.innerText;
    if (currentSign === ">" || currentSign === "&gt;") {
        el.innerHTML = "&lt;";
    } else {
        el.innerHTML = "&gt;";
    }
    return false;
}

function psign(newsign) {
    const el = document.getElementById("psign");
    if (!el) return false;
    if (newsign) {
        el.innerHTML = newsign;
        return;
    }
    const currentSign = el.innerText;
    if (currentSign === ">" || currentSign === "&gt;") {
        el.innerHTML = "&lt;";
    } else {
        el.innerHTML = "&gt;";
    }
    return false;
}

function showLink(mac1, name1, host1, mac2, name2, host2) {
    document.getElementById("mac1").value = mac1;
    document.getElementById("mac2").value = mac2;
    const linkText = document.getElementById("linkText");
    if (linkText) {
        linkText.innerHTML = `${name1} (${host1})<br>and<br>${name2}(${host2})`;
    }
    openModal(document.getElementById("linkDialog"));
}

// Read the radio buttons form to see what is selected for linking devices
async function recordLink() {
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
    try {
        const result = await postForm("duplicatesjoin", sendData);
        if (result !== "ok") {
            console.error(result);
        } else {
            reloadTable();
        }
    } catch (error) {
        console.error("Record link failed:", error);
    }
}