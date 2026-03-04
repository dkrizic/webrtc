# Using .env File for Credentials

## Overview

The `.env` file allows you to keep sensitive credentials out of version control while easily managing environment variables for Docker Compose.

## Setup Instructions

### Step 1: Copy the Template File

```bash
cd /Users/dkrizic/Repository/dkrizic/webrtc
cp .env.example .env
```

### Step 2: Edit .env with Your Credentials

```bash
nano .env
# or use your favorite editor
```

Your `.env` file should look like:
```
SIP_SERVER=K4493.reg.cloud-cfg.com:6050
SIP_USERNAME=K4493PM9MK
SIP_PASSWORD=YOUR_ACTUAL_PASSWORD_HERE
SIP_DOMAIN=K4493.reg.cloud-cfg.com
LISTEN_ADDR=:8080
LOG_LEVEL=debug
API_BASE_PATH=/api
```

Replace `YOUR_ACTUAL_PASSWORD_HERE` with your actual NFON SIP password.

### Step 3: Start Docker Compose

```bash
docker-compose up --build
```

Docker Compose will automatically load variables from the `.env` file.

## How It Works

1. **`.env.example`** — Template file (committed to git) showing what variables are needed
2. **`.env`** — Your actual credentials (NOT committed, in .gitignore)
3. **`docker-compose.yml`** — References variables from `.env` using `${VARIABLE_NAME}`

## File Status

```
✅ .env.example    → Committed (safe, has placeholder password)
❌ .env            → Not committed (contains real password)
```

Check your .gitignore:
```bash
cat .gitignore | grep env
```

Should show:
```
.env
.env.*
```

## Verify Setup

### Before Starting
```bash
# Check .env exists
ls -la .env

# Check .env is ignored by git
git status .env
# Should show: nothing (not tracked)

# Check values are set
cat .env
```

### While Running
```bash
# View actual environment variables in container
docker-compose ps
docker exec [container-id] env | grep SIP

# View logs
docker-compose logs backend
```

## Updating Credentials

To change any credential:

1. Edit `.env`
2. Run `docker-compose down`
3. Run `docker-compose up --build`

The new values will be loaded automatically.

## Security Best Practices

✅ **Do:**
- Keep `.env` in .gitignore
- Use unique passwords for each environment
- Rotate credentials periodically
- Use environment variables in production
- Document what each variable does (in `.env.example`)

❌ **Don't:**
- Commit `.env` to git
- Share `.env` files unencrypted
- Use same credentials in dev and prod
- Hardcode credentials in code or config files
- Store backups of `.env` in public places

## Troubleshooting

### "File not found: .env"
```bash
# Create it from template
cp .env.example .env

# Then edit and add your password
nano .env
```

### "Variables not loading"
```bash
# Verify .env file exists
ls -la .env

# Check it's readable
file .env

# Check docker-compose references it
grep "env_file" docker-compose.yml
```

### "Still seeing placeholder password"
```bash
# Make sure you edited .env (not .env.example)
grep "YOUR_SIP_PASSWORD_HERE" .env
# Should be empty

grep "YOUR_SIP_PASSWORD_HERE" .env.example
# Should show the line (this is OK)
```

## Multiple Environments

For multiple environments, create separate .env files:

```bash
# Development
cp .env.example .env.dev
nano .env.dev

# Production
cp .env.example .env.prod
nano .env.prod

# Use with docker-compose
docker-compose --env-file .env.dev up --build
docker-compose --env-file .env.prod up --build
```

## CI/CD Integration

For continuous integration, set environment variables directly:

```bash
# GitHub Actions
export SIP_PASSWORD=${{ secrets.SIP_PASSWORD }}
docker-compose up --build

# GitLab CI
docker-compose up --build
# Variables set via GitLab UI secrets

# Jenkins
environment {
    SIP_PASSWORD = credentials('nfon-sip-password')
}
docker-compose up --build
```

## Summary

| File | Purpose | Committed |
|------|---------|-----------|
| `.env.example` | Template for variables | ✅ Yes |
| `.env` | Your credentials | ❌ No |
| `.gitignore` | Ignore rules | ✅ Yes |

## Next Steps

1. Copy template: `cp .env.example .env`
2. Edit file: `nano .env`
3. Add your password
4. Start app: `docker-compose up --build`
5. Verify: `docker-compose logs backend | grep "registration"`

That's it! Your credentials are now safely externalized. 🔒

