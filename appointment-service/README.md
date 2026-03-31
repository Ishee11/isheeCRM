# appointment-service CI/CD

## What is in the repo now

- `.env.example` with all required runtime variables
- `docker-compose-app.yml` for server deployment from a published image
- `docker-compose.yml` for local/dev with PostgreSQL
- `deploy.sh` as the single deployment entrypoint on the server
- `.github/workflows/deploy.yml` for build, push and deploy

## Runtime model

The image is built in GitHub Actions and pushed to GHCR with two tags:

- commit SHA
- `main`

The server keeps its own `.env` file and deploys a specific image tag.

## Server setup

1. Clone the repository on the server.
2. Go to `appointment-service`.
3. Copy `.env.example` to `.env`.
4. Fill real values for:
   - `APP_IMAGE`
   - `IMAGE_TAG`
   - `DB_HOST`
   - `DB_PORT`
   - `DB_USER`
   - `DB_PASSWORD`
   - `DB_NAME`
5. Make the script executable:

```bash
chmod +x deploy.sh
```

6. If you deploy only the app container, make sure the external Docker network from `docker-compose-app.yml` already exists:

```bash
docker network create appointment-service_app-network
```

## Manual deploy on the server

```bash
cd appointment-service
IMAGE_TAG=<git-sha> ./deploy.sh
```

## GitHub Actions secrets

Create these repository secrets:

- `SSH_HOST`
- `SSH_PORT`
- `SSH_USER`
- `SSH_PRIVATE_KEY`
- `DEPLOY_PATH`
- `REGISTRY_USERNAME`
- `REGISTRY_PASSWORD`

`DEPLOY_PATH` is the absolute path to the repository on the server, for example:

```text
/opt/isheeCRM
```

## Notes

- Keep production secrets only on the server, never in the repository.
- The workflow assumes the repository already exists on the server.
- The deploy step updates the repo on the server, writes the current image tag into `.env` and runs `deploy.sh`.
- For GHCR, `REGISTRY_USERNAME` is usually your GitHub username and `REGISTRY_PASSWORD` should be a PAT with package read access.
