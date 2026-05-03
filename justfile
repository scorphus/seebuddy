default:
    @just --list

# Run the backend dev server (regenerates sqlc + TS client first).
backend: gen
    encore run

# Run the frontend dev server (regenerates sqlc + TS client first).
frontend: gen
    cd frontend && npm run dev

# One-time install of frontend deps.
frontend-install:
    cd frontend && npm install

# Deploy the Cloudflare Worker that triggers Poll every 15 min.
worker-deploy:
    cd worker && npx wrangler deploy

# One-time install of worker deps.
worker-install:
    cd worker && npm install

test:
    encore test ./...

lint:
    go vet ./...
    @if [ -n "$(gofmt -l .)" ]; then echo "gofmt found unformatted files:"; gofmt -l .; exit 1; fi

gen: gen-sqlc gen-client

gen-sqlc:
    sqlc generate

gen-client:
    encore gen client seebuddy-y3mi --output=./frontend/src/client.ts --env=local --lang=typescript
