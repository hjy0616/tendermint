package ethereum

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"sync/atomic"

	"github.com/tendermint/tendermint/abci/example/eireneapp/ethereum/rlp"
	"github.com/zenanetwork/go-zenanet/common"
	"github.com/zenanetwork/go-zenanet/crypto"
	"golang.org/x/crypto/sha3"
)

var (
	ErrInvalidSig           = errors.New("invalid transaction v, r, s values")
	ErrInvalidTxType        = errors.New("transaction type not supported")
	ErrTxTypeNotSupported   = errors.New("transaction type not supported")
	ErrInvalidChainID       = errors.New("invalid chain id for signer")
	ErrNegativeValue        = errors.New("negative value")
	ErrExtractSignature     = errors.New("could not extract signature")
	ErrGasPriceTooLow       = errors.New("gas price too low")
	
	// 트랜잭션 타입 상수
	LegacyTxType      = 0x00
	AccessListTxType  = 0x01
	DynamicFeeTxType  = 0x02
)

// TxData 인터페이스는 모든 유형의 이더리움 트랜잭션에서 구현해야 하는 메서드를 정의합니다.
type TxData interface {
	txType() byte // 트랜잭션 유형 식별자 (0x00: legacy, 0x01: access list, 0x02: dynamic fee)
	copy() TxData // TxData의 딥 카피 생성
	
	chainID() *big.Int // 체인 ID 반환
	accessList() AccessList // 액세스 리스트 반환
	data() []byte // 트랜잭션 데이터 반환
	gas() uint64 // 가스 제한 반환
	gasPrice() *big.Int // 가스 가격 반환
	gasTipCap() *big.Int // 팁 상한 반환
	gasFeeCap() *big.Int // 수수료 상한 반환
	value() *big.Int // 트랜잭션 값 반환
	nonce() uint64 // 계정 논스 반환
	to() *common.Address // 수신자 주소 반환
	
	rawSignatureValues() (v, r, s *big.Int) // 서명 값 그대로 반환
	setSignatureValues(chainID, v, r, s *big.Int) // 서명 값 설정
}

