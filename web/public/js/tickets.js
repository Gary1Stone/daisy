/* global Metro */

// Remembering the filter settings from screen to screen, not session to session.
// Should we automatically filter for this user?
$(document).ready(function () {
    const searchBox = document.querySelector('.table-component .table-top .table-search-block .input input[type="text"]');
    const searchString = sessionStorage.getItem('searchString');
    if (searchString === null) {
        searchBox.value = $("#fullname").val();
    } else {
        searchBox.value = searchString;
    }
    const event = new Event('change');
    searchBox.dispatchEvent(event);
    // remember any changes the user makes to the seach box
    searchBox.addEventListener('change', (event) => {
        sessionStorage.setItem('searchString', event.target.value);
    });
    tableChangeWatcher();
    addRowClick();
});



// Select the table and all rows
// Loop through each row and add a click event listener
function addRowClick() {
    const table = document.getElementById('tickettable');
    const rows = table.getElementsByTagName('tr');
    Array.from(rows).forEach(row => {
        if (row.parentNode.tagName.toLowerCase() === 'thead') {
          return;
        }
        row.removeEventListener('click', handleRowClick);
        row.addEventListener('click', handleRowClick);
    });
}

// Add a onclick event to each row of the table
// To navigate to the ticket
// First cell is rownum, second is checkbox, third cell has the AID
function handleRowClick(event) {
    const row = event.currentTarget;
    // Check if the row has a second cell (last row does not)
    if (row.cells.length < 3) {
        return;
    }
    const aid = txt2Int(row.cells[2].innerText); // 
    window.location.href =  encodeURI("ticket.html?aid=" + aid);
}

// Add a mutation observer to watch for any table changes
// Start observing the target node for configured mutations
function tableChangeWatcher() {
    const targetNode = document.getElementById('tickettable');
    const observer = new MutationObserver((mutationsList, observer) => {
        for (let mutation of mutationsList) {
            if (mutation.type === 'childList') {
                addRowClick();
            }
        }
    });
    const config = { childList: true, subtree: true };
    observer.observe(targetNode, config);
}
