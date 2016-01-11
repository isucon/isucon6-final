var WIDTH = '800';
var HEIGHT = '600';
var XMLNS = 'http://www.w3.org/2000/svg';

function Stroke(svg, width, red, green, blue, alpha) {
    this.xs = [];
    this.ys = [];
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

    var points = '';
    for (var i = 0; i < this.xs.length; i++) {
        points += this.xs[i] + ',' + this.ys[i] + ' ';
    }
    this.elem.setAttribute('points', points);
}

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
        stroke = null;
    });
});
