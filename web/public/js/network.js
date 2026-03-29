//Network.js

function goToOnline() {
    window.location.href = encodeURI("online.html");
    return false;
}

function goToHistory() {
    mid = $("#mid").val();
    window.location.href = encodeURI("history.html?mid=" + mid);
    return false;
}

function showHelp() {
    Metro.dialog.open("#helpDialog");
}

function goToDuplicates() {
    window.location.href = encodeURI("duplicates.html");
    return false;
}

function goToCorrealtion() {
    window.location.href = encodeURI("correlation.html");
    return false;
}
