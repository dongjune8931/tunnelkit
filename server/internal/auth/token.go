package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateToken은 암호학적으로 안전한 랜덤 토큰을 생성한다.
func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("토큰 생성 실패: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
