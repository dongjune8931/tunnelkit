package dashboard

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tunnelkit/server/internal/auth"
	"github.com/tunnelkit/server/internal/ws"
)

type Handler struct {
	db  *sql.DB
	hub *ws.Hub
}

func NewHandler(db *sql.DB, hub *ws.Hub) *Handler {
	return &Handler{db: db, hub: hub}
}

// ListTunnels는 활성 터널 목록을 반환한다.
func (h *Handler) ListTunnels(c *gin.Context) {
	subs := h.hub.List()
	type TunnelInfo struct {
		Subdomain   string `json:"subdomain"`
		Connected   bool   `json:"connected"`
	}
	result := make([]TunnelInfo, 0, len(subs))
	for _, s := range subs {
		result = append(result, TunnelInfo{Subdomain: s, Connected: true})
	}
	c.JSON(http.StatusOK, result)
}

// CreateInviteToken은 초대 링크용 토큰을 생성한다.
func (h *Handler) CreateInviteToken(c *gin.Context) {
	sub := c.Param("sub")

	// 세션 확인
	var sessionID string
	err := h.db.QueryRow("SELECT id FROM sessions WHERE subdomain = ?", sub).Scan(&sessionID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "터널을 찾을 수 없습니다"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "세션 조회 실패"})
		return
	}

	token, err := auth.GenerateToken(16)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 생성 실패"})
		return
	}

	label := c.DefaultQuery("label", "초대 링크")
	_, err = h.db.Exec(
		"INSERT INTO access_tokens (token, session_id, label) VALUES (?, ?, ?)",
		token, sessionID, label,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 저장 실패"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"url":   fmt.Sprintf("http://%s.localhost:8080?token=%s", sub, token),
	})
}

// LogStream은 SSE로 실시간 요청 로그를 스트리밍한다.
func (h *Handler) LogStream(c *gin.Context) {
	sub := c.Param("sub")
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastID string
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			rows, err := h.db.Query(
				`SELECT id, method, path, status, duration_ms, created_at
				 FROM request_logs WHERE session_id = (SELECT id FROM sessions WHERE subdomain = ?)
				 AND id > ? ORDER BY created_at DESC LIMIT 20`,
				sub, lastID,
			)
			if err != nil {
				continue
			}

			type LogEntry struct {
				ID         string    `json:"id"`
				Method     string    `json:"method"`
				Path       string    `json:"path"`
				Status     int       `json:"status"`
				DurationMs int       `json:"duration_ms"`
				CreatedAt  time.Time `json:"created_at"`
			}

			var entries []LogEntry
			for rows.Next() {
				var e LogEntry
				rows.Scan(&e.ID, &e.Method, &e.Path, &e.Status, &e.DurationMs, &e.CreatedAt)
				entries = append(entries, e)
			}
			rows.Close()

			if len(entries) > 0 {
				lastID = entries[0].ID
				data, _ := json.Marshal(entries)
				fmt.Fprintf(c.Writer, "data: %s\n\n", data)
				c.Writer.Flush()
			}
		}
	}
}

// LogRequest는 요청 로그를 DB에 저장한다.
func (h *Handler) LogRequest(sessionID, method, path string, status, durationMs int) {
	h.db.Exec(
		`INSERT INTO request_logs (id, session_id, method, path, status, duration_ms) VALUES (?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), sessionID, method, path, status, durationMs,
	)
}
