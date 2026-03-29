function getReport(task, devType="") {
    let sendData = { task, devType };
    $.post("reports", sendData).then(response => {
        $("#report").html(response);
    });
}
