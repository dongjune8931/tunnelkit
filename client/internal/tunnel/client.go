package tunnel

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// 메시지 타입 (서버와 동일)
const (
	MsgTypeWelcome    = "welcome"
	MsgTypeRequest    = "request"
	MsgTypeResponse   = "response"
	MsgTypePing       = "ping"
	MsgTypePong       = "pong"
	MsgTypeDisconnect = "disconnect"
)

type TunnelMessage struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type WelcomePayload struct {
	SessionID   string `json:"session_id"`
	AccessToken string `json:"access_token"`
	TunnelURL   string `json:"tunnel_url"`
}

type TunnelRequest struct {
	RequestID string            `json:"request_id"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
}

type TunnelResponse struct {
	RequestID  string            `json:"request_id"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// OnWelcome은 서버 연결 성공 시 콜백이다.
type OnWelcome func(payload WelcomePayload)

// Client는 서버와의 WebSocket 터널 연결을 관리한다.
type Client struct {
	serverURL string
	subdomain string
	authToken string
	forwarder *Forwarder
	onWelcome OnWelcome
}

func NewClient(serverURL, subdomain, authToken string, localPort int, onWelcome OnWelcome) *Client {
	return &Client{
		serverURL: serverURL,
		subdomain: subdomain,
		authToken: authToken,
		forwarder: NewForwarder(localPort),
		onWelcome: onWelcome,
	}
}

// Connect는 서버에 연결하고 메시지 루프를 실행한다 (재연결 포함).
func (c *Client) Connect() error {
	for {
		if err := c.connect(); err != nil {
			log.Printf("연결 실패 (5초 후 재시도): %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
	}
}

func (c *Client) connect() error {
	// WebSocket URL 구성: ws://localhost:8080/ws/tunnel
	wsURL := fmt.Sprintf("%s/ws/tunnel", c.serverURL)

	header := http.Header{}
	header.Set("X-Subdomain", c.subdomain)
	header.Set("X-Auth-Token", c.authToken)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return fmt.Errorf("WebSocket 연결 실패: %w", err)
	}
	defer conn.Close()

	log.Printf("서버 연결 성공: %s", wsURL)

	for {
		var msg TunnelMessage
		if err := conn.ReadJSON(&msg); err != nil {
			return fmt.Errorf("메시지 읽기 오류: %w", err)
		}

		switch msg.Type {
		case MsgTypeWelcome:
			var payload WelcomePayload
			if err := json.Unmarshal(msg.Payload, &payload); err == nil && c.onWelcome != nil {
				c.onWelcome(payload)
			}

		case MsgTypeRequest:
			var req TunnelRequest
			if err := json.Unmarshal(msg.Payload, &req); err != nil {
				log.Printf("요청 언마샬 실패: %v", err)
				continue
			}
			// 비동기로 로컬 서버에 전달
			go func(r TunnelRequest) {
				resp := c.forwarder.Forward(&r)
				payload, err := json.Marshal(resp)
				if err != nil {
					log.Printf("응답 직렬화 실패: %v", err)
					return
				}
				respMsg := TunnelMessage{
					Type:      MsgTypeResponse,
					RequestID: r.RequestID,
					Payload:   payload,
				}
				if err := conn.WriteJSON(respMsg); err != nil {
					log.Printf("응답 전송 실패: %v", err)
				}
			}(req)

		case MsgTypePing:
			conn.WriteJSON(TunnelMessage{Type: MsgTypePong})

		case MsgTypeDisconnect:
			return fmt.Errorf("서버에서 연결 종료 요청")
		}
	}
}
