package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/relentlessworks/slugkit/internal/slug"
	"github.com/relentlessworks/slugkit/internal/store"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	Store *store.Store
}

// New creates a new API handler.
func New(s *store.Store) *Handler {
	return &Handler{Store: s}
}

// RegisterRoutes wires all endpoints onto the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.root)
	mux.HandleFunc("/help", h.help)
	mux.HandleFunc("/.well-known/agent.md", h.help)
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/slug", h.slugHandler)
	mux.HandleFunc("/validate", h.validateHandler)
	mux.HandleFunc("/mcp", h.mcp)
}

// --- Helpers ---

func wantsJSON(r *http.Request) bool {
	if r.URL.Query().Get("format") == "json" {
		return true
	}
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "application/json")
}

func writeError(w http.ResponseWriter, r *http.Request, msg, hint string, code int) {
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]string{
			"error": msg,
			"hint":  hint,
		})
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "error: %s | hint: %s\n", msg, hint)
}

func writeText(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(text))
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// --- Handlers ---

func (h *Handler) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	writeText(w, "slugkit — agentic-first URL slug generation service | hint: GET /help for the operating manual\n")
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeText(w, "ok\n")
}

// slugRequest holds the parameters for a slug generation request.
type slugRequest struct {
	Text      string `json:"text"`
	Separator string `json:"separator"`
	Lowercase *bool  `json:"lowercase"`
	MaxLength int    `json:"maxLength"`
}

func (h *Handler) slugHandler(w http.ResponseWriter, r *http.Request) {
	var text, separator string
	lowercase := true
	maxLength := 0

	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, r, "failed to read request body", "ensure the request body is valid", http.StatusBadRequest)
			return
		}

		var req slugRequest
		if err := json.Unmarshal(body, &req); err == nil && req.Text != "" {
			text = req.Text
			separator = req.Separator
			if req.Lowercase != nil {
				lowercase = *req.Lowercase
			}
			maxLength = req.MaxLength
		} else {
			values, err := url.ParseQuery(string(body))
			if err == nil {
				text = values.Get("text")
				separator = values.Get("separator")
				if lc := values.Get("lowercase"); lc == "false" {
					lowercase = false
				}
				if ml := values.Get("maxLength"); ml != "" {
					if n, err := strconv.Atoi(ml); err == nil && n >= 0 {
						maxLength = n
					}
				}
			} else {
				writeError(w, r, "invalid request body", "send JSON {\"text\":\"...\"} or form-encoded text=...", http.StatusBadRequest)
				return
			}
		}
	} else if r.Method == http.MethodGet {
		text = r.URL.Query().Get("text")
		separator = r.URL.Query().Get("separator")
		if lc := r.URL.Query().Get("lowercase"); lc == "false" {
			lowercase = false
		}
		if ml := r.URL.Query().Get("maxLength"); ml != "" {
			if n, err := strconv.Atoi(ml); err == nil && n >= 0 {
				maxLength = n
			} else {
				writeError(w, r, "invalid maxLength parameter", "maxLength must be a non-negative integer, e.g. /slug?text=Hello&maxLength=20", http.StatusBadRequest)
				return
			}
		}
	} else {
		writeError(w, r, "method not allowed", "use GET /slug?text=... or POST /slug with JSON body", http.StatusMethodNotAllowed)
		return
	}

	if text == "" {
		writeError(w, r, "missing text parameter", "provide the text to slugify, e.g. /slug?text=Hello+World or POST /slug with JSON {\"text\":\"Hello World\"}", http.StatusBadRequest)
		return
	}

	h.Store.IncrSlug()

	opts := slug.Options{
		Separator: separator,
		Lowercase: lowercase,
		MaxLength: maxLength,
	}
	result := slug.Generate(text, opts)

	if wantsJSON(r) {
		writeJSON(w, map[string]interface{}{
			"slug": result,
			"text": text,
		})
		return
	}

	writeText(w, result+"\n")
}

func (h *Handler) validateHandler(w http.ResponseWriter, r *http.Request) {
	var slugStr, separator string

	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, r, "failed to read request body", "ensure the request body is valid", http.StatusBadRequest)
			return
		}

		var req struct {
			Slug      string `json:"slug"`
			Separator string `json:"separator"`
		}
		if err := json.Unmarshal(body, &req); err == nil && req.Slug != "" {
			slugStr = req.Slug
			separator = req.Separator
		} else {
			values, err := url.ParseQuery(string(body))
			if err == nil {
				slugStr = values.Get("slug")
				separator = values.Get("separator")
			} else {
				writeError(w, r, "invalid request body", "send JSON {\"slug\":\"...\"} or form-encoded slug=...", http.StatusBadRequest)
				return
			}
		}
	} else if r.Method == http.MethodGet {
		slugStr = r.URL.Query().Get("slug")
		separator = r.URL.Query().Get("separator")
	} else {
		writeError(w, r, "method not allowed", "use GET /validate?slug=... or POST /validate with JSON body", http.StatusMethodNotAllowed)
		return
	}

	if slugStr == "" {
		writeError(w, r, "missing slug parameter", "provide the slug to validate, e.g. /validate?slug=hello-world", http.StatusBadRequest)
		return
	}

	h.Store.IncrValid()

	valid := slug.Validate(slugStr, separator)

	if wantsJSON(r) {
		writeJSON(w, map[string]interface{}{
			"slug":   slugStr,
			"valid":  valid,
		})
		return
	}

	writeText(w, fmt.Sprintf("slug=%s valid=%t\n", slugStr, valid))
}

func (h *Handler) help(w http.ResponseWriter, r *http.Request) {
	writeText(w, helpText)
}

const helpText = `slugkit — agentic-first URL slug generation service

Generate URL-safe slugs from arbitrary text. Supports unicode transliteration
(accented chars, Greek, Cyrillic), configurable separators, max length, and
slug validation. No database needed — pure stateless computation.

AUTH:
  If SLUGKIT_API_KEY is set, send Authorization: Bearer <key> header.
  If not set, no auth is required.

ENDPOINTS:

  GET /slug?text=Hello+World
    Generate a slug from text.
    Params: text (required), separator (- _ .), lowercase (true/false), maxLength (int)
    Example: GET /slug?text=Hello+World → hello-world

  POST /slug
    Generate a slug from JSON body.
    Body: {"text":"Hello World","separator":"-","lowercase":true,"maxLength":0}
    Example response: {"slug":"hello-world","text":"Hello World"}

  GET /validate?slug=hello-world
    Validate whether a string is a proper slug.
    Params: slug (required), separator (- _ .)
    Example: GET /validate?slug=hello-world → slug=hello-world valid=true

  POST /validate
    Validate a slug from JSON body.
    Body: {"slug":"hello-world","separator":"-"}

  GET /help  (or /.well-known/agent.md)
    This operating manual.

  GET /health
    Health check.

  POST /mcp
    MCP (Model Context Protocol) endpoint for chat client integrations.

RESPONSE FORMAT:
  Plain text by default. JSON via Accept: application/json header or ?format=json.

ERRORS:
  error: message | hint: what to do next

CONFIG:
  SLUGKIT_ADDR     Listen address (default :8472)
  SLUGKIT_API_KEY  API key for auth (optional, no auth if empty)
  -addr            Override listen address
  -api-key          Override API key

BUILD:
  make build    # CGO_ENABLED=0, single static binary
  make test     # go test -race
  make vet      # go vet
  make run      # build + run
`
