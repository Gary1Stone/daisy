/* global Metro */
// ajax.js

// The select with onchange event used to load a dependant selects 
// unfortunately triggers a fetch when they are first loaded.
// This prevents them from requesting data (new select) when the page is loaded.
let pageLoading = true;

$(document).ready(function () {
    pageLoading = false;
});

/**
 * Asynchronously fetches html 'controls (select, input, button...)' data using $.post and updates a target element's HTML.
 *
 * @param {string} target - The ID of the HTML element to update (without the '#').
 * @param {object} ctrlData - The data object to send with the POST request.
 */
async function getCtrl(target, ctrlData) {
    if (typeof pageLoading !== 'undefined' && pageLoading) {
        return;
    }
    // Validate input parameters
    if (typeof target !== 'string' || target.trim() === '') {
        toast("Error in getCtrl: 'target' must be a non-empty string.", "alert");
        return;
    }
    if (typeof ctrlData !== 'object' || ctrlData === null) {
        toast("Error in getCtrl: 'ctrlData' must be a valid object.", "alert");
        return;
    }
    const $targetElement = $("#" + target);
    if ($targetElement.length === 0) {
        toast(`Error in getCtrl: Target element '#${target}' not found.`, "alert");
        return;
    }
    // Show a loading indicator in the target element
     $targetElement.html(`<span class="mif-spinner4 ani-spin"></span>`);
    try {
        const response = await $.post("droplist", ctrlData);
        if (typeof response === 'string' || typeof response === 'number') {
            $targetElement.html(response);
        } else {
            toast(`Received invalid data for '#${target}'.`, "warning");
        }
    } catch (error) {
        console.error(`Error posting data to "droplist" for target '#${target}':`, error);
        let userMessage = `Failed to load data for '#${target}'.`;
        if (error.statusText) {
            userMessage += ` Status: ${error.statusText}`;
        }
        toast(userMessage, "alert");
        $targetElement.html(``);
    }
}
