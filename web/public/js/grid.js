// grid.js

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
}