// table.js

document.addEventListener('DOMContentLoaded', function() {
    // Find all tables with data-sortable='true'
    const sortableTables = document.querySelectorAll('[data-sortable="true" i]');
    sortableTables.forEach(table => {
        buildTable(table.id);
    });
});


// For the html table, enable sorting by columns and search by table data contents.
function buildTable(tableID) {
    const table = document.getElementById(tableID);
    if (!table) return;

    const tbody = table.querySelector('tbody');
    const headers = table.querySelectorAll('thead th');
    // Common search input ID used across the application (e.g., in devices.js)
    const searchInput = document.getElementById('txtSearch');

    // --- Filtering Logic ---
    if (searchInput) {
        searchInput.addEventListener('input', function() {
            const query = this.value.toLowerCase();
            Array.from(tbody.rows).forEach(row => {
                const match = Array.from(row.cells).some(td => 
                    td.textContent.toLowerCase().includes(query)
                );
                row.style.display = match ? '' : 'none';
            });
        });
    }

    // --- Sorting Logic ---
    headers.forEach((th, index) => {
        th.style.cursor = 'pointer';
        th.title = "Click to sort";
        th.addEventListener('click', () => {
            const rows = Array.from(tbody.rows);
            const isAsc = th.getAttribute('data-sort') === 'asc';
            const direction = isAsc ? -1 : 1;

            //modify aria-sort="none" for all columns, then set aria-sort="ascending" or aria-sort="decending" for the sorted column
            headers.forEach(h => h.setAttribute('aria-sort', 'none'));
            th.setAttribute('aria-sort', isAsc ? 'ascending' : 'descending');

            rows.sort((a, b) => {
                const aCol = a.cells[index].textContent.trim();
                const bCol = b.cells[index].textContent.trim();

                // Numeric: true handles IDs and numbers naturally within strings
                return aCol.localeCompare(bCol, undefined, { numeric: true, sensitivity: 'base' }) * direction;
            });

            // Reset other headers and toggle current sort direction attribute
            headers.forEach(h => h.removeAttribute('data-sort'));
            th.setAttribute('data-sort', isAsc ? 'desc' : 'asc');
            rows.forEach(row => tbody.appendChild(row));
        });
    });
    tableChangeWatcher(tableID);
    addRowClick(tableID);
}

// Add a mutation observer to watch for any table changes
// Start observing the target node for configured mutations
function tableChangeWatcher(tableId) {
    if (!tableId) return;
    const targetNode = document.getElementById(tableId);
    if (!targetNode) return;
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


// Select the table and all rows
// Loop through each row and add a click event listener
function addRowClick(tableId) {
    if (!tableId) return;
    const table = document.getElementById(tableId);
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
  const id = row.dataset.id;
  if (!id) return;
  addRecord(id);
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