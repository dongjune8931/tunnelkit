package tui

import (
	"fmt"
	"time"
)

// Display는 터미널 출력을 담당한다.
type Display struct{}

func New() *Display { return &Display{} }

func (d *Display) Connected(tunnelURL, accessToken string) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║           previewd - 터널 연결 성공              ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Printf("║  터널 URL:  %-36s ║\n", tunnelURL)
	fmt.Printf("║  토큰:      %-36s ║\n", accessToken)
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("  요청 로그:")
}

func (d *Display) LogRequest(method, path string, status, durationMs int) {
	now := time.Now().Format("15:04:05")
	statusStr := fmt.Sprintf("%d", status)
	fmt.Printf("  [%s] %s %-6s %-40s %sms\n", now, statusStr, method, path, fmt.Sprintf("%d", durationMs))
}

func (d *Display) Error(msg string) {
	fmt.Printf("  [오류] %s\n", msg)
}
