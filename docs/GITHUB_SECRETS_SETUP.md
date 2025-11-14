# GitHub Secrets & Actions Setup Guide

This guide shows you how to add Cloudflare D1 credentials to GitHub for CI/CD workflows.

## Why You Need GitHub Secrets

GitHub Secrets allow you to:
- Store sensitive credentials (API tokens, passwords) securely
- Use them in GitHub Actions workflows without exposing them
- Deploy your Docker container to production automatically
- Run tests against your real D1 database in CI/CD

## Required Secrets for This Project

For Docker deployment with Cloudflare D1, you need these secrets:

| Secret Name | Description | How to Get It |
|-------------|-------------|---------------|
| `CLOUDFLARE_ACCOUNT_ID` | Your Cloudflare account ID | See instructions below |
| `CLOUDFLARE_API_TOKEN` | API token with D1 permissions | See instructions below |
| `CLOUDFLARE_DB_NAME` | Name of your D1 database | Usually `cloudflaredb` |
| `DOCKERHUB_USERNAME` | (Optional) Your Docker Hub username | For pushing images |
| `DOCKERHUB_TOKEN` | (Optional) Docker Hub access token | For pushing images |

## Step-by-Step: Adding Secrets to GitHub

### 1. Get Your Cloudflare Credentials

**A. Get Your Account ID:**

1. Go to https://dash.cloudflare.com/
2. Look at the URL or right sidebar
3. Copy your Account ID (looks like: `a1b2c3d4e5f6789...`)

**B. Create API Token:**

1. Go to https://dash.cloudflare.com/profile/api-tokens
2. Click **"Create Token"**
3. Choose **"Edit Cloudflare Workers"** template
   - Or create custom token with permissions:
     - Account > D1 > Edit
     - Account > Workers Scripts > Edit (if deploying Workers)
4. Click **"Continue to summary"**
5. Click **"Create Token"**
6. **IMPORTANT:** Copy the token immediately! You can only see it once!

**C. Get Database Name:**

Run this command in your terminal:
```bash
npx wrangler d1 list
```

Output will show your database name (usually `cloudflaredb`).

### 2. Add Secrets to GitHub Repository

#### Option A: Via GitHub Web Interface (Easiest)

1. **Go to your repository on GitHub**
   - Example: `https://github.com/your-username/cloudflaredb`

2. **Click on "Settings"**
   - It's in the top menu of your repository
   - If you don't see it, you might not have admin access

3. **Navigate to Secrets**
   - In the left sidebar, click **"Secrets and variables"**
   - Then click **"Actions"**

4. **Add Each Secret**

   For each secret, do the following:

   a. Click **"New repository secret"** button (green button on the right)

   b. Enter the secret details:
   - **Name:** `CLOUDFLARE_ACCOUNT_ID`
   - **Value:** Your actual account ID (paste it here)

   c. Click **"Add secret"**

   d. Repeat for each secret:

   ```
   Name: CLOUDFLARE_ACCOUNT_ID
   Value: a1b2c3d4e5f6789abcdef...
   ```

   ```
   Name: CLOUDFLARE_API_TOKEN
   Value: xyz789abc123def456...
   ```

   ```
   Name: CLOUDFLARE_DB_NAME
   Value: cloudflaredb
   ```

5. **Verify Secrets Were Added**
   - You should see all secrets listed
   - ⚠️ You won't be able to view the secret values again (only update them)

#### Option B: Via GitHub CLI (Advanced)

If you have GitHub CLI installed:

```bash
# Login to GitHub CLI (if not already)
gh auth login

# Add each secret
gh secret set CLOUDFLARE_ACCOUNT_ID --body "your_account_id_here"
gh secret set CLOUDFLARE_API_TOKEN --body "your_api_token_here"
gh secret set CLOUDFLARE_DB_NAME --body "cloudflaredb"
```

## Visual Guide: Adding Secrets Step-by-Step

### Step 1: Navigate to Settings
```
GitHub Repository → Settings (top menu)
```

### Step 2: Find Secrets Section
```
Settings (left sidebar) → Secrets and variables → Actions
```

### Step 3: Add New Secret
```
Click "New repository secret" button
```