// Transaction은 이더리움 트랜잭션을 나타냅니다.
type Transaction struct {
	data txdata
	// 캐시
	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

type txdata struct {
	AccountNonce uint64          `json:"nonce"    gencodec:"required"`
	Price        *big.Int        `json:"gasPrice" gencodec:"required"`
	GasLimit     uint64          `json:"gas"      gencodec:"required"`
	Recipient    *common.Address `json:"to"       rlp:"nil"` // nil이면 계약 생성
	Amount       *big.Int        `json:"value"    gencodec:"required"`
	Payload      []byte          `json:"input"    gencodec:"required"`

	// 서명 값
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	// ChainID는 EIP155에 대한 것
	ChainID *big.Int `json:"chainId,omitempty"`
}

// NewTransaction은 새 트랜잭션을 생성합니다.
func NewTransaction(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	return newTransaction(nonce, &to, amount, gasLimit, gasPrice, data)
}

// NewContractCreation은 계약 생성 트랜잭션을 생성합니다.
func NewContractCreation(nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	return newTransaction(nonce, nil, amount, gasLimit, gasPrice, data)
}

func newTransaction(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := txdata{
		AccountNonce: nonce,
		Recipient:    to,
		Payload:      data,
		Amount:       new(big.Int),
		GasLimit:     gasLimit,
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}

	return &Transaction{data: d}
}

// Type은 트랜잭션 유형을 반환합니다.
func (tx *Transaction) Type() uint8 {
	return tx.data.txType()
}

// ChainID는 트랜잭션의 체인 ID를 반환합니다.
func (tx *Transaction) ChainID() *big.Int {
	return deriveChainID(tx.data.V)
}

// Nonce는 트랜잭션 발신자의 논스를 반환합니다.
func (tx *Transaction) Nonce() uint64 {
	return tx.data.AccountNonce
}

// To는 트랜잭션 수신자 주소를 반환합니다. nil은 컨트랙트 생성을 의미합니다.
func (tx *Transaction) To() *common.Address {
	if tx.data.Recipient == nil {
		return nil
	}
	to := *tx.data.Recipient
	return &to
}

// Value는 트랜잭션에 포함된 원본 금액을 반환합니다.
func (tx *Transaction) Value() *big.Int {
	return new(big.Int).Set(tx.data.Amount)
}

// Gas는 트랜잭션에 설정된 가스 제한을 반환합니다.
func (tx *Transaction) Gas() uint64 {
	return tx.data.GasLimit
}

// GasPrice는 트랜잭션에 설정된 가스 가격을 반환합니다.
func (tx *Transaction) GasPrice() *big.Int {
	return new(big.Int).Set(tx.data.Price)
}

// GasTipCap는 트랜잭션에 설정된 팁 상한을 반환합니다.
func (tx *Transaction) GasTipCap() *big.Int {
	return new(big.Int).Set(tx.data.Price)
}

// GasFeeCap는 트랜잭션에 설정된 수수료 상한을 반환합니다.
func (tx *Transaction) GasFeeCap() *big.Int {
	return new(big.Int).Set(tx.data.Price)
}

// Data는 트랜잭션의 입력 데이터를 반환합니다.
func (tx *Transaction) Data() []byte {
	return common.CopyBytes(tx.data.Payload)
}

// AccessList는 트랜잭션의 액세스 리스트를 반환합니다.
func (tx *Transaction) AccessList() AccessList {
	return tx.data.accessList()
}

// RawSignatureValues는 서명 값 v, r, s를 반환합니다.
func (tx *Transaction) RawSignatureValues() (v, r, s *big.Int) {
	return tx.data.rawSignatureValues()
}

// Hash는 트랜잭션의 해시를 반환합니다.
func (tx *Transaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	
	v := rlpHash(tx)
	tx.hash.Store(v)
	return v
}

// Size는 트랜잭션의 직렬화된 크기를 바이트 단위로 반환합니다.
func (tx *Transaction) Size() int {
	if size := tx.size.Load(); size != nil {
		return size.(int)
	}
	c := writeCounter(0)
	rlp.Encode(&c, tx)
	tx.size.Store(int(c))
	return int(c)
}

// 발신자 주소를 설정합니다.
func (tx *Transaction) SetFrom(addr common.Address) {
	tx.from.Store(addr)
}

// From은 발신자 주소를 반환합니다.
func (tx *Transaction) From() (common.Address, error) {
	if from := tx.from.Load(); from != nil {
		return from.(common.Address), nil
	}
	
	addr, err := Sender(tx)
	if err != nil {
		return common.Address{}, err
	}
	tx.from.Store(addr)
	return addr, nil
}

// SignatureValues는 트랜잭션 서명 값 v, r, s를 반환하고 서명 값이 유효한지 확인합니다.
func (tx *Transaction) SignatureValues() (v, r, s *big.Int, err error) {
	v, r, s = tx.data.rawSignatureValues()
	
	if isProtectedV(v) {
		chainID := deriveChainID(v)
		if chainID.Cmp(tx.ChainID()) != 0 {
			return nil, nil, nil, ErrInvalidChainID
		}
	}
	
	if !crypto.ValidateSignatureValues(v, r, s, false) {
		return nil, nil, nil, ErrInvalidSig
	}
	
	return v, r, s, nil
}

// WithSignature는 지정된 서명으로 새 트랜잭션을 반환합니다.
func (tx *Transaction) WithSignature(signer Signer, sig []byte) (*Transaction, error) {
	r, s, v, err := signer.SignatureValues(tx, sig)
	if err != nil {
		return nil, err
	}
	cpy := tx.data.copy()
	cpy.setSignatureValues(signer.ChainID(), v, r, s)
	return &Transaction{data: cpy}, nil
}

// LegacyTx는 일반 이더리움 트랜잭션을 나타냅니다.
type LegacyTx struct {
	Nonce    uint64          // 발신자 계정 논스
	GasPrice *big.Int        // 가스당 Wei
	Gas      uint64          // 가스 제한
	To       *common.Address // 수신자 주소 (nil = 컨트랙트 생성)
	Value    *big.Int        // 전송할 wei 금액
	Data     []byte          // 컨트랙트 데이터
	V, R, S  *big.Int        // 서명 값
}

// txType은 레거시 트랜잭션임을 나타내는 0x00을 반환합니다.
func (tx *LegacyTx) txType() byte { return LegacyTxType }

// copy는 tx의 깊은 복사본을 만듭니다.
func (tx *LegacyTx) copy() TxData {
	cpy := &LegacyTx{
		Nonce:    tx.Nonce,
		GasPrice: new(big.Int).Set(tx.GasPrice),
		Gas:      tx.Gas,
		To:       nil,
		Value:    new(big.Int).Set(tx.Value),
		Data:     common.CopyBytes(tx.Data),
		V:        new(big.Int).Set(tx.V),
		R:        new(big.Int).Set(tx.R),
		S:        new(big.Int).Set(tx.S),
	}
	if tx.To != nil {
		to := *tx.To
		cpy.To = &to
	}
	return cpy
}

// accessList는 레거시 트랜잭션에는 액세스 리스트가 없으므로 nil을 반환합니다.
func (tx *LegacyTx) accessList() AccessList { return nil }

// data는 트랜잭션 데이터를 반환합니다.
func (tx *LegacyTx) data() []byte { return tx.Data }

// gas는 가스 제한을 반환합니다.
func (tx *LegacyTx) gas() uint64 { return tx.Gas }

// gasPrice는 가스 가격을 반환합니다.
func (tx *LegacyTx) gasPrice() *big.Int { return tx.GasPrice }

// gasTipCap은 레거시 트랜잭션에서는 가스 가격과 같습니다.
func (tx *LegacyTx) gasTipCap() *big.Int { return tx.GasPrice }

// gasFeeCap은 레거시 트랜잭션에서는 가스 가격과 같습니다.
func (tx *LegacyTx) gasFeeCap() *big.Int { return tx.GasPrice }

// value는 송금되는 이더 금액을 반환합니다.
func (tx *LegacyTx) value() *big.Int { return tx.Value }

// nonce는 발신자 계정의 논스를 반환합니다.
func (tx *LegacyTx) nonce() uint64 { return tx.Nonce }

// to는 수신자 주소를 반환합니다.
func (tx *LegacyTx) to() *common.Address { return tx.To }

// chainID는 체인 ID를 v 값에서 추출합니다.
func (tx *LegacyTx) chainID() *big.Int {
	if tx.V.Sign() != 0 && isProtectedV(tx.V) {
		return deriveChainID(tx.V)
	}
	return nil
}

// rawSignatureValues는 서명 값을 반환합니다.
func (tx *LegacyTx) rawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

// setSignatureValues는 서명 값을 설정합니다.
func (tx *LegacyTx) setSignatureValues(chainID, v, r, s *big.Int) {
	if chainID.Sign() != 0 {
		v = new(big.Int).Add(v, new(big.Int).Mul(chainID, big.NewInt(2)))
		v = v.Add(v, big.NewInt(35))
	}
	tx.V, tx.R, tx.S = v, r, s
}

// AccessList는 EIP-2930 트랜잭션의 액세스 리스트입니다.
type AccessList []AccessTuple

// AccessTuple은 주소와 스토리지 키의 쌍입니다.
type AccessTuple struct {
	Address     common.Address `json:"address"`
	StorageKeys []common.Hash  `json:"storageKeys"`
}

// Helper 함수들

// isProtectedV는 v 값이 보호된(EIP-155) 트랜잭션인지 확인합니다.
func isProtectedV(v *big.Int) bool {
	if v.BitLen() <= 8 {
		v := v.Uint64()
		return v != 27 && v != 28
	}
	// 체인 ID가 0인 경우 실제로 보호되지 않음
	return true
}

// deriveChainID는 v 값에서 체인 ID를 추출합니다.
func deriveChainID(v *big.Int) *big.Int {
	if v.Sign() != 1 {
		return new(big.Int)
	}
	if v.Cmp(big.NewInt(35)) <= 0 {
		return new(big.Int)
	}
	// V = 2*chainID + 35 or V = 2*chainID + 36
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, big.NewInt(2))
}

