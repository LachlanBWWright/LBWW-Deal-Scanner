This project is split into two packages:

- `server/`: the Fastify + Discord bot server
- `frontend/`: the Vue control panel

Run them separately:

- `cd server && npm install`, then `npm run dev` for the server
- `cd frontend && npm install`, then `npm run dev` for the Vue UI

The server exposes the local API and Swagger UI, and the frontend consumes the generated OpenAPI client types in `frontend/src/api/`.
