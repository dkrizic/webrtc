# Enhanced Logging Guide

This document explains the improved logging that has been added to the WebRTC SIP gateway for easier diagnosis of connection and call issues.

## Overview

The logging now includes:
- **Registration Status Tracking**: Know when SIP is connected
- **Call Flow Logging**: See incoming calls, call acceptance, rejection, and hangups
- **Debug Information**: Detailed step-by-step logs for troubleshooting
- **Emoji Indicators**: Visual indicators for quick scanning

## Log Levels

Configure logging with the `LOG_LEVEL` environment variable:

- **debug** — Detailed step-by-step logs for development/troubleshooting
- **info** — Important events (registration, incoming calls, etc.)
- **warn** — Warnings and non-critical issues
- **error** — Errors that need attention

## Frontend Yellow Status Issue

The frontend was showing yellow (unregistered) because the status endpoint didn't track actual SIP registration. This is now **FIXED**:

✅ The backend now tracks SIP registration state
✅ The status endpoint returns the actual registration status
✅ Frontend can now show green (✅ registered) or yellow (🟡 unregistered)

## Key Log Messages

### Registration Flow

```
SIP: starting registration
  ↓
SIP: user agent created
SIP: client created
SIP: REGISTER request built
SIP: contact header set
SIP: initial REGISTER response (status_code: 401)
SIP: received 401 Unauthorized, attempting digest authentication
SIP: digest challenge parsed (realm: K4493.reg.cloud-cfg.com)
SIP: digest credentials computed
SIP: sending authenticated REGISTER request
SIP: authenticated REGISTER response (status_code: 200)
✅ SIP: registration successful (username: K4493PM9MK)
```

**On Success:**
```json
{"level":"info","msg":"SIP: ✅ registration successful","username":"K4493PM9MK"}
```

**On Failure:**
```json
{"level":"error","msg":"SIP: REGISTER failed with error response","status_code":401,"reason":"Unauthorized"}
```

### Incoming Call Flow

```
SIP: 🔔 incoming INVITE received (from: caller-id)
  ↓
SIP: SDP offer extracted
SIP: sent 180 Ringing
SIP: incoming call forwarded to bridge
Bridge: 🔔 forwarding incoming call to WebSocket clients (from: caller-id)
Bridge: incoming call broadcasted
(Frontend shows incoming call notification)
```

### Call Acceptance Flow

```
(User accepts in frontend)
  ↓
Bridge: ✅ received answer from WebSocket, accepting SIP call
SIP: ✅ call accepted, 200 OK sent
```

### Call Rejection Flow

```
(User rejects in frontend)
  ↓
Bridge: ❌ received hangup from WebSocket, ending SIP call
SIP: ❌ call rejected, 486 Busy sent
```

### Call End Flow

```
SIP: 📞 BYE received (call ended)
  ↓
SIP: sent 200 OK to BYE
```

## Log Message Format

All logs are in JSON format with fields:

```json
{
  "level": "info|warn|error|debug",
  "msg": "Message text",
  "timestamp": "...",
  // Additional context fields like:
  "server": "K4493.reg.cloud-cfg.com:6050",
  "username": "K4493PM9MK",
  "from": "caller-id",
  "status_code": 200,
  "error": "error details"
}
```

## Troubleshooting with Logs

### Issue: Frontend always shows yellow (unregistered)

**Check logs for:**
```
SIP: starting registration
```

If you don't see this, the registration goroutine might not have started.

**Look for registration success:**
```
✅ SIP: registration successful
```

If missing, check for auth errors:
```
SIP: REGISTER failed with error response (status_code: 401)
```

**Solutions:**
1. Verify `SIP_USERNAME` and `SIP_PASSWORD` are correct
2. Verify `SIP_SERVER` and `SIP_DOMAIN` are correct for NFON
3. Check if the registrar server is reachable: `ping K4493.reg.cloud-cfg.com`
4. Increase log level to `debug` for more details

### Issue: Incoming calls not received

**Check logs for:**
```
SIP: 📞 listener starting (addr: 0.0.0.0:5060)
```

If missing, the listener didn't start.

**Check if calls arrive:**
```
SIP: 🔔 incoming INVITE received
```

If missing, calls aren't reaching the server.

**Solutions:**
1. Verify port 5060 UDP is open and reachable
2. Check firewall rules
3. Verify SIP server has your registration contact address
4. Increase log level to `debug` for detailed SIP messages

### Issue: Call accepted but no audio

**Check logs for:**
```
Bridge: ✅ received answer from WebSocket, accepting SIP call
SIP: ✅ call accepted, 200 OK sent
```

If these appear, the SIP signaling is working. Issue is likely with WebRTC/audio layer.

## Running with Logs

### Docker Compose

Logs are automatically captured. View with:
```bash
docker-compose up --build
```

The backend logs will display in the terminal with JSON formatting.

### Viewing Logs

Filter by component:
```bash
docker-compose logs backend | grep "SIP:"
docker-compose logs backend | grep "Bridge:"
```

Watch real-time logs:
```bash
docker-compose logs -f backend
```

### Log Level Configuration

In `docker-compose.yml`:
```yaml
environment:
  - LOG_LEVEL=debug  # verbose debugging
  - LOG_LEVEL=info   # normal operation (default)
  - LOG_LEVEL=warn   # warnings only
```

## Status Endpoint

The `/api/status` endpoint now returns actual SIP registration status:

**Registered:**
```json
{
  "sip": "registered",
  "phone_number": "K4493PM9MK"
}
```

**Unregistered:**
```json
{
  "sip": "unregistered",
  "phone_number": "K4493PM9MK"
}
```

The frontend uses this to determine the status indicator color:
- 🟢 **Green** = "registered"
- 🟡 **Yellow** = "unregistered"

## Performance Impact

The enhanced logging has minimal performance impact:
- Debug logs only appear at `debug` level
- Info/warn/error logs are minimal
- All logging is asynchronous (non-blocking)

## Summary

The improved logging provides clear visibility into:
1. ✅ **Registration Status** - Easily see when SIP is connected
2. 🔔 **Incoming Calls** - Clear indication when calls arrive
3. 📞 **Call Flow** - Track accept, reject, hangup operations
4. 🐛 **Debug Info** - Detailed logs at each step for troubleshooting

This should eliminate the mystery of why the frontend was showing yellow, and make it much easier to diagnose any SIP-related issues.

