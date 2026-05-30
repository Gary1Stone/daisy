//devices.js

let page = 0;
let endReached = false; // Flag to indicate if all data has been loaded

// When the page is finished loading
document.addEventListener("DOMContentLoaded", function() {
  initCustomSelects();
});

// Asynchronously posts data to the server and updates the target element
async function postData(task, target) {
  if (pageLoading) return; // Prevent multiple requests
  try {
      if (task === "get_first_page") {
          page = 0;
          endReached = false;
          const targetElement = document.getElementById(target);
          if (targetElement) targetElement.innerHTML = "";
      }
      const sendData = getFormData();
      sendData.task = task;

      const response = await postJSON("devices", sendData);
      document.getElementById(target).innerHTML = response;
  } catch (error) {
    console.error("Error while posting data:", error);
  }
}

// Add a new record
function addRecord() {
  window.location.href = 'device.html?cid=0';
}

// Search Filters show
function popFilters() {
  const modal = document.getElementById("filterDialog");
  if (modal) modal.showModal();
}

function search() {
  if (pageLoading) return;
  page = 0;
  const sendData = getFormData();
  sendData.task = "get_first_page";
  postJSON("devices", sendData).then(response => {
      document.getElementById("cards").innerHTML = response;
      if (response.length < 10) {
          toast("No device matches found!", "alert");
      } else {
        const dialog = document.getElementById("filterDialog");
        if (dialog) dialog.close();
      }
  });
}


/**************************************************/
/*         Droplist handling                      */
/**************************************************/
function getUserSearchCtrl() {
  getCtrl("selectSearchUser", ctrlData("USERSEARCH"));
}

function getOfficeSearchCtrl() {
  getCtrl("selectOffice", ctrlData("OFFICESEARCH"));
}

function ctrlData(task) {
  const droplistRequest = { 
    task: task, 
    isTicket: false,
    isWizard: false,
    cid: 0,
    gid: txt2Int(document.getElementById("groupSearch").value),
    uid: txt2Int(document.getElementById("userSearch").value),
    site: document.getElementById("siteSearch").value,
    office: document.getElementById("officeSearch").value,
    impact: "",
    trouble: "",
    wizard: "", // This field is not used in this context
    type: "",
    inform_gid: 0,
    isReadonly: false,
  }
  return droplistRequest;
}
/**************************************************/

// Check if the user has scrolled to the bottom of the page
window.addEventListener('scroll', function() {
  if (window.pageYOffset + window.innerHeight >= document.documentElement.scrollHeight - 200) {
    if (!endReached && !pageLoading) {
      pageLoading = true;
      page++;
      getNextBlock();
    }
  }
});

// Fetches the next block of alerts and appends it to the existing content
async function getNextBlock() {
  try {
      const sendData = getFormData();
      sendData.task = "get_next_page";
      const response = await postJSON("devices", sendData);
      if (response.length === 0) {
          endReached = true;
      } else {
          document.getElementById("cards").insertAdjacentHTML('beforeend', response);
      } 
      // Wait 1/2 second before getting next block
      setTimeout(() => {
          pageLoading = false;
      }, 500);
  } catch (error) {
    pageLoading = false;
    console.error("Error while posting data:", error);
  }
}


function seeIconClick() { // This function seems to be related to a button with ID "btnSeeState"
  let curState = document.getElementById("btnSeeState").value;
  const devType = document.getElementById("type").value;
  if (!(devType === "DESKTOP" || devType === "LAPTOP" || devType === "") && curState === "see") {
    curState = "late";
  }
  if (curState === "off") { // Starting state is off, switching to See activated
    document.getElementById("btnSeeState").value = "see";
    document.getElementById("mif-eye").classList.remove("fg-white"); document.getElementById("mif-eye").classList.add("fg-red");
    document.getElementById("ismissing").value = "1";
    document.getElementById("islate").value = "0";
  } else if (curState === "see") {  // See activated, switching to late backups
    document.getElementById("btnSeeState").value = "late";
    document.getElementById("mif-copy").style.display = "";
    document.getElementById("mif-eye").style.display = "none";
    document.getElementById("ismissing").value = "0";
    document.getElementById("islate").value = "1";
  } else if (curState === "late") {    // Late activated, switching off
    document.getElementById("btnSeeState").value = "off";
    document.getElementById("mif-copy").style.display = "none";
    document.getElementById("mif-eye").style.display = "";
    document.getElementById("mif-eye").classList.remove("fg-red"); document.getElementById("mif-eye").classList.add("fg-white");
    document.getElementById("islate").value = "0";
    document.getElementById("ismissing").value = "0";
  }
  postData("get_first_page", "cards");
}

function popWizards(cid, devName, devType) {
  if (cid < 1) return
  const modal = document.getElementById("wizardDialog");
  if (modal) modal.showModal();
  document.getElementById("wizcid").value = cid;
  document.getElementById("deviceName").value = devName;
  if (devType.toUpperCase() === "DESKTOP" || devType.toUpperCase() === "LAPTOP") {
    document.getElementById("removePick").style.display = "block";
    document.getElementById("backupPick").style.display = "block";
  } else {
    document.getElementById("removePick").style.display = "none";
    document.getElementById("backupPick").style.display = "none";
  }
}

function showWizard(wiz) {
  window.location.href = encodeURI("wizard.html?wizkey=" + wiz + "&cid=" +  document.getElementById("wizcid").value);
}

function getFormData() {
  let sendData = { 
    task: "get_first_page", 
    page: page, // 'page' is a global variable
    cid: 0,
    devtype:  document.getElementById("type").value,
    site: document.getElementById("siteSearch").value,
    office: document.getElementById("officeSearch").value,
    gid : txt2Int(document.getElementById("groupSearch").value),
    uid: txt2Int(document.getElementById("userSearch").value),
    searchtxt: document.getElementById("textSearch").value,
    islate: false,
    ismissing: false
  }
  return sendData;
}
