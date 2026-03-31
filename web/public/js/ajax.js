/* global Metro */
// ajax.js

// The select with onchange event used to load a dependant selects 
// unfortunately triggers a fetch when they are first loaded.
// This prevents them from requesting data (new select) when the page is loaded.
let pageLoading = true;

// When the page is finished loading
document.addEventListener("DOMContentLoaded", function() {
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




//Ajax.js

// postJSON: Send JSON, Receive JSON or string
// response.status = true/false
// response.statusText - JSON reply
async function postJSON(url, data = {}, successCallback) {
    try {
        const response = await fetch(url, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json",
            },
            body: JSON.stringify(data),
        });
        
        if (!response.ok) {
            const errorText = response.statusText || `HTTP error ${response.status}`;
            console.error("AJAX HTTP Error:", `Status: ${response.status}`, `Message: ${errorText}`);
            throw new Error(`HTTPError ${response.status}: ${errorText}`);
        }

        const contentType = response.headers.get("content-type");
        let responseData;
        if (contentType && contentType.includes("application/json")) {
            responseData = await response.json();
        } else {
            responseData = await response.text();
        }

        if (typeof successCallback === 'function') {
            successCallback(responseData);
        }
        return responseData;
    } catch (error) {
        if (error.message?.startsWith("HTTPError ")) {
            console.debug("Fetch chain aborted due to HTTP error:", error.message);
        } else {
            console.error("AJAX Request Failed (Network/Other):", error);
        }
        throw error; // Let the caller handle it if needed
    }
}

// fetch html and replace the contents of targetId
async function htmx(url, formData, targetId) {
    try {
        await postJSON(url, formData, (htmlResponse) => {
            const target = document.getElementById(targetId);
            if (target) {
                if (htmlResponse != "error") {
                    target.innerHTML = htmlResponse;
                } else {
                    console.error("Failed to get HTML:")
                    target.innerHTML = "";
                }
            } else {
                console.warn("No container found for response");
            }
        });
    } catch (error) {
        console.error("Failed to get HTML:", error);
    }
}

function showMsg(msg) {
    const msgDiv = document.getElementById("msg");
    if (msgDiv) {
        msgDiv.innerHTML = msg;
        setTimeout(() => {
            msgDiv.innerHTML = "&nbsp;";
        }, 10000);
    }
}
