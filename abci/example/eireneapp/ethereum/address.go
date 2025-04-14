package ethereum

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"github.com/zenanetwork/go-zenanet/common"
	"github.com/zenanetwork/go-zenanet/crypto"
	"golang.org/x/crypto/sha3"
)

// Address는 이더리움 20바이트 주소를 나타냅니다.
type Address [20]byte

// BytesToAddress는 바이트 슬라이스를 Address로 변환합니다.
func BytesToAddress(b []byte) Address {
	var a Address
	// 20바이트보다 클 경우 뒤에서 20바이트만 사용
	if len(b) > len(a) {
		b = b[len(b)-20:]
	}
	copy(a[20-len(b):], b)
	return a
}

// HexToAddress는 hex 문자열을 Address로 변환합니다.
func HexToAddress(s string) Address {
	if len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		s = s[2:]
	}
	
	if len(s)%2 == 1 {
		s = "0" + s
	}
	
	b, _ := hex.DecodeString(s)
	return BytesToAddress(b)
}

// Bytes는 Address를 바이트 슬라이스로 변환합니다.
func (a Address) Bytes() []byte {
	return a[:]
}

// Hex는 Address를 "0x" 접두사가 있는 hex 문자열로 변환합니다.
func (a Address) Hex() string {
	return fmt.Sprintf("0x%x", a[:])
}

// String은 Address의 문자열 표현을 반환합니다.
func (a Address) String() string {
	return a.Hex()
}

// IsZero는 주소가 0인지 확인합니다.
func (a Address) IsZero() bool {
	return bytes.Equal(a[:], make([]byte, 20))
}

// PublicKeyToAddress는 ECDSA 공개키에서 이더리움 주소를 도출합니다.
func PublicKeyToAddress(pubKey *ecdsa.PublicKey) Address {
	pubBytes := crypto.FromECDSAPub(pubKey)
	h := sha3.NewLegacyKeccak256()
	h.Write(pubBytes[1:]) // 앞의 0x04 바이트 제거
	hash := h.Sum(nil)
	return BytesToAddress(hash[12:]) // 마지막 20바이트 사용
}

// PrivateKeyToAddress는 ECDSA 개인키에서 이더리움 주소를 도출합니다.
func PrivateKeyToAddress(privateKey *ecdsa.PrivateKey) Address {
	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	return PublicKeyToAddress(publicKey)
}

// VerifySignature는 서명이 지정된 주소에서 유효한지 확인합니다.
func VerifySignature(address Address, data []byte, signature []byte) bool {
	// 서명 형식: [R || S || V]
	if len(signature) != 65 {
		return false
	}
	
	// recover 함수를 통해 서명자의 공개키 복구
	hash := crypto.Keccak256Hash(data)
	sigPublicKey, err := crypto.Ecrecover(hash.Bytes(), signature)
	if err != nil {
		return false
	}
	
	// 복구된 공개키에서 주소 계산
	pubKey, err := crypto.UnmarshalPubkey(sigPublicKey)
	if err != nil {
		return false
	}
	
	recoveredAddr := PublicKeyToAddress(pubKey)
	return bytes.Equal(address[:], recoveredAddr[:])
}

// NewAddressFromPrivateKey는 개인키에서 새 주소를 생성합니다.
func NewAddressFromPrivateKey(privateKeyHex string) (Address, error) {
	if len(privateKeyHex) >= 2 && privateKeyHex[0] == '0' && (privateKeyHex[1] == 'x' || privateKeyHex[1] == 'X') {
		privateKeyHex = privateKeyHex[2:]
	}
	
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return Address{}, err
	}
	
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return Address{}, err
	}
	
	return PrivateKeyToAddress(privateKey), nil
}

// EthereumAddressToCommonAddress는 우리의 Address 타입을 go-ethereum의 common.Address로 변환합니다.
func EthereumAddressToCommonAddress(addr Address) common.Address {
	var commAddr common.Address
	copy(commAddr[:], addr[:])
	return commAddr
}

// CommonAddressToEthereumAddress는 go-ethereum의 common.Address를 우리의 Address 타입으로 변환합니다.
func CommonAddressToEthereumAddress(commAddr common.Address) Address {
	var addr Address
	copy(addr[:], commAddr[:])
	return addr
} 