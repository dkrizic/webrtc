// app.js - Main application logic
(function() {
    let currentCall = null;

    // Wire signaling events
    Signaling.on('onopen', () => {
        UI.setConnected(true);
        UI.addCallLogEntry('Connected to server');
    });

    Signaling.on('onclose', () => {
        UI.setConnected(false);
        UI.addCallLogEntry('Disconnected from server');
    });

    Signaling.on('onmessage', (msg) => {
        switch (msg.type) {
            case 'status':
                console.log('[App] status:', msg.data);
                break;
            case 'answer':
                WebRTCClient.handleAnswer(msg.payload);
                break;
            case 'ice':
                WebRTCClient.addIceCandidate(msg.payload);
                break;
            case 'incoming':
                currentCall = msg.payload && msg.payload.from;
                UI.showIncomingCall(currentCall || 'Unknown');
                break;
            case 'hangup':
                WebRTCClient.hangup();
                UI.hideActiveCall();
                UI.addCallLogEntry('Call ended with ' + (currentCall || 'Unknown'));
                currentCall = null;
                break;
            case 'error':
                console.error('[App] server error:', msg.data);
                break;
        }
    });

    // Dial button
    document.getElementById('dial-btn').addEventListener('click', async () => {
        const number = document.getElementById('phone-number').value.trim();
        if (!number) return;
        currentCall = number;
        UI.addCallLogEntry('Dialing ' + number);
        const offer = await WebRTCClient.startCall((candidate) => {
            Signaling.send({ type: 'ice', payload: candidate });
        });
        Signaling.send({ type: 'dial', payload: { to: number } });
        Signaling.send({ type: 'offer', payload: offer });
        UI.showActiveCall(number);
    });

    // Hangup button
    document.getElementById('hangup-btn').addEventListener('click', () => {
        Signaling.send({ type: 'hangup', payload: {} });
        WebRTCClient.hangup();
        UI.hideActiveCall();
        UI.addCallLogEntry('Hung up on ' + (currentCall || 'Unknown'));
        currentCall = null;
    });

    // Accept button
    document.getElementById('accept-btn').addEventListener('click', async () => {
        UI.hideIncomingCall();
        const offer = await WebRTCClient.startCall((candidate) => {
            Signaling.send({ type: 'ice', payload: candidate });
        });
        Signaling.send({ type: 'answer', payload: offer });
        UI.showActiveCall(currentCall || 'Incoming');
    });

    // Reject button
    document.getElementById('reject-btn').addEventListener('click', () => {
        UI.hideIncomingCall();
        Signaling.send({ type: 'hangup', payload: {} });
        UI.addCallLogEntry('Rejected call from ' + (currentCall || 'Unknown'));
        currentCall = null;
    });

    // Start signaling
    Signaling.connect();
})();
