var WIDTH = '1028';
var HEIGHT = '768';
var XMLNS = 'http://www.w3.org/2000/svg';

function Stroke(svg, width, red, green, blue, alpha) {
    this.xs = [];
    this.ys = [];
    this.width = width;
    this.red = red;
    this.green = green;
    this.blue = blue;
    this.alpha = alpha;
    this.elem = document.createElementNS(XMLNS, 'polyline');
    this.elem.setAttribute('stroke', 'rgba('+red+','+green+','+blue+','+alpha+')');
    this.elem.setAttribute('stroke-width', width);
    this.elem.setAttribute('stroke-linecap', 'round');
    this.elem.setAttribute('fill', 'none');
    svg.appendChild(this.elem);
}

Stroke.prototype.move = function (x, y) {
    this.xs.push(x);
    this.ys.push(y);
    this.draw();
};

Stroke.prototype.setPoints = function(xs, ys) {
    this.xs = xs;
    this.ys = ys;
    this.draw();
}

Stroke.prototype.draw = function () {
    var points = '';
    for (var i = 0; i < this.xs.length; i++) {
        points += this.xs[i] + ',' + this.ys[i] + ' ';
    }
    this.elem.setAttribute('points', points);
};

Stroke.prototype.remove = function() {
    return this.elem.parentNode.removeChild(this.elem);
};

Stroke.prototype.send = function() {
    var it = this;
    var xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/stroke');
    xhr.responseType = 'json';
    xhr.onload = function() {
        myStrokes.push(xhr.response);
    };
    xhr.onerror = function() {
        it.remove();
    };
    xhr.send(JSON.stringify({
        width: this.width,
        red: this.red,
        green: this.green,
        blue: this.blue,
        alpha: this.alpha,
        xs: this.xs,
        ys: this.ys,
    }));
};

var ourStrokes = [];
var myStrokes = [];

var stroke = null;

document.addEventListener('DOMContentLoaded', function () {
    var canvas = document.querySelector('#canvas');
    canvas.setAttribute('style', 'width:' + WIDTH + 'px;height:' + HEIGHT + 'px;border:solid black 1px;');

    var svg = document.createElementNS(XMLNS, 'svg');
    svg.setAttribute('viewBox', '0 0 ' + WIDTH + ' ' + HEIGHT);
    svg.setAttribute('width', WIDTH);
    svg.setAttribute('height', HEIGHT);
    canvas.appendChild(svg);

    canvas.addEventListener('mousedown', function (e) {
        e.preventDefault();

        stroke = new Stroke(svg, 5, 128, 128, 128, 0.9);
        stroke.move(e.offsetX, e.offsetY);
    });

    canvas.addEventListener('mousemove', function (e) {
        if (stroke === null) return;
        e.preventDefault();

        stroke.move(e.offsetX, e.offsetY);
    });

    canvas.addEventListener('mouseup', function (e) {
        if (stroke === null) return;
        e.preventDefault();

        stroke.move(e.offsetX, e.offsetY);
        stroke.send();
        stroke = null;
    });
    
    var es = new EventSource('/api/events');
    es.onmessage = function(ev) {
        var data = JSON.parse(ev.data);
        // TODO: look at event type
        var stroke = new Stroke(svg,
            data.width, data.red, data.green, data.blue, data.alpha);
        stroke.setPoints(data.xs, data.ys);
    };
    es.onerror = function(err) {
        console.log(err);
    }
});
