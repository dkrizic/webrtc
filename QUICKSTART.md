# Quick Start - After Improvements

## What Was Fixed

✅ **Frontend yellow status issue** - Now shows real SIP registration state  
✅ **Enhanced logging** - Easy to diagnose connection and call issues  
✅ **Registration state tracking** - Backend knows if SIP is connected

## Quick Start (5 minutes)

### 1. Verify Configuration

Check `docker-compose.yml` has correct NFON credentials:

```yaml
environment:
  - SIP_SERVER=K4493.reg.cloud-cfg.com:6050
  - SIP_USERNAME=K4493PM9MK
  - SIP_PASSWORD=YOUR_SIP_PASSWORD_HERE
  - SIP_DOMAIN=K4493.reg.cloud-cfg.com
  - LOG_LEVEL=debug
```

**Note:** Temporarily using `LOG_LEVEL=debug` for detailed logs. Change to `info` later.

### 2. Start the Application

```bash
cd /Users/dkrizic/Repository/dkrizic/webrtc
docker-compose up --build
```

### 3. Watch for Registration Success

Look for this log message:

```
SIP: ✅ registration successful
```

If you see it → **Success! ✅**  
If you don't see it after 5 seconds → **Check the error logs**

### 4. Open Frontend

Open browser: **http://localhost**

**Expected:**
- Status indicator should show 🟢 **green** (registered)
- Phone number should display your SIP username

**If still yellow 🟡:**
- Check backend logs for errors
- Verify credentials in docker-compose.yml
- See Troubleshooting section below

### 5. Test Incoming Call

Call your NFON SIP number from a phone.

**Expected logs:**
```
SIP: 🔔 incoming INVITE received (from: [caller-id])
Bridge: 🔔 forwarding incoming call to WebSocket clients
```

**Frontend shows:** Incoming call notification

### 6. Accept the Call

Click "Accept" button in frontend.

**Expected logs:**
```
Bridge: ✅ received answer from WebSocket, accepting SIP call
SIP: ✅ call accepted, 200 OK sent
```

**Result:** Call is active, audio flows

### 7. End the Call

Click "Hangup" button or caller hangs up.

**Expected logs:**
```
SIP: 📞 BYE received (call ended)
SIP: 📞 call ended (hangup)
```

## Key Logs to Watch

| Log Message | Meaning |
|-------------|---------|
| `✅ registration successful` | SIP is connected ✅ |
| `REGISTER failed with error` | Auth problem or network issue ❌ |
| `🔔 incoming INVITE received` | Call incoming 📞 |
| `✅ call accepted` | Call accepted by frontend ✅ |
| `❌ call rejected` | Call rejected by frontend ❌ |
| `📞 BYE received` | Call ended by caller 📞 |

## Troubleshooting

### Frontend shows 🟡 yellow (unregistered)

**Check logs:**
```bash
docker-compose logs backend | grep "registration"
```

**If you see "REGISTER failed":**

1. **Wrong password?**
   ```bash
   docker-compose logs backend | grep "Forbidden"
   ```
   → Update `SIP_PASSWORD` in docker-compose.yml

2. **Network unreachable?**
   ```bash
   docker-compose logs backend | grep "transaction terminated"
   ```
   → Test connectivity: `ping K4493.reg.cloud-cfg.com`
   → Check firewall for port 6050 UDP

3. **Wrong server/domain?**
   → Verify both `SIP_SERVER` and `SIP_DOMAIN` match NFON config

### No incoming calls received

**Check logs:**
```bash
docker-compose logs backend | grep "📞 listener"
```

If not there, listener didn't start.

**Check logs:**
```bash
docker-compose logs backend | grep "🔔 incoming"
```

If not there, calls aren't reaching server.

**Solutions:**
1. Verify port 5060 UDP is open: `netstat -un | grep 5060`
2. Check firewall rules for UDP 5060
3. Verify registration contact address in logs

### Calls received but no audio

If logs show `✅ call accepted` but no audio:

→ Not a SIP signaling issue  
→ Check WebRTC/audio layer  
→ Check browser audio permissions

## Log Levels

### For Development/Troubleshooting
```yaml
LOG_LEVEL=debug  # Detailed step-by-step logs
```

### For Production
```yaml
LOG_LEVEL=info   # Only important events
```

### For Quiet Operation
```yaml
LOG_LEVEL=warn   # Only warnings and errors
```

Change in `docker-compose.yml` and restart.

## View Logs

### Real-time logs:
```bash
docker-compose logs -f backend
```

### Filter by keyword:
```bash
docker-compose logs backend | grep "SIP:"
docker-compose logs backend | grep "registration"
docker-compose logs backend | grep "ERROR"
```

### Filter by emoji (fun!):
```bash
docker-compose logs backend | grep "🔔\|✅\|❌\|📞"
```

## Files Changed

If you're curious what was modified:

- `backend/internal/sip/client.go` — Registration tracking + logging
- `backend/internal/api/status.go` — Real registration status
- `backend/internal/api/router.go` — Pass SIP status to endpoint
- `backend/cmd/server/main.go` — Wire SIP client to API
- `backend/internal/bridge/bridge.go` — Call routing logs

## Documentation

- **LOGGING_GUIDE.md** — Detailed logging reference
- **IMPROVEMENTS_SUMMARY.md** — Technical details of changes
- **EXAMPLE_LOGS.md** — Sample log outputs

## Next Steps

1. ✅ Start: `docker-compose up --build`
2. ✅ Check logs for `✅ registration successful`
3. ✅ Verify frontend shows 🟢 green
4. ✅ Test incoming call
5. ✅ Test accept/reject/hangup
6. ✅ Change `LOG_LEVEL` to `info` for production

## Support

If you still have issues:

1. **Get full logs:**
   ```bash
   docker-compose logs backend > logs.txt
   ```

2. **Check for these patterns:**
   - `registration successful` → Registration worked
   - `REGISTER failed` → Auth or network issue
   - `🔔 incoming` → Calls arriving
   - `ERROR` → Look for specific error messages

3. **Verify configuration:**
   - `SIP_SERVER` with port
   - `SIP_DOMAIN` without port
   - Username and password correct
   - Network connectivity to registrar

That's it! You should now have a fully working WebRTC-to-SIP gateway with clear visibility into what's happening. 🎉

