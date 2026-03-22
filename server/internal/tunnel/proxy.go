package tunnel

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tunnelkit/server/internal/inject"
	"github.com/tunnelkit/server/internal/ws"
)

// ProxyHandler는 외부 HTTP 요청을 터널을 통해 로컬로 전달한다.
type ProxyHandler struct {
	hub      *ws.Hub
	registry *Registry
	injector *inject.Injector
}

func NewProxyHandler(hub *ws.Hub, registry *Registry, injector *inject.Injector) *ProxyHandler {
	return &ProxyHandler{hub: hub, registry: registry, injector: injector}
}

func (h *ProxyHandler) Handle(c *gin.Context) {
	subdomain := c.Param("subdomain")
	if subdomain == "" {
		// Host 헤더에서 서브도메인 추출: myproject.localhost:8080 → myproject
		host := c.Request.Host
		parts := strings.SplitN(host, ".", 2)
		if len(parts) < 2 {
			c.String(http.StatusBadRequest, "서브도메인을 확인할 수 없습니다")
			return
		}
		subdomain = parts[0]
	}

	client, ok := h.hub.Get(subdomain)
	if !ok {
		c.String(http.StatusServiceUnavailable, "터널이 연결되어 있지 않습니다: %s", subdomain)
		return
	}

	// 요청 본문 읽기
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, "요청 본문 읽기 실패")
		return
	}

	// 헤더 수집
	headers := make(map[string]string)
	for key, vals := range c.Request.Header {
		headers[key] = strings.Join(vals, ", ")
	}

	requestID := uuid.New().String()
	tunReq := &TunnelRequest{
		RequestID: requestID,
		Method:    c.Request.Method,
		Path:      c.Request.RequestURI,
		Headers:   headers,
		Body:      base64.StdEncoding.EncodeToString(bodyBytes),
	}

	msg, err := BuildRequestMsg(tunReq)
	if err != nil {
		c.String(http.StatusInternalServerError, "메시지 생성 실패")
		return
	}

	// 대기 등록
	pending := h.registry.Add(requestID)
	defer h.registry.Remove(requestID)

	// CLI로 전송
	if err := client.WriteJSON(msg); err != nil {
		c.String(http.StatusBadGateway, "CLI 전송 실패: %v", err)
		return
	}

	// 응답 대기
	select {
	case resp := <-pending.ch:
		h.writeResponse(c, resp)
	case <-time.After(RequestTimeout):
		c.String(http.StatusGatewayTimeout, "터널 응답 시간 초과 (30초)")
	}
}

func (h *ProxyHandler) writeResponse(c *gin.Context, resp *TunnelResponse) {
	// 헤더 설정
	for k, v := range resp.Headers {
		lk := strings.ToLower(k)
		if lk == "content-length" || lk == "transfer-encoding" {
			continue
		}
		c.Header(k, v)
	}

	// 본문 디코딩
	body, err := base64.StdEncoding.DecodeString(resp.Body)
	if err != nil {
		log.Printf("응답 본문 디코딩 실패: %v", err)
		c.Status(resp.StatusCode)
		return
	}

	// HTML이면 오버레이 주입
	contentType := resp.Headers["Content-Type"]
	if strings.Contains(contentType, "text/html") && h.injector != nil {
		body = h.injector.Inject(body)
		c.Header("Content-Length", fmt.Sprintf("%d", len(body)))
	}

	c.Data(resp.StatusCode, contentType, body)
}

// DeliverResponse는 CLI에서 온 응답을 대기 중인 요청에 전달한다.
func DeliverResponse(registry *Registry, data json.RawMessage) {
	var resp TunnelResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		log.Printf("응답 언마샬 실패: %v", err)
		return
	}
	if !registry.Deliver(resp.RequestID, &resp) {
		log.Printf("대기 요청 없음: %s (타임아웃됐을 가능성)", resp.RequestID)
	}
}
