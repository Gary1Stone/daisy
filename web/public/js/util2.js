// util2.js

function txt2Int(value) {
    const result = parseInt(value, 10);
    return isNaN(result) ? 0 : result;
}

/*
Snackbar Usage:

Snackbar.push({
    message: "Saved successfully!", // mandatory: message
    type: "success",            // optional: "success", "error", "warning", or "info"
    duration: 6000              // optional: milliseconds, default is 3 seconds
    actionText: "Undo",         // optional: provide button for user to click, and onAction runs when they click it
    onAction: () => {
        console.log("Undo clicked");
    }
});

*/
