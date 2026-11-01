//Network.js

const UI = {
    form: () => document.getElementById("theForm"),
    site: () => document.getElementById("site"),
    name: () => document.getElementById("name"),
    cid: () => document.getElementById("cid"),
    parent: () => document.getElementById("parent"),
    kind: () => document.getElementById("kind"),
    office: () => document.getElementById("office"),
    popEdit: () => document.getElementById("popEdit"),
    iParent: () => document.getElementsByClassName("iParent"),
    iOffice: () => document.getElementsByClassName("iOffice"),
    iKind: () => document.getElementsByClassName("iKind"),
    treeview: () => document.getElementsByClassName("treeview"),
    helpDialog: () => document.getElementById("helpDialog")
};

document.addEventListener('DOMContentLoaded', function() {
    if (UI.parent()) {
        UI.parent().addEventListener("change", parentChanged);
    }
    if (UI.kind()) {
        UI.kind().addEventListener("change", kindChanged);
    }
});

function parentChanged() {
    const selected = UI.parent().value;
    htmx("/icons", {ctrl: "parent", selected: selected}, 'iParent');
}

function kindChanged() {
    const selected = UI.kind().value;
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
        const site = site().value;
        postJSON("/tree/show", {cid: cid, site: site}, (response) => {
            UI.cid().value = response.cid || "";
            UI.name().innerText = response.name || "";
            setDropdownValue('kind', response.kind || "");
            setDropdownValue('office', response.office || "");
            setDropdownValue('parent', response.parent || "");
            openModal(UI.popEdit());
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
        Cid: parseInt(UI.cid().value),
        Name: UI.name().innerText,
        Icon: "",
		Parent: parseInt(UI.parent().value),
		Office: UI.office().value,
		Kind: UI.kind().value,
        Site: UI.site().value,
    };
    
    try {
        htmx("/tree/update", formData, "treeview");
        closeModal(UI.popEdit());
    } catch (error) {
        alert("Failed to update device details: " + error);
        console.error("Failed to get response from sever:", error);
    }
}

function showHelp() {
    openModal(UI.helpDialog());
}

function goto(url) {
    if (!url) return
    window.location.href = encodeURI(url);
}

function newSiteSelected() {
    window.location.href = encodeURI("network.html?site=" + UI.site().value);
}
