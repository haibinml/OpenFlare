# Deployment

This page summarizes the OpenFlare deployment baseline, integration flow, upgrade entry points, and Agent install scripts.

## Requirements

Server:

* Go 1.24+
* Node.js 18+
* Writable SQLite directory or reachable PostgreSQL instance

Agent:

* Go 1.24+
* Writable Agent data directory
* Local mode requires `openresty -t` and `openresty -s reload`
* Docker mode requires Docker access

## Docker Compose

PostgreSQL is recommended for production:

```yaml
services:
  postgres:
    image: postgres:17-alpine
    restart: unless-stopped
    environment:
      POSTGRES_DB: openflare
      POSTGRES_USER: openflare
      POSTGRES_PASSWORD: replace-with-strong-password
    volumes:
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U openflare -d openflare"]
      interval: 10s
      timeout: 5s
      retries: 5

  openflare:
    image: ghcr.io/rain-kl/openflare:latest
    container_name: openflare
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
    ports:
      - "3000:3000"
    environment:
      SESSION_SECRET: replace-with-random-string
      SQLITE_PATH: /data/openflare.db
      DSN: postgres://openflare:replace-with-strong-password@postgres:5432/openflare?sslmode=disable
      GIN_MODE: release
      LOG_LEVEL: info
    volumes:
      - openflare-data:/data

volumes:
  postgres-data:
  openflare-data:
```

```bash
docker compose up -d
```

Open `http://localhost:3000`. The default account is `root` / `123456`; change the password immediately.

## Agent Install

Using `discovery_token`:

```bash
curl -fsSL https://raw.githubusercontent.com/Rain-kl/OpenFlare/main/scripts/install-agent.sh | bash -s -- \
  --server-url http://your-server:3000 \
  --discovery-token YOUR_DISCOVERY_TOKEN
```

Using node-specific `agent_token`:

```bash
curl -fsSL https://raw.githubusercontent.com/Rain-kl/OpenFlare/main/scripts/install-agent.sh | bash -s -- \
  --server-url http://your-server:3000 \
  --agent-token YOUR_AGENT_TOKEN
```

Supported options include `--server-url`, `--discovery-token`, `--agent-token`, `--install-dir`, `--repo`, and `--no-service`.

## Validation

1. Prepare `agent_token` or `discovery_token` in the console.
2. Start Agent and confirm the node is online.
3. Add an enabled reverse proxy site.
4. Publish and activate a new version.
5. Confirm Agent pulls, validates, reloads, and reports the result.

## Uninstall Agent

```bash
curl -fsSL https://raw.githubusercontent.com/Rain-kl/OpenFlare/main/scripts/uninstall-agent.sh | bash
```
