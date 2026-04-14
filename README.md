# PikaBiasalah Bot

This repository defines the KataML configuration (`bot.yml`) and Go backend powering PikaBiasalah, a Telegram-style chatbot that greets users, registers their names, and returns Pokemon information via the hosted `hc4k4ccck4kwskc00gw44ggk.triki.cloud` APIs.

## Prerequisites

- [Go 1.20+](https://go.dev/doc/install) (to run the API server)
- [Node.js](https://nodejs.org/) + `kata-cli` (for KataML flows)
- MySQL (the Go server expects a database; use Docker or a local instance)

## Backend setup

1. Copy or set the usual MySQL environment variables (`DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME`). The server falls back to `root`, `3306`, `localhost`, and `pokebot` if values are missing.
2. Run `go mod tidy` if dependencies are new, then `go run .` to start the API server. Logs confirm the database connection and expose `/api/v1/register`, `/api/v1/users`, `/api/v1/users/:id`, `/api/v1/pokemon`, and `/health`.
3. Point the KataML bot actions at `https://hc4k4ccck4kwskc00gw44ggk.triki.cloud` (already wired in `bot.yml`). If you run your own backend, update the URIs accordingly before deploying.

## KataML development flow

1. Install `kata-cli` globally: `npm install -g kata-cli`.
2. Use `kata push` to push `bot.yml` changes, then `kata create-deployment` and `kata update-environment <version>` to publish the flows.
3. Interact with the bot via `kata console` (choose Development, send `/start`, follow the prompts).

## Testing & validation

- `curl https://hc4k4ccck4kwskc00gw44ggk.triki.cloud/health` should return the hosted service health, providing an additional manual verification point when running the bot.
- Use `kata console` to step through `/start`, name confirmation, and the Pokemon query flow, checking that `pokeLost` triggers for unknown names.

## Notes

- The Go backend calls the public `pokeapi.co` API; ensure outbound HTTPS access is allowed when running locally or in CI.
- `bot.yml` remains the source of truth for bot behavior—adjust the flows there, then redeploy via the standard Kata commands listed above.
