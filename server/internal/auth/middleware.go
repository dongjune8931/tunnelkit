package auth

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TokenMiddleware는 ?token= 쿼리 파라미터로 접근을 제어한다.
func TokenMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			token = c.GetHeader("X-Access-Token")
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "접근 토큰이 필요합니다. ?token=xxx 파라미터를 추가하세요.",
			})
			return
		}

		var sessionID string
		err := db.QueryRow(
			"SELECT session_id FROM access_tokens WHERE token = ?", token,
		).Scan(&sessionID)
		if err == sql.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "유효하지 않은 토큰입니다"})
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "토큰 검증 실패"})
			return
		}

		c.Set("session_id", sessionID)
		c.Next()
	}
}
