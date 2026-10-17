// ticket.js

function addRecord(id = 0) {
  window.location.href = encodeURI("ticket.html?aid=" + id);
}
