// Inspiration for path drawing rather than pixel shading comes from
// http://jsfiddle.net/m1erickson/AEYYq/ , as does the code for undo.

var canvas = document.getElementById("imgCanvas");
var context = canvas.getContext("2d");
context.globalAlpha = 1.0;
context.imageSmoothingEnabled = false;
context.fillStyle = "rgba(0, 0, 0, 1)";

// Brush variables
var brush = "stroke";
var brushSize = 1;
var brushColor = "#000000";

// Undo
var lastX;
var lastY;
var points = [];

// Helpers
function getMousePos(canvas, evt) {
    const newLocal = canvas.getBoundingClientRect();
    var rect = newLocal;
    return {
        x: (evt.clientX - rect.left) / (rect.right - rect.left) * canvas.width,
        y: (evt.clientY - rect.top) / (rect.bottom - rect.top) * canvas.height
    };
}

function setBrushSize(size) {
    brush = "stroke";
    brushSize = size;
}

function setBrush(newBrush) {
    brush = newBrush;
}

// Handlers for the position of the mouse, and whether or not the mouse button
// has been held down.
var mouseDown = false;
function stop(e) {
    fullyShade();

    var pos = getMousePos(canvas, e);

    points.push({
        x: pos.x,
        y: pos.y,
        brush: brush,
        size: brushSize,
        color: brushColor,
        mode: "end"
    });
    lastX = pos.x;
    lastY = pos.y;

    mouseDown = false;
}

function start(e) {
    var pos = getMousePos(canvas, e);

    lastX = pos.x;
    lastY = pos.y;

    if(brush == "fill") {
        console.log("Fill");

        const newLocal = canvas.getBoundingClientRect();
        var rect = newLocal;

        floodFill({r: 0x0, g: 0x0, b: 0x0, a: 0xff}, pos.x, pos.y);

        return false
    }

    context.beginPath();
    
    // Ensure that brush vars are set
    if(context.lineWidth != brushSize) {
        context.lineWidth = brushSize;
    }

    context.moveTo(pos.x, pos.y);
    
    points.push({
        x: pos.x,
        y: pos.y,
        brush: brush,
        size: brushSize,
        color: brushColor,
        mode: "begin"
    });

    mouseDown = true;
}

function draw(e) {
    if(!mouseDown) {
        return
    }

    var pos = getMousePos(canvas, e);

    //context.fillStyle = "#000000";

    context.lineTo(pos.x, pos.y);
    context.stroke();

    // command pattern stuff
    points.push({
        x: pos.x,
        y: pos.y,
        brush: brush,
        size: brushSize,
        color: brushColor,
        mode: "draw"
    });
    lastX = pos.x;
    lastY = pos.y;
}

canvas.addEventListener('mousemove', draw, false);
canvas.addEventListener('mouseout', stop, false);
canvas.addEventListener('mouseup', stop, false);
canvas.addEventListener('mousedown', start, false);

function saveCanvas() {
    var base64Image = CanvasToBMP.toDataURL(document.getElementById('imgCanvas'));

    document.getElementById("imgBase64").value = base64Image;
    document.getElementById("saveImage").submit();
}

function ajaxSaveCanvas(imageIndex) {
    var base64Image = CanvasToBMP.toDataURL(document.getElementById('imgCanvas'));

    $.ajax({
        type: "POST",
        url: "/traceoverlay/" + imageIndex,
        data: { 
            imgBase64: base64Image
        }
    }).done(function(o) {
        console.log(base64Image);
        console.log('saved'); 
    });
}

function downloadCanvas() {

    var link = document.createElement('a');
    link.download = 'canvas.bmp';
    link.href = CanvasToBMP.toBlob(document.getElementById('imgCanvas'));
    link.click();

    return
}

function fullyShade() {
    var imageData = context.getImageData(0,0,canvas.width, canvas.height);
    var pixels = imageData.data;
    var numPixels = pixels.length;

    context.clearRect(0, 0, canvas.width, canvas.height);

    for (var i = 0; i < numPixels; i++) {
        if (pixels[i*4+3] <= 1) {
            pixels[i*4+3] = 0;
        } else {
            pixels[i*4+3] = 255;
        }
    }
    context.putImageData(imageData, 0, 0);
}

function redrawAll() {

    if (points.length == 0) {
        return;
    }

    context.clearRect(0, 0, canvas.width, canvas.height);

    for (var i = 0; i < points.length; i++) {

        var pt = points[i];

        var begin = false;

        if (context.lineWidth != pt.size) {
            context.lineWidth = pt.size;
            begin = true;
        }
        if (context.strokeStyle != pt.color) {
            context.strokeStyle = pt.color;
            begin = true;
        }
        if (pt.mode == "begin" || begin) {
            context.beginPath();
            context.moveTo(pt.x, pt.y);
        }
        context.lineTo(pt.x, pt.y);
        if (pt.mode == "end" || (i == points.length - 1)) {
            context.stroke();
        }
    }
    context.stroke();
}

var interval;
document.getElementById("undo").addEventListener('mouseout', undoStop, false);
document.getElementById("undo").addEventListener('mouseup', undoStop, false);
document.getElementById("undo").addEventListener('mousedown', undoStart, false);

function undoStop() {
    clearInterval(interval);
}

function undoStart() {
    interval = setInterval(undoLast, 20);
}

function undoLast() {
    points.pop();
    redrawAll();
}