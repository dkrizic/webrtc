// webrtc.js - RTCPeerConnection management
const WebRTCClient = (() => {
    let pc = null;
    const audioEl = document.getElementById('remote-audio');

    const config = {
        iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
    };

    function createPeerConnection(onIceCandidate) {
        pc = new RTCPeerConnection(config);

        pc.onicecandidate = (event) => {
            if (event.candidate) {
                onIceCandidate(event.candidate);
            }
        };

        pc.ontrack = (event) => {
            console.log('[WebRTC] remote track received');
            if (audioEl && event.streams[0]) {
                audioEl.srcObject = event.streams[0];
            }
        };

        pc.oniceconnectionstatechange = () => {
            console.log('[WebRTC] ICE state:', pc.iceConnectionState);
        };

        return pc;
    }

    async function startCall(onIceCandidate) {
        createPeerConnection(onIceCandidate);
        try {
            const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
            stream.getTracks().forEach(track => pc.addTrack(track, stream));
        } catch (e) {
            console.error('[WebRTC] getUserMedia failed', e);
            pc.close();
            pc = null;
            throw e;
        }
        const offer = await pc.createOffer();
        await pc.setLocalDescription(offer);
        return offer;
    }

    async function handleOffer(offer, onIceCandidate) {
        createPeerConnection(onIceCandidate);
        try {
            const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
            stream.getTracks().forEach(track => pc.addTrack(track, stream));
        } catch (e) {
            console.error('[WebRTC] getUserMedia failed', e);
            pc.close();
            pc = null;
            throw e;
        }
        await pc.setRemoteDescription(new RTCSessionDescription(offer));
        const answer = await pc.createAnswer();
        await pc.setLocalDescription(answer);
        return answer;
    }

    async function handleAnswer(answer) {
        if (!pc) return;
        await pc.setRemoteDescription(new RTCSessionDescription(answer));
    }

    async function addIceCandidate(candidate) {
        if (!pc) return;
        await pc.addIceCandidate(new RTCIceCandidate(candidate));
    }

    function hangup() {
        if (pc) {
            pc.close();
            pc = null;
        }
    }

    return { startCall, handleOffer, handleAnswer, addIceCandidate, hangup };
})();
