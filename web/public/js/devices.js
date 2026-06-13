//devices.js

let page = 0;
let endReached = false; // Flag to indicate if all data has been loaded

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
    const dialog = document.getElementById("filterDialog");
    const dialogMsg = document.getElementById("dialogMsg");
    const cards = document.getElementById("cards");
      if (response.length < 10) {
        if (dialogMsg) dialogMsg.innerHTML = "No device matches found!";
      } else {
        if (dialogMsg) dialogMsg.innerHTML = "";
        if (cards) cards.innerHTML = response;
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
    islate: (document.getElementById("isLate").checked) ? true : false,
    ismissing: (document.getElementById("isMissing").checked) ? true : false
  }
  return sendData;
}
