// iconbar.js

let btnSave = {
    id: "btnSave",
    state: "on",
    on: function () {
        const canSave = document.getElementById("canSave");
        if (canSave && canSave.value === "1") {
            setDisplay(document.getElementById(this.id), true);
            this.state = "on";
        }
    },
    off: function () {
        setDisplay(document.getElementById(this.id), false);
        this.state = "off";
    }
};

let btnNew = {
    id: "btnNew",
    state: "on",
    on: function () {
        const canNew = document.getElementById("canNew");
        if (canNew && canNew.value === "1") {
            setDisplay(document.getElementById(this.id), true);
            this.state = "on";
        }
    },
    off: function () {
        setDisplay(document.getElementById(this.id), false);
        this.state = "off";
    }
};

let btnDelete = {
    id: "btnDelete",
    state: "on",
    on: function () {
        const canDelete = document.getElementById("canDelete");
        if (canDelete && canDelete.value === "1") {
            setDisplay(document.getElementById(this.id), true);
            this.state = "on";
        } else {
            this.off();
        }
    },
    off: function () {
            setDisplay(document.getElementById(this.id), false);
        this.state = "off";
    }
};
