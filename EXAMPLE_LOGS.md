# Example Log Output

This document shows example logs you'll see when running the application with the improvements.

## Successful Registration Sequence

```json
{"time":"2026-03-04T10:15:30.123Z","level":"INFO","msg":"starting webrtc-backend","listen":":8080","log-level":"debug","api-base-path":"/api"}
{"time":"2026-03-04T10:15:30.125Z","level":"INFO","msg":"Bridge: initialized and starting"}
{"time":"2026-03-04T10:15:30.126Z","level":"DEBUG","msg":"Bridge: SIP client available, starting incoming call forwarder"}
{"time":"2026-03-04T10:15:30.127Z","level":"INFO","msg":"server started","addr":":8080"}
{"time":"2026-03-04T10:15:30.128Z","level":"INFO","msg":"SIP: starting registration","server":"K4493.reg.cloud-cfg.com:6050","username":"K4493PM9MK","domain":"K4493.reg.cloud-cfg.com"}
{"time":"2026-03-04T10:15:30.129Z","level":"DEBUG","msg":"SIP: user agent created"}
{"time":"2026-03-04T10:15:30.130Z","level":"DEBUG","msg":"SIP: client created"}
{"time":"2026-03-04T10:15:30.131Z","level":"DEBUG","msg":"SIP: REGISTER request built","uri":"sip:K4493PM9MK@K4493.reg.cloud-cfg.com:6050"}
{"time":"2026-03-04T10:15:30.132Z","level":"DEBUG","msg":"SIP: contact header set","local_ip":"172.18.0.3"}
{"time":"2026-03-04T10:15:30.200Z","level":"DEBUG","msg":"SIP: initial REGISTER response","status_code":401}
{"time":"2026-03-04T10:15:30.201Z","level":"INFO","msg":"SIP: received 401 Unauthorized, attempting digest authentication"}
{"time":"2026-03-04T10:15:30.202Z","level":"DEBUG","msg":"SIP: digest challenge parsed","realm":"K4493.reg.cloud-cfg.com"}
{"time":"2026-03-04T10:15:30.203Z","level":"DEBUG","msg":"SIP: digest credentials computed"}
{"time":"2026-03-04T10:15:30.204Z","level":"DEBUG","msg":"SIP: sending authenticated REGISTER request"}
{"time":"2026-03-04T10:15:30.280Z","level":"DEBUG","msg":"SIP: authenticated REGISTER response","status_code":200}
{"time":"2026-03-04T10:15:30.281Z","level":"INFO","msg":"SIP: ✅ registration successful","username":"K4493PM9MK"}
{"time":"2026-03-04T10:15:30.282Z","level":"DEBUG","msg":"Bridge: waiting for incoming SIP calls..."}
{"time":"2026-03-04T10:15:30.283Z","level":"INFO","msg":"SIP: 📞 listener starting","addr":"0.0.0.0:5060"}
```

**Frontend Status Check:**
```json
{"time":"2026-03-04T10:15:31.500Z","level":"DEBUG","msg":"status check","sip_status":"registered"}
```

**Frontend Response:**
```json
{
  "sip": "registered",
  "phone_number": "K4493PM9MK"
}
```

## Incoming Call Sequence

```json
{"time":"2026-03-04T10:20:15.450Z","level":"INFO","msg":"SIP: 🔔 incoming INVITE received","from":"1234567890"}
{"time":"2026-03-04T10:20:15.451Z","level":"DEBUG","msg":"SIP: SDP offer extracted","from":"1234567890"}
{"time":"2026-03-04T10:20:15.452Z","level":"DEBUG","msg":"SIP: sent 180 Ringing","from":"1234567890"}
{"time":"2026-03-04T10:20:15.453Z","level":"DEBUG","msg":"SIP: incoming call forwarded to bridge","from":"1234567890"}
{"time":"2026-03-04T10:20:15.454Z","level":"INFO","msg":"Bridge: 🔔 forwarding incoming call to WebSocket clients","from":"1234567890"}
{"time":"2026-03-04T10:20:15.455Z","level":"DEBUG","msg":"Bridge: incoming call broadcasted","from":"1234567890"}
```

## Call Acceptance Sequence

```json
{"time":"2026-03-04T10:20:20.100Z","level":"INFO","msg":"Bridge: ✅ received answer from WebSocket, accepting SIP call"}
{"time":"2026-03-04T10:20:20.101Z","level":"INFO","msg":"SIP: ✅ call accepted, 200 OK sent"}
```

**Active call logs during conversation:**
```
(call is active, minimal logging)
```

## Call Rejection Sequence

```json
{"time":"2026-03-04T10:20:18.200Z","level":"INFO","msg":"Bridge: ❌ received hangup from WebSocket, ending SIP call"}
{"time":"2026-03-04T10:20:18.201Z","level":"INFO","msg":"SIP: ❌ call rejected, 486 Busy sent"}
```

## Call End (BYE) Sequence

```json
{"time":"2026-03-04T10:20:45.300Z","level":"INFO","msg":"SIP: 📞 BYE received (call ended)"}
{"time":"2026-03-04T10:20:45.301Z","level":"DEBUG","msg":"SIP: sent 200 OK to BYE"}
{"time":"2026-03-04T10:20:45.302Z","level":"INFO","msg":"Bridge: ❌ received hangup from WebSocket, ending SIP call"}
{"time":"2026-03-04T10:20:45.303Z","level":"INFO","msg":"SIP: 📞 call ended (hangup)"}
```

