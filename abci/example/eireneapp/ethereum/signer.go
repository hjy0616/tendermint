package ethereum

import (
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/zenanetwork/go-zenanet/common"
	"github.com/zenanetwork/go-zenanet/crypto"
)

// EIP155Transaction implements Signer using the EIP155 rules.
type EIP155Signer struct {
	chainId, chainIdMul *big.Int
}

// NewEIP155Signer는 특정 체인 ID로 EIP155Signer를 생성합니다.
func NewEIP155Signer(chainId *big.Int) EIP155Signer {
	if chainId == nil {
		chainId = new(big.Int)
	}
	return EIP155Signer{
		chainId:    chainId,
		chainIdMul: new(big.Int).Mul(chainId, big.NewInt(2)),
	}
}

// ChainID는 서명자의 체인 ID를 반환합니다.
func (s EIP155Signer) ChainID() *big.Int {
	return s.chainId
}

// Equal은 두 서명자가 동일한지 비교합니다.
func (s EIP155Signer) Equal(s2 Signer) bool {
	eip155, ok := s2.(EIP155Signer)
	return ok && eip155.chainId.Cmp(s.chainId) == 0
}

// Hash는 서명에 사용될 트랜잭션 해시를 계산합니다.
func (s EIP155Signer) Hash(tx *Transaction) common.Hash {
	return rlpHash([]interface{}{
		tx.Nonce(),
		tx.GasPrice(),
		tx.Gas(),
		tx.To(),
		tx.Value(),
		tx.Data(),
		s.chainId, uint(0), uint(0),
	})
}

// SignatureValues는 서명 바이트를 R, S, V 값으로 변환합니다.
func (s EIP155Signer) SignatureValues(tx *Transaction, sig []byte) (R, S, V *big.Int, err error) {
	if len(sig) != crypto.SignatureLength {
		return nil, nil, nil, errors.New("invalid signature length")
	}
	
	R = new(big.Int).SetBytes(sig[:32])
	S = new(big.Int).SetBytes(sig[32:64])
	V = new(big.Int).SetBytes([]byte{sig[64] + 27})
	
	if s.chainId.Sign() != 0 {
		V = new(big.Int).Add(V, s.chainIdMul)
		V = new(big.Int).Add(V, big.NewInt(8))
	}
	
	return R, S, V, nil
}

// Sender는 트랜잭션에서 발신자 주소를 추출합니다.
func (s EIP155Signer) Sender(tx *Transaction) (common.Address, error) {
	if from := tx.from.Load(); from != nil {
		return from.(common.Address), nil
	}
	
	v, r, s := tx.RawSignatureValues()
	
	// EIP-155 검증
	if s.chainId.Cmp(new(big.Int)) != 0 {
		v = new(big.Int).Sub(v, s.chainIdMul)
		v = new(big.Int).Sub(v, big.NewInt(8))
	}
	
	// V 값은 27 또는 28이어야 합니다.
	if v.BitLen() > 8 {
		return common.Address{}, ErrInvalidSig
	}
	
	v = new(big.Int).Sub(v, big.NewInt(27))
	
	msg := s.Hash(tx)
	sig := make([]byte, 65)
	copy(sig[:32], r.Bytes())
	copy(sig[32:64], s.Bytes())
	sig[64] = byte(v.Uint64())
	
	pub, err := crypto.Ecrecover(msg[:], sig)
	if err != nil {
		return common.Address{}, err
	}
	
	if len(pub) == 0 || pub[0] != 4 {
		return common.Address{}, errors.New("invalid public key")
	}
	
	var addr common.Address
	copy(addr[:], crypto.Keccak256(pub[1:])[12:])
	tx.from.Store(addr)
	return addr, nil
}

// SignTx는 주어진 개인 키로 트랜잭션에 서명합니다.
func (s EIP155Signer) SignTx(tx *Transaction, prv *ecdsa.PrivateKey) (*Transaction, error) {
	h := s.Hash(tx)
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return nil, err
	}
	return tx.WithSignature(s, sig)
} 