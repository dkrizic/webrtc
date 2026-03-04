# Configuration Setup Instructions

## Before Running the Application

You need to replace the placeholder password with your actual NFON SIP password.

### Step 1: Get Your Credentials

From your NFON support ticket, you should have:
- **Siplogin** (username): K4493PM9MK
- **SIP Kennwort** (password): [Your actual password]
- **SIP Registrar**: K4493.reg.cloud-cfg.com
- **Port**: 6050

### Step 2: Update docker-compose.yml

Edit `/Users/dkrizic/Repository/dkrizic/webrtc/docker-compose.yml`

Find this section:
```yaml
environment:
  - SIP_SERVER=K4493.reg.cloud-cfg.com:6050
  - SIP_USERNAME=K4493PM9MK
  - SIP_PASSWORD=YOUR_SIP_PASSWORD_HERE    ← Replace this
  - SIP_DOMAIN=K4493.reg.cloud-cfg.com
```

### Step 3: Verify Configuration

Double-check:
- ✅ `SIP_SERVER` = `K4493.reg.cloud-cfg.com:6050` (with port)
- ✅ `SIP_USERNAME` = `K4493PM9MK` (your SIP login)
- ✅ `SIP_PASSWORD` = Your actual password (not placeholder)
- ✅ `SIP_DOMAIN` = `K4493.reg.cloud-cfg.com` (without port)

### Step 4: Start Application

```bash
cd /Users/dkrizic/Repository/dkrizic/webrtc
docker-compose up --build
```

### Step 5: Verify Registration

Look for this log message:
```
✅ SIP: registration successful
```

If you see it → Configuration is correct! 🎉

## Security Notes

⚠️ **Important:**
- Never commit the actual password to version control
- The `docker-compose.yml` file in git contains `YOUR_SIP_PASSWORD_HERE` as a placeholder
- Only your local copy should have the real password
- Consider using `.env` files for production

## Alternative: Using Environment Variables

Instead of editing docker-compose.yml, you can set environment variables:

```bash
export SIP_PASSWORD="your_actual_password"
docker-compose up --build
```

Or create a `.env` file:

```bash
cat > .env << EOF
SIP_SERVER=K4493.reg.cloud-cfg.com:6050
SIP_USERNAME=K4493PM9MK
SIP_PASSWORD=your_actual_password
SIP_DOMAIN=K4493.reg.cloud-cfg.com
LOG_LEVEL=debug
EOF

docker-compose up --build
```

Then docker-compose will automatically load from `.env`.

## Troubleshooting Configuration

### If you see "Forbidden" error:
```json
{"level":"error","msg":"SIP: REGISTER failed with error response","status_code":403}
```

→ Wrong password! Double-check in docker-compose.yml

### If you see "Unauthorized" error:
```json
{"level":"info","msg":"SIP: received 401 Unauthorized, attempting digest authentication"}
```

→ This is normal! The server is challenging for auth (digest auth)
→ If followed by "registration successful", all is good ✅

### If no response after timeout:
```json
{"level":"error","msg":"SIP: no response to initial REGISTER"}
```

→ Network issue or firewall blocking port 6050 UDP
→ Check: `ping K4493.reg.cloud-cfg.com`

## What Not To Do

❌ Don't commit the real password to git
❌ Don't share docker-compose.yml with real password
❌ Don't use the same password elsewhere
❌ Don't leave plaintext password in shell history

## Clean Up

After testing, remove the plaintext password if you committed it:

```bash
# Remove from git history
git filter-branch --env-filter '...'

# Or use git-secret
gem install git-secret
git secret init
git secret add docker-compose.yml
```

## Summary

| Step | Action |
|------|--------|
| 1 | Get password from NFON |
| 2 | Edit `docker-compose.yml` |
| 3 | Replace `YOUR_SIP_PASSWORD_HERE` |
| 4 | Run `docker-compose up --build` |
| 5 | Look for `✅ registration successful` |

Once configured correctly, you'll see:
- Frontend status: 🟢 green
- Logs show successful registration
- Incoming calls work

