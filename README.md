# tunnelkit

로컬 개발 환경을 팀과 공유하는 셀프 호스팅 터널 도구.

배포 전 PM과 디자이너가 개발자의 `localhost`를 직접 확인하고 피드백을 남길 수 있습니다. 외부 서버에 데이터를 보내지 않고, 직접 운영하는 서버 위에서 동작합니다.

---

## why?

기존 ngrok 같은 터널 서비스는 유료 플랜 없이는 URL이 매번 바뀌고, 팀 협업 기능이 없으며, 모든 트래픽이 외부 서버를 경유합니다. tunnelkit은 이 문제를 셀프 호스팅으로 해결합니다.

- 고정 서브도메인: `myproject.localhost:8080`
- 초대 링크 + 토큰 기반 접근 제어
- 피드백 오버레이 내장 (HTML에 자동 주입)
- 모든 데이터가 자체 서버에만 저장

---

## 아키텍처

```
[외부 사용자 브라우저]
        |
        | HTTP  myproject.localhost:8080
        v
[터널 서버 :8080]  <-->  [대시보드 API :8081]
        |
        | WebSocket (멀티플렉싱)
        v
[CLI 클라이언트] -- 개발자 로컬 머신
        |
        v
[localhost:3000] -- 개발 중인 서버
```

---

## 빠른 시작

### 요구사항

- Docker, Docker Compose
- Go 1.22+ (CLI 빌드 시)

### 서버 실행

```bash
git clone https://github.com/dongjune8931/tunnelkit
cd tunnelkit
cp .env.example .env
make docker-up
```

서버가 뜨면 `http://localhost:8081`에서 대시보드에 접근할 수 있습니다.

### CLI 설치 및 사용

```bash
make client
./bin/previewd start --port 3000 --subdomain myproject
```

터미널에 출력되는 초대 링크를 PM/디자이너에게 공유하면 됩니다.

```
╔══════════════════════════════════════════════════╗
║           previewd - 터널 연결 성공              ║
╠══════════════════════════════════════════════════╣
║  터널 URL:  http://myproject.localhost:8080?...  ║
╚══════════════════════════════════════════════════╝
```

### 개발 환경 실행 (서버 + 웹 동시)

```bash
make dev
```

---

## 기능

| 기능 | 설명 |
|------|------|
| 터널 프록시 | WebSocket 기반 HTTP 요청 멀티플렉싱 |
| 고정 서브도메인 | CLI 실행 시 지정한 이름으로 고정 |
| 초대 링크 | 토큰 기반 접근 제어, QR 코드 포함 |
| 피드백 오버레이 | HTML 페이지에 자동 삽입, 엘리먼트 클릭으로 피드백 |
| 실시간 로그 | SSE로 HTTP 요청 로그 스트리밍 |
| 셀프 호스팅 | Docker Compose 단일 명령으로 운영 |

---

## 기술 스택

- **서버**: Go + Gin + gorilla/websocket + modernc.org/sqlite
- **CLI**: Go + Cobra + Viper
- **프론트엔드**: React + TypeScript + Vite + Zustand + TanStack Query (Claude로 작성)
- **피드백 오버레이**: 순수 TypeScript IIFE 번들 (Claude로 작성)
- **배포**: Docker Compose

> 프론트엔드(대시보드 UI, 피드백 오버레이)는 Claude를 사용해 작성했습니다.
> 이 프로젝트의 핵심 관심사는 Go 언어와 터널링 메커니즘입니다.

---

## 구현 상세 — 터널링과 Go

### WebSocket 요청 멀티플렉싱

tunnelkit의 핵심은 단일 WebSocket 연결 위에서 수십 개의 HTTP 요청을 동시에 처리하는 멀티플렉싱입니다.

```
PM 브라우저 → GET /page
    → 서버: UUID requestID 생성
    → pendingRequests[requestID] = make(chan *Response, 1)
    → CLI로 WS 메시지 전송 {type:"request", request_id, method, path, body(base64)}
    → CLI: localhost:3000으로 HTTP 요청 → 응답 수신
    → CLI → 서버: {type:"response", request_id, status_code, body(base64)}
    → pendingRequests[requestID] 채널에 응답 전달
    → PM 브라우저에 HTTP 응답 반환
```

Go의 goroutine과 channel이 이 구조를 자연스럽게 표현합니다. 각 HTTP 요청은 독립된 goroutine에서 처리되고, 채널로 WebSocket 응답을 기다립니다. 뮤텍스와 UUID만으로 요청-응답을 정확히 매칭할 수 있습니다.

```go
// 요청마다 응답 채널을 등록
pending := registry.Add(requestID)
defer registry.Remove(requestID)

// CLI로 요청 전송
client.WriteJSON(msg)

// 채널에서 CLI 응답 대기 (30초 타임아웃)
select {
case resp := <-pending.ch:
    writeResponse(c, resp)
case <-time.After(30 * time.Second):
    c.String(504, "터널 응답 시간 초과")
}
```

