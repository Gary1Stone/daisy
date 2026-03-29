// home.js

$(document).ready(function () {
    const hrs = new Date().getHours();
    let greeting = "Good evening";
    if (hrs < 10) {
        greeting = "Good morning";
    } else if (hrs < 20) {
        greeting = "Good day";
    }
    $("#greeting").html(greeting);
    //Send long/lat to be saved
    const geoString = sessionStorage.getItem('geo');
    if (geoString !== null) {
        let geo = JSON.parse(geoString);
        sessionStorage.removeItem("geo");
        geo.task = "save_lon_lat";
        $.post("home", geo).then(response => {
            if (response !== "ok") {
                console.log(response);
            }
        });
    }
});

function ackAlert(aid = 0) {
    let sendData = {
        task: "get_alerts", 
        aid: aid
    };
    $.post("home", sendData).then(response => {
        $("#alerts").html(response);
    });
}
