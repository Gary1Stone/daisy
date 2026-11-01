// device_mac.js

const UI = {
    site: () => document.getElementById("site"),
    cid: () => document.getElementById("cid"),
    txtSearch: () => document.getElementById("txtSearch"),
    searchListing: () => document.getElementById("searchListing"),
    popAssociate: () => document.getElementById("popAssociate"),
    mid: () => document.getElementById("mid")
};

document.addEventListener('DOMContentLoaded', function() {
    setupTableSearch(); //On the Find popup
});

function showHelp() {
    const helpDialog = document.getElementById("helpDialog");
    if (helpDialog) {
        openModal(helpDialog);
    }
}

function removeMac(cid, mid) {
    if (!mid || !cid) return;
    htmx("/devmacs/removemac", {cid: cid, mid: mid}, "associated").then(() => reloadPage(cid));
}

function getDeviceList(id) {
    if (!id) return;
    cid = txt2Int(id);
    if (cid < 1) return;
    const site = UI.site.value;
    htmx("/devmacs/getdevicelist", {cid: cid, site: site}, "deviceListing");
}

function reloadPage(id) {
    if (!id) return;
    cid = txt2Int(id);
    if (cid < 1) return;
    window.location.href = encodeURI("devmacs.html?cid=" + cid);
}

function setupTableSearch() {
    UI.searchListing.addEventListener('input', function () {
        const filter = this.value.toLowerCase();
        const rows = document.querySelectorAll('#listingTable tbody tr');
        rows.forEach(row => {
            // Get all cell text values inside the current row
            const rowText = row.textContent.toLowerCase();
                
            // If the search query exists in the row text, show it. Otherwise, hide it.
            if (rowText.includes(filter)) {
                row.style.display = '';
            } else {
                row.style.display = 'none';
            }
        });
    });
}

//Show popup of association recomendations for selected MAC
function rowClicked(id = 0) {
    if (!id) return;
    const mid = txt2Int(id);
    const cid = txt2Int(UI.cid.value);
    if (mid < 1 || cid < 1) return;
    UI.popAssociate.showModal();
    htmx("/devmacs/associates", {cid: cid, mid: mid}, "associateDetails");
}

// Make association
function associatedMac() {
    const cid = txt2Int(UI.cid.value);
    const associateMac = document.getElementById("associateMac");
    if (!associateMac) return;
    const mid = txt2Int(associateMac.value);
    if (mid < 1 || cid < 1) return;
    postJSON("/devmacs/associated", {cid: cid, mid: mid}, (reply) => {
        if (reply !== "okay") {
           console.log(reply);
        } else {
            reloadPage(cid);
        }
    });
}