default:
    @just --list

# Run the backend dev server (regenerates sqlc + TS client first).
backend: gen
    encore run

test:
    encore test ./...

lint:
    go vet ./...
    @if [ -n "$(gofmt -l .)" ]; then echo "gofmt found unformatted files:"; gofmt -l .; exit 1; fi

gen: gen-sqlc gen-client

gen-sqlc:
    sqlc generate

gen-client:
    encore gen client muenchner-see-buddy-an4i --output=./frontend/src/client.ts --env=local --lang=typescript
