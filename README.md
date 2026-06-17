# mochi-card-mcp

A [Model Context Protocol](https://modelcontextprotocol.io) (MCP) server for the
[Mochi Cards](https://mochi.cards) spaced-repetition app, written in Go using the
official [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk).

It lets Claude (and any other MCP client) create, read, update, delete, and
search your Mochi flashcards and decks. It is a Go port of
[louis030195/mochi-mcp](https://github.com/louis030195/mochi-mcp) and exposes the
same set of tools.

## Tools

| Tool | Description | Arguments |
| --- | --- | --- |
| `mochi_list_cards` | List cards, optionally filtered by deck | `deck_id?`, `bookmark?`, `limit?` |
| `mochi_list_due_cards` | List cards due for review | `deck_id?`, `date?` |
| `mochi_get_card` | Get a single card | `card_id` |
| `mochi_create_card` | Create a card (`name` = front, `content` = back) | `name`, `content`, `deck_id?` |
| `mochi_update_card` | Update a card's front and/or back | `card_id`, `name?`, `content?` |
| `mochi_delete_card` | Delete a card | `card_id` |
| `mochi_search_cards` | Search cards by content (case-insensitive) | `query`, `bookmark?` |
| `mochi_list_decks` | List decks | `bookmark?` |
| `mochi_get_deck` | Get a single deck | `deck_id` |
| `mochi_create_deck` | Create a deck | `name`, `parent_id?` |
| `mochi_update_deck` | Update a deck's name, parent, and/or sort | `deck_id`, `name?`, `parent_id?`, `sort?` |
| `mochi_delete_deck` | Delete a deck | `deck_id` |
| `mochi_list_templates` | List card templates | `bookmark?` |
| `mochi_get_template` | Get a single template | `template_id` |
| `mochi_create_template` | Create a template | `name`, `content`, `fields?` |

Mochi stores both sides of a card in a single Markdown field separated by a
`---` line. This server hides that detail: `name` is the front and `content` is
the back. When updating, you can change just one side and the other is preserved.

List and search results are paginated. When a response includes a non-empty
`bookmark`, pass it back in the next call to fetch the following page.

## API coverage

Status of each [Mochi API](https://mochi.cards/docs/api/) endpoint. Keep this
table in sync when adding or removing tools.

| Resource | Endpoint | Status | Tool / Issue |
| --- | --- | --- | --- |
| Cards | `GET /cards` | ✅ Supported | `mochi_list_cards` |
| Cards | `GET /cards/:id` | ✅ Supported | `mochi_get_card` |
| Cards | `POST /cards` | ✅ Supported | `mochi_create_card` |
| Cards | `POST /cards/:id` | ✅ Supported | `mochi_update_card` |
| Cards | `DELETE /cards/:id` | ✅ Supported | `mochi_delete_card` |
| Cards | _(client-side search)_ | ✅ Supported | `mochi_search_cards` |
| Cards | `POST /cards/:card-id/attachments/:filename` | ❌ Unsupported | [#8](https://github.com/dudleycarr/mochi-card-mcp/issues/8) |
| Cards | `DELETE /cards/:card-id/attachments/:filename` | ❌ Unsupported | [#12](https://github.com/dudleycarr/mochi-card-mcp/issues/12) |
| Decks | `GET /decks` | ✅ Supported | `mochi_list_decks` |
| Decks | `POST /decks` | ✅ Supported | `mochi_create_deck` |
| Decks | `GET /decks/:id` | ✅ Supported | `mochi_get_deck` |
| Decks | `POST /decks/:id` | ✅ Supported | `mochi_update_deck` |
| Decks | `DELETE /decks/:id` | ✅ Supported | `mochi_delete_deck` |
| Templates | `GET /templates` | ✅ Supported | `mochi_list_templates` |
| Templates | `GET /templates/:id` | ✅ Supported | `mochi_get_template` |
| Templates | `POST /templates` | ✅ Supported | `mochi_create_template` |
| Due cards | `GET /due`, `GET /due/:deck-id` | ✅ Supported | `mochi_list_due_cards` |

## Installation

### Download a release

Grab the archive for your platform from the
[Releases](https://github.com/dudleycarr/mochi-card-mcp/releases) page (macOS
Intel/Apple Silicon, Linux amd64/arm64, Windows amd64), extract it, and place the
`mochi-card-mcp` binary on your `PATH`.

### Build from source

Requires Go 1.26+.

```sh
go install github.com/dudleycarr/mochi-card-mcp@latest
```

or, from a checkout:

```sh
go build -o mochi-card-mcp .
```

## Configuration

The server needs a Mochi API key, supplied via the `MOCHI_API_KEY` environment
variable. Create one under **Account Settings → API** in Mochi (a Mochi Pro
subscription is required for API access).

### Claude Desktop / Claude Code

Add the server to your MCP configuration (e.g. `claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "mochi": {
      "command": "mochi-card-mcp",
      "env": {
        "MOCHI_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

Or, with the Claude Code CLI:

```sh
claude mcp add mochi --env MOCHI_API_KEY=your-api-key-here -- mochi-card-mcp
```

The server communicates over stdio.

## Development

```sh
go test ./...        # run the tests
go vet ./...         # static checks
gofmt -l .           # formatting (should print nothing)
```

The project layout:

- `internal/mochi` — a thin client for the Mochi Cards REST API.
- `internal/server` — the MCP server wiring the API to tools.
- `main.go` — the stdio entrypoint.

## License

See [LICENSE](LICENSE).
