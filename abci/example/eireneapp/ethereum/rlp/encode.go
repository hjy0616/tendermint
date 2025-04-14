package rlp

// EncodeToBytes encodes a value to bytes using RLP.
// 이 함수는 단순화된 버전으로, 실제 구현은 더 복잡할 수 있습니다.
func EncodeToBytes(val interface{}) ([]byte, error) {
	// 실제 구현은 go-ethereum/rlp 패키지를 사용하거나 직접 구현할 수 있습니다.
	// 임시 목적으로는 빈 구현을 제공합니다.
	// TODO: 실제 RLP 인코딩 구현
	return []byte{}, nil
}

// DecodeBytes decodes RLP encoded bytes to a value.
func DecodeBytes(b []byte, val interface{}) error {
	// 실제 구현은 go-ethereum/rlp 패키지를 사용하거나 직접 구현할 수 있습니다.
	// 임시 목적으로는 빈 구현을 제공합니다.
	// TODO: 실제 RLP 디코딩 구현
	return nil
} 