// tree.js

const UI = {
    form: null, site: null, name: null, cid: null, parent: null, kind: null, 
    office: null, popEdit: null, iParent: null, iOffice: null, iKind: null
};

document.addEventListener('DOMContentLoaded', function() {
    UI.form = document.getElementById("theForm");
    UI.site = document.getElementById("site");
    UI.name = document.getElementById("name");
    UI.cid = document.getElementById("cid");
    UI.parent = document.getElementById("parent");
    UI.kind = document.getElementById("kind");
    UI.office = document.getElementById("office");
    UI.popEdit = document.getElementById("popEdit");
    UI.iParent = document.getElementsByClassName("iParent");
    UI.iOffice = document.getElementsByClassName("iOffice");
    UI.iKind = document.getElementsByClassName("iKind");
    if (UI.parent) {
        UI.parent.addEventListener("change", parentChanged);
    }
    if (UI.kind) {
        UI.kind.addEventListener("change", kindChanged);
    }
});

function parentChanged() {
    const selected = UI.parent.value;
    htmx("/icons", {ctrl: "parent", selected: selected}, 'iParent');
}

function kindChanged() {
    const selected = UI.kind.value;
    htmx("/icons", {ctrl: "kind", selected: selected}, 'iKind');
}

function showDetail(cid) {
    if (!cid) {
        alert("Missing id parameter");
        console.error("Missing Mid parameter");
        return;
    }
    try {
        postJSON("/tree/show", {cid: cid}, (response) => {
            UI.cid.value = response.cid || "";
            UI.name.innerText = response.name || "";
            UI.parent.value = response.parent || "";
            UI.kind.value = response.kind || "";
            UI.office.value = response.office || "";
            openModal(UI.popEdit);
            parentChanged();
            kindChanged();
        });
    } catch (error) {
        alert("Failed to get response from sever: " + error);
        console.error("Failed to get response from sever:", error);
    }
}

function saveDetail() {
    const formData = {
        Cid: parseInt(UI.cid.value),
        Name: UI.name.innerText,
        Icon: "",
		Parent: parseInt(UI.parent.value),
		Office: UI.office.value,
		Kind: UI.kind.value
    };
    
    try {
        postJSON("/tree/update", formData, (response) => {
            if (response == "OK") {
                closeModal(UI.popEdit);
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