### Step 4: Fill in Secret Details
```
┌─────────────────────────────────────────┐
│ Name *                                  │
│ ┌─────────────────────────────────────┐ │
│ │ CLOUDFLARE_ACCOUNT_ID               │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ Secret *                                │
│ ┌─────────────────────────────────────┐ │
│ │ a1b2c3d4e5f6789abcdef123456789012   │ │
│ └─────────────────────────────────────┘ │
│                                         │
│        [Add secret]                     │
└─────────────────────────────────────────┘
```

### Step 5: Repeat for All Secrets

After adding all secrets, you should see:

```
Repository secrets

CLOUDFLARE_ACCOUNT_ID        Updated 1 minute ago
CLOUDFLARE_API_TOKEN         Updated 1 minute ago
CLOUDFLARE_DB_NAME          Updated 1 minute ago
```

## Using Secrets in GitHub Actions

Once secrets are added, you can use them in your workflows:

### Example Workflow File

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Build Docker Image
        run: docker build -t cloudflaredb:latest .

      - name: Run Migrations
        env:
          CLOUDFLARE_ACCOUNT_ID: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
          CLOUDFLARE_DB_NAME: ${{ secrets.CLOUDFLARE_DB_NAME }}
        run: |
          npm install -g wrangler
          npx wrangler d1 execute ${{ secrets.CLOUDFLARE_DB_NAME }} \
            --remote \
            --file=migrations/001_create_users_table.sql

      - name: Deploy Docker Container
        env:
          DATABASE_DSN: d1://${{ secrets.CLOUDFLARE_ACCOUNT_ID }}:${{ secrets.CLOUDFLARE_API_TOKEN }}@${{ secrets.CLOUDFLARE_DB_NAME }}
        run: |
          docker run -d \
            -p 8080:8080 \
            -e DATABASE_DRIVER=cfd1 \
            -e DATABASE_DSN="$DATABASE_DSN" \
            cloudflaredb:latest
```

### How to Access Secrets in Workflows

Use this syntax:
```yaml
${{ secrets.SECRET_NAME }}
```

Example:
```yaml
env:
  CLOUDFLARE_ACCOUNT_ID: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
```

## Docker Compose with GitHub Actions

If you want to use docker-compose in GitHub Actions:

```yaml
- name: Deploy with Docker Compose
  env:
    CLOUDFLARE_ACCOUNT_ID: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
    CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
    CLOUDFLARE_DB_NAME: ${{ secrets.CLOUDFLARE_DB_NAME }}
  run: |
    docker-compose up -d
```

The secrets are automatically available as environment variables!

## Security Best Practices

### ✅ DO:

1. **Use Repository Secrets for sensitive data**
   - API tokens
   - Passwords
   - Database credentials

2. **Rotate tokens regularly**
   - Create new API token every 3-6 months
   - Update the secret on GitHub

3. **Use least privilege**
   - Only grant D1:Edit permission
   - Don't use account-wide tokens

4. **Use environment-specific secrets**
   - Different secrets for staging vs production
   - Use GitHub Environments for this

### ❌ DON'T:

1. **Never commit secrets to git**
   - Don't put tokens in `.env` files that are committed
   - Always use `.gitignore`

2. **Never log secrets**
   - Don't `echo` or `print` secret values in workflows
   - GitHub will try to mask them, but don't risk it

3. **Never share secrets**
   - Each team member should have their own API token
   - Don't share tokens via Slack/email

4. **Never use secrets in pull requests from forks**
   - Secrets aren't available to fork PRs (security feature)
   - This is intentional to prevent secret theft

## Environment Secrets (Advanced)

For staging vs production:

### 1. Create Environments

In GitHub:
1. Go to **Settings → Environments**
2. Click **"New environment"**
3. Create environments:
   - `production`
   - `staging`
   - `development`

### 2. Add Environment-Specific Secrets

1. Click on an environment (e.g., "production")
2. Add secrets specific to that environment
3. Example:
   - Production: `CLOUDFLARE_DB_NAME=cloudflaredb-prod`
   - Staging: `CLOUDFLARE_DB_NAME=cloudflaredb-staging`

### 3. Use in Workflows

```yaml
jobs:
  deploy-production:
    runs-on: ubuntu-latest
    environment: production  # ← Uses production secrets

    steps:
      - name: Deploy
        env:
          DB_NAME: ${{ secrets.CLOUDFLARE_DB_NAME }}  # Gets prod value
        run: echo "Deploying to $DB_NAME"
