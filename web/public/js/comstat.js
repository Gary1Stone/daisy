
let btnNew;

document.addEventListener('DOMContentLoaded', function() {
  btnNew = new Button("btnNew");
});

function addRecord(id = 0) {
  window.location.href = encodeURI("device.html?cid=" + txt2Int(id));
}
