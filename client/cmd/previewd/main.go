package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tunnelkit/client/internal/config"
	tunnelclient "github.com/tunnelkit/client/internal/tunnel"
	"github.com/tunnelkit/client/internal/tui"
)

func main() {
	root := &cobra.Command{
		Use:   "previewd",
		Short: "로컬 개발 환경을 팀과 공유하는 터널 도구",
	}

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "터널 연결 시작",
		RunE:  runStart,
	}

	startCmd.Flags().IntP("port", "p", 3000, "로컬 서버 포트")
	startCmd.Flags().StringP("subdomain", "s", "", "서브도메인 (예: myproject)")
	startCmd.Flags().StringP("server", "", "ws://localhost:8080", "터널 서버 URL")
	startCmd.Flags().StringP("token", "t", "", "인증 토큰")

	viper.BindPFlag("local_port", startCmd.Flags().Lookup("port"))
	viper.BindPFlag("subdomain", startCmd.Flags().Lookup("subdomain"))
	viper.BindPFlag("server_url", startCmd.Flags().Lookup("server"))
	viper.BindPFlag("auth_token", startCmd.Flags().Lookup("token"))

	root.AddCommand(startCmd)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runStart(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("설정 로드 실패: %w", err)
	}

	if cfg.Subdomain == "" {
		// 디렉토리명을 기본 서브도메인으로 사용
		cwd, _ := os.Getwd()
		parts := strings.Split(cwd, "/")
		cfg.Subdomain = parts[len(parts)-1]
	}
	if cfg.AuthToken == "" {
		cfg.AuthToken = "devtoken"
	}

	display := tui.New()

	client := tunnelclient.NewClient(
		cfg.ServerURL,
		cfg.Subdomain,
		cfg.AuthToken,
		cfg.LocalPort,
		func(payload tunnelclient.WelcomePayload) {
			display.Connected(payload.TunnelURL, payload.AccessToken)
		},
	)

	log.Printf("터널 연결 중: %s → localhost:%d", cfg.Subdomain, cfg.LocalPort)
	return client.Connect()
}
