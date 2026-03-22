.PHONY: server client web overlay docker-up docker-down dev tidy

# 서버 빌드 및 실행
server:
	cd server && go run ./cmd/server

# CLI 클라이언트 빌드
client:
	cd client && go build -o ../bin/previewd ./cmd/previewd

# 웹 대시보드 개발 서버
web:
	cd web && npm install && npm run dev

# 오버레이 JS 번들 빌드
overlay:
	cd web && npm run build:overlay

# 전체 웹 빌드 (대시보드 + 오버레이)
build-web:
	cd web && npm install && npm run build && npm run build:overlay

# Docker Compose 실행
docker-up:
	cp -n .env.example .env || true
	docker compose up --build -d

# Docker Compose 중지
docker-down:
	docker compose down

# 개발 환경 전체 실행 (서버 + 웹)
dev:
	@echo "서버와 웹을 동시에 실행합니다..."
	@make -j2 server web

# Go 모듈 정리
tidy:
	cd server && go mod tidy
	cd client && go mod tidy

# 바이너리 디렉토리 생성
bin:
	mkdir -p bin
