// Concept: Have all pages aligned in a single row on a slider.
// Animate the slider back and forth depending on page number.
// Squish vertically all non-displayed pages so they take up the 
// space horizontally (set to screen width), but not vertically.
//
// After animation completes (1/2 sec), 
// set previous page to squish down (css class: offscreen) and 
// set the global variable (curPage) to the new current page.

let curPage = 0;    // Current displayed page
let conv2 = { cad: 1.37, vnd: 25450.00, khr: 4090.26, lak: 21469.95 }; // Currency Conversion Rates
let startX;         // Touch start X axis
let startY;         // Touch start Y axis
const nFormat = new Intl.NumberFormat(undefined, {minimumFractionDigits: 0}); // Add commas to currency
let country = localStorage.getItem("country") || "ca"; // Currency convert from
let dollars = { usd: 0.0, vnd: 0.0, khr: 0.0, lak: 0.0, cad: 0.0 }; // Amount converted to
const titles = [ "Indochina", "Departure", "Hanoi", "Hanoi", "Hanoi", "Hanoi", "Ha Long Bay", 
                "Hue", "Hue", "Hoi An", "Hoi An", "Ho Chi Minh", "Ho Chi Minh", "Phnom Penh", 
                "Phnom Penh", "Siem Reap", "Siem Reap", "Luang Prabang", "Lang Prabang", 
                "Lang Prabang", "Vientiane", "Vientiane", "Hanoi", "Flying", "Information", 
                "Search", "Hotels", "Currency"];
const url = "aHR0cHM6Ly92Ni5leGNoYW5nZXJhdGUtYXBpLmNvbS92Ni8zNmMzOGYyMWY5OTVjNWYyZWQwYzVjZGIvbGF0ZXN0L1VTRA==";
const pages = document.querySelectorAll('.page');
const slider = document.getElementById('slider');
const pageIds = ["3.1", "4.1", "4.2", "4.3", "4.4", "4.5", "4.6", "4.7", "5.1", "5.2", "5.3", 
        "5.4", "5.5", "5.6", "6.1", "6.2", "6.3", "6.4", "7.1", "7.2", "8.1", "8.2", "8.3", 
        "8.4", "8.5", "9.1", "9.2", "9.3", "9.4", "9.5", "9.6", "9.7", "9.8", "9.9", "9.10", 
        "10.1", "10.2", "10.3", "10.4", "11.1", "11.2", "11.3", "11.4", "11.5", "13.1", "12.1", 
        "12.2", "12.3", "12.4", "14.1", "14.2", "14.3", "14.4", "14.5", "15.1", "16.1", "16.2", 
        "16.3", "16.4", "16.5", "16.6", "17.1", "18.1", "18.2", "18.3", "18.4", "18.5", "18.6", 
        "18.7", "19.1", "19.2", "19.3", "20.1", "20.2", "20.3", "20.4", "21.1", "21.2", "21.3"];
const words = ["Z2FyeQ==", "YmFyYg==", "cGV0ZQ==", "bGl6"];

// Set the event listeners
window.onload = () => {
    //Navbar dropdown menu
    const dropdownItems = document.querySelectorAll('.dropdown-item');
    const dropdownContent = document.querySelector('.dropdown-content');
    const dropbtn = document.querySelector('.dropbtn');

    // Add event listeners to dropdown items to hide the dropdown content
    dropdownItems.forEach(item => {
        item.addEventListener('click', () => {
            dropdownContent.classList.remove('show');
        });
    });

    // Add an event listener to the dropbtn to toggle the dropdown content visibility
    dropbtn.addEventListener('click', () => {
        dropdownContent.classList.toggle('show');
    });

    // Hide the dropdown content if clicked outside
    document.addEventListener('click', (event) => {
        if (!event.target.closest('.dropdown')) {
            dropdownContent.classList.remove('show');
        }
    });

    const table = document.getElementById("searchTable");
    const rows = table.getElementsByTagName("tr");
    
    for (let i = 1; i < rows.length; i++) { // Start from 1 to skip the header row
        rows[i].addEventListener("click", function() {
            let selectedItem = pageIds[i-1];
            jumpTo(selectedItem);
        });
    }
    document.addEventListener('touchstart', handleTouchStart, false);
    document.addEventListener('touchmove', handleTouchMove, false);

    //Prompt for passcode on the document links
    const pdfs = document.querySelectorAll('.pdf');
    pdfs.forEach(link => {
        link.addEventListener('click', (event) => {
            event.preventDefault(); // Always prevent navigation initially
            let word = prompt('Passcode?');
            if (word) {
                word = btoa(word.toLowerCase());
                if (words.includes(word)) {
                    window.location.href =  encodeURI(link.href); // Navigate to the URL if the user confirms
                } else {
                    window.location.href =  encodeURI("https://en.wikipedia.org/wiki/Vietnam");
                }
            } else {
                window.location.href =  encodeURI("https://en.wikipedia.org/wiki/Vietnam");
            }
        });
    });

    // Wait 5 seconds before getting currency conversion rates
    setTimeout(() => {
        apiFetchCurrency();
        pickSource(country);
    }, 5000);
}

