# Simple CI/CD for Staging

A straightforward CI/CD pipeline that automatically deploys to staging on every push to `main`.

## Quick Setup (3 Steps)

### Step 1: Add GitHub Secrets

Go to: `https://github.com/lac-hong-legacy/TechYouth-Be/settings/secrets/actions`

Click **"New repository secret"** and add these 5 secrets:

```
DOCKER_USERNAME = your_dockerhub_username
DOCKER_PASSWORD = your_dockerhub_token

STAGING_HOST = 192.168.1.100  # Your staging server IP
STAGING_USER = deploy
STAGING_SSH_KEY = -----BEGIN OPENSSH PRIVATE KEY-----
                  <paste your entire private key here>
                  -----END OPENSSH PRIVATE KEY-----
```

### Step 2: Setup Staging Server

SSH into your staging server:

```bash
ssh root@your-staging-server

# Download and run setup script
wget https://raw.githubusercontent.com/lac-hong-legacy/TechYouth-Be/main/scripts/setup-staging.sh
chmod +x setup-staging.sh
./setup-staging.sh
```

Follow the prompts:
1. Edit `/opt/ven-api/.env` with your configuration
2. Add your SSH public key to `~/.ssh/authorized_keys`
3. Start the app: `cd /opt/ven-api && docker compose up -d`

### Step 3: Push to GitHub

```bash
git add .
git commit -m "feat: enable CI/CD"
git push origin main
```

✅ **Done!** Your app will automatically deploy to staging.

---

## What Happens Automatically

### On Push to `main`:
1. ✅ Run tests with PostgreSQL + Redis
2. ✅ Run code linting
3. ✅ Build Docker image and push to Docker Hub
4. ✅ Deploy to staging server
5. ✅ Run health check

### On Pull Requests:
1. ✅ Run tests
2. ✅ Run code linting
3. ❌ No deployment

---

## Generate SSH Keys

On your local machine:

```bash
# Generate SSH key pair
ssh-keygen -t ed25519 -C "deploy@staging" -f ~/.ssh/staging-deploy

# View public key (add this to server's ~/.ssh/authorized_keys)
cat ~/.ssh/staging-deploy.pub

# View private key (add this to GitHub secrets as STAGING_SSH_KEY)
cat ~/.ssh/staging-deploy
```

On staging server:

```bash
# Add public key
mkdir -p ~/.ssh
chmod 700 ~/.ssh
nano ~/.ssh/authorized_keys
# Paste the public key, save and exit

chmod 600 ~/.ssh/authorized_keys
```

Test connection:

```bash
ssh -i ~/.ssh/staging-deploy deploy@your-staging-server
```

---

## Useful Commands

### Check Deployment Status

```bash
# On GitHub: Actions tab → Latest workflow run

# Or use GitHub CLI
gh run list
gh run watch
```

### Check Staging Server

```bash
ssh deploy@staging-server

# View running containers
docker compose ps

# View logs
cd /opt/ven-api
docker compose logs -f ven-api

# Restart service
docker compose restart ven-api
```

### Manual Deployment

```bash
# On GitHub: Actions → Deploy to Staging → Run workflow

# Or on server
ssh deploy@staging-server
cd /opt/ven-api
git pull origin main
docker compose pull
docker compose up -d
```

### Rollback to Previous Version

```bash
ssh deploy@staging-server
cd /opt/ven-api

# Go back one commit
git reset --hard HEAD~1
docker compose restart

# Or go to specific commit
git reset --hard <commit-hash>
docker compose restart
```

---

## Troubleshooting

### Deployment Fails

**Check GitHub Actions logs:**
- Go to: Actions tab → Failed workflow → View logs

**Check staging server:**
```bash
ssh deploy@staging-server
cd /opt/ven-api
docker compose logs
docker compose ps
```

### Can't Connect to Server

**Verify SSH key:**
```bash
# Test connection
ssh -i ~/.ssh/staging-deploy deploy@staging-server

# Check GitHub secret is correct
# Settings → Secrets → STAGING_SSH_KEY
```

**Check firewall:**
```bash
# On server
sudo ufw status
sudo ufw allow 22/tcp  # Allow SSH
```

### Health Check Fails

```bash
ssh deploy@staging-server
cd /opt/ven-api

# Check service status
docker compose ps

# Check logs
docker compose logs ven-api

# Restart
docker compose restart ven-api

# Check health endpoint manually
curl http://localhost:8000/health
```

### Tests Fail

```bash
# Run tests locally first
go test -v ./...

# Check test logs in GitHub Actions
# Actions → Failed workflow → test job
```

---

## Monitoring

### Check App Health

```bash
# From anywhere
curl http://your-staging-server:8000/health

# Expected response:
{"status":"ok"}
```

### View Logs in Real-time

```bash
ssh deploy@staging-server
cd /opt/ven-api
docker compose logs -f ven-api
```

### Check Resource Usage

```bash
ssh deploy@staging-server

# CPU and memory
docker stats

# Disk space
df -h
```

---

## That's It! 🎉

Your simple staging CI/CD is ready. Every push to `main` automatically deploys to staging.

**Need help?** Open a GitHub issue or check the troubleshooting section above.
