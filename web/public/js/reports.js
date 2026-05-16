function getReport(task, devType = "") {
    const sendData = { task, devType };
    fetch("reports", {
        method: "POST",
        body: new URLSearchParams(sendData)
    })
    .then(response => response.text())
    .then(html => {
        const reportEl = document.getElementById("report");
        if (reportEl) reportEl.innerHTML = html;
    })
    .catch(err => console.error("Report fetch failed:", err));
}