## Registration Failure - Wrong Password

```json
{"time":"2026-03-04T10:15:30.128Z","level":"INFO","msg":"SIP: starting registration","server":"K4493.reg.cloud-cfg.com:6050","username":"K4493PM9MK","domain":"K4493.reg.cloud-cfg.com"}
{"time":"2026-03-04T10:15:30.129Z","level":"DEBUG","msg":"SIP: user agent created"}
{"time":"2026-03-04T10:15:30.130Z","level":"DEBUG","msg":"SIP: client created"}
{"time":"2026-03-04T10:15:30.131Z","level":"DEBUG","msg":"SIP: REGISTER request built","uri":"sip:K4493PM9MK@K4493.reg.cloud-cfg.com:6050"}
{"time":"2026-03-04T10:15:30.132Z","level":"DEBUG","msg":"SIP: contact header set","local_ip":"172.18.0.3"}
{"time":"2026-03-04T10:15:30.200Z","level":"DEBUG","msg":"SIP: initial REGISTER response","status_code":401}
{"time":"2026-03-04T10:15:30.201Z","level":"INFO","msg":"SIP: received 401 Unauthorized, attempting digest authentication"}
{"time":"2026-03-04T10:15:30.202Z","level":"DEBUG","msg":"SIP: digest challenge parsed","realm":"K4493.reg.cloud-cfg.com"}
{"time":"2026-03-04T10:15:30.203Z","level":"DEBUG","msg":"SIP: digest credentials computed"}
{"time":"2026-03-04T10:15:30.204Z","level":"DEBUG","msg":"SIP: sending authenticated REGISTER request"}
{"time":"2026-03-04T10:15:30.280Z","level":"DEBUG","msg":"SIP: authenticated REGISTER response","status_code":403}
{"time":"2026-03-04T10:15:30.281Z","level":"ERROR","msg":"SIP: REGISTER failed with error response","status_code":403,"reason":"Forbidden"}
{"time":"2026-03-04T10:15:30.282Z","level":"WARN","msg":"SIP registration failed","error":"SIP REGISTER failed with status 403: Forbidden"}
```

**Frontend Status Check:**
```json
{
  "sip": "unregistered",
  "phone_number": "K4493PM9MK"
}
```

Frontend shows 🟡 yellow

## Registration Failure - Network Unreachable

```json
{"time":"2026-03-04T10:15:30.128Z","level":"INFO","msg":"SIP: starting registration","server":"K4493.reg.cloud-cfg.com:6050","username":"K4493PM9MK","domain":"K4493.reg.cloud-cfg.com"}
{"time":"2026-03-04T10:15:30.129Z","level":"DEBUG","msg":"SIP: user agent created"}
{"time":"2026-03-04T10:15:30.130Z","level":"DEBUG","msg":"SIP: client created"}
{"time":"2026-03-04T10:15:30.131Z","level":"DEBUG","msg":"SIP: REGISTER request built","uri":"sip:K4493PM9MK@K4493.reg.cloud-cfg.com:6050"}
{"time":"2026-03-04T10:15:30.132Z","level":"DEBUG","msg":"SIP: contact header set","local_ip":"172.18.0.3"}
{"time":"2026-03-04T10:15:35.200Z","level":"ERROR","msg":"SIP: no response to initial REGISTER","error":"transaction terminated"}
{"time":"2026-03-04T10:15:35.201Z","level":"WARN","msg":"SIP registration failed","error":"SIP REGISTER response error: transaction terminated"}
```

**Indicates:** Network connectivity issue or firewall blocking port 6050 UDP

## Log Filtering Examples

### See only SIP events:
```bash
docker-compose logs backend | grep "SIP:"
```

### See only errors:
```bash
docker-compose logs backend | grep "ERROR"
```

### See only registration:
```bash
docker-compose logs backend | grep "registration"
```

### See incoming calls:
```bash
docker-compose logs backend | grep "🔔"
```

### See call lifecycle:
```bash
docker-compose logs backend | grep -E "🔔|✅|❌|📞"
```

### Watch logs in real-time:
```bash
docker-compose logs -f backend
```

## Status Over Time

### With current setup (before fix):
```
Frontend status: 🟡 unregistered
(even though SIP actually registered)
Status endpoint always returned: {"sip": "unregistered"}
```

### After fix:
```
Frontend status: 🟡 unregistered (during startup)
  ↓ (after ~1 second, registration succeeds)
Frontend status: 🟢 registered
Status endpoint returns: {"sip": "registered"}
```

## Emoji Legend

- 🟢 **Green** = SIP registered (frontend)
- 🟡 **Yellow** = SIP unregistered (frontend)
- 🔔 = Incoming call
- ✅ = Call accepted / Success
- ❌ = Call rejected / Error
- 📞 = Call ended / BYE
- 💥 = Critical error (in error context)

## Performance Notes

- Debug logs: Some overhead, good for troubleshooting
- Info logs: Minimal overhead, recommended for normal operation
- Timestamp on every log: Helps track timing of events
- JSON format: Easy to parse and analyze programmatically

## Next Steps

1. Start the application: `docker-compose up --build`
2. Compare your logs to these examples
3. Look for ✅ registration success message
4. Frontend should show 🟢 green
5. Make a test call to verify incoming call flow
6. Change `LOG_LEVEL` to `info` in docker-compose.yml for production

