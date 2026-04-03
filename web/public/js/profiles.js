// Profiles.js

window.addEventListener("load", initialize);

function initialize() {
  buildTable('profileTable');
  tableChangeWatcher();
  addRowClick();
}

// Select the table and all rows
// Loop through each row and add a click event listener
function addRowClick() {
  const table = document.getElementById('profileTable');
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
  const uid = txt2Int(row.cells[0].innerText); // 
  window.location.href =  encodeURI("profile.html?uid=" + uid);
}

// Add a mutation observer to watch for any table changes
// Start observing the target node for configured mutations
function tableChangeWatcher() {
  const targetNode = document.getElementById('profileTable');
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


function addRecord() {
  location.href='profile.html?uid=0';
}
