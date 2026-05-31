//Network.js

function goToOnline() {
    window.location.href = encodeURI("online.html");
    return false;
}

function goToHistory() {
    const midElement = document.getElementById("mid");
    const mid = midElement ? midElement.value : "";
    window.location.href = encodeURI("history.html?mid=" + mid);
    return false;
}

function showHelp() {
    openModal(document.getElementById("helpDialog"));
}

function goToDuplicates() {
    window.location.href = encodeURI("duplicates.html");
    return false;
}

function goToCorrealtion() {
    window.location.href = encodeURI("correlation.html");
    return false;
}
