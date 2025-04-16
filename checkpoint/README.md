# Tendermint Checkpoint

Tendermint의 체크포인트 메커니즘을 구현한 패키지입니다. 이 패키지는 블록체인 상태의 스냅샷을 저장하고 검증하는 기능을 제공합니다.

## 주요 기능

- 지정된 블록 높이에서 체크포인트 생성
- 체크포인트 조회 및 검증
- RPC 및 HTTP 인터페이스 지원
- GRPC 인터페이스 지원

## 구조

- `types.go`: 기본 데이터 구조 및 인터페이스 정의
- `manager.go`: 체크포인트 매니저 구현체
- `rpc.go`: JSON-RPC 핸들러
- `grpc.go`: GRPC 서버 구현
- `routes.go`: HTTP 라우트 핸들러
- `example.go`: 사용 예제

## 사용 예시

### 초기화

```go
import (
    "github.com/tendermint/tendermint/checkpoint"
    "github.com/tendermint/tendermint/state"
    "github.com/tendermint/tendermint/store"
    tmdb "github.com/tendermint/tm-db"
)

// 데이터베이스 초기화
db := tmdb.NewMemDB()
blockStore := store.NewBlockStore(db)
stateStore := state.NewStore(db, state.StoreOptions{})

// 체크포인트 매니저 생성
manager := checkpoint.NewManager(
    db,
    blockStore,
    stateStore,
    &checkpoint.Config{
        CheckpointInterval: 1000,
    },
)
```

### 체크포인트 생성

```go
// 특정 높이에서 체크포인트 생성
checkpoint, err := manager.CreateCheckpoint(context.Background(), 1000)
if err != nil {
    // 오류 처리
}

fmt.Printf("체크포인트 생성: 번호=%d, 높이=%d\n",
    checkpoint.Number, checkpoint.Data.Height)
```

### 체크포인트 조회

```go
// 체크포인트 번호로 조회
checkpoint, err := manager.GetCheckpoint(1)
if err != nil {
    // 오류 처리
}

// 최신 체크포인트 조회
latestCP, err := manager.GetLatestCheckpoint()
if err != nil {
    // 오류 처리
}

// 체크포인트 개수 조회
count, err := manager.GetCheckpointCount()
if err != nil {
    // 오류 처리
}
```

### RPC 서버 등록

```go
import (
    "net/http"
    "github.com/tendermint/tendermint/checkpoint"
    "github.com/tendermint/tendermint/libs/log"
)

// RPC 핸들러 생성
handler := checkpoint.NewRPCHandler(manager)
handler.SetLogger(log.NewTMLogger(os.Stdout))

// HTTP 핸들러 등록
mux := http.NewServeMux()
checkpoint.RegisterHTTPHandlers(mux, handler, logger)

// 서버 시작
http.ListenAndServe(":8080", mux)
```

## API 엔드포인트

- `GET /checkpoint?number={number}`: 특정 번호의 체크포인트 조회
- `GET /checkpoint/latest`: 최신 체크포인트 조회
- `GET /checkpoint/count`: 체크포인트 개수 조회
- `GET /checkpoint/create?height={height}`: 특정 높이에서 체크포인트 생성

## GRPC 인터페이스

- `FetchCheckpoint`: 특정 번호의 체크포인트 조회
- `FetchCheckpointCount`: 체크포인트 개수 조회
- `CreateCheckpoint`: 체크포인트 생성

## 구성 옵션

- `CheckpointInterval`: 자동 체크포인트 생성 간격 (블록 수)
- `PrivateKey`: 체크포인트 서명에 사용할 개인키
- `PublicKey`: 체크포인트 검증에 사용할 공개키

## 실행

예제 코드를 실행하려면:

```
go run example.go --help
```

## 테스트

테스트를 실행하려면:

```
go test -v
```