// rlpHash는 x의 RLP 해시를 계산합니다.
func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

// writeCounter는 쓰기 작업을 개수만 세는 io.Writer입니다.
type writeCounter int

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}

// Signer는 트랜잭션에 서명하고 서명을 검증하는 인터페이스입니다.
type Signer interface {
	// 서명 관련 메소드
	ChainID() *big.Int
	Hash(tx *Transaction) common.Hash
	Sender(tx *Transaction) (common.Address, error)
	SignatureValues(tx *Transaction, sig []byte) (r, s, v *big.Int, err error)
	
	// 추가적인 서명 메소드
	SignTx(tx *Transaction, prv *ecdsa.PrivateKey) (*Transaction, error)
}

// Sender는 트랜잭션 서명으로부터 발신자 주소를 복구합니다.
func Sender(tx *Transaction) (common.Address, error) {
	v, r, s := tx.data.V, tx.data.R, tx.data.S
	
	if !crypto.ValidateSignatureValues(v, r, s, true) {
		return common.Address{}, ErrInvalidSig
	}
	
	// EIP155 서명: v = CHAIN_ID * 2 + 35/36 + {0,1}
	var chainID *big.Int
	if isProtectedV(v) {
		chainID = deriveChainID(v)
		v = new(big.Int).Sub(v, new(big.Int).Add(
			new(big.Int).Mul(chainID, big.NewInt(2)),
			big.NewInt(35),
		))
	}
	
	// 27/28을 0/1로 변환하여 secp256k1 라이브러리 호출을 준비
	v = new(big.Int).Sub(v, big.NewInt(27))
	
	msg := RLPSignHash(tx)
	pub, err := crypto.Ecrecover(msg[:], SignatureBytes(v, r, s))
	if err != nil {
		return common.Address{}, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return common.Address{}, errors.New("invalid public key")
	}
	
	var addr common.Address
	copy(addr[:], crypto.Keccak256(pub[1:])[12:])
	return addr, nil
}

