function switchTheme(el) { document.documentElement.setAttribute('data-theme', el.value) }
    
function switchCSS(cssid, el){ document.getElementById(cssid).href = el.value; }

function addFontSize(addPx){
    html = document.querySelector('html');
    currentSize = parseFloat(window.getComputedStyle(html, null)
        .getPropertyValue('font-size'));
    html.style.fontSize = (currentSize + addPx) + 'px';
}

function toggleDarkMode(el){
    var theme=''
    if (el.innerText == '☪'){
        el.innerText = '☀'; theme='dark';
    } else {
        el.innerText = '☪'; theme='light';
    }
    document.documentElement.setAttribute('data-theme', theme)
    fetch("/api/v1?theme="+theme)
}

var currH2 = 0;//current visible doc header

window.addEventListener('load', function() {
    // update current visible doc
    h2s = document.getElementsByTagName('h2')
    for (i = 0; i < h2s.length; i++) {
        if (isVisible(h2s[i])) {
            currH2 = i;
            //console.log(currH2,h2s[i]);
            break;
        }
    }
})

function isVisible(elem) {
    if (!(elem instanceof Element)) throw Error('DomUtil: elem is not an element.');
    const style = getComputedStyle(elem);
    if (style.display === 'none') return false;
    if (style.visibility !== 'visible') return false;
    if (style.opacity < 0.1) return false;
    if (elem.offsetWidth + elem.offsetHeight + elem.getBoundingClientRect().height +
        elem.getBoundingClientRect().width === 0) {
        return false;
    }
    const elemCenter   = {
        x: elem.getBoundingClientRect().left + elem.offsetWidth / 2,
        y: elem.getBoundingClientRect().top + elem.offsetHeight / 2
    };
    if (elemCenter.x < 0) return false;
    if (elemCenter.x > (document.documentElement.clientWidth || window.innerWidth)) return false;
    if (elemCenter.y < 0) return false;
    if (elemCenter.y > (document.documentElement.clientHeight || window.innerHeight)) return false;
    let pointContainer = document.elementFromPoint(elemCenter.x, elemCenter.y);
    do {
        if (pointContainer === elem) return true;
    } while (pointContainer = pointContainer.parentNode);
    return false;
}

document.onkeypress = function (e) {
    var evt = window.event || e;
    switch (evt.keyCode) {
        case 32:
            currH2++;
            doc = document.getElementsByTagName('h2')[currH2];
            if (doc === undefined) {
                break;
            }
            evt.preventDefault();
            window.scrollTo({ top: doc.offsetTop-20, behavior: 'smooth'});
            break;
    }
}