# slugkit

> Agentic-first URL slug generation service. Generate URL-safe slugs from text with unicode transliteration, configurable separators, and validation. Plain text API, agent-driven, single Go binary.

## Quick Start

```bash
# Build
make build

# Run
./slugkit

# Generate a slug
curl "http://localhost:8472/slug?text=Hello+World"

# Generate a slug via POST (JSON)
curl -X POST http://localhost:8472/slug \
  -H "Content-Type: application/json" \
  -d '{"text":"Hello World"}'

# Generate with options
curl "http://localhost:8472/slug?text=Hello+World&separator=_&maxLength=10"

# Validate a slug
curl "http://localhost:8472/validate?slug=hello-world"

# Get JSON output
curl -H "Accept: application/json" "http://localhost:8472/slug?text=Hello+World"
```

## API Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/help` | Operating manual (also at `/.well-known/agent.md`) |
| GET | `/health` | Health check |
| GET | `/slug` | Generate a slug (params: `text`, `separator`, `lowercase`, `maxLength`) |
| POST | `/slug` | Generate a slug (JSON body: `text`, `separator`, `lowercase`, `maxLength`) |
| GET | `/validate` | Validate a slug (params: `slug`, `separator`) |
| POST | `/validate` | Validate a slug (JSON body: `slug`, `separator`) |
| POST | `/mcp` | MCP (Model Context Protocol) endpoint |

### Response Format

**Plain text (default):**

`/slug` returns the slug directly:
```
hello-world
```

`/validate` returns a one-line result:
```
slug=hello-world valid=true
```

**JSON:** Send `Accept: application/json` header or `?format=json` query param.

```json
{"slug":"hello-world","text":"Hello World"}
```

### Errors

Errors include a hint for self-correction:

```
error: missing text parameter | hint: provide the text to slugify, e.g. /slug?text=Hello+World
```

### Unicode Support

slugkit transliterates accented characters, Greek, and Cyrillic:

| Input | Output |
|-------|--------|
| `Héllo Wörld` | `hello-world` |
| `Café Münster` | `cafe-munster` |
| `Привет мир` | `privet-mir` |
| `Αθήνα` | `athina` |
| `Über` | `uber` |

## Configuration

| Source | Variable | Default | Description |
|--------|----------|---------|-------------|
| Env | `SLUGKIT_ADDR` | `:8472` | Listen address |
| Env | `SLUGKIT_API_KEY` | (empty) | API key for auth (no auth if empty) |
| Flag | `-addr` | `:8472` | Override listen address |
| Flag | `-api-key` | (empty) | Override API key |

Priority: defaults < env vars < flags.

## MCP Integration

slugkit speaks Model Context Protocol at `POST /mcp` for chat client integrations (Claude, Cursor, etc.).

**Tools:**
- `generate_slug` — Generate a URL-safe slug from text
- `validate_slug` — Validate whether a string is a proper URL slug

## Build

```bash
make build    # CGO_ENABLED=0, single static binary
make test     # go test -race
make vet      # go vet
make run      # build + run
```

## Design Principles

- **Agent IS the interface** — No UI, no SDK. The API is the product.
- **Plain text by default** — Token-cheap, grepable, one record per line.
- **Instructive errors** — Every 4xx includes a hint for self-correction.
- **Self-documenting** — `GET /help` returns the full operating manual.
- **Single static binary** — Go, zero external dependencies, CGO_ENABLED=0.
- **Zero config** — Runs out of the box with sensible defaults.

## License

MIT