// RLPSignHash는 서명할 트랜잭션 해시를 반환합니다.
func RLPSignHash(tx *Transaction) common.Hash {
	return rlpHash([]interface{}{
		tx.data.AccountNonce,
		tx.data.Price,
		tx.data.GasLimit,
		tx.data.Recipient,
		tx.data.Amount,
		tx.data.Payload,
		// EIP155 체인 ID 포함
		tx.ChainID(),
		uint(0),
		uint(0),
	})
}

// SignatureBytes는 v, r, s 값을 단일 서명 바이트로 결합합니다.
func SignatureBytes(v, r, s *big.Int) []byte {
	sig := make([]byte, 65)
	copy(sig[32-len(r.Bytes()):32], r.Bytes())
	copy(sig[64-len(s.Bytes()):64], s.Bytes())
	sig[64] = byte(v.Uint64())
	return sig
}

// ValidateSignatureValues는 서명 값이 유효한지 확인합니다.
func ValidateSignatureValues(v, r, s *big.Int, homestead bool) bool {
	if r.Sign() != 1 || s.Sign() != 1 {
		return false
	}
	
	// 2^256-1 제한 대신 경량 체크
	if r.BitLen() > 256 || s.BitLen() > 256 {
		return false
	}
	
	// EIP-2: homestead 기간 동안 s 값은 secp256k1n/2보다 작아야 함
	if homestead && s.Cmp(secp256k1HalfN) > 0 {
		return false
	}
	
	// 0 < r, s < secp256k1n 확인
	if r.Cmp(secp256k1N) >= 0 || s.Cmp(secp256k1N) >= 0 {
		return false
	}

	// v 값이 유효한지 확인
	if v.BitLen() > 8 {
		return true
	}
	return v.Uint64() == 27 || v.Uint64() == 28
}

// Sign은 지정된 개인 키로 트랜잭션에 서명합니다.
func (tx *Transaction) Sign(prv *ecdsa.PrivateKey, chainID *big.Int) (*Transaction, error) {
	h := RLPSignHash(tx)
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return nil, err
	}
	
	// v, r, s 값 설정
	v, r, s := sig[64], sig[:32], sig[32:64]
	tx.data.R = new(big.Int).SetBytes(r)
	tx.data.S = new(big.Int).SetBytes(s)
	
	if chainID != nil { // EIP155 서명
		tx.data.V = new(big.Int).Add(
			new(big.Int).Add(
				new(big.Int).Mul(chainID, big.NewInt(2)),
				big.NewInt(35),
			),
			new(big.Int).SetBytes([]byte{v}),
		)
	} else { // 레거시 서명
		tx.data.V = new(big.Int).Add(big.NewInt(27), new(big.Int).SetBytes([]byte{v}))
	}
	
	return tx, nil
}

// secp256k1 곡선 파라미터
var (
	secp256k1N     = crypto.S256().Params().N // 곡선 위수
	secp256k1HalfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
)

// txType은 레거시 트랜잭션 유형을 반환합니다.
func (tx *txdata) txType() byte {
	return LegacyTxType
}

// copy는 txdata의 딥 카피를 생성합니다.
func (tx *txdata) copy() txdata {
	cpy := *tx
	if tx.Price != nil {
		cpy.Price = new(big.Int).Set(tx.Price)
	}
	if tx.Amount != nil {
		cpy.Amount = new(big.Int).Set(tx.Amount)
	}
	if tx.V != nil {
		cpy.V = new(big.Int).Set(tx.V)
	}
	if tx.R != nil {
		cpy.R = new(big.Int).Set(tx.R)
	}
	if tx.S != nil {
		cpy.S = new(big.Int).Set(tx.S)
	}
	if tx.ChainID != nil {
		cpy.ChainID = new(big.Int).Set(tx.ChainID)
	}
	if tx.Payload != nil {
		cpy.Payload = common.CopyBytes(tx.Payload)
	}
	if tx.Recipient != nil {
		cpy := *tx.Recipient
		cpy.Recipient = &cpy
	}
	return cpy
}

// accessList는 빈 액세스 리스트를 반환합니다.
func (tx *txdata) accessList() AccessList {
	return AccessList{}
}

// rawSignatureValues는 서명 값 v, r, s를 반환합니다.
func (tx *txdata) rawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

// setSignatureValues는 서명 값을 설정합니다.
func (tx *txdata) setSignatureValues(chainID, v, r, s *big.Int) {
	if chainID != nil {
		tx.ChainID = chainID
	}
	tx.V, tx.R, tx.S = v, r, s
} 