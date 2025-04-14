package eireneapp

import (
	"github.com/tendermint/tendermint/abci/types"
)

// Validator는 검증자 정보를 나타냅니다.
type Validator struct {
	PubKey []byte `json:"pub_key"` // 검증자 공개키
	Power  int64  `json:"power"`   // 검증자 투표력
}

// BlockInfo는 현재 처리 중인 블록 정보를 담고 있습니다.
type BlockInfo struct {
	Height     int64             // 블록 높이
	Time       int64             // 블록 타임스탬프
	Hash       []byte            // 블록 해시
	ProposerID []byte            // 블록 제안자 ID
	TxCount    int               // 트랜잭션 수
}

// TxResult는 트랜잭션 실행 결과를 나타냅니다.
type TxResult struct {
	Hash   []byte            // 트랜잭션 해시
	Status uint32            // 트랜잭션 상태 (0: 실패, 1: 성공)
	Data   []byte            // 반환 데이터
	Events []types.Event     // 이벤트
	Gas    uint64            // 사용된 가스량
}

// AppState는 애플리케이션의 상태를 나타냅니다.
type AppState struct {
	Height     int64       `json:"height"`     // 현재 블록 높이
	AppHash    []byte      `json:"app_hash"`   // 애플리케이션 상태 해시
	Validators []Validator `json:"validators"` // 현재 검증자 목록
	StateRoot  []byte      `json:"state_root"` // 상태 루트 해시 (이더리움 호환)
}

// Codec 간단한 코덱 인터페이스
type Codec interface {
	MarshalJSON(v interface{}) ([]byte, error)
	UnmarshalJSON(bz []byte, ptr interface{}) error
} 