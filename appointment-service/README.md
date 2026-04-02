# appointment-service CI/CD

## What is in the repo now

- `.env.example` with all required runtime variables
- `docker-compose-app.yml` for server deployment from a published image
- `docker-compose.yml` for local/dev with PostgreSQL
- `docker-compose-postgres.yml` for production PostgreSQL 17.6 with persistent host storage
- `backup-db.sh` for logical PostgreSQL backups
- `restore-db.sh` for restoring PostgreSQL from a dump
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

7. For production PostgreSQL, set these values in `.env`:
   - `POSTGRES_IMAGE=postgres:17.6-bookworm`
   - `POSTGRES_BIND_ADDRESS=127.0.0.1`
   - `POSTGRES_DATA_DIR=/var/lib/isheecrm/postgres`
   - `POSTGRES_BACKUP_DIR=/var/backups/isheecrm`
   - `POSTGRES_BACKUP_RETENTION_DAYS=7`

## Production database flow

### 1. First boot of PostgreSQL on the server

Use a bind mount so data survives container recreation:

```bash
cd appointment-service
docker network create appointment-service_app-network || true
docker compose --env-file .env -f docker-compose-postgres.yml up -d
```

The database files will live on the host in `POSTGRES_DATA_DIR`, not inside the container.

### 2. Restore from the dump in the repository root

The repository root currently contains a PostgreSQL custom-format dump:

```text
../isheecrm_20260329030001.dump
```

This dump was made by PostgreSQL `17.6`, so the server database container should stay on PostgreSQL `17.x`. Restore it like this:

```bash
cd appointment-service
./restore-db.sh ../isheecrm_20260329030001.dump
```

The script will:

- stop the app container if it is running
- start PostgreSQL 17.6
- drop and recreate the target database
- restore from the dump
- run `ALTER DATABASE ... REFRESH COLLATION VERSION`
- run `REINDEX DATABASE`
- start the app again

### 3. Normal operation

For normal app deploys, keep using:

```bash
cd appointment-service
IMAGE_TAG=<git-sha> ./deploy.sh
```

This only updates the app container. The PostgreSQL data stays in `POSTGRES_DATA_DIR` and survives `docker compose down` or container recreation as long as you do not remove the host directory.

## Backups

Create a backup:

```bash
cd appointment-service
./backup-db.sh
```

The backup is written to `POSTGRES_BACKUP_DIR` in PostgreSQL custom format (`.dump`), which is suitable for `pg_restore`.

Recommended production policy:

- keep backups outside the container filesystem
- copy backups to remote object storage or another server
- test restore regularly on a clean PostgreSQL 17 container
- never rely on the Docker container itself as a backup

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
