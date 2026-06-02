function getReport(task, devType = "") {
    const sendData = { task, devType };
    postForm("reports", sendData, (html) => {
        const reportEl = document.getElementById("report");
        if (reportEl) {
            
            reportEl.innerHTML = html;

            // Find all tables within the report div
            const tables = reportEl.querySelectorAll('table');

            // Extract the IDs into an array, filtering out empty IDs
            const tableIds = Array.from(tables)
            .map(table => table.id)
            .filter(id => id !== "");

            // Initalize the tables for sorting
            for (let i = 0; i < tableIds.length; i++) {
                buildTable(tableIds[i]);
            }
        }
    });
}
