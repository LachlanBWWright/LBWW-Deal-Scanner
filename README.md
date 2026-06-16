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

## Docker / Cloud Run deployment

The root `Dockerfile` builds the Vue control panel and the Go server into a single container. The Go server serves the built Vue assets from `/app/frontend/dist`, so one Cloud Run service is enough for the API and site.

Build and run locally:

```bash
docker compose up --build
```

Provide these environment variables through `.env` locally or Cloud Run service secrets:

```bash
TURSO_DATABASE_URL=libsql://your-database-org.turso.io
TURSO_AUTH_TOKEN=your-token
API_SECRET=your-api-secret
```

The Go server requires Turso. `DATABASE_URL` and local SQLite files are not used by the containerized app.

Deploying to Cloud Run only requires the container image and the environment variables above. Cloud Run sets `PORT` automatically, and the server listens on it.
