# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

An MCP (Model Context Protocol) server that exposes the [Mochi Cards](https://mochi.cards) REST API as tools for Claude and other MCP clients. Built in Go on the official [go-sdk](https://github.com/modelcontextprotocol/go-sdk). Communicates over stdio; reads the Mochi API key from `MOCHI_API_KEY`.

## Commands

```sh
go test ./...                              # all tests
go test ./internal/mochi -run TestListCards -v   # single test
go test -race ./...                        # what CI runs
go vet ./...
gofmt -l .                                 # must print nothing (CI fails otherwise)

go build -o mochi-card-mcp .               # local binary
MOCHI_API_KEY=xxx ./mochi-card-mcp         # run (serves on stdio)
```

Release builds stamp the version via ldflags: `-ldflags "-X main.version=1.2.3"`. The `version` var in `main.go` is a hardcoded fallback used only for un-stamped local builds; the published version comes from the git tag (see CI/CD).

## Architecture

Three layers, dependencies pointing downward:

- **`main.go`** — entrypoint. Reads `MOCHI_API_KEY`, constructs the client and server, runs `srv.Run(ctx, &mcp.StdioTransport{})`.
- **`internal/server`** — MCP tool definitions and handlers. Depends on the unexported `cardAPI` interface, not the concrete client, so handlers are tested with an in-memory fake.
- **`internal/mochi`** — a thin, transport-agnostic REST client. Knows nothing about MCP.

### Key design decisions (read these before changing behavior)

- **Card sides live in one field.** Mochi stores both sides of a card in a single Markdown `content` field separated by a line containing only `---`. The **client stays "pure"**: its `CreateCard`/`UpdateCard` take raw `content`. The **server layer owns the front/back abstraction**: tool inputs use `name` (front) + `content` (back), and `internal/mochi/content.go`'s `JoinSides`/`SplitSides` convert. `mochi_update_card` first fetches the card, splits it, overrides only the provided side, and rejoins — so a caller can update one side without clobbering the other.

- **`cardAPI` interface is the seam for tests.** When adding a new tool, you must add its method to the `cardAPI` interface in `internal/server/server.go` AND to the `fakeAPI` in `server_test.go`, or the package won't compile. The end-to-end test (`TestServerToolsRegistered`) asserts an exact tool count — update its `want` map when adding/removing tools.

- **The README "API coverage" table is the source of truth for endpoint support.** It maps every Mochi API endpoint to ✅ Supported (with its tool name) or ❌ Unsupported (with a tracking issue link). **Always update this table before committing/opening a PR** whenever you add, remove, or rename a tool, or change which endpoints are covered — move the row to Supported and reference the implementing tool. Unimplemented endpoints each have an open GitHub issue; close/link it from the PR.

- **Search is client-side.** Mochi has no search endpoint. `SearchCards` fetches one page of `/cards` and filters by case-insensitive substring. A page can return zero matches while still returning a non-empty `bookmark` to continue — this is expected, not a bug.

- **Pagination = opaque bookmarks.** `/cards` and `/decks` return `{"docs": [...], "bookmark": "..."}`; pass a non-empty `bookmark` back to get the next page. `mochi_list_cards`/`mochi_search_cards`/`mochi_list_decks` thread it through input and output.

- **The due-cards endpoint is the exception.** `/due/` returns `{"cards": [...]}` (note: `cards`, not `docs`) and is **not paginated** — no `bookmark`, no `limit`. `mochi_list_due_cards` deliberately exposes only `deck_id?` and `date?`. Keep it that way; don't add pagination params the API ignores.

### MCP SDK conventions

- Tools are registered with `mcp.AddTool(server, &mcp.Tool{...}, handler)`. Handler signature: `func(ctx, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error)`.
- Return `nil, out, nil` on success — the SDK auto-populates both structured content and JSON text content from the typed `Out`.
- Return a plain `error` for tool failures; the SDK wraps it as a `CallToolResult` with `IsError=true` (don't hand-build error results).
- Input/output structs use `json` + `jsonschema` struct tags; the schema is generated from them. Field types matter: avoid `json.RawMessage` for pass-through fields (it generates a byte-array schema) — use `any` instead (see `Card.CreatedAt`/`UpdatedAt`).

### Mochi API specifics

Reference: <https://mochi.cards/docs/api/> (the "Get all due cards" endpoint is at <https://mochi.cards/docs/api/#get-all-due-cards>). Consult it before changing request shapes or adding endpoints.

- Base URL `https://app.mochi.cards/api`. Auth is HTTP Basic with the API key as the **username and an empty password** (`req.SetBasicAuth(key, "")`).
- Request bodies use **hyphenated keys**: `deck-id`, `parent-id`.
- Non-2xx responses become a typed `*mochi.APIError{StatusCode, Body}`.

## CI/CD

- **`.github/workflows/ci.yml`** — gofmt check, vet, build, `go test -race` on push/PR to `main`.
- **`.github/workflows/release.yml`** — triggered by `v*` tags. Cross-compiles macOS (intel/arm64), Linux (amd64/arm64), Windows (amd64), packages with README/LICENSE, generates checksums, and publishes a GitHub release. Version comes from the tag (`${GITHUB_REF_NAME#v}`).

To release: merge to `main`, then `git tag -a vX.Y.Z <commit> && git push origin vX.Y.Z`.

## Repo conventions

- Workflow is branch + PR into `main`; releases are cut by tagging the merge commit.
- Before committing or opening a PR that changes tool coverage, update the README "API coverage" table (see the design note above).
- Commit messages are Conventional-Commits style (`feat:`, `feat(mochi):`, `ci:`, `docs:`, `chore:`).
- This repo is configured to sign commits via a 1Password SSH agent, which requires interactive desktop approval and fails in non-interactive sessions. When committing programmatically here, use `--no-gpg-sign` (and `-c tag.gpgSign=false` for tags). Existing history is intentionally unsigned for this reason.
