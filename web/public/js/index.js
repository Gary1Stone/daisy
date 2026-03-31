// index.js

// geo is the object that holds the geographic data such as longitude and latitude
// We store it in session so we don't have to repeat calls when page is reloaded
// and it is used from sessison on the login page and the home page
let geo = { task: "", tzoff: 0, lon: 0, lat: 0, ip: "", sec: 0, err: true, source: 0, timezone: "" };

const msgs = {
  support: "Access Denied! Browser must support geolocation.",
  latlon: "Access Denied! Latitude/Longitude invalid.",
  device: "Sorry, this device/broswer is not supported.",
  security: "Access Denied! Browser Settings > Privacy and Security > Location to allow access.",
  geoloc: "Ly9nZW9sb2NhdGlvbi1kYi5jb20vanNvbi83YTliMWI2MC02Y2Q2LTExZWQtYTVjNy0xMTA0Njg3NTYwYTk=",
  extreme: "Ly9leHRyZW1lLWlwLWxvb2t1cC5jb20vanNvbi8/a2V5PVE5Q0dnTFFNRnpkSlBqcTlwOWtV",
  denied: "User denied the request for geolocation.",
  unavailable: "Location information is unavailable.",
  timeout: "The request to get user location timed out.",
  unknown: "An unknown error occurred."
};

window.onload = () => {   
  const geoString = sessionStorage.getItem('geo');
  if (geoString) {
      geo = JSON.parse(geoString);
  }

  isWebAuthnAvailable().then(isAvailable => {
    if (isAvailable) {
      //enable button
      document.getElementById("btn").disabled = false;
      document.getElementById("btn").style.display = "block"; // Show the accept button
    } else {
      showMsg(msgs.device, true); //Give we are sorry message
    } 
  });
}

// We store GEO in session storage so it can be used by the next few pages
// which might be login or registration or home.
// We may not have the user ID yet, so it cannot be stored in the DB yet
// If last time we were here is within 15 mins, don't get the geo data again
function doAccept() {
  document.getElementById("btn").disabled = true;
  document.getElementById("btn").setAttribute("aria-busy", "true"); //Show spinner
  const utcSec = Math.floor(new Date().getTime() / 1000);
  if (utcSec - geo.sec > 900) {
    tryGeoBrowser();
  } else {
    geoCheck();
  }
  return false;
}


//Most accurate and must be this if user is geofenced
function tryGeoBrowser() {
    if (!navigator.geolocation) {
      geo.err = true;
      toast(msgs.support, "error");
      console.error("Geolocation is not supported by this browser.");
      setTimeout(tryGeoSource2, 111); //Try alternative to getting location
      return;
    }
    // Geolocation is supported
    navigator.geolocation.getCurrentPosition(
        // Success callback
        (position) => {
            geo.lat = position.coords.latitude;
            geo.lon =  position.coords.longitude;
            geo.ip = ""        
            geo.err = false;
            geo.source = 1;
            geo.timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
            if (!geoCheck()) {       
                geo.err = true;
                geo.source = 0;
                setTimeout(() => tryGeoSource2(), 111); 
            }
        },
        // Error callback
        (error) => { 
          handleGeoError(error);
            setTimeout(() => tryGeoSource2(), 111); 
        }
    );
}

//geolocation-db
function tryGeoSource2() {
  fetchGeoData(atob(msgs.geoloc), 2, tryGeoSource3);
}

//extreme-ip-lookup
function tryGeoSource3() {
  fetchGeoData(atob(msgs.extreme), 3, () => toast(msgs.latlon));
}

function fetchGeoData(url, source, fallback) {
  fetch(location.protocol + url)
    .then(response => response.json())
    .then(data => {
      geo.lon = data.longitude || data.lon;
      geo.lat = data.latitude || data.lat;
      geo.ip = data.IPv4 || data.query;
      geo.err = false;
      geo.source = source;
      geo.timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
      if (!geoCheck()) {
        geo.err = true;
        geo.source = 0;
        fallback();
      }
    })
    .catch(error => {
      console.error(`Error fetching geolocation data from source ${source}: ${error.message}`);
      geo.err = true;
      fallback();
    });
}

// If a user reloads the page over and over, we don't want to do geofetch every time
// Just store the last set of good long/lat
// If user logs out, get a new set of long/lat when landing on this page
function geoCheck() {
  if (!isValidCoordinates()) {
    return false;
  }
  const now = new Date();
  geo.tzoff = now.getTimezoneOffset() * 60;
  geo.sec = Math.floor(now.getTime() / 1000);
  sessionStorage.setItem('geo', JSON.stringify(geo));

  const pageField = document.getElementById("nextpage");
  const page = pageField?.value?.trim();
  if (!page) {
      return true;
  }
  if (page.length > 0) {
      window.location.href =  encodeURI(page); // Redirect to the specified page
  }
  return true;
}

function isValidCoordinates() {
  return geo.lat !== 0 && geo.lon !== 0 && geo.lon <= 180 && geo.lon >= -180 && geo.lat <= 90 && geo.lat >= -90;
}

// Check if user verifying platform authenticator is available
async function isWebAuthnAvailable() {
    if (window.PublicKeyCredential) {
        try {
            const isUVPAA = await PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable();
            return isUVPAA;
        } catch (error) {
            console.error('Error checking for WebAuthn availability:', error);
            return false;
        }
    } else {
        return false;
    }
}

function handleGeoError(error) {
  let msg = msgs.security + " ";
  switch (error.code) {
    case error.PERMISSION_DENIED:
      msg += msgs.denied;
      break;
    case error.POSITION_UNAVAILABLE:
      msg += msgs.unavailable;
      break;
    case error.TIMEOUT:
      msg += msgs.timeout;
      break;
    default:
      msg += msgs.unknown;
      break;
  }
  geo.err = true;
  toast(msg, "error");
}

function toast(msg, type = "info") {
    Snackbar.push({
      message: msg,
      type: type
    });
}
