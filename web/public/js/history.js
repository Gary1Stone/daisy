//History.js

document.addEventListener('DOMContentLoaded', function() {
    const mid = document.getElementById("mid");
    if (mid) {
        mid.addEventListener("change", function () {
            const val = mid.value;
            window.location.href = encodeURI("history.html?mid=" + val);
        });
    }
});

function showHelp() {
    openModal(document.getElementById("helpDialog"));
}
