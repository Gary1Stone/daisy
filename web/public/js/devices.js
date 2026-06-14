//devices.js

// Cache DOM elements using getters to ensure they are available when needed
const UI = {
    filterDialog: () => document.getElementById("filterDialog"),
    dialogMsg: () => document.getElementById("dialogMsg"),
    cards: () => document.getElementById("cards"),
    type: () => document.getElementById("type"),
    siteSearch: () => document.getElementById("siteSearch"),
    officeSearch: () => document.getElementById("officeSearch"),
    groupSearch: () => document.getElementById("groupSearch"),
    userSearch: () => document.getElementById("userSearch"),
    textSearch: () => document.getElementById("textSearch"),
    isLate: () => document.getElementById("isLate"),
    isMissing: () => document.getElementById("isMissing"),
    wizcid: () => document.getElementById("wizcid"),
    wizardDialog: () => document.getElementById("wizardDialog"),
    deviceName: () => document.getElementById("deviceName"),
    removePick: () => document.getElementById("removePick"),
    backupPick: () => document.getElementById("backupPick")
};

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
  const modal = UI.filterDialog();
  if (modal) modal.showModal();
}

function search() {
  if (pageLoading) return;
  page = 0;
  const sendData = getFormData();
  sendData.task = "get_first_page";
  postJSON("devices", sendData).then(response => {
    const dialog = UI.filterDialog();
    const dialogMsg = UI.dialogMsg();
    const cards = UI.cards();
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
    gid: txt2Int(UI.groupSearch().value),
    uid: txt2Int(UI.userSearch().value),
    site: UI.siteSearch().value,
    office: UI.officeSearch().value,
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
          UI.cards().insertAdjacentHTML('beforeend', response);
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
  const modal = UI.wizardDialog();
  if (modal) modal.showModal();
  UI.wizcid().value = cid;
  UI.deviceName().value = devName;
  if (devType.toUpperCase() === "DESKTOP" || devType.toUpperCase() === "LAPTOP") {
    setDisplay(UI.removePick(), true);
    setDisplay(UI.backupPick(), true);
  } else {
    setDisplay(UI.removePick(), false);
    setDisplay(UI.backupPick(), false);
  }
}

function showWizard(wiz) {
  window.location.href = encodeURI("wizard.html?wizkey=" + wiz + "&cid=" +  UI.wizcid().value);
}

function getFormData() {
  let sendData = { 
    task: "get_first_page", 
    page: page, // 'page' is a global variable
    cid: 0,
    devtype:  UI.type().value,
    site: UI.siteSearch().value,
    office: UI.officeSearch().value,
    gid : txt2Int(UI.groupSearch().value),
    uid: txt2Int(UI.userSearch().value),
    searchtxt: UI.textSearch().value,
    islate: UI.isLate().checked,
    ismissing: UI.isMissing().checked
  }
  return sendData;
}
