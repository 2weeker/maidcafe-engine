# Getting started

## Requirements

- Docker with the Compose plugin
- GNU Make
- Go 1.27 (only for local development: editor support, tests, `make vet`)

Everything runs in containers; you don't need Go or Postgres installed to run the site.

## Run it

```sh
make up
```

This builds the app image (templ generation happens inside the build), starts Postgres, waits
for it to be healthy, then starts the app. Open <http://localhost:8080>.

On startup the app:

1. loads `config.yaml` (see [configuration.md](configuration.md)),
2. connects to Postgres (retrying for up to 30 seconds),
3. applies any SQL migrations it hasn't applied yet,
4. listens on `:8080`.

## Make targets

| Target | What it does |
| --- | --- |
| `make help` | List targets (the default) |
| `make up` | Build and start everything in the background |
| `make down` | Stop everything (data is kept) |
| `make build` | Build the app image only |
| `make rebuild` | Rebuild and restart the app container (use after code changes); recreates the db container too if it's missing, without touching its data |
| `make restart` | Restart containers without rebuilding (use after editing `config.yaml`) |
| `make logs` | Follow the app's logs |
| `make ps` | Container status |
| `make psql` | Open a `psql` shell on the database |
| `make generate` | Run `templ generate` locally so editors see the generated Go |
| `make fmt` | `templ fmt` the templates and `gofmt` the Go code |
| `make vet` | Generate, then `go vet ./...` |
| `make clean` | Stop everything **and delete the database volume** |

## Common tasks

**Change board limits or page sizes:** edit `config.yaml`, then `make restart`. A typo or invalid
value stops the app with a clear error in `make logs`.

**Change the port:** `APP_PORT=9000 make up`.

**Start over with an empty database:** `make clean && make up`. This deletes every account and post.

**Look at the data:** `make psql`, then e.g. `\dt`, `SELECT username FROM users;`.

**Run the tests:**

```sh
make generate
go test ./...
```

## Running a second, throwaway copy

Useful for trying things without touching your real data. Run it from a separate directory as
a precaution: during development, the live `db` container disappeared a few times around the
start of a throwaway stack. The cause was never pinned down (it did not reproduce in isolation),
the `pgdata` volume was never affected, and `make rebuild` recreates the container.

```sh
mkdir -p /tmp/mctest
sed 's|build: \.|build: /path/to/maidcafe-engine|; s|\./config.yaml|/path/to/maidcafe-engine/config.yaml|' \
  docker-compose.yml > /tmp/mctest/docker-compose.yml
cd /tmp/mctest
APP_PORT=8090 docker compose -p mctest up -d --build
# ... http://localhost:8090 ...
APP_PORT=8090 docker compose -p mctest down -v   # removes it and its data
```

If the live database container ever goes missing (`make ps` shows only `app`), `make rebuild`
or `docker compose up -d` brings it back from the `maidcafe-engine_pgdata` volume with all data.

## Files you'll touch most

| Path | Purpose |
| --- | --- |
| `config.yaml` | Runtime configuration (mounted read-only into the container) |
| `app/` | Go code: config, database, models, controllers, middleware, routes |
| `web/templates/*.templ` | Views |
| `web/static/css/style.css` | All styling |
| `web/markup/` | The post formatter |
| `app/database/migrations/` | Schema changes, applied in filename order |
