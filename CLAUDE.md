# CLAUDE.md

The **content ingestion tool** for prarthana-service's catalogue. Despite the name "script", it is a long-running Go + Gin service (base path `ingestion/prarthana`, port `:8080`) with a small embedded HTML UI. It reads devotional content from **Zoho Sheet**, writes it into prarthana-service's MongoDB, indexes it into OpenSearch/Elasticsearch, and can generate shloka translations via OpenAI.

**It writes directly into another service's database.** `MongoConfig.Database` points at `prarthana_service_stage` (prod equivalent in the deployed config) and it writes the very collections prarthana-service reads: `prarthanas`, `deities`, `shloks`, `stotras`, `pooja_catalogue`, `prarthana_collections`. There is no API boundary — treat every change here as a change to prarthana-service's data contract, and check `~/go/src/prarthana-service/entity/` when you touch an entity.

## Content model & ingestion order
`shloks` (individual verses) → `stotras` (compositions of shloks) → `prarthanas` (playable prayer items) → `deities`. The four ingestion endpoints mirror that hierarchy and **must be run in dependency order** — a prarthana referencing a stotra that hasn't been ingested yet produces dangling references, not an error.

`entity/sheet.go` defines the spreadsheet row shapes; the other entity files are the Mongo documents. The Zoho sheet is the **source of truth for content**, Mongo is the derived store.

## API surface (`server/router.go`)
Under `ingestion/prarthana/v1`:
- `POST /shloks`, `POST /stotras`, `POST /prarthanas`, `POST /deities` — sheet → Mongo. All behind `ZohoAuthMiddleware`.
- `POST /shlokas-translation` — OpenAI-generated translations, also behind Zoho auth.
- `GET /deities-search`, `GET /prarthanas-search`, `GET /pooja-search` — reindex Mongo → OpenSearch. **These have no auth middleware.**

Plus `GET /ingestion/prarthana/prarthana.html`, a Gin-rendered operator UI (`ingestion/index.html`, loaded via `LoadHTMLGlob`) parameterised by `UIConfig.BackendHost`. That page is how content ops actually trigger ingestion — so the HTML template and the route paths are coupled; changing a path breaks the UI silently.

## The Zoho integration (`service/zoho/`)
- `ZohoAuthMiddleware` **refreshes the Zoho access token on every request** using `ZohoConfig.RefreshToken`, then puts it in the Gin context and on the request header as `zoho-access-token`. It returns 401 if the refresh fails. Consequence: a stale/revoked refresh token breaks all ingestion at once, and the middleware is doing a network call per request.
- `GetSheetData(sheetName, response)` posts `method=worksheet.records.fetch` to `https://sheet.zoho.in/api/v2/{SheetId}` — **one workbook (`ZohoConfig.SheetId`), many worksheets by name.** The worksheet names are the contract; renaming a tab in Zoho breaks ingestion.
- `SetSheetData` and `AddUUIDToSheet` **write back to the sheet** — after inserting a document, the generated UUID is written into the row so re-running ingestion updates instead of duplicating. That write-back is the idempotency mechanism; if it fails, the next run creates duplicates. `TmpId`/empty-`Id` checks in the Mongo repository are the other half of that logic.
- Scopes are `zohosheet.dataapi.update,zohosheet.dataapi.read` on the **`.in` Zoho region** (`accounts.zoho.in`, `sheet.zoho.in`) — not `.com`.

## Search indexing
`ESConfig` names the indices explicitly per environment: `staging-deity`, `staging-prarthana`, `staging-pooja`, against the **VPC Elasticsearch domain shared with panchanga-service** (`vpc-panchanga-service-*.ap-south-1.es.amazonaws.com`). Only reachable from inside the VPC. Note the index names are environment-prefixed in config, so a misconfigured deploy can reindex the wrong environment's index — check `ESConfig` before running a reindex endpoint.

`repository/es/prarthana` here writes the indices that prarthana-service's `repository/es/*` reads.

## Config (`configuration/config.json`, Viper — note the package is `configuration/`, not `config/`)
`ServerConfig`, `MongoConfig` (**20s timeout — longer than the fleet's usual 5s, because bulk ingestion queries are slow**), `ZohoConfig`, `ESConfig`, `OpenAIConfig`, `UIConfig`. All secrets are `XXXX` placeholders in the repo. Config `init()` panics if the file isn't found → run from repo root (the `LoadHTMLGlob("ingestion/*.html")` also requires it).

## Build / run
```bash
go build ./... && go vet ./...
go run .        # repo root only; then open /ingestion/prarthana/prarthana.html
```
Go 1.22.4 on an **old commons pseudo-version** (`v0.0.8-0.20241110151102-...`) — much older than the services; don't assume newer commons helpers exist here. Deps include `go-audio/wav` and `hajimehoshi/go-mp3`, so audio-duration inspection happens during prarthana ingestion. No tests.

Docker → ECR → ECS (`ap-south-1`) via `.github/workflows/aws.yml`. Current branch `fix/docker-buildkit-image-index`: the build needs `DOCKER_BUILDKIT=1 --provenance=false` — don't change those flags.

## Before running anything
1. Confirm which environment `MongoConfig.Database` and `ESConfig.*Index` point at. This tool has no dry-run mode.
2. Prod content writes follow the usual OIT rule — prepare the change and hand it over rather than pointing this at prod yourself.
3. Ingestion is not transactional: a mid-run failure leaves partially written documents plus rows whose UUID write-back didn't happen. Re-running is the recovery path, which is why the UUID write-back matters.

## Where to look
- `service/{shlok,stotra,prarthana,deity}_ingestion/` — one service per content type; sheet parsing + Mongo upsert.
- `service/search_ingestion/` — Mongo → ES reindex.
- `service/shlok_translation/` + `repository/openai/` — translation generation.
- `repository/mongo/prarthana_data/` — all the upsert logic, including the `Id`/`TmpId` dedupe.
