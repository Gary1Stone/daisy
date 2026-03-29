/* global Metro */

let page = 0;
let endReached = false;


// Asynchronously posts data to the server and updates the target element
async function postData(task, target) {
  if (pageLoading) return;
  try {
      if (task === "get_first_page") {
          page = 0;
          endReached = false;
          $("#" + target).html("");
      }
      const sendData = getFormData();
      sendData.task = task;

      const response = await $.post("devices", sendData);
      $("#" + target).html(response);
  } catch (error) {
    console.error("Error while posting data:", error);
  }
}

// Add a new record
function addRecord() {
  location.href='device.html?cid=0';
}

// Search Filters show
function popFilters() {
  Metro.dialog.open("#searchDialog");
}

function search() {
  if (pageLoading) return;
  page = 0;
  sendData = getFormData("get_first_page");
  $.post("devices", sendData).then((response) => {
      $("#cards").html(response);
      let msg = "";
      if (response.length < 10) {
          msg = "No device matches found!";
          Metro.dialog.open("#searchDialog");
      }
      $("#searchError").html(msg);
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
  let droplistRequest = { 
    task: task, 
    isTicket: false,
    isWizard: false,
    cid: 0,
    gid: txt2Int($("#groupSearch").val()),
    uid: txt2Int($("#userSearch").val()),
    site: $("#siteSearch").val(),
    office: $("#officeSearch").val(),
    impact: "",
    trouble: "",
    wizard: "",
    type: "",
    inform_gid: 0,
    isReadonly: false,
  }
  return droplistRequest;
}
/**************************************************/

// Check if the user has scrolled to the bottom of the page
$(window).scroll(function() {
  if ($(window).scrollTop() + $(window).height() >= $(document).height() - 200) {
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
      const response = await $.post("devices", sendData);
      if (response.length === 0) {
          endReached = true;
      } else {
          $("#cards").append(response);
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


function seeIconClick() {
  let curState = $(btnSeeState).val();
  const devType = $("#type").val();
  if (!(devType === "DESKTOP" || devType === "LAPTOP" || devType === "") && curState === "see") {
    curState = "late";
  }
  if (curState === "off") { // Starting state is off, switching to See activated
    $("#btnSeeState").val("see");
    $("#mif-eye").removeClass("fg-white").addClass("fg-red");
    $("#ismissing").val("1");
    $("#islate").val("0");
  } else if (curState === "see") {  // See activated, switching to late backups
    $("#btnSeeState").val("late");
    $("#mif-copy").show();
    $("#mif-eye").hide();
    $("#ismissing").val("0");
    $("#islate").val("1");
  } else if (curState === "late") {    // Late activated, switching off
    $("#btnSeeState").val("off");
    $("#mif-copy").hide();
    $("#mif-eye").show();
    $("#mif-eye").removeClass("fg-red").addClass("fg-white");
    $("#islate").val("0");
    $("#ismissing").val("0");
  }
  postData("get_first_page", "cards");
}

function popWizards(cid, devName, devType) {
  if (cid < 1) return
  Metro.dialog.open('#chooseWizard');
  $("#wizcid").val(cid);
  $("#deviceName").val(devName);
  if (devType.toUpperCase() === "DESKTOP" || devType.toUpperCase() === "LAPTOP") {
    $("#installPick").show();
    $("#removePick").show();
    $("#backupPick").show();
  } else {
    $("#installPick").hide();
    $("#removePick").hide();
    $("#backupPick").hide();
  }
}

function showWizard(wiz) {
  window.location.href = encodeURI("wizard.html?wizkey=" + wiz + "&cid=" +  $("#wizcid").val());
}

function getFormData() {
  let sendData = { 
    task: "get_first_page", 
    page: page,
    cid: 0,
    devtype:  $("#type").val(),
    site: $("#siteSearch").val(),
    office: $("#officeSearch").val(),
    gid : txt2Int($("#groupSearch").val()),
    uid: txt2Int($("#userSearch").val()),
    searchtxt: $("#txtSearch").val(),
    islate: $("#islate").val() === "1" ? true: false,
    ismissing: $("#ismissing").val() === "1" ? true: false
  }
  return sendData;
}
