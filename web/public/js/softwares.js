// softwares.js

let btnNew;

document.addEventListener('DOMContentLoaded', function() {
  btnNew = new Button("btnNew");
});

function addRecord(id = 0) {
  window.location.href = encodeURI("software.html?sid=" + txt2Int(id));
}

// Send ajax request to add unwanted software and replace the unwanted table
function addUnwanted(id, name) {
  if (!id || !name) return
  const cell = document.getElementById(id);
  if (cell) {
    cell.innerHTML = "👎";
  }
  htmx("/unwanted/add", {id: 0, name: name}, "unwantedDiv");
}

// Send ajax request to delete unwanted software and replace the unwanted table
function deleteUnwanted(id) {
  if (!id) return
  htmx("/unwanted/delete", {id: id, name: ""}, "unwantedDiv");
}
