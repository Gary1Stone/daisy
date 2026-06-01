// Control.js

function showActiveUsers() {
    const content = document.getElementById("popupContent");
    const title = document.getElementById("popupTitle");
    if (content) content.innerHTML = "";
    if (title) title.innerHTML = "";
    getActiveUsers();
    const modal = document.getElementById("popup");
    if (modal) modal.showModal();
}

async function endSession(id) {
    const sendData = getFormData("end_session", id);
    try {
        const response = await fetch("control", {
            method: "POST",
            body: new URLSearchParams(sendData)
        });
        const html = await response.text();
        const content = document.getElementById("popupContent");
        const title = document.getElementById("popupTitle");
        if (content) content.innerHTML = html;
        if (title) title.innerHTML = "Active Users";
    } catch (error) {
        console.error("End session failed:", error);
    }
}

async function getActiveUsers() {
   const sendData = getFormData("get_active_users");
    try {
        const response = await fetch("control", {
            method: "POST",
            body: new URLSearchParams(sendData)
        });
        const html = await response.text();
        const content = document.getElementById("popupContent");
        const title = document.getElementById("popupTitle");
        if (content) content.innerHTML = html;
        if (title) title.innerHTML = "Active Users";
    } catch (error) {
        console.error("Get active users failed:", error);
    }
}

async function getAttacks(duration) {
    const sendData = getFormData("get_attacks", duration);
    let title = "Attacks ";
    if (duration === 1) { title += "(Day)"; } 
    else if (duration === 7) { title += "(Week)"; } 
    else if (duration === 30) { title += "(Month)"; }
    
    try {
        const response = await fetch("control", {
            method: "POST",
            body: new URLSearchParams(sendData)
        });
        const html = await response.text();
        const content = document.getElementById("popupContent");
        const titleEl = document.getElementById("popupTitle");
        if (content) content.innerHTML = html;
        if (titleEl) titleEl.innerHTML = title;
        openModal(document.getElementById("popup"));
    } catch (error) {
        console.error("Get attacks failed:", error);
    }
}


function getFormData(task, id=0) {
    return {task: task, id: id};
}