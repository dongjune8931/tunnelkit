package tunnel

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Forwarder는 서버에서 받은 요청을 로컬 서버로 전달한다.
type Forwarder struct {
	LocalPort int
	client    *http.Client
}

func NewForwarder(localPort int) *Forwarder {
	return &Forwarder{
		LocalPort: localPort,
		client: &http.Client{
			Timeout: 25 * time.Second,
		},
	}
}

// Forward는 TunnelRequest를 로컬 서버에 전달하고 TunnelResponse를 반환한다.
func (f *Forwarder) Forward(req *TunnelRequest) *TunnelResponse {
	bodyBytes, err := base64.StdEncoding.DecodeString(req.Body)
	if err != nil {
		return errorResponse(req.RequestID, "요청 본문 디코딩 실패: "+err.Error())
	}

	url := fmt.Sprintf("http://localhost:%d%s", f.LocalPort, req.Path)
	httpReq, err := http.NewRequest(req.Method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return errorResponse(req.RequestID, "HTTP 요청 생성 실패: "+err.Error())
	}

	// 헤더 복사 (Host는 로컬로 교체)
	for k, v := range req.Headers {
		if strings.ToLower(k) == "host" {
			continue
		}
		httpReq.Header.Set(k, v)
	}
	httpReq.Header.Set("Host", fmt.Sprintf("localhost:%d", f.LocalPort))

	resp, err := f.client.Do(httpReq)
	if err != nil {
		return errorResponse(req.RequestID, "로컬 서버 요청 실패: "+err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorResponse(req.RequestID, "응답 본문 읽기 실패: "+err.Error())
	}

	// 응답 헤더 수집
	headers := make(map[string]string)
	for k, vals := range resp.Header {
		headers[k] = strings.Join(vals, ", ")
	}

	return &TunnelResponse{
		RequestID:  req.RequestID,
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       base64.StdEncoding.EncodeToString(respBody),
	}
}

func errorResponse(requestID, msg string) *TunnelResponse {
	return &TunnelResponse{
		RequestID:  requestID,
		StatusCode: http.StatusBadGateway,
		Headers:    map[string]string{"Content-Type": "text/plain"},
		Body:       base64.StdEncoding.EncodeToString([]byte(msg)),
	}
}
