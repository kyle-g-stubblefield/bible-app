
var themeToggleDarkIcon = document.getElementById('theme-toggle-dark-icon');
var themeToggleLightIcon = document.getElementById('theme-toggle-light-icon');

// Change the icons inside the button based on previous settings
if (localStorage.getItem('color-theme') === 'dark' || (!('color-theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    themeToggleLightIcon.classList.remove('hidden');
    document.getElementById("theme-toggle").checked = true;
} else {
    themeToggleDarkIcon.classList.remove('hidden');
    document.getElementById("theme-toggle").checked = false;
}

var themeToggleBtn = document.getElementById('theme-toggle');

themeToggleBtn.addEventListener('click', function() {

    // toggle icons inside button
    themeToggleDarkIcon.classList.toggle('hidden');
    themeToggleLightIcon.classList.toggle('hidden');

    // if set via local storage previously
    if (localStorage.getItem('color-theme')) {
        if (localStorage.getItem('color-theme') === 'light') {
            document.documentElement.classList.add('dark');
            localStorage.setItem('color-theme', 'dark');
        } else {
            document.documentElement.classList.remove('dark');
            localStorage.setItem('color-theme', 'light');
        }

    // if NOT set via local storage previously
    } else {
        if (document.documentElement.classList.contains('dark')) {
            document.documentElement.classList.remove('dark');
            localStorage.setItem('color-theme', 'light');
        } else {
            document.documentElement.classList.add('dark');
            localStorage.setItem('color-theme', 'dark');
        }
    }
    
});


const searchHistory = new Set();

function addHistoryItem(verse) {
    document.getElementById("history").innerHTML += "<div class=\"history-item\">" + 
    "<button class=\"cursor-pointer underline\" onclick=\"useHistory('" + verse + "')\" >" + verse + "</button>"
    + "</div>";
}

function createHistory() {
    document.getElementById("history").innerHTML = "";
    searchHistory.forEach(addHistoryItem);
}

function useHistory(verse) {
    document.getElementById("search").value = verse;
    verseLookup();
}

function clearSearchHistory() {
    document.getElementById('history').innerHTML = '';
    searchHistory.clear();
}

async function verseLookup() {
    var verse = document.getElementById("search").value;
    var headings = document.getElementById("headings").checked;
    var extras = document.getElementById("extras").checked;
    var numbers = document.getElementById("numbers").checked;
    var url = "/api?verse=" + verse + "&headings=" + headings + "&extras=" + extras + "&numbers=" + numbers;
    fetch(url)
        .then(response => response.json())
        .then(data => {
            document.getElementById("verse").innerHTML = data.passages.join("");
            searchHistory.add(data.query);
            createHistory();
            wrapText();
        });
    window.scrollTo(0,0);
}

var inputField = document.getElementById('search');
inputField.addEventListener('keydown', function(event) {
    if (event.key === 'Enter') {
        verseLookup();
    }
});

["headings", "extras", "numbers"].forEach(function(id) {
    document.getElementById(id).addEventListener('change', verseLookup);
});

function wrapText() {
    const anchors = document.querySelectorAll("a.va");
    anchors.forEach((anchor) => {
        const wrapper = document.createElement("span");
        wrapper.classList.add("verse");
        const verseId = anchor.getAttribute("rel"); // e.g., "v40001007"
        if (verseId) {
            wrapper.setAttribute("data-verse", verseId);
        }

        let current = anchor;
        const parent = anchor.parentNode;

        // Add the anchor itself to the wrapper
        wrapper.appendChild(current.cloneNode(true));
        let next = current.nextSibling;

        // Wrap content until the next anchor
        while (next && !(next.nodeType === 1 &&
            (next.matches("a.va")
                || next.classList.contains("verse-num")
                || next.tagName.toLowerCase() === "p"))) {
            const sibling = next.nextSibling;
            wrapper.appendChild(next);
            next = sibling;
        }
        // Insert the wrapper before the anchor, then remove the original
        parent.insertBefore(wrapper, anchor);
        anchor.remove();

        // Add click-to-highlight functionality
        wrapper.addEventListener("click", () => {
            wrapper.classList.toggle("highlighted");
            const id = wrapper.getAttribute("data-verse");
            if (wrapper.classList.contains("highlighted")) {
                console.log(`Highlighted: ${id}`);
                // You can send this ID to your backend
            } else {
                console.log(`Unhighlighted: ${id}`);
                // Remove it from your backend
            }
        });
    });
}


function hasClass(ele, cls) {
    return !!ele.className.match(new RegExp('(\\s|^)' + cls + '(\\s|$)'));
}

function addClass(ele, cls) {
    if (!hasClass(ele, cls)) ele.className += " " + cls;
}

function removeClass(ele, cls) {
    if (hasClass(ele, cls)) {
        var reg = new RegExp('(\\s|^)' + cls + '(\\s|$)');
        ele.className = ele.className.replace(reg, ' ');
    }
}

//Add event from js the keep the marup clean
function init() {
    document.getElementById("open-menu").addEventListener("click", toggleMenu);
    document.getElementById("body-overlay").addEventListener("click", toggleMenu);
}

//The actual fuction
function toggleMenu() {
    var ele = document.getElementsByTagName('body')[0];
    if (!hasClass(ele, "menu-open")) {
        addClass(ele, "menu-open");
    } else {
        removeClass(ele, "menu-open");
    }
}

//Prevent the function to run before the document is loaded
document.addEventListener('readystatechange', function() {
    if (document.readyState === "complete") {
        init();
    }
});