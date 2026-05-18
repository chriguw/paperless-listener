# Paperless Listener

Go webhook service that reads a Paperless document, derives a normalized title, and updates the document metadata.

## Required runtime config

The service resolves the config path in this order:

1. `PAPERLESS_CONFIG_PATH` (if set)
2. `config.json` (current working directory)
3. `../../config.json` (useful when running from `cmd/paperless-listener`)
4. `/app/config.json` (Docker)

1. Copy the example and adjust values:

```bash
cp configs/config.example.json config.json
```

2. Fill in your real Paperless host, token, and title mappings.

Run examples:

```bash
# from repository root
go run ./cmd/paperless-listener

# from cmd directory (will fallback to ../../config.json)
cd cmd/paperless-listener
go run .

# explicit override (works from any directory)
PAPERLESS_CONFIG_PATH=/Users/tgdwuch2/source/Paperless/config.json go run ./cmd/paperless-listener
```

## Make targets (recommended)

```bash
make help
make test
make lint
make build
make docker-build
make docker-run
make compose-up
make buildx-amd64
make buildx-arm64
make docker-buildx-local
make buildx-multiarch
make clean
```

Override image settings if needed:

```bash
make docker-build IMAGE=your-registry/paperless-listener TAG=latest
make buildx-multiarch IMAGE=your-registry/paperless-listener TAG=latest
make docker-buildx-local IMAGE=your-registry/paperless-listener TAG=latest MULTIARCH_OUTPUT=dist/paperless.oci.tar
```

## Build and run locally with Docker

```bash
docker build -t paperless-listener:local .
docker run --rm -p 8080:8080 -v "$(pwd)/config.json:/app/config.json:ro" paperless-listener:local
```

## Run with Docker Compose

```bash
docker compose up --build
```

## Build multi-arch image (AMD64 + ARM64)

Create and use a buildx builder once:

```bash
docker buildx create --name paperless-builder --use
docker buildx inspect --bootstrap
```

Build and push both platforms:

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t your-registry/paperless-listener:latest \
  --push \
  .
```

## Optional: build a single platform with buildx

```bash
docker buildx build --platform linux/amd64 -t paperless-listener:amd64 --load .
docker buildx build --platform linux/arm64 -t paperless-listener:arm64 --load .
```

## Prerequisites in Paperless-ngx

the following Environment variables needs to be set in Paperless-ngx that it works

```bash
PAPERLESS_WORKFLOW_ENABLED: "true"
PAPERLESS_URL: http://localhost:8081
```
## Configuration in Paperless-ngx
Create a workflow in paperless

- Name: Add Title
- Trigger: Document added
- Action: Webhook
  
  Webhook-Url http://<your-ip>:8080/webhook/1
  
  Parameter for Webhook-content: True
  
  Sent Webhook-Payload: True

  Webhook-parameter -> pfad: {doc-url}
