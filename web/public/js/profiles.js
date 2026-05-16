// Profiles.js

let btnNew;

document.addEventListener('DOMContentLoaded', function() {
  btnNew = new Button("btnNew");
  buildTable('profileTable');
  tableChangeWatcher();
  addRowClick();
});

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
// To navigate to the profile
function handleRowClick(event) {
  const row = event.currentTarget;
  const uid = row.dataset.uid;
  if (!uid) return;
  window.location.href = encodeURI("profile.html?uid=" + uid);
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

function popSearch() {
  const searchInput = document.getElementById("txtSearch")
  if (window.getComputedStyle(searchInput).display === "none") {
    searchInput.style.display = "block";
    searchInput.focus();
  } else {
    // Clear searchInput 
    searchInput.value = "";
  // Create and dispatch an 'input' event
    const inputEvent = new Event('input', {
      bubbles: true, // Allows the event to bubble up the DOM tree
      cancelable: true // Allows the event to be cancelled
    });
    searchInput.dispatchEvent(inputEvent);
    searchInput.style.display = "none";
  }
}