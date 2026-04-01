/* 
 *  Author: StoneSoft
 *  Copyright 17 Jan 2022
 *  All rights Reserved
 * 
 *  registration.js
 */

const loginPage = "login.html";
const homePage = "home.html";

window.addEventListener('load', () => {
    const username = localStorage.getItem('username');
    const emailField = document.getElementById('username');
    if (username && emailField && !emailField.value.trim()) {
        emailField.value = username;
    }
});

function okayClick() {
    const div = document.getElementById('enterPassscode');
    if (div) {
    // It's better to check for a class or data attribute than style, but this is a small improvement.
        if (div.style.display === "block") {
            register();
        } else {
            requestCode()
        }
    }
}

// Verify the username is filled in properly
function checkEmailField() {
    const emailField = document.getElementById('username');
    const username = emailField?.value?.trim();
    if (!username || !emailField.checkValidity()) {
        return "";
    }
    return username;
}

// Verify the passcode field is filled in properly
function checkPasscodeField() {
    const passcodeField = document.getElementById('passcode');
    const passcode = passcodeField?.value?.trim();
    if (!passcode || !passcodeField.checkValidity()) {
        showPasscodeForm();
        return "";
    }
    return passcode;
}

async function requestCode() {
    const username = checkEmailField();
    if (username.length === 0) {
        return;
    }
    const apicode = document.getElementById("apicode")?.value?.trim();
    if (!apicode) {
        return;
    }
    localStorage.setItem("username", username);
    try {
        const response = await fetch('/api/passkey/requestCode', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({username: username, apicode: apicode})
        });
        if (!response.ok) {
            const reply = await response.json();
            toast(reply.msg);
        } else {
            const reply = await response.json();
            if (reply.msg === "goLogin") {
                window.location.href =  encodeURI(loginPage);
            } else if (reply.msg === "goHome") {
                window.location.href =  encodeURI(homePage);
            } else {
                showPasscodeForm();
                startCountdown();
            }
        }
    } catch (error) {
        toast(error);
    }
}

function showPasscodeForm() {
    document.getElementById('enterPassscode').style.display = 'block';
    const passcodeField = document.getElementById('passcode');
    passcodeField.disabled = false;
    passcodeField.focus();
}

async function register() {
    let username = checkEmailField();
    if (username.length === 0) {
        return;
    }
    let passcode = checkPasscodeField();
    if (passcode.length === 0) {
        return;
    }
    const apicode = document.getElementById("apicode")?.value?.trim();
    if (!apicode) {
        return;
    }
    try {
        // Send geo/tz data to be saved in the session for automatic login
        let toSend = {username: username, passcode: passcode, apicode: apicode, tzoff: 0, lon: 0.0, lat: 0.0, timezone: ""};
        const geoString = sessionStorage.getItem('geo');
        if (geoString) {
            let geo = JSON.parse(geoString);
            toSend.tzoff = geo.tzoff;
            toSend.lon = geo.lon;
            toSend.lat = geo.lat;
            toSend.timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
        }

        // Get registration options from your server. Here, we also receive the challenge.
        const response = await fetch('/api/passkey/registerStart', {
            method: 'POST', 
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(toSend)
        });

        // Check if the registration options are ok.
        if (!response.ok) {
            const msg = await response.json();
            throw new Error('User registered or bad options: ' + msg);
        }

        // Convert the registration options to JSON.
        const options = await response.json();

        // This triggers the browser to display the passkey / WebAuthn modal (e.g. Face ID, Touch ID, Windows Hello).
        // A new attestation is created. This also means a new public-private-key pair is created.
        const attestationResponse = await SimpleWebAuthnBrowser.startRegistration({ optionsJSON: options.publicKey });

        // Send attestationResponse back to server for verification and storage.
        const verificationResponse = await fetch('/api/passkey/registerFinish', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(attestationResponse)
        });

        const reply = await verificationResponse.json();
        console.log(reply.msg);
        if (verificationResponse.ok) {
            sessionStorage.removeItem("geo");
            window.location.href =  encodeURI(homePage);
        }
    } catch (error) {
        toast(error);
    }
}

function startCountdown() {
    var remainingMinutes = 15;
    var countdown = setInterval(function() {
        const minsSpan = document.getElementById('mins');
        minsSpan.textContent = remainingMinutes;
        remainingMinutes--;
        if (remainingMinutes === 0) {
            clearInterval(countdown);
            minsSpan.textContent = "Expired!";
        }
    }, 60000);
}
