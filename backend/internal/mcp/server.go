package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"fraud-detection/internal/api/dto"
	"fraud-detection/internal/service"
)

// JSON-RPC 2.0 types

type request struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
}

type response struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Result  any              `json:"result,omitempty"`
	Error   *rpcError        `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP protocol types

type initializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    serverCapabilities `json:"capabilities"`
	ServerInfo      serverInfo         `json:"serverInfo"`
}

type serverCapabilities struct {
	Tools map[string]any `json:"tools"`
}

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema inputSchema `json:"inputSchema"`
}

type inputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

type property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type toolsListResult struct {
	Tools []tool `json:"tools"`
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type toolResult struct {
	Content []contentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type contentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Tool argument types

type getRecentFraudsArgs struct {
	HoursBack int `json:"hours_back"`
	Limit     int `json:"limit"`
}

type checkUserStatusArgs struct {
	UserID string `json:"user_id"`
}

// Server

type Server struct {
	svc *service.TransactionService
}

func NewServer(svc *service.TransactionService) *Server {
	return &Server{svc: svc}
}

func (s *Server) Run() error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}

		// Notifications (no id) — no response required
		if req.ID == nil {
			continue
		}

		resp := s.handle(req)
		data, _ := json.Marshal(resp)
		fmt.Printf("%s\n", data)
	}
	return scanner.Err()
}

func (s *Server) handle(req request) response {
	switch req.Method {
	case "initialize":
		return response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: initializeResult{
				ProtocolVersion: "2024-11-05",
				Capabilities:    serverCapabilities{Tools: map[string]any{}},
				ServerInfo:      serverInfo{Name: "fraud-detection", Version: "1.0.0"},
			},
		}

	case "tools/list":
		return response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  toolsListResult{Tools: s.tools()},
		}

	case "tools/call":
		var params toolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return errResponse(req.ID, -32602, "invalid params")
		}
		result, err := s.callTool(params)
		if err != nil {
			return errResponse(req.ID, -32603, err.Error())
		}
		return response{JSONRPC: "2.0", ID: req.ID, Result: result}

	default:
		return errResponse(req.ID, -32601, "method not found")
	}
}

func (s *Server) tools() []tool {
	return []tool{
		{
			Name:        "get_recent_frauds",
			Description: "Son N saatteki fraud işlemlerini döner. Fraud analizi ve anomali tespiti için kullanılır.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]property{
					"hours_back": {
						Type:        "integer",
						Description: "Kaç saat geriye bakılacağı (varsayılan: 24, maks: 720)",
					},
					"limit": {
						Type:        "integer",
						Description: "Döndürülecek maksimum kayıt sayısı (varsayılan: 10, maks: 100)",
					},
				},
			},
		},
		{
			Name:        "check_user_status",
			Description: "Bir kullanıcının güven skoru, risk seviyesi ve işlem istatistiklerini döner.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]property{
					"user_id": {
						Type:        "string",
						Description: "Sorgulanacak kullanıcının ID'si",
					},
				},
				Required: []string{"user_id"},
			},
		},
	}
}

func (s *Server) callTool(params toolCallParams) (toolResult, error) {
	switch params.Name {
	case "get_recent_frauds":
		return s.getRecentFrauds(params.Arguments)
	case "check_user_status":
		return s.checkUserStatus(params.Arguments)
	default:
		return toolResult{}, fmt.Errorf("unknown tool: %s", params.Name)
	}
}

func (s *Server) getRecentFrauds(raw json.RawMessage) (toolResult, error) {
	var args getRecentFraudsArgs
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &args); err != nil {
			return toolResult{}, fmt.Errorf("invalid arguments: %w", err)
		}
	}

	if args.HoursBack <= 0 {
		args.HoursBack = 24
	}
	if args.HoursBack > 720 {
		args.HoursBack = 720
	}
	if args.Limit <= 0 {
		args.Limit = 10
	}
	if args.Limit > 100 {
		args.Limit = 100
	}

	now := time.Now().UTC()
	from := now.Add(-time.Duration(args.HoursBack) * time.Hour)

	page := dto.PageRequest{Page: 1, Limit: int64(args.Limit)}
	result, err := s.svc.GetFraudsBetween(context.Background(), from, now, page)
	if err != nil {
		return toolResult{}, fmt.Errorf("get frauds: %w", err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return toolResult{
		Content: []contentItem{{Type: "text", Text: string(data)}},
	}, nil
}

func (s *Server) checkUserStatus(raw json.RawMessage) (toolResult, error) {
	var args checkUserStatusArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return toolResult{}, fmt.Errorf("invalid arguments: %w", err)
	}
	if args.UserID == "" {
		return toolResult{}, fmt.Errorf("user_id is required")
	}

	score, err := s.svc.GetUserTrustScore(context.Background(), args.UserID)
	if err != nil {
		return toolResult{}, fmt.Errorf("get trust score: %w", err)
	}

	data, _ := json.MarshalIndent(score, "", "  ")
	return toolResult{
		Content: []contentItem{{Type: "text", Text: string(data)}},
	}, nil
}

func errResponse(id *json.RawMessage, code int, msg string) response {
	return response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: msg},
	}
}