function showPage(newPage) {
    const totalPages = pages.length;
    newPage = (newPage + totalPages) % totalPages;
    if (newPage === curPage) return;
    pages[newPage].classList.remove('offscreen');
    let slide = (newPage * window.innerWidth) + "px";
    slider.style.transform = `translateX(-${slide})`;
    setTimeout((newerPage) => {
        pages[curPage].classList.add('offscreen');
        curPage = newerPage;
        const titleblock = document.getElementById("title");
        titleblock.innerHTML = titles[curPage];
    }, 100, newPage);
}

function nextPage() {
    showPage(curPage + 1);
}

function backPage() {
    showPage(curPage - 1);
}

function handleTouchStart(event) {
    startX = event.touches[0].clientX;
    startY = event.touches[0].clientY;
}

// Try to determin if the user is swiping
// left or right but not up or down
function handleTouchMove(event) {
    if (!startX && !startY) return;
    const currentX = event.touches[0].clientX;
    const currentY = event.touches[0].clientY;
    const diffX = startX - currentX;
    const diffY = Math.abs(startY - currentY);
    if (diffX > 100 && diffY < 50) {
        nextPage();
        startX = null;
        startY = null;
    } else if (diffX < -100 && diffY < 50) {
        backPage();
        startX = null;
        startY = null;
    }
}

//Search functionality:
//Note: have to wait for item's page to be rendered before scrolling to it
function jumpTo(selectedItem) {
    if (selectedItem.length < 3) return;
    const parts = selectedItem.split(".");
    const page = parseInt(parts[0]);
    showPage(page);
    setTimeout((selectedItem) => {
        const element = document.getElementById(selectedItem);
        if (element) {
            element.scrollIntoView({ behavior: 'smooth' });
        }
    }, 500, selectedItem);
}

//Search table
function mySeach() {
    const input = document.getElementById("searchInput");
    const filter = input.value.toUpperCase();
    const table = document.getElementById("searchTable");
    const rows = table.getElementsByTagName("tr");
    // Loop through all table rows, and hide those who don't match the search query
    for (let i = 0; i < rows.length; i++) {
        const td = rows[i].getElementsByTagName("td")[0];
        const city = rows[i].getElementsByTagName("td")[1];
        if (td) {
            const txtValue = td.textContent || td.innerText;
            const cityValue = city.textContent || city.innerText;
            if (txtValue.toUpperCase().indexOf(filter) > -1 || cityValue.toUpperCase().indexOf(filter) > -1) {
                rows[i].style.display = "";
            } else {
                rows[i].style.display = "none";
            }
        }
    }
}

// Currency Converter

// Called to pick the country currency to convert from
function pickSource(newCountry) {
  const fromFlag = document.getElementById("fromFlag");
  const fromCur = document.getElementById("fromCur");
  document.getElementById("us_cur").classList.remove('hide');
  document.getElementById("vn_cur").classList.remove('hide');
  document.getElementById("kh_cur").classList.remove('hide');
  document.getElementById("la_cur").classList.remove('hide');
  document.getElementById("ca_cur").classList.remove('hide');
    switch (newCountry) {
        case "us":
          fromFlag.src = "tour/us.svg";
          fromCur.innerText = "USD";
          document.getElementById("us_cur").classList.add('hide');
          break;
        case "ca":
          fromFlag.src = "tour/ca.svg";
          fromCur.innerText = "CAD";
          document.getElementById("ca_cur").classList.add('hide');
          break;
        case "vn":
          fromFlag.src = "tour/vn.svg";
          fromCur.innerText = "Dong";
          document.getElementById("vn_cur").classList.add('hide');
          break;
        case "kh":
          fromFlag.src = "tour/kh.svg";
          fromCur.innerText = "Riel";
          document.getElementById("kh_cur").classList.add('hide');
          break;
        case "la":
          fromFlag.src = "tour/la.svg";
          fromCur.innerText = "Kip";
          document.getElementById("la_cur").classList.add('hide');
          break;
    }
  country = newCountry;
  localStorage.setItem("country", country);
  key("");
}

