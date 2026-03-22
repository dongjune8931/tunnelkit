package tunnel

import (
	"encoding/json"
	"fmt"
	"time"
)

// 메시지 타입 상수
const (
	MsgTypeWelcome    = "welcome"
	MsgTypeRequest    = "request"
	MsgTypeResponse   = "response"
	MsgTypePing       = "ping"
	MsgTypePong       = "pong"
	MsgTypeDisconnect = "disconnect"
)

// TunnelMessage는 서버↔CLI 간 WebSocket 메시지다.
type TunnelMessage struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// WelcomePayload는 연결 환영 메시지 페이로드다.
type WelcomePayload struct {
	SessionID   string `json:"session_id"`
	AccessToken string `json:"access_token"`
	TunnelURL   string `json:"tunnel_url"`
}

// TunnelRequest는 서버→CLI 방향 HTTP 요청 정보다.
type TunnelRequest struct {
	RequestID string            `json:"request_id"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"` // base64
}

// TunnelResponse는 CLI→서버 방향 HTTP 응답 정보다.
type TunnelResponse struct {
	RequestID  string            `json:"request_id"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"` // base64
}

// BuildRequestMsg는 TunnelRequest를 WebSocket 메시지로 직렬화한다.
func BuildRequestMsg(req *TunnelRequest) (*TunnelMessage, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	return &TunnelMessage{
		Type:      MsgTypeRequest,
		RequestID: req.RequestID,
		Payload:   payload,
	}, nil
}

// RequestTimeout은 CLI 응답 대기 최대 시간이다.
const RequestTimeout = 30 * time.Second
