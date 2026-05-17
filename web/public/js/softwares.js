// softwares.js

document.addEventListener('DOMContentLoaded', function() {
  btnNew = new Button("btnNew");
  buildTable('softwaretable');
});

function addRecord(id = 0) {
  window.location.href = encodeURI("software.html?sid=" + txt2Int(id));
}
