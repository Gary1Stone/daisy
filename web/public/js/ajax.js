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
 * Standardized AJAX request handler.
 * @param {string} url - Target URL.
 * @param {object} options - Fetch options.
 * @returns {Promise<any>}
 */
async function apiRequest(url, options = {}) {
    try {
        const response = await fetch(url, options);

        if (!response.ok) {
            const errorMsg = await response.text().catch(() => response.statusText);
            console.error("AJAX HTTP Error:", `Status: ${response.status}`, `URL: ${url}`, `Message: ${errorMsg}`);
            throw new Error(`HTTPError ${response.status}: ${errorMsg || response.statusText}`);
        }

        const contentType = response.headers.get("content-type");
        return contentType && contentType.includes("application/json") 
            ? await response.json() 
            : await response.text();
    } catch (error) {
        if (!error.message.startsWith("HTTPError")) {
            console.error(`AJAX Request Failed (Network/Other) to ${url}:`, error);
        }
        throw error;
    }
}

/**
 * Asynchronously fetches html 'controls (select, input, button...)' data using post and updates a target element's HTML.
 *
 * @param {string} target - The ID of the HTML element to update (without the '#').
 * @param {object} ctrlData - The data object to send with the POST request.
 */
async function getCtrl(target, ctrlData) {
    if (pageLoading) {
        return;
    }

    const targetElement = document.getElementById(target);
    if (!targetElement) {
        toast(`Error in getCtrl: Target element '#${target}' not found.`, "alert");
        return;
    }

    // Show a loading indicator in the target element
    targetElement.innerHTML = `<span class="mif-spinner4 ani-spin"></span>`;

    try {
        const responseData = await apiRequest("droplist", {
            method: "POST",
            body: new URLSearchParams(ctrlData)
        });
        targetElement.innerHTML = responseData;
    } catch (error) {
        toast(`Failed to load control: ${error.message}`, "alert");
        targetElement.innerHTML = ``;
    }
}

/**
 * Sends JSON to the server and returns a response.
 * Supports both Promise .then() usage and success callbacks.
 * 
 * @param {string} url - Target URL.
 * @param {object} data - Data to stringify and send.
 * @param {function} successCallback - Optional callback for the response data.
 */
async function postJSON(url, data = {}, successCallback) {
    try {
        const responseData = await apiRequest(url, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json",
            },
            body: JSON.stringify(data),
        });

        if (typeof successCallback === 'function') {
            successCallback(responseData);
        }
        return responseData;
    } catch (error) {
        throw error;
    }
}

/**
 * Fetches HTML via postJSON and replaces the content of the targetId.
 * 
 * @param {string} url - Target URL.
 * @param {object} formData - Data to send.
 * @param {string} targetId - ID of the element to update.
 */
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
