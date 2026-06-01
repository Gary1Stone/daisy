// duplicates.js

document.addEventListener('DOMContentLoaded', function() {
    const avoidSelect = document.getElementById("avoidSelect");
    if (avoidSelect) {
        avoidSelect.addEventListener("change", async function () {
        const optionSelected = avoidSelect.value;
        const macsArray = optionSelected.split("_");
        const sendData = { 
            mac1: macsArray[0], 
            mac2: macsArray[1]
        }
        try {
            const response = await fetch("duplicates", {
                method: "POST",
                body: new URLSearchParams(sendData)
            });
            const html = await response.text();
            document.getElementById("chart").innerHTML = html;
        } catch (e) { console.error(e); }
    })};
});

function showHelp() {
    openModal(document.getElementById("helpDialog"));
}

function showLink() {
    const avoid = document.getElementById("avoidSelect");
    const text = avoid.options[avoid.selectedIndex].text;
    const linkText = document.getElementById("linkText");
    if (linkText) linkText.textContent = text;''
    openModal(document.getElementById("linkDialog"));
}

// Read the radio buttons form to see what is selected for linking devices
async function recordLink() {
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
    const optionSelected = document.getElementById("avoidSelect").value;
    const macsArray = optionSelected.split("_");
    const sendData = { 
        mac1: macsArray[0], 
        mac2: macsArray[1],
        isSame: isSame,
        isIgnore: isIgnore
    }
    try {
        const response = await fetch("duplicatesjoin", {
            method: "POST",
            body: new URLSearchParams(sendData)
        });
        const result = await response.text();
        if (result !== "ok") {
            console.log(result);
        } else {
            window.location.reload();
        }
    } catch (e) { console.error(e); }
}
