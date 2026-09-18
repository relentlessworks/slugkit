package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/relentlessworks/slugkit/internal/slug"
)

// mcpRequest represents a Model Context Protocol request.
type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

// mcpResponse represents a Model Context Protocol response.
type mcpResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *mcpError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (h *Handler) mcp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, r, "method not allowed", "POST /mcp with JSON-RPC body", http.StatusMethodNotAllowed)
		return
	}

	var req mcpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, mcpResponse{
			JSONRPC: "2.0",
			Error:   &mcpError{Code: -32700, Message: "parse error"},
		})
		return
	}

	switch req.Method {
	case "initialize":
		writeJSON(w, mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"serverInfo": map[string]string{
					"name":    "slugkit",
					"version": "0.1.0",
				},
				"capabilities": map[string]interface{}{
					"tools": map[string]bool{"listChanged": false},
				},
			},
		})

	case "tools/list":
		writeJSON(w, mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"tools": []map[string]interface{}{
					{
						"name":        "generate_slug",
						"description": "Generate a URL-safe slug from text",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"text":      map[string]string{"type": "string", "description": "Text to slugify"},
								"separator": map[string]string{"type": "string", "description": "Separator: - _ or . (default: -)"},
								"lowercase": map[string]string{"type": "boolean", "description": "Lowercase the slug (default: true)"},
								"maxLength": map[string]string{"type": "integer", "description": "Max slug length (0 = unlimited)"},
							},
							"required": []string{"text"},
						},
					},
					{
						"name":        "validate_slug",
						"description": "Validate whether a string is a proper URL slug",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"slug":      map[string]string{"type": "string", "description": "Slug to validate"},
								"separator": map[string]string{"type": "string", "description": "Separator: - _ or . (default: -)"},
							},
							"required": []string{"slug"},
						},
					},
				},
			},
		})

	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			writeJSON(w, mcpResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &mcpError{Code: -32602, Message: "invalid params"},
			})
			return
		}

		switch params.Name {
		case "generate_slug":
			text, _ := params.Arguments["text"].(string)
			if text == "" {
				writeJSON(w, mcpResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Error:   &mcpError{Code: -32602, Message: "missing text argument"},
				})
				return
			}
			separator, _ := params.Arguments["separator"].(string)
			lowercase := true
			if lc, ok := params.Arguments["lowercase"].(bool); ok {
				lowercase = lc
			}
			maxLength := 0
			if ml, ok := params.Arguments["maxLength"].(float64); ok {
				maxLength = int(ml)
			}

			h.Store.IncrSlug()
			result := slug.Generate(text, slug.Options{
				Separator: separator,
				Lowercase: lowercase,
				MaxLength: maxLength,
			})

			writeJSON(w, mcpResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"content": []map[string]string{
						{"type": "text", "text": result},
					},
				},
			})

		case "validate_slug":
			slugStr, _ := params.Arguments["slug"].(string)
			if slugStr == "" {
				writeJSON(w, mcpResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
					Error:   &mcpError{Code: -32602, Message: "missing slug argument"},
				})
				return
			}
			separator, _ := params.Arguments["separator"].(string)

			h.Store.IncrValid()
			valid := slug.Validate(slugStr, separator)

			writeJSON(w, mcpResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"content": []map[string]string{
						{"type": "text", "text": fmt.Sprintf("valid=%t", valid)},
					},
				},
			})

		default:
			writeJSON(w, mcpResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &mcpError{Code: -32601, Message: "unknown tool: " + params.Name},
			})
		}

	default:
		writeJSON(w, mcpResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &mcpError{Code: -32601, Message: "method not found: " + req.Method},
		})
	}
}
