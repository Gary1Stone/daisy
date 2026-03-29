/* global Metro */

// iconbar.js

let btnSave = {
    id: "btnSave",
    state: "on",
    on: function () {
        if ($("#canSave").val() === "1") {
            $("#btnSave").show();
            this.state = "on";
        }
    },
    off: function () {
        $("#btnSave").hide();
        this.state = "off";
    }
};

let btnNew = {
    id: "btnNew",
    state: "on",
    on: function () {
        if ($("#canNew").val() === "1") {
            $("#btnNew").show();
            this.state = "on";
        }
    },
    off: function () {
        $("#btnNew").hide();
        this.state = "off";
    }
};

let btnDelete = {
    id: "btnDelete",
    state: "on",
    on: function () {
        if ($("#canDelete").val() === "1") {
            $("#btnDelete").show();
            this.state = "on";
        } else {
            this.off();
        }
    },
    off: function () {
            $("#btnDelete").hide();
        this.state = "off";
    }
};
