// device_mac.js

const UI = {
    site: null, cid: null, txtSearch: null, searchListing: null, popAssociate: null, mid: null
};

document.addEventListener('DOMContentLoaded', function() {
    UI.site = document.getElementById("site");
    UI.cid = document.getElementById("cid");
    UI.txtSearch = document.getElementById("txtSearch");
    UI.searchListing = document.getElementById("searchListing");
    UI.popAssociate = document.getElementById("popAssociate");
    UI.mid = document.getElementById("mid");
    setupTableSearch(); //On the Find popup

});

function removeMac(cid, mid) {
    if (!mid || !cid) return;
    htmx("/devmacs/removemac", {cid: cid, mid: mid}, "macListing");
}

function getDeviceList(cid) {
    const site = UI.site.value;
    htmx("/devmacs/getdevicelist", {cid: txt2Int(cid), site: site}, "deviceListing")
        .then(() => buildTable("listingTable"));
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

function rowClicked(id=0) {
    if (!id) return;
    const mid = txt2Int(id);
    if (mid < 1) return;
    UI.mid.value = mid;
    UI.popAssociate.showModal();
}



function associateMac() {
    UI.popAssociate.close();
    UI.mid.value = 0;

}