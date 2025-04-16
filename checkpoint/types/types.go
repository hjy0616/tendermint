package types

import (
	"encoding/json"
	"time"

	"github.com/tendermint/tendermint/crypto"
)

// Checkpoint 구조체는 블록체인의 특정 높이에서의 체크포인트 정보를 담습니다.
type Checkpoint struct {
	// 체크포인트의 고유 ID
	ID string `json:"id"`

	// 체크포인트가 생성된 블록 높이
	Height int64 `json:"height"`

	// 체크포인트 생성 시간
	Timestamp time.Time `json:"timestamp"`

	// 블록 해시
	BlockHash []byte `json:"block_hash"`

	// 검증자 서명 목록
	Signatures []ValidatorSignature `json:"signatures"`

	// 상태 루트 해시
	StateRootHash []byte `json:"state_root_hash"`
}

// ValidatorSignature 구조체는 검증자의 서명 정보를 담습니다.
type ValidatorSignature struct {
	// 검증자 주소
	ValidatorAddress crypto.Address `json:"validator_address"`

	// 서명 데이터
	Signature []byte `json:"signature"`
}

// Marshal은 Checkpoint를 JSON 바이트 배열로 직렬화합니다.
func (c *Checkpoint) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

// Unmarshal은 JSON 바이트 배열을 Checkpoint 객체로 역직렬화합니다.
func Unmarshal(data []byte) (*Checkpoint, error) {
	var checkpoint Checkpoint
	err := json.Unmarshal(data, &checkpoint)
	if err != nil {
		return nil, err
	}
	return &checkpoint, nil
}
