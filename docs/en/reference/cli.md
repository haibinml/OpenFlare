# Commands and Scripts

You will learn: Common commands for starting, building, testing, installing, and uninstalling the OpenFlare Server, management console frontend, Agent, Swagger, and documentation site.

## Server

Start from source:

```bash
cd openflare_server
export SESSION_SECRET='replace-with-random-string'
export SQLITE_PATH='./openflare.db'
export LOG_LEVEL='info'
go run .
```

Specify listening port and log directory:

```bash
go run . --port 3000 --log-dir ./logs
```

Test:

```bash
cd openflare_server
GOCACHE=/tmp/openflare-go-cache go test ./...
```

## Frontend

Development:

```bash
cd openflare_server/web
pnpm install
pnpm dev
```

Build static artifacts:

```bash
cd openflare_server/web
pnpm build
```

Checks:

```bash
cd openflare_server/web
pnpm lint
pnpm typecheck
pnpm test
```

## Agent

Run from source:

```bash
cd openflare_agent
go run ./cmd/agent -config /path/to/agent.json
```

Compile:

```bash
cd openflare_agent
go build -o openflare-agent ./cmd/agent
```

Test:

```bash
cd openflare_agent
GOCACHE=/tmp/openflare-go-cache go test ./...
```

## Install Agent

```bash
curl -fsSL https://raw.githubusercontent.com/Rain-kl/OpenFlare/main/scripts/install-agent.sh | bash -s -- \
  --server-url http://your-server:3000 \
  --agent-token YOUR_AGENT_TOKEN
```

## Uninstall Agent

```bash
curl -fsSL https://raw.githubusercontent.com/Rain-kl/OpenFlare/main/scripts/uninstall-agent.sh | bash
```

## Swagger

Regenerate Swagger documentation:

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4
cd openflare_server
swag init -g main.go -o docs
```

## Docs

Local preview:

```bash
cd docs
pnpm dev
```

Build:

```bash
cd docs
pnpm build
```
