//History.js

document.addEventListener('DOMContentLoaded', function() {
    $("#mid").change(function () {
        mid = $("#mid").val();
        window.location.href = encodeURI("history.html?mid=" + mid);
    });
});

function showHelp() {
    Metro.dialog.open("#helpDialog");
}
