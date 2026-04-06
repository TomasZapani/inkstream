// Canvas principal y overlay de cursores
const canvas = document.getElementById('canvas');
const ctx = canvas.getContext('2d');
const cursorCanvas = document.getElementById('cursors');
const cursorCtx = cursorCanvas.getContext('2d');

// Toolbar
const colorPicker = document.getElementById('colorPicker');
const brushSize = document.getElementById('brushSize');
const clearBtn = document.getElementById('clearBtn');

// Estado local
let isDrawing = false;
let lastX = 0;
let lastY = 0;
const remoteCursors = new Map(); // clientId -> {x, y}
let cursorAnimFrame = null;

// Ajustar tamaño del canvas al viewport
function resizeCanvas() {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
    cursorCanvas.width = window.innerWidth;
    cursorCanvas.height = window.innerHeight;
}
resizeCanvas();
window.addEventListener('resize', resizeCanvas);

// Conexión WebSocket
const ws = new WebSocket(`ws://${location.host}/ws`);

ws.onopen = () => console.log('Conectado a Inkstream');
ws.onclose = () => console.log('Desconectado');
ws.onerror = (e) => console.error('Error WS:', e);

ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    handleMessage(msg);
};

// Despacha mensajes entrantes según tipo
function handleMessage(msg) {
    switch (msg.type) {
        case 'sync':
            msg.strokes.forEach(s => drawSegment(ctx, s));
            break;
        case 'draw':
            drawSegment(ctx, msg);
            break;
        case 'cursor':
            remoteCursors.set(msg.clientId, { x: msg.x, y: msg.y });
            scheduleCursorRedraw();
            break;
        case 'clear':
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            break;
    }
}

// Dibuja un segmento en el contexto dado
function drawSegment(context, s) {
    context.beginPath();
    context.moveTo(s.prevX, s.prevY);
    context.lineTo(s.x, s.y);
    context.strokeStyle = s.color;
    context.lineWidth = s.width;
    context.lineCap = 'round';
    context.lineJoin = 'round';
    context.stroke();
}

// Redibuja los cursores remotos (coalesced con rAF)
function scheduleCursorRedraw() {
    if (cursorAnimFrame) return;
    cursorAnimFrame = requestAnimationFrame(() => {
        cursorAnimFrame = null;
        cursorCtx.clearRect(0, 0, cursorCanvas.width, cursorCanvas.height);
        remoteCursors.forEach(({ x, y }) => {
            cursorCtx.beginPath();
            cursorCtx.arc(x, y, 6, 0, Math.PI * 2);
            cursorCtx.fillStyle = 'rgba(124, 106, 247, 0.8)';
            cursorCtx.fill();
        });
    });
}

// Eventos de dibujo
canvas.addEventListener('mousedown', (e) => {
    isDrawing = true;
    lastX = e.offsetX;
    lastY = e.offsetY;
});

canvas.addEventListener('mousemove', (e) => {
    const x = e.offsetX;
    const y = e.offsetY;

    // Mandar posición del cursor siempre
    ws.send(JSON.stringify({ type: 'cursor', x, y }));

    if (!isDrawing) return;

    const stroke = {
        type: 'draw',
        x,
        y,
        prevX: lastX,
        prevY: lastY,
        color: colorPicker.value,
        width: parseInt(brushSize.value),
    };

    // Dibujar localmente (feedback inmediato)
    drawSegment(ctx, stroke);

    // Mandar al servidor
    ws.send(JSON.stringify(stroke));

    lastX = x;
    lastY = y;
});

canvas.addEventListener('mouseup', () => { isDrawing = false; });
canvas.addEventListener('mouseleave', () => { isDrawing = false; });

// Botón limpiar
clearBtn.addEventListener('click', () => {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ws.send(JSON.stringify({ type: 'clear' }));
});
