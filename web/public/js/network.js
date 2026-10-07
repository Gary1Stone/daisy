//Network.js

const UI = {
    form: null, site: null, name: null, cid: null, parent: null, kind: null, office: null, 
    popEdit: null, iParent: null, iOffice: null, iKind: null, treeview: null
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
    UI.treeview = document.getElementsByClassName("treeview");
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

function getOfficeCtrl() {
    return;
}

function showDetail(cid) {
    if (!cid) {
        alert("Missing id parameter");
        console.error("Missing Mid parameter");
        return;
    }
    try {
        site = document.getElementById("site").value;
        postJSON("/tree/show", {cid: cid, site: site}, (response) => {
            UI.cid.value = response.cid || "";
            UI.name.innerText = response.name || "";
            UI.parent.value = response.parent || "";
            setDropdownValue('kind', response.kind || "");
            setDropdownValue('office', response.office || "");
            // UI.kind.value = response.kind || "";
            // UI.office.value = response.office || "";
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
		Kind: UI.kind.value,
        Site: UI.site.value,
    };
    
    try {
        htmx("/tree/update", formData, "treeview")
        closeModal(UI.popEdit);
    } catch (error) {
        alert("Failed to update device details: " + error);
        console.error("Failed to get response from sever:", error);
    }
}

function showHelp() {
    openModal(document.getElementById("helpDialog"));
}

function goto(url) {
    if (!url) return
    window.location.href = encodeURI(url);
}

function newSiteSelected() {
    const newSite = UI.site.value;
    window.location.href = encodeURI("network.html?site=" + newSite);
}
