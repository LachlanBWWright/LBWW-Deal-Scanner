This project is split into these active packages:

- `goserver/`: the Go API, Discord bot, scanner runtime, and static Vue host
- `frontend/`: the Vue control panel
- `archived-ts/`: the archived TypeScript scanner/server implementation

Run the active Go backend and Vue frontend together:

```bash
./run-go.sh
```

The Go server exposes the local API and serves the built Vue app in the container. The frontend consumes the generated OpenAPI client types in `frontend/src/api/`.

Scheduled scanner mode can be enabled with `.env` variables:

- `SCHEDULED_SCANNER_MODE=true` runs the scanner without starting the API or Discord bot.
- `SCHEDULED_SCANNER_DURATION_MS=300000` controls how long the scheduled scanner runs before stopping. The default is 5 minutes.

The scanner stores its resume counters and scheduled runtime totals in the database so the next scheduled start continues from the previous scan state.

Reset the Turso database to the current Go model schema:

```bash
set -a
source .env
set +a
go -C goserver run cmd/reset-db/main.go
```

## Docker / VM deployment

The root `Dockerfile` builds the Vue control panel and the Go server into a single container. The Go server serves the built Vue assets from `/app/frontend/dist`, so one Cloud Run service is enough for the API and site.

Build and run locally:

```bash
docker compose up --build
```

Provide these environment variables through `.env` locally or Cloud Run service secrets:

```bash
DATABASE_URL=file:/var/lib/dealscanner/dealscanner.db
API_SECRET=your-api-secret
```

The Go server uses `DATABASE_URL`. Use a `file:` URL for local SQLite; Turso remains supported temporarily when `DATABASE_URL` is a `libsql://` or `https://` URL and `TURSO_AUTH_TOKEN` is set.

The production workflow builds the image in GitHub Actions, pushes it to GitHub Container Registry, SSHes into the VM, and runs Docker Compose to pull and restart the container. This keeps Docker builds off the VM and avoids running a GitHub Actions runner on the production host.

Create these repository secrets:

```bash
ORACLE_HOST=your-vm-ip-or-hostname
ORACLE_USER=opc
ORACLE_PORT=22
ORACLE_SSH_KEY=private-key-that-can-ssh-to-the-vm
```

On the VM, install Docker with the Compose plugin. Oracle Linux uses `dnf`, not `apt`:

```bash
sudo dnf install -y dnf-utils curl
sudo dnf config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo dnf install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo systemctl enable --now docker
```

The CI deployment creates `/opt/dealscanner`, constructs `app.env` from individual GitHub secrets and variables, and authenticates to GitHub Container Registry with the workflow's short-lived `GITHUB_TOKEN`. Required application secrets are `API_SECRET`, `TURSO_DATABASE_URL`, `TURSO_AUTH_TOKEN`, and `DISCORD_TOKEN`. Scanner flags, Discord IDs, channel IDs, and role IDs are configured as GitHub repository or environment variables.

Manual deploys can be triggered from GitHub Actions with the `Deploy Oracle VM` workflow. Pushes to `master` deploy automatically after the image is built and pushed.
