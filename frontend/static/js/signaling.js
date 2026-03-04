// signaling.js - WebSocket connection to backend /ws
const Signaling = (() => {
    let socket = null;
    let reconnectTimer = null;
    const RECONNECT_DELAY = 3000;
    const handlers = {};

    function connect() {
        const wsUrl = (location.protocol === 'https:' ? 'wss:' : 'ws:') + '//' + location.host + '/ws';
        console.log('[Signaling] connecting to', wsUrl);
        socket = new WebSocket(wsUrl);

        socket.onopen = () => {
            console.log('[Signaling] connected');
            if (handlers.onopen) handlers.onopen();
        };

        socket.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);
                console.log('[Signaling] received', msg);
                if (handlers.onmessage) handlers.onmessage(msg);
            } catch (e) {
                console.error('[Signaling] failed to parse message', e);
            }
        };

        socket.onclose = () => {
            console.log('[Signaling] disconnected, reconnecting in', RECONNECT_DELAY, 'ms');
            if (handlers.onclose) handlers.onclose();
            reconnectTimer = setTimeout(connect, RECONNECT_DELAY);
        };

        socket.onerror = (err) => {
            console.error('[Signaling] error', err);
        };
    }

    function send(msg) {
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify(msg));
        } else {
            console.warn('[Signaling] not connected, cannot send', msg);
        }
    }

    function on(event, handler) {
        handlers[event] = handler;
    }

    return { connect, send, on };
})();
