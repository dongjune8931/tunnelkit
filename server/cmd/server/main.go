package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/tunnelkit/server/internal/auth"
	"github.com/tunnelkit/server/internal/config"
	"github.com/tunnelkit/server/internal/dashboard"
	"github.com/tunnelkit/server/internal/db"
	"github.com/tunnelkit/server/internal/feedback"
	"github.com/tunnelkit/server/internal/inject"
	"github.com/tunnelkit/server/internal/tunnel"
	"github.com/tunnelkit/server/internal/ws"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("설정 로드 실패: %v", err)
	}

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("DB 열기 실패: %v", err)
	}
	if err := db.Migrate(database); err != nil {
		log.Fatalf("DB 마이그레이션 실패: %v", err)
	}

	hub := ws.NewHub()
	registry := tunnel.NewRegistry()
	dashboardURL := fmt.Sprintf("http://%s:%d", cfg.ServerHost, cfg.DashboardPort)
	injector := inject.New(dashboardURL)
	proxyHandler := tunnel.NewProxyHandler(hub, registry, injector)
	dashHandler := dashboard.NewHandler(database, hub)
	fbHandler := feedback.NewHandler(database)

	// 터널 프록시 서버 (포트 8080)
	tunnelRouter := gin.Default()
	tunnelRouter.Use(corsMiddleware())

	// WebSocket 핸드셰이크 엔드포인트 (CLI 전용)
	tunnelRouter.GET("/ws/tunnel", func(c *gin.Context) {
		handleTunnelWS(c, hub, registry, dashHandler)
	})

	// 모든 HTTP 요청을 터널로 프록싱 (서브도메인 기반)
	tunnelRouter.NoRoute(proxyHandler.Handle)

	// 대시보드 API 서버 (포트 8081)
	apiRouter := gin.Default()
	apiRouter.Use(corsMiddleware())

	// 정적 파일 (오버레이 JS)
	apiRouter.Static("/static", "./web/dist")

	api := apiRouter.Group("/api")
	{
		api.GET("/tunnels", dashHandler.ListTunnels)
		api.POST("/tunnels/:sub/invite", dashHandler.CreateInviteToken)
		api.GET("/tunnels/:sub/logs/stream", dashHandler.LogStream)
		api.POST("/tunnels/:sub/feedback", fbHandler.Create)
		api.GET("/tunnels/:sub/feedback", fbHandler.List)
		api.PATCH("/feedback/:id/resolve", fbHandler.Resolve)
	}

	// 두 서버 동시 실행
	go func() {
		addr := fmt.Sprintf(":%d", cfg.DashboardPort)
		log.Printf("대시보드 API 서버 시작: http://%s:%d", cfg.ServerHost, cfg.DashboardPort)
		if err := apiRouter.Run(addr); err != nil {
			log.Fatalf("대시보드 서버 오류: %v", err)
		}
	}()

	addr := fmt.Sprintf(":%d", cfg.TunnelPort)
	log.Printf("터널 프록시 서버 시작: http://%s:%d", cfg.ServerHost, cfg.TunnelPort)
	if err := tunnelRouter.Run(addr); err != nil {
		log.Fatalf("터널 서버 오류: %v", err)
	}
}

// handleTunnelWS는 CLI의 WebSocket 연결을 처리한다.
func handleTunnelWS(c *gin.Context, hub *ws.Hub, registry *tunnel.Registry, dash *dashboard.Handler) {
	subdomain := c.GetHeader("X-Subdomain")
	authToken := c.GetHeader("X-Auth-Token")
	if subdomain == "" {
		c.String(http.StatusBadRequest, "X-Subdomain 헤더 필요")
		return
	}
	if authToken == "" {
		c.String(http.StatusUnauthorized, "X-Auth-Token 헤더 필요")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 업그레이드 실패: %v", err)
		return
	}
	defer conn.Close()

	client := ws.NewClient(subdomain, conn)
	hub.Register(subdomain, client)
	defer hub.Unregister(subdomain)

	sessionID := uuid.New().String()
	accessToken, _ := auth.GenerateToken(16)

	log.Printf("터널 연결: %s (세션: %s)", subdomain, sessionID)

	// 환영 메시지 전송
	welcome := tunnel.TunnelMessage{
		Type: tunnel.MsgTypeWelcome,
	}
	payload, _ := json.Marshal(tunnel.WelcomePayload{
		SessionID:   sessionID,
		AccessToken: accessToken,
		TunnelURL:   fmt.Sprintf("http://%s.localhost:8080?token=%s", subdomain, accessToken),
	})
	welcome.Payload = payload
	client.WriteJSON(welcome)

	// Ping 루프
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	go func() {
		for range pingTicker.C {
			client.WriteJSON(tunnel.TunnelMessage{Type: tunnel.MsgTypePing})
		}
	}()

	// 메시지 수신 루프
	for {
		var msg tunnel.TunnelMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("터널 연결 종료: %s (%v)", subdomain, err)
			}
			break
		}

		switch msg.Type {
		case tunnel.MsgTypeResponse:
			tunnel.DeliverResponse(registry, msg.Payload)
		case tunnel.MsgTypePong:
			// 핑퐁 정상
		default:
			log.Printf("알 수 없는 메시지 타입: %s", msg.Type)
		}
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-Access-Token, X-Auth-Token, X-Subdomain")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