```

## Troubleshooting

### Secret Not Found Error

**Error:** `Error: Secret CLOUDFLARE_ACCOUNT_ID not found`

**Solutions:**
1. Check the secret name is spelled correctly (case-sensitive!)
2. Make sure you added it as a Repository Secret (not Environment Secret)
3. Verify you're an admin on the repository
4. Try re-adding the secret

### Secret Value Looks Wrong

**Error:** Deployment fails with authentication error

**Solutions:**
1. Verify you copied the entire token (no spaces/newlines)
2. Check the token hasn't expired
3. Verify token has D1 permissions
4. Try creating a new token and updating the secret

### Can't See Settings Tab

**Problem:** No "Settings" option in repository

**Solution:** You need admin access. Ask the repository owner to:
- Add you as an admin, or
- Add the secrets themselves

### Secrets Not Working in Pull Requests

**Problem:** Secrets are undefined in PR workflows

**Solution:** This is normal for PRs from forks (security feature). Options:
1. Only use secrets on `push` events to main branch
2. Use GitHub Environments with required reviewers
3. Have maintainers run workflows manually

## Updating Secrets

To update a secret:

1. Go to **Settings → Secrets and variables → Actions**
2. Click on the secret name
3. Click **"Update secret"**
4. Enter new value
5. Click **"Update secret"**

Or via CLI:
```bash
gh secret set CLOUDFLARE_API_TOKEN --body "new_token_value"
```

## Verifying Secrets Work

Create a simple workflow to test:

```yaml
# .github/workflows/test-secrets.yml
name: Test Secrets

on:
  workflow_dispatch  # Manually trigger this

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Check Secrets Exist
        env:
          ACCOUNT_ID: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
          DB_NAME: ${{ secrets.CLOUDFLARE_DB_NAME }}
        run: |
          if [ -z "$ACCOUNT_ID" ]; then
            echo "❌ CLOUDFLARE_ACCOUNT_ID is not set"
            exit 1
          fi
          if [ -z "$DB_NAME" ]; then
            echo "❌ CLOUDFLARE_DB_NAME is not set"
            exit 1
          fi
          echo "✅ All secrets are configured"
          # Don't print the actual values!
          echo "Account ID length: ${#ACCOUNT_ID}"
          echo "DB Name: $DB_NAME"
```

Run this workflow manually:
1. Go to **Actions** tab
2. Select "Test Secrets" workflow
3. Click **"Run workflow"**
4. Check the output

## Quick Reference

| Task | Location |
|------|----------|
| Add secrets | Settings → Secrets and variables → Actions |
| View secrets | Settings → Secrets and variables → Actions (values are hidden) |
| Update secret | Click secret name → Update secret |
| Delete secret | Click secret name → Remove secret |
| Use in workflow | `${{ secrets.SECRET_NAME }}` |

## Example: Complete Setup Checklist

- [ ] Get Cloudflare Account ID from dashboard
- [ ] Create Cloudflare API token with D1 permissions
- [ ] Get D1 database name from `wrangler d1 list`
- [ ] Go to GitHub repository → Settings
- [ ] Navigate to Secrets and variables → Actions
- [ ] Add `CLOUDFLARE_ACCOUNT_ID` secret
- [ ] Add `CLOUDFLARE_API_TOKEN` secret
- [ ] Add `CLOUDFLARE_DB_NAME` secret
- [ ] Create `.env` file locally (not committed)
- [ ] Test GitHub Actions workflow
- [ ] Verify deployment works

## Need Help?

If you're stuck:

1. Check the [Troubleshooting](#troubleshooting) section
2. Verify you have admin access to the repository
3. Make sure token has D1 permissions
4. Try the test workflow above
5. Check GitHub Actions logs for specific errors

## Additional Resources

- [GitHub Secrets Documentation](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Cloudflare API Tokens](https://dash.cloudflare.com/profile/api-tokens)
- [Docker Deployment Guide](DOCKER_D1_SETUP.md)
