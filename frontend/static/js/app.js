// app.js - Main application logic
(function() {
    let currentCall = null;
    let pendingOffer = null;
    let statusPollInterval = null;

    // Display frontend version from version.js constant
    UI.setFrontendVersion(FRONTEND_VERSION);

    function fetchBackendVersion() {
        fetch('/api/version')
            .then(r => r.json())
            .then(data => UI.setBackendVersion(data.version || 'unknown'))
            .catch(() => UI.setBackendVersion('unknown'));
    }

    function pollStatus() {
        fetch('/api/status')
            .then(r => r.json())
            .then(data => {
                UI.setSipStatus(data.sip || 'unregistered');
                UI.setPhoneNumber(data.phone_number || '');
            })
            .catch(err => console.warn('[App] status poll failed:', err));
    }

    // Wire signaling events
    Signaling.on('onopen', () => {
        UI.setSipStatus('unregistered');
        UI.addCallLogEntry('Connected to server');
        pollStatus();
        fetchBackendVersion();
        statusPollInterval = setInterval(pollStatus, 5000);
    });

    Signaling.on('onclose', () => {
        UI.setSipStatus('disconnected');
        UI.addCallLogEntry('Disconnected from server');
        if (statusPollInterval) { clearInterval(statusPollInterval); statusPollInterval = null; }
    });

    Signaling.on('onmessage', (msg) => {
        switch (msg.type) {
            case 'status':
                console.log('[App] status:', msg.data);
                if (msg.data) { UI.setSipStatus(msg.data); }
                break;
            case 'answer':
                WebRTCClient.handleAnswer(msg.payload);
                break;
            case 'ice':
                WebRTCClient.addIceCandidate(msg.payload);
                break;
            case 'incoming':
                currentCall = msg.payload && msg.payload.from;
                pendingOffer = msg.payload && msg.payload.offer;
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
        try {
            const offer = await WebRTCClient.startCall((candidate) => {
                Signaling.send({ type: 'ice', payload: candidate });
            });
            Signaling.send({ type: 'dial', payload: { to: number } });
            Signaling.send({ type: 'offer', payload: offer });
            UI.showActiveCall(number);
        } catch (e) {
            console.error('[App] dial failed', e);
            UI.addCallLogEntry('Failed to dial ' + number + ': ' + e.message);
            currentCall = null;
        }
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
        try {
            const answer = await WebRTCClient.handleOffer(pendingOffer, (candidate) => {
                Signaling.send({ type: 'ice', payload: candidate });
            });
            Signaling.send({ type: 'answer', payload: answer });
            UI.showActiveCall(currentCall || 'Incoming');
            pendingOffer = null;
        } catch (e) {
            console.error('[App] accept call failed', e);
            UI.addCallLogEntry('Failed to accept call: ' + e.message);
        }
    });

    // Reject button
    document.getElementById('reject-btn').addEventListener('click', () => {
        UI.hideIncomingCall();
        Signaling.send({ type: 'hangup', payload: {} });
        UI.addCallLogEntry('Rejected call from ' + (currentCall || 'Unknown'));
        currentCall = null;
        pendingOffer = null;
    });

    // Start signaling
    Signaling.connect();
})();
