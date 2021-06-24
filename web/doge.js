function switchTheme(el) { document.documentElement.setAttribute('data-theme', el.value) }
    
function switchCSS(cssid, el){ document.getElementById(cssid).href = el.value; }

function addFontSize(addPx){
html = document.querySelector('html');
currentSize = parseFloat(window.getComputedStyle(html, null)
    .getPropertyValue('font-size'));
html.style.fontSize = (currentSize + addPx) + 'px';
}

function toggleDarkMode(el){
    var theme='light'
    if (el.innerText == '☪'){
        el.innerText = '☀'; theme='dark';
    } else {
        el.innerText = '☪'; theme='light';
    }
    document.documentElement.setAttribute('data-theme', theme)
}