//When the user presses a key, add it to the display
function key(digit) {
    const display = document.getElementById('display');
    let str = display.textContent;
    if (str.length > 16) return;
    if (digit.length === 0) {
        str = str.replace(/,/g, '');
    } else if (str === "0") {
        display.textContent = digit;
        str = digit;
    } else {      
      str = str.replace(/,/g, '');
      str += digit;
      display.textContent = nFormat.format(str);
    }
    doConversion(str);
}

//When the user presses the backspace key
function backspace() {
    const display = document.getElementById('display');
    let str = display.textContent;
    str = str.replace(/,/g, '');
    str = str.slice(0, -1);
    if (str === "") {
        str = "0";
    }
    display.textContent = nFormat.format(str);
    doConversion(str);
}

//When the user presses the clear display key
function clearDisplay() {
    document.getElementById('display').textContent = '0';
    doConversion('0');
}

// Calculate the Currency Conversion 
function doConversion(amount) {
  let cash = parseFloat(amount) || 0.0
    switch (country) {
        case "us":
            usd = cash;
            break;
        case "ca":
            usd = cash / conv2.cad;
            break;
        case "vn":
            usd = cash / conv2.vnd;
            break;
        case "kh":
            usd = cash / conv2.khr;
            break;
        case "la":
            usd = cash / conv2.lak;
            break;
    }
    dollars.usd = Number(usd).toFixed();
    dollars.cad = Number(usd * conv2.cad).toFixed();
    dollars.vnd = Number(usd * conv2.vnd).toFixed();
    dollars.khr = Number(usd * conv2.khr).toFixed();
    dollars.lak = Number(usd * conv2.lak).toFixed();
    document.getElementById("us").innerText = nFormat.format(dollars.usd);
    document.getElementById("ca").innerText = nFormat.format(dollars.cad);
    document.getElementById("vn").innerText = nFormat.format(dollars.vnd);
    document.getElementById("kh").innerText = nFormat.format(dollars.khr);
    document.getElementById("la").innerText = nFormat.format(dollars.lak);
}

function apiFetchCurrency() {
    const now = Math.floor(Date.now() / 1000);
    const lastSaved = parseInt(localStorage.getItem("saved")) || 0;
    conv2.cad = parseFloat(localStorage.getItem("usd2cad")) || conv2.cad;
    conv2.vnd = parseFloat(localStorage.getItem("usd2vnd")) || conv2.vnd;
    conv2.khr = parseFloat(localStorage.getItem("usd2khr")) || conv2.khr;
    conv2.lak = parseFloat(localStorage.getItem("usd2lak")) || conv2.lak;
    // If more than 12 hours elapsed, then fetch new values
    if (now - lastSaved > 43200) { 
        fetch(atob(url))
        .then(response => {
            if (!response.ok) {    // response not in 200 range
                throw new Error("cannot fetch new exchange rates");
            }
            return response.json();
        }).then(msg => {
            conv2.cad = msg.conversion_rates.CAD > 0.0 ? msg.conversion_rates.CAD : conv2.cad;
            conv2.vnd = msg.conversion_rates.VND > 0.0 ? msg.conversion_rates.VND : conv2.vnd;
            conv2.khr = msg.conversion_rates.KHR > 0.0 ? msg.conversion_rates.KHR : conv2.khr;
            conv2.lak = msg.conversion_rates.LAK > 0.0 ? msg.conversion_rates.LAK : conv2.lak;
            // Save for later
            localStorage.setItem("usd2cad", conv2.cad);
            localStorage.setItem("usd2vnd", conv2.vnd);
            localStorage.setItem("usd2khr", conv2.khr);
            localStorage.setItem("usd2lak", conv2.lak);
            localStorage.setItem("saved", now);
        }).catch(err => {
            console.warn('Error in the exchange rates API call.', err);
        });
    }
};
