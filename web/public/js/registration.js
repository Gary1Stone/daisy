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

function okayClick(event) {
    if (event) event.preventDefault();
    const div = document.getElementById('enterPasscode');
    if (div) {
        // Check computed style if inline style isn't set yet
        if (div.style.display === "block" || window.getComputedStyle(div).display === "block") {
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
        toast("Please enter a valid email address", "warning");
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
        toast("Please enter the code sent to your email", "warning");
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

    const btn = document.getElementById("btnSubmit");
    if (btn) {
        btn.disabled = true;
        btn.setAttribute("aria-busy", "true");
    }

    localStorage.setItem("username", username);
    try {
        const response = await fetch('/api/passkey/requestCode', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({username: username, apicode: apicode})
        });

        const reply = await response.json();
        if (!response.ok) {
            toast(reply.msg, "error");
            return;
        }

        if (reply.msg === "goLogin") {
            window.location.href =  encodeURI(loginPage);
        } else if (reply.msg === "goHome") {
            window.location.href =  encodeURI(homePage);
        } else {
            showPasscodeForm();
            startCountdown();
        }
    } catch (error) {
        toast(error, "error");
    } finally {
        if (btn) {
            btn.disabled = false;
            btn.setAttribute("aria-busy", "false");
        }
    }
}

function showPasscodeForm() {
    document.getElementById('enterPasscode').style.display = 'block';
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

    const btn = document.getElementById("btnSubmit");
    if (btn) {
        btn.disabled = true;
        btn.setAttribute("aria-busy", "true");
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
            throw new Error(msg.msg || 'User registered or bad options');
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
        if (!verificationResponse.ok) {
            throw new Error(reply.msg || "Failed to verify registration");
        }        
        toast(reply.msg, "success");
        sessionStorage.removeItem("geo");
        window.location.href = encodeURI(homePage);
    } catch (error) {
        console.error(error);
        toast(error.message || error, "error");
    } finally {
        if (btn) {
            btn.disabled = false;
            btn.setAttribute("aria-busy", "false");
        }
    }
}

function startCountdown() {
    var remainingMinutes = 15;
    const minsSpan = document.getElementById('mins');
    if (minsSpan) minsSpan.textContent = remainingMinutes;
    
    var countdown = setInterval(function() {
        const minsSpan = document.getElementById('mins');
        if (!minsSpan) return clearInterval(countdown);
        
        remainingMinutes--;
        minsSpan.textContent = remainingMinutes;
        if (remainingMinutes === 0) {
            clearInterval(countdown);
            minsSpan.textContent = "Expired!";
        }
    }, 60000);
}
