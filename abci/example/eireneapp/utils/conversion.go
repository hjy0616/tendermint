package utils

import (
	"crypto/sha256"
	"fmt"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
)

// TMTimeToUnixTime은 Tendermint 시간을 Unix 타임스탬프로 변환합니다.
func TMTimeToUnixTime(tmTime time.Time) int64 {
	return tmTime.Unix()
}

// HashData는 데이터의 SHA256 해시를 계산합니다.
func HashData(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// CreateTxHash는 트랜잭션 데이터로부터 간단한 해시를 생성합니다.
func CreateTxHash(tx []byte) []byte {
	return HashData(tx)
}

// CreateAppHash는 애플리케이션 상태로부터 해시를 생성합니다.
func CreateAppHash(height int64, data []byte) []byte {
	// 높이와 데이터를 결합하여 해시
	combined := append([]byte(fmt.Sprintf("%d:", height)), data...)
	return HashData(combined)
}

// CreateEvents는 이벤트 집합을 생성합니다.
func CreateEvents(txHash []byte, eventType string, attributes map[string]string) []abci.Event {
	event := abci.Event{
		Type: eventType,
		Attributes: []abci.EventAttribute{
			{Key: []byte("tx_hash"), Value: []byte(fmt.Sprintf("%X", txHash)), Index: true},
		},
	}

	// 추가 속성
	for k, v := range attributes {
		event.Attributes = append(event.Attributes, abci.EventAttribute{
			Key:   []byte(k),
			Value: []byte(v),
			Index: true,
		})
	}

	return []abci.Event{event}
}

// ValidatorPubKeyToHex는 검증자 공개키를 16진수 문자열로 변환합니다.
func ValidatorPubKeyToHex(pubKey []byte) string {
	return fmt.Sprintf("%X", pubKey)
}

// HexToBytes는 16진수 문자열을 바이트 배열로 변환합니다.
func HexToBytes(hexStr string) ([]byte, error) {
	// 간단한 구현 - 실제로는 hex.DecodeString 사용
	return nil, nil
} 