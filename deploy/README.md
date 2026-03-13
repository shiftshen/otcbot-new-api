# OTCBOT deployment

This folder contains the deployment path for `api.otcbot.com`.

## Server-side script

Run `deploy/otcbot-deploy.sh` on the server to:

- pull the latest GitHub code
- rebuild the Docker image
- replace the running `new-api-master` container

The script defaults to the current OTCBOT server layout and can be overridden with environment variables when needed.

## GitHub Actions

The workflow in `.github/workflows/deploy-otcbot.yml` is designed for push-to-deploy through a self-hosted GitHub Actions runner on the OTCBOT server.

Runner labels expected by the workflow:

- `self-hosted`
- `linux`
- `x64`
- `otcbot`

## First-time server bootstrap

Clone your GitHub repository to `/home/happy/new-api-repo`, register the self-hosted runner, and make the deploy script executable:

```bash
cd /home/happy
git clone <your-github-repo-url> new-api-repo
cd /home/happy/new-api-repo
chmod +x deploy/otcbot-deploy.sh
```
