# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

**sybot** is a Kakao chatbot backend written in Go that surfaces Lost Ark (로스트아크) character data. It receives webhook calls from the Kakao chatbot platform, parses slash commands, queries the Lost Ark public API (with Redis caching), and returns formatted text responses.

## Commands

```bash
# Build
go build -o sybot ./cmd

# Run (requires env vars)
LOSTARK_API_KEY=<key> REDIS_ADDR=localhost:6379 go run ./cmd

# Test
go test ./...

# Single package test
go test ./internal/command/...

# Docker build
docker build -t sybot .

# Lint (if golangci-lint installed)
golangci-lint run
```

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `LOSTARK_API_KEY` | yes | — | Lost Ark developer API key |
| `REDIS_ADDR` | no | `localhost:6379` | Redis address for caching |

### 로컬 개발 시 주입 방법

앱 자체는 `.env`를 읽지 않으므로 shell에서 주입한다.

```bash
# 1) .env.example 복사 후 키 입력
cp .env.example .env
# .env 편집: LOSTARK_API_KEY=실제키

# 2) 실행 시 env 파일 소싱
set -a && source .env && set +a
go run ./cmd

# 또는 한 줄로
env $(grep -v '^#' .env | xargs) go run ./cmd
```

`.env`는 `.gitignore`에 등록되어 있어 커밋되지 않는다.

### 환경별 주입 방식

| 환경 | 방식 |
|---|---|
| 로컬 | `.env` 파일 소싱 후 `go run` |
| Docker | `docker run --env-file .env sybot` |
| Kubernetes | `sybot-secrets` Secret → `LOSTARK_API_KEY` envFrom |

## Lost Ark API 참고 문서

공식 API 문서: `lostarkapi-docs.md` (프로젝트 루트)

핵심 스펙:
- Base URL: `https://developer-lostark.game.onstove.com`
- 인증: `Authorization: bearer {JWT}` — `bearer` 다음 공백 필수, 중괄호 없이 토큰 그대로
- Rate Limit: **분당 100 요청**, 초과 시 429. 응답 헤더 `X-RateLimit-Remaining` / `X-RateLimit-Reset` 확인
- 점검 중: 503 반환. 추가 요청 자제

## Architecture

Request flow: **Kakao platform → `POST /webhook/kakao` → `KakaoHandler` → `command.Parse` → `lostark.Client` → Lost Ark API**

- **`cmd/main.go`** — wires dependencies and starts the HTTP server on `:8080`. Also exposes `GET /healthz`.
- **`internal/handler/kakao.go`** — decodes `KakaoRequest`, enforces per-user rate limiting via `ratelimit.Limiter`, dispatches to the appropriate command handler, and encodes `KakaoResponse` (Kakao i Open Builder v2 format).
- **`internal/command/parser.go`** — parses slash-prefixed utterances (`/캐릭터 <name>` or `/character <name>`). Returns `ErrUnknownCommand` for unrecognized input.
- **`internal/lostark/client.go`** — calls `GET /characters/{name}/siblings` on the Lost Ark API, picks the matching entry from the siblings array, and caches the result in Redis for 5 minutes (key: `character:<name>`).
- **`internal/lostark/types.go`** — `CharacterInfo` (used for API responses and cache) and `ArmoryProfile` (defined but not yet used by any command).
- **`internal/cache/redis.go`** — thin JSON-serializing wrapper around `go-redis/v9`. `Get`/`Set`/`Ping`.
- **`internal/ratelimit/limiter.go`** — in-memory per-user token bucket (10 req/s, burst 100) using `golang.org/x/time/rate`. Limiters are never evicted; long-running deployments with many unique user IDs will grow unbounded.

## Deployment

Kubernetes manifests are in `k8s/` (raw YAML) and `helm/sybot/` (Helm chart). The deployment runs 2 replicas. `LOSTARK_API_KEY` is injected from a Kubernetes Secret named `sybot-secrets` (key: `lostark-api-key`). Redis is expected at `redis:6379` in-cluster.

## Adding New Commands

1. Add a new `CmdType` constant in `internal/command/parser.go` and a case in `Parse`.
2. Add a corresponding `case` in `KakaoHandler.Handle` in `internal/handler/kakao.go`.
3. Add any new Lost Ark API methods to `internal/lostark/client.go` and types to `internal/lostark/types.go`.
