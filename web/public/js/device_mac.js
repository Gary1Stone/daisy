// tree.js

const UI = {
    site: null, cid: null, selected: null, name: null, hostname: null, devicelist: null, model: null, hostname: null,
    kind: null, office: null
};

document.addEventListener('DOMContentLoaded', function() {
    UI.site = document.getElementById("site");
    UI.cid = document.getElementById("cid");
    UI.selected = document.getElementById("selected");
    UI.name = document.getElementById("name");
    UI.hostname = document.getElementById("hostname");
    UI.devicelist = document.getElementById("devicelist");
    UI.model = document.getElementById("model");
    UI.hostname = document.getElementById("hostname");
    UI.kind = document.getElementById("kind");
    UI.office = document.getElementById("office");

    // if (UI.parent) {
    //     UI.parent.addEventListener("change", parentChanged);
    // }
    // if (UI.kind) {
    //     UI.kind.addEventListener("change", kindChanged);
    // }
});

// function parentChanged() {
//     const selected = UI.parent.value;
//     htmx("/icons", {ctrl: "parent", selected: selected}, 'iParent');
// }

function cidSelected(cid) {
    if (!cid || cid <= 0) {
        console.error("Missing Mid parameter");
        return;
    }
    UI.selected.innerHTML = document.getElementById("cid" + cid).innerHTML;
    UI.cid.value = cid;
    UI.devicelist.removeAttribute('open');
    const site = UI.site.value;

    try {
        postJSON("/tree/show", {cid: cid, site: site}, (response) => {
            UI.name.innerHTML = response.name || "";
            UI.model.innerHTML = response.model || "";
            UI.kind.innerHTML = response.kindtitle || "";
            UI.office.innerHTML = response.officetitle || "";
        });
    } catch (error) {
        console.error("Failed to get response from server:", error);
    }
}

// function saveDetail() {
//     const formData = {
//         Cid: parseInt(UI.cid.value),
//         Name: UI.name.innerText,
//         Icon: "",
// 		Parent: parseInt(UI.parent.value),
// 		Office: UI.office.value,
// 		Kind: UI.kind.value
//     };
    
//     try {
//         postJSON("/tree/update", formData, (response) => {
//             if (response == "OK") {
//                 closeModal(UI.popEdit);
//                 location.reload();
//             } else {
//                 alert("Failed to update device details: " + response);
//             }
//         });
//     } catch (error) {
//         alert("Failed to update device details: " + error);
//         console.error("Failed to get response from server:", error);
//     }
// }
