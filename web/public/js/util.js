/* global Metro */
// util.js

function txt2Int(value) {
    const result = parseInt(value, 10);
    return isNaN(result) ? 0 : result;
}

// Colours are primary, secondary, success, alert, warning, yellow, info, light
function toast(msg, colour) {
    let txt = "";
    if (typeof colour === 'undefined') {
        colour = "secondary";
    }
    // Check if msg is an error object
    if (msg instanceof Error) {
        txt = msg.message;
        colour = "alert";
    // Check if msg is an object and has a message property
    } else if (msg && typeof msg === 'object' && 'message' in msg) { 
        txt = msg.message;
    } else if (typeof msg === 'string' && msg.trim() !== '') {
        txt = msg;  // Directly use msg if it's a valid string
    } else {
        return; // Exit if none of the conditions are met
    }
    var options = {showTop: true, distance: 160};
    // Assuming Metro UI is loaded
    if (typeof Metro !== 'undefined' && Metro.toast) {
        Metro.toast.create(txt, null, 5000, colour, options);
    //  Metro.toast.create(message, callback, timeout, color, options)
    } else {
        console.warn("Metro UI toast function not available. Message:", txt);
        alert(txt); // Fallback
    }
}
