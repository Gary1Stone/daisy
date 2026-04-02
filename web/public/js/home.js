// home.js
window.addEventListener("load", initialize);

function initialize() {
    const hrs = new Date().getHours();
    let greeting = "Good evening";
    if (hrs < 10) {
        greeting = "Good morning";
    } else if (hrs < 20) {
        greeting = "Good day";
    }
    const greetingEl = document.getElementById("greeting");
    if (greetingEl) {
        greetingEl.innerHTML = greeting;
    }

    //Send long/lat to be saved
    const geoString = sessionStorage.getItem('geo');
    if (geoString) {
        let geo = JSON.parse(geoString);
        sessionStorage.removeItem("geo");
        geo.task = "save_lon_lat";

        fetch("home", {
            method: "POST",
            body: new URLSearchParams(geo)
        }).then(response => response.text()).then(response => {
            if (response !== "ok") {
                toast(response);
            }
        });
    }
}

function ackAlert(aid = 0) {
    const sendData = {
        task: "get_alerts", 
        aid: aid
    };
    fetch("home", {
        method: "POST",
        body: new URLSearchParams(sendData)
    }).then(response => response.text()).then(response => {
        const alertsEl = document.getElementById("alerts");
        if (alertsEl) {
            alertsEl.innerHTML = response;
        }
    });
}

function startWizard() {
    const wizkey = document.getElementById("wizkey");
    if (wizkey) {
        const selected = wizkey.options[wizkey.selectedIndex].value;
        if (selected) {
            window.location.href = encodeURI("wizard.html?wizkey=" + selected);
        }
    }
}
