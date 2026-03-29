/* 
 *  Author: StoneSoft
 *  Copyright 1 April 2021
 *  All rights Reserved
 */

// Retrieve the username from the input field
function okayClick() {
    toast("");
    const username = checkEmailField();
    if (username.length > 0) {
        login(username);
    }
    return false;
}

//Verify the username is filled in properly
function checkEmailField() {
    let isValid = true;
    const emailField = document.getElementById('username');
    if (!emailField.checkValidity()) {
        isValid = false;
    }
    const username = emailField?.value?.trim();
    if (!username) {
        isValid = false;
    }
    if (!isValid) {
        return ""
    }
    return username;
}

async function login(username) {
    try {
        //Send long/lat to be saved
        let toSend ={username:"", tzoff: 0, lon:0.0, lat:0.0, timezone: ""};
        const geoString = sessionStorage.getItem('geo');
        if (geoString) {
            let geo = JSON.parse(geoString);
            toSend.username = username;
            toSend.tzoff = geo.tzoff;
            toSend.lon = geo.lon;
            toSend.lat = geo.lat;
            toSend.timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
        }

        // Get login options from your server. Here, we also receive the challenge.
        const response = await fetch('/api/passkey/loginStart', {
            method: 'POST', 
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(toSend)
        });
        // Check if the login options are ok.
        if (!response.ok) {
            const msg = await response.json();
            throw new Error('Failed to get login options from server: ' + msg);
        }
        // Convert the login options to JSON.
        const options = await response.json();

        // The browser can throw an error if `allowCredentials` is too large.
        // A client-side workaround is to truncate the array.
        // The ideal fix is for the server to not send so many credentials.
        if (options.publicKey?.allowCredentials && options.publicKey.allowCredentials.length > 64) {
            console.warn(`'allowCredentials' contains ${options.publicKey.allowCredentials.length} credentials, truncating to 64.`);
            options.publicKey.allowCredentials = options.publicKey.allowCredentials.slice(0, 64);
        }
        
        // This triggers the browser to display the passkey / WebAuthn modal (e.g. Face ID, Touch ID, Windows Hello).
        // A new assertionResponse is created. This also means that the challenge has been signed.
        const assertionResponse = await SimpleWebAuthnBrowser.startAuthentication({ optionsJSON: options.publicKey });

        // Send assertionResponse back to server for verification.
        const verificationResponse = await fetch('/api/passkey/loginFinish', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(assertionResponse)
        });

        const reply = await verificationResponse.json();
        toast(reply.msg);
        if (verificationResponse.ok) {
            sessionStorage.removeItem("geo");
            window.location.href = encodeURI("home.html");
        }
    } catch (error) {
        console.error(error);
        toast(error);
    }
}

