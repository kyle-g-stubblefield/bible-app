var allowHighlighting = false;
var buttonClicked = "";
var hightlightColor = "";
const highlightColors = [
    "bg-red-600",
    "bg-orange-600",
    "bg-amber-600",
    "bg-yellow-600",
    "bg-lime-600",
    "bg-green-600",
    "bg-emerald-600",
    "bg-teal-600",
    "bg-cyan-600",
    "bg-sky-600",
    "bg-rose-600",
    "bg-indigo-600",
    "bg-violet-600",
    "bg-fuchsia-600",
    "bg-pink-600",
    "bg-blue-600"
];

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
            if (allowHighlighting) {
                if (hasClass(wrapper, hightlightColor)) {
                    removeClass(wrapper, hightlightColor);
                    //wrapper.children.forEach(child => {
                    //    if(hasClass)
                    //})
                } else {
                    highlightColors.forEach((color) => {
                        removeClass(wrapper, color);
                    });
                    addClass(wrapper, hightlightColor);
                }
            }
        });
    });
}

function HighlightButtonClicked(me) {
    //console.log(me.id);
    var buttons = document.querySelectorAll(".highlight-button");
    buttons.forEach(button => {
        removeClass(button, "border-stone-200");
        addClass(button, "border-transparent");
    });
    if (allowHighlighting && me.id === buttonClicked) {
        allowHighlighting = false;
    } else {
        removeClass(me, "border-transparent");
        addClass(me, "border-stone-200");
        buttonClicked = me.id;
        allowHighlighting = true;
        hightlightColor = "bg-" + me.id.split("-")[1] + "-600";
        //console.log(hightlightColor);

    }
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