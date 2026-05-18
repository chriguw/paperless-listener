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
  
  Webhook-Url http://paperless-listener.local:8080/webhook/1
  
  Parameter for Webhook-content: True
  
  Sent Webhook-Payload: True

  Webhook-parameter -> pfad: {doc-url}

## config.json
in config json the mapping takes place

paperlessHost is the servicename from you paperless instance in the docker-compose, note that you need there the internal port

token is the token from your paperless

the amountCustomfieldId is the id of the currentcy_customfield which you have to define in paperless if you like to have also the ammount of a bill 
if you don't have this customfield just let a 0 there

```json
  "paperlessHost": "http://webserver:8000",
  "token": "replace-with-api-token-from-paperless",
  "amountCustomFieldId": 0,
```

titles

the script reads line by line from the ocr comment of the document, if the line contains the word on the left side, title1 to title4 will contain the word on the right side

```json
  "title1": {
      "Rechtsschutzversicherung": "Rechtsschutzversicherung",
      "Haushaltversicherung": "Haushaltversicherung",
      "Motorfahrzeugversicherung": "Motorfahrzeugversicherung",
      "Gebäudeversicherung": "Gebäudeversicherug",
      "CH00 0000 0000 0000 0000 0": "Lohnkonto"
    }
```

```json
  "title2": {
      "Prämienrechnung": "Prämienrechnung",
      "Leistungsabrechnung": "Leistungsabrechnung",
      "Endabrechnung":"Endabrechnung"
    }
```

```json
  "title3": {
      "Vertragsänderung": "Vertragsänderung",
      "Rechnungs-Nr.":"Rechnung",
      "Akontorechnung vom": "Akontorechnung"
    }
```
here you could par example also use a contractnumber and map it to a person
```json
  "title4": {
      "00 0000 000": "Klaus",
      "11 1111 111": "Berta"
    }
```
yearKeyword

here it tries to extract the year from a date which is placed on the right side of one of this words

par example "Kontoabschluss per 02.02.2026" would return 2026

```json
  "yearKeywords": [
    "Kalenderjahr",
    "Steuer ab",
    "Kontoabschluss per"
  ]
```
amountKeywords

here it tries to extract an amount placed on the right side of one of this words

par example "Total zu Ihren Lasten CHF 15.50" would return 15.50

```json
  "amountKeywords": [
    "Total zu Ihren Lasten CHF",
    "zu Ihren Lasten"
  ]
```

The final title of the document would then be as follows

title1_title2_title3_title4_year

if title2 is empty, it would be 

title1_title3_title4_year

there is an additional function which tries to extract year an month of the dates in the document

if a line in the document contains 2. Mai 2026 it will return 2026_05, if this is found it will replace the year from above

if you did changes on your configfile, you can reload the config like this

```
curl --request GET \
  --url http://localhost:8080/reload
```

to trigger a single document you can do it like this while the 760 in the body needs to be the document id

```
curl --request POST \
  --url http://localhost:8080/webhook/1 \
  --header 'content-type: application/json' \
  --data '{"pfad":"http://localhost:8081/documents/760/"}

```

to add it to your existing docker compose file, here is an example

please note that in the example the mac image is used, there is also an amd64 image

```
paperless-listener:
   image: mydoidfortest/paperless-listener:arm64-1.2
   container_name: listener
   ports:
     - "8080:8080"
   volumes:
     - ./config.json:/app/config.json:ro
   restart: unless-stopped
   networks:
    default:
      aliases:
        - paperless-listener.local
```

the docker images can be found here 
https://hub.docker.com/r/mydoidfortest/paperless-listener/tags

or you can create them by yourself as described above