### sync.RWMutex 설계

WebSocket 허브는 동시에 여러 goroutine이 읽고 쓰는 구조입니다. `sync.Mutex` 대신 `sync.RWMutex`를 선택한 이유는, 대부분의 접근이 "읽기"(특정 서브도메인의 클라이언트 조회)이기 때문입니다. `RWMutex`는 읽기 잠금이 동시에 여러 개 허용되어 처리량이 높습니다.

또한 `WriteJSON`은 Client 레벨에서 별도의 `sync.Mutex`로 보호합니다. 30초마다 ping을 보내는 goroutine과 HTTP 요청을 처리하는 goroutine이 동시에 WebSocket 연결에 쓰려 할 수 있기 때문입니다.

### CGO 없는 SQLite

`database/sql` 표준 인터페이스에 `modernc.org/sqlite` 드라이버를 연결했습니다. C 라이브러리 없이 순수 Go로 컴파일되어 `CGO_ENABLED=0`으로 alpine 이미지에서 바로 실행됩니다. Docker 빌드 시 C 컴파일 환경이 필요 없어 이미지 크기와 빌드 속도가 개선됩니다.

### 피드백 오버레이 주입

프록시 핸들러에서 Content-Type이 `text/html`인 응답을 감지하면 `</body>` 직전에 스크립트 태그를 삽입합니다. 클라이언트 코드 수정 없이 모든 페이지에 오버레이가 붙습니다. 삽입 후 Content-Length를 재계산해 브라우저가 잘린 응답을 받지 않도록 합니다.

### 자동 재연결

CLI 클라이언트는 서버 연결이 끊어지면 5초 후 자동으로 재연결을 시도합니다. 개발 중 서버를 재시작하거나 네트워크가 잠깐 끊겨도 CLI를 다시 실행할 필요가 없습니다.

---

## 앞으로의 방향 — 터널링과 Go 심화

이 프로젝트의 개발 방향은 Go 언어의 동시성 모델과 네트워크 레이어를 깊이 파고드는 것에 집중합니다. 프론트엔드는 현재 수준에서 유지합니다.

### 1. TCP 레벨 터널링으로 전환

현재는 HTTP 요청을 WebSocket 메시지로 직렬화해 전달하는 방식입니다. 다음 단계는 HTTP뿐 아니라 임의의 TCP 스트림을 터널링하는 것입니다.

- WebSocket 연결 위에서 TCP 스트림을 양방향 파이프로 처리
- `io.Copy`와 goroutine 두 개로 전이중(full-duplex) 터널 구현
- WebSocket over HTTP, HTTP/2 CONNECT, gRPC 스트리밍 각각의 성능 비교

### 2. 연결 풀링 및 멀티플렉싱 최적화

현재는 CLI 클라이언트 하나당 WebSocket 연결 하나입니다. 고부하 상황에서는 연결 오버헤드가 병목이 됩니다.

- 단일 WebSocket 위에 논리적 스트림을 여러 개 운영하는 커스텀 멀티플렉싱 레이어 구현
- HTTP/2의 스트림 개념을 WebSocket 위에서 직접 구현해보는 실험
- QUIC 프로토콜 기반 터널링 (Go의 `quic-go` 라이브러리 활용)

### 3. Go 런타임 수준의 성능 튜닝

- `pprof`로 goroutine 누수, 메모리 할당, GC 압력 측정
- 요청 본문 base64 인코딩 제거 — WebSocket binary 프레임으로 직접 전송해 CPU 절감
- `sync.Pool`로 메시지 버퍼 재사용, GC 압력 감소
- 고루틴 수 상한선 설정 및 워커 풀 패턴 적용

### 4. 분산 서버 구성

현재는 단일 서버 프로세스입니다. 수평 확장을 위해:

- 서버 간 터널 세션 공유: Redis pub/sub으로 어느 서버에 CLI가 연결됐는지 브로드캐스트
- 일관된 해싱(Consistent Hashing)으로 서브도메인을 특정 서버 노드에 고정
- Go의 `net` 패키지를 직접 써서 커스텀 로드밸런서 구현

### 5. 커스텀 프로토콜 설계

JSON 기반 메시지 포맷을 바이너리 프로토콜로 교체합니다.

- Protocol Buffers 또는 직접 설계한 TLV(Type-Length-Value) 프레임 포맷 적용
- 압축 (zstd) 적용으로 대용량 HTML/JSON 응답 전송 효율 개선
- 프로토콜 버전 협상 핸드셰이크 구현

---

## 라이선스

MIT

---

## 기여

이슈와 PR은 환영합니다. 단, 이 프로젝트의 개발 방향은 위에서 설명한 터널링 및 Go 네트워크 레이어에 집중합니다.
