# igozhang

Single-binary Go web server: **article site** (Markdown + SQLite FTS), **directory browse/download**, **HTTP(S) fetch and browser upload** into a transfer folder, **text stash**, **MP3 streaming** with cover metadata, plus small games and a typing trainer. Default listen **`:3003`**. No extra runtime; HTML/JS/CSS are `//go:embed`. Production site is usually **`https://igozhang.cn/`** (no `BASE_PATH`).

Full specification (rules, limits, every API): **[README.md](README.md)** (Chinese). Deploy: **[deploy/deploy.md](deploy/deploy.md)**. Release checks: **[checklist.md](checklist.md)**.

**Scale (approx., excluding tests):** `cmd/igozhang/` ~16k Go / ~11k embedded JS / ~4.3k HTML; `internal/` ~2.4k Go. Image is `CGO_ENABLED=0` static binary; **`assets/` is not in the image** (hostPath / `DATA_ROOT`). Go **1.22**. CI: `go vet` + `go test`.

---

## Features

| Area | Notes |
|------|--------|
| **Articles** | `GET /` category + pager (`[doc] home_latest` per page); `GET /{slug}`; Mermaid on demand; `/igoadmin` CRUD, batch tags/categories, ZIP export/import (merge). |
| **Browse** | `/fs` after IGO key auth; list dirs whose names start with **`f`** (e.g. `fs`); public `GET /files/…` and `GET /preview/…`; authenticated `DELETE /api/file`. |
| **Transfer** | `POST /api/fetch` (NDJSON progress), `POST /api/upload`; 2 GiB cap; 7-day TTL; 9 GiB dir cap. |
| **Stash** | `GET/POST/DELETE /api/stash` with ETag / 409 conflict; 200 KiB / 12 h TTL. |
| **MP3** | Stream `/stream/mp3/…`; same-origin **gate cookie** (`GET /api/mp3/gate`); dual-slot player + Media Session. |
| **Auth** | Shared IGO key for `/ops`, `/igoadmin`, `/fs`, and `[secret-auth]` articles. Memory token (no cookie). |
| **Ops** | Remote SSH shell, visitor board, IP blacklist UI. IP **whitelist** (`secret07/ip-whitelist.conf`) gates `/ops` `/igoadmin` `/fs` `/api/auth` when non-empty. |
| **Games** | `/gomoku`, `/xiangqi`, `/texas`, `/zhijian`, `/mermaid`. |

---

## Environment

| Variable | Default | Meaning |
|----------|---------|---------|
| `LISTEN` | `:3003` | Bind address |
| `DATA_ROOT` | `/igozhang.cn` | Data root (`assets/`, `fs/`, `secret07/`, …) |
| `TRANSFER_SUBDIR` | `fs` | Transfer directory name |
| `BASE_PATH` | (empty) | HTTP prefix; leave empty for site root |
| `DOC_ASSETS_SEED` | `assets` | Seed `DATA_ROOT/assets/` when `md/` is empty |

Limits and TTLs live in **`DATA_ROOT/secret07/igo.conf`** (source: `cmd/igozhang/igo.conf`; `go generate ./cmd/igozhang/...` copies to `deploy/secret07/`). Restart the process after edits (no hot reload).

---

## Build and run

```bash
go build -o igozhang ./cmd/igozhang
DATA_ROOT=/tmp/igozhang.cn ./igozhang
# http://127.0.0.1:3003/
go vet ./... && go test ./...
```

Docker (BuildKit):

```bash
DOCKER_BUILDKIT=1 docker build -t krccr.ccs.tencentyun.com/igokorea/igozhang:$(date +%Y%m%d-%H%M%S) .
```

K8s manifests: `deploy/k8s/` (namespace `igo`, port **3003**, hostPath **`/data/igozhang.cn`**). Copy repo `assets/` onto the data volume on first deploy.

---

## Layout

| Path | Role |
|------|------|
| `cmd/igozhang/` | Server, routes, embedded UI, `igo.conf` |
| `internal/` | docindex, markdown, ipgeo, mermaid 11.4.1 |
| `assets/` | Markdown, pictures, MP3, phrase/word lists |
| `deploy/` | K8s, Nginx sample, `igo.conf` copy |
| `scripts/` | `go generate` helpers |
