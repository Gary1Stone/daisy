// tree.js

const UI = {
    form: () => document.getElementById("theForm"),
    site: () => document.getElementById("site"),
    name: () => document.getElementById("name"),
    cid: () => document.getElementById("cid"),
    parent: () => document.getElementById("parent"),
    kind: () => document.getElementById("kind"),
    office: () => document.getElementById("office"),
    popEdit: () => document.getElementById("popEdit"),
};

function showDetail(cid) {
    if (!cid) {
        alert("Missing id parameter");
        console.error("Missing Mid parameter");
        return;
    }
    try {
        postJSON("/tree/show", {cid: cid}, (response) => {
            UI.cid().value = response.cid || "";
            UI.name().innerText = response.name || "";
            UI.parent().value = response.parent || "";
            UI.kind().value = response.kind || "";
            UI.office().value = response.office || "";
            openModal(UI.popEdit());
        });
    } catch (error) {
        alert("Failed to get response from sever: " + error);
        console.error("Failed to get response from sever:", error);
    }
}

function saveDetail() {
    formData = {
        Cid: parseInt(UI.cid().value),
        Name: UI.name().innerText,
        Icon: "",
		Parent: parseInt(UI.parent().value),
		Office: UI.office().value,
		Kind: UI.kind().value
    };
    
    try {
        postJSON("/tree/update", formData, (response) => {
            if (response == "OK") {
                closeModal(UI.popEdit());
                location.reload();
            } else {
                alert("Failed to update device details: " + response);
            }
        });
    } catch (error) {
        alert("Failed to update device details: " + error);
        console.error("Failed to get response from sever:", error);
    }
}
