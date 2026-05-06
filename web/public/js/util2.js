// util2.js

function txt2Int(value) {
    const result = parseInt(value, 10);
    return isNaN(result) ? 0 : result;
}
