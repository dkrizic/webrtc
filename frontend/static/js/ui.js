// ui.js - DOM manipulation and status display
const UI = (() => {
    const sipStatus = document.getElementById('sip-status');
    const sipPhoneNumber = document.getElementById('sip-phone-number');
    const incomingCallDiv = document.getElementById('incoming-call');
    const callerIdSpan = document.getElementById('caller-id');
    const activeCallDiv = document.getElementById('active-call');
    const callPartySpan = document.getElementById('call-party');
    const callDurationSpan = document.getElementById('call-duration');
    const callLogList = document.getElementById('call-log-list');

    let callTimer = null;
    let callSeconds = 0;

    function setSipStatus(status) {
        if (status === 'registered') {
            sipStatus.textContent = 'SIP: Connected ✅';
            sipStatus.className = 'status-indicator connected';
        } else if (status === 'disconnected') {
            sipStatus.textContent = 'SIP: Disconnected ❌';
            sipStatus.className = 'status-indicator disconnected';
        } else {
            sipStatus.textContent = 'SIP: Unregistered 🟡';
            sipStatus.className = 'status-indicator uncertain';
        }
    }

    function setPhoneNumber(number) {
        sipPhoneNumber.textContent = number ? ' | ' + number : '';
    }

    function showIncomingCall(callerId) {
        callerIdSpan.textContent = callerId;
        incomingCallDiv.classList.remove('hidden');
    }

    function hideIncomingCall() {
        incomingCallDiv.classList.add('hidden');
    }

    function showActiveCall(party) {
        callPartySpan.textContent = party;
        callSeconds = 0;
        callDurationSpan.textContent = '00:00';
        activeCallDiv.classList.remove('hidden');
        callTimer = setInterval(() => {
            callSeconds++;
            const m = String(Math.floor(callSeconds / 60)).padStart(2, '0');
            const s = String(callSeconds % 60).padStart(2, '0');
            callDurationSpan.textContent = m + ':' + s;
        }, 1000);
    }

    function hideActiveCall() {
        activeCallDiv.classList.add('hidden');
        if (callTimer) { clearInterval(callTimer); callTimer = null; }
    }

    function addCallLogEntry(entry) {
        const li = document.createElement('li');
        li.textContent = new Date().toLocaleTimeString() + ' - ' + entry;
        callLogList.insertBefore(li, callLogList.firstChild);
    }

    return { setSipStatus, setPhoneNumber, showIncomingCall, hideIncomingCall, showActiveCall, hideActiveCall, addCallLogEntry };
})();
