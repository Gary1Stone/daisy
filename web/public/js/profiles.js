// Profiles.js

let btnNew;

document.addEventListener('DOMContentLoaded', function() {
  btnNew = new Button("btnNew");
});

function addRecord(id = 0) {
  window.location.href = encodeURI("profile.html?uid=" + txt2Int(id));
}
