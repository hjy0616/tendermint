# Eirene ABCI 애플리케이션

Eirene는 Tendermint 합의 엔진을 이용한 Ethereum 호환 ABCI(Application Blockchain Interface) 애플리케이션입니다. 이 애플리케이션은 Tendermint의 합의 메커니즘과 Ethereum의 상태 관리 및 스마트 컨트랙트 기능을 통합하는 브릿지 역할을 합니다.

## 기능

- Tendermint ABCI 인터페이스 구현
- Ethereum 호환 트랜잭션 처리
- PoS(Proof of Stake) 검증자 관리
- 상태 동기화 및 체크포인트 메커니즘

## 구조

- `app.go`: 핵심 ABCI 애플리케이션 구현
- `types.go`: 기본 데이터 타입 정의
- `validators/`: 검증자 관리 시스템
- `utils/`: 유틸리티 함수
- `cmd/`: 실행 바이너리

## 빌드 및 실행

### 사전 요구사항

- Go 1.16 이상
- Tendermint Core

### 빌드

```bash
make build
```

### 실행

```bash
# 기본 설정으로 실행
make run

# 개발 모드로 실행 (디버그 로그 활성화)
make dev
```

### Tendermint와 함께 실행

1. Eirene ABCI 애플리케이션 실행:
```bash
make run
```

2. 별도의 터미널에서 Tendermint 노드 실행:
```bash
make tendermint
```

## 커스터마이징

### 설정 옵션

- `-socket`: ABCI 서버 주소 (기본값: "tcp://127.0.0.1:26658")
- `-db`: 데이터베이스 디렉토리 (기본값: "./data")
- `-log`: 로그 레벨 (debug, info, error) (기본값: "info")

### 예시

```bash
./build/eireneapp -socket="tcp://0.0.0.0:26658" -log="debug" -db="./mydata"
```

## 향후 개발 계획

- EVM(Ethereum Virtual Machine) 통합
- 이더리움 JSON-RPC API 지원
- 스마트 컨트랙트 배포 및 실행 기능
- 상태 증명 및 경량 클라이언트 지원 