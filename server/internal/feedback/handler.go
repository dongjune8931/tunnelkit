package feedback

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Create(c *gin.Context) {
	sessionID := c.Param("sub")

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fb := &Feedback{
		ID:         uuid.New().String(),
		SessionID:  sessionID,
		PageURL:    req.PageURL,
		ElementCSS: req.ElementCSS,
		XPercent:   req.XPercent,
		YPercent:   req.YPercent,
		Comment:    req.Comment,
		AuthorName: req.AuthorName,
		CreatedAt:  time.Now(),
	}

	_, err := h.db.Exec(
		`INSERT INTO feedbacks (id, session_id, page_url, element_css, x_percent, y_percent, comment, author_name, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fb.ID, fb.SessionID, fb.PageURL, fb.ElementCSS,
		fb.XPercent, fb.YPercent, fb.Comment, fb.AuthorName, fb.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "피드백 저장 실패"})
		return
	}

	c.JSON(http.StatusCreated, fb)
}

func (h *Handler) List(c *gin.Context) {
	sessionID := c.Param("sub")

	rows, err := h.db.Query(
		`SELECT id, session_id, page_url, element_css, x_percent, y_percent, comment, author_name, resolved, created_at
		 FROM feedbacks WHERE session_id = ? ORDER BY created_at DESC`, sessionID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "피드백 조회 실패"})
		return
	}
	defer rows.Close()

	feedbacks := make([]*Feedback, 0)
	for rows.Next() {
		fb := &Feedback{}
		var resolved int
		if err := rows.Scan(&fb.ID, &fb.SessionID, &fb.PageURL, &fb.ElementCSS,
			&fb.XPercent, &fb.YPercent, &fb.Comment, &fb.AuthorName, &resolved, &fb.CreatedAt); err != nil {
			continue
		}
		fb.Resolved = resolved == 1
		feedbacks = append(feedbacks, fb)
	}

	c.JSON(http.StatusOK, feedbacks)
}

func (h *Handler) Resolve(c *gin.Context) {
	id := c.Param("id")

	result, err := h.db.Exec("UPDATE feedbacks SET resolved = 1 WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "피드백 업데이트 실패"})
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "피드백을 찾을 수 없습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"resolved": true})
}
