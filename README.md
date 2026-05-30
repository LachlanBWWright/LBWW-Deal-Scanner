This project is split into two packages:

- `server/`: the Fastify + Discord bot server
- `frontend/`: the Vue control panel

Run them separately:

- `cd server && npm install`, then `npm run dev` for the server
- `cd frontend && npm install`, then `npm run dev` for the Vue UI

The server exposes the local API and Swagger UI, and the frontend consumes the generated OpenAPI client types in `frontend/src/api/`.

Scheduled scanner mode can be enabled with server `.env` variables:

- `SCHEDULED_SCANNER_MODE=true` runs the scanner without starting the API or Discord bot.
- `SCHEDULED_SCANNER_DURATION_MS=300000` controls how long the scheduled scanner runs before stopping. The default is 5 minutes.

The scanner stores its resume counters and scheduled runtime totals in the database so the next scheduled start continues from the previous scan state.
