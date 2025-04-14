package ethereum

import (
	"errors"
	"math/big"

	"github.com/zenanetwork/go-zenanet/common"
	"github.com/zenanetwork/go-zenanet/core/vm"
	"github.com/zenanetwork/go-zenanet/params"
)

var (
	ErrOutOfGas           = errors.New("out of gas")
	ErrCodeStoreOutOfGas  = errors.New("contract creation code storage out of gas")
	ErrDepth              = errors.New("max call depth exceeded")
	ErrContractAddressCollision = errors.New("contract address collision")
	ErrExecutionReverted  = errors.New("execution reverted")
	ErrMaxCodeSizeExceeded = errors.New("max code size exceeded")
)

// Context는 EVM 실행에 필요한 컨텍스트 정보를 포함합니다.
type Context struct {
	CanTransfer    func(StateDB, common.Address, *big.Int) bool
	Transfer       func(StateDB, common.Address, common.Address, *big.Int)
	GetHash        func(uint64) common.Hash
	Origin         common.Address
	Coinbase       common.Address
	BlockNumber    *big.Int
	Time           *big.Int
	Difficulty     *big.Int
	GasLimit       uint64
	GasPrice       *big.Int
	ChainID        *big.Int
}

// NewEVMContext는 새로운 EVM 컨텍스트를 생성합니다.
func NewEVMContext(origin common.Address, coinbase common.Address, blockNumber, time, gasLimit, gasPrice *big.Int) Context {
	return Context{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:     func(uint64) common.Hash { return common.Hash{} },
		Origin:      origin,
		Coinbase:    coinbase,
		BlockNumber: blockNumber,
		Time:        time,
		Difficulty:  big.NewInt(0),
		GasLimit:    gasLimit.Uint64(),
		GasPrice:    gasPrice,
		ChainID:     big.NewInt(1), // 기본 체인 ID
	}
}

// StateDB는 이더리움 상태와 상호작용하는 인터페이스입니다.
type StateDB interface {
	CreateAccount(common.Address)
	SubBalance(common.Address, *big.Int)
	AddBalance(common.Address, *big.Int)
	GetBalance(common.Address) *big.Int
	GetNonce(common.Address) uint64
	SetNonce(common.Address, uint64)
	GetCodeHash(common.Address) common.Hash
	GetCode(common.Address) []byte
	SetCode(common.Address, []byte)
	GetState(common.Address, common.Hash) common.Hash
	SetState(common.Address, common.Hash, common.Hash)
	Suicide(common.Address) bool
	HasSuicided(common.Address) bool
	Exist(common.Address) bool
	Empty(common.Address) bool
	RevertToSnapshot(int)
	Snapshot() int
	AddRefund(uint64)
	GetRefund() uint64
	AddLog(*vm.Log)
}

// StateAdapter는 State를 StateDB 인터페이스로 변환합니다.
type StateAdapter struct {
	state *State
}

// NewStateAdapter는 새로운 StateAdapter를 생성합니다.
func NewStateAdapter(state *State) *StateAdapter {
	return &StateAdapter{state: state}
}

// CreateAccount는 새 계정을 생성합니다.
func (s *StateAdapter) CreateAccount(addr common.Address) {
	s.state.CreateAccount(addr)
}

// SubBalance는 계정 잔액을 감소시킵니다.
func (s *StateAdapter) SubBalance(addr common.Address, amount *big.Int) {
	_ = s.state.SubBalance(addr, amount)
}

// AddBalance는 계정 잔액을 증가시킵니다.
func (s *StateAdapter) AddBalance(addr common.Address, amount *big.Int) {
	s.state.AddBalance(addr, amount)
}

// GetBalance는 계정 잔액을 반환합니다.
func (s *StateAdapter) GetBalance(addr common.Address) *big.Int {
	return s.state.GetBalance(addr)
}

// GetNonce는 계정의 논스를 반환합니다.
func (s *StateAdapter) GetNonce(addr common.Address) uint64 {
	return s.state.GetNonce(addr)
}

// SetNonce는 계정의 논스를 설정합니다.
func (s *StateAdapter) SetNonce(addr common.Address, nonce uint64) {
	s.state.SetNonce(addr, nonce)
}

// GetCodeHash는 계정 코드의 해시를 반환합니다.
func (s *StateAdapter) GetCodeHash(addr common.Address) common.Hash {
	// 실제 구현에서는 계정의 코드 해시를 반환해야 합니다.
	return common.Hash{}
}

// GetCode는 계정 코드를 반환합니다.
func (s *StateAdapter) GetCode(addr common.Address) []byte {
	return s.state.GetCode(addr)
}

// SetCode는 계정 코드를 설정합니다.
func (s *StateAdapter) SetCode(addr common.Address, code []byte) {
	s.state.SetCode(addr, code)
}

// GetState는 계정 스토리지에서 키에 대한 값을 반환합니다.
func (s *StateAdapter) GetState(addr common.Address, key common.Hash) common.Hash {
	value := s.state.GetState(addr, key.Hex())
	if value == nil {
		return common.Hash{}
	}
	
	var hash common.Hash
	copy(hash[:], value)
	return hash
}

// SetState는 계정 스토리지의 키에 값을 설정합니다.
func (s *StateAdapter) SetState(addr common.Address, key common.Hash, value common.Hash) {
	s.state.SetState(addr, key.Hex(), value[:])
}

// Suicide는 계정을 자살(삭제) 상태로 표시합니다.
func (s *StateAdapter) Suicide(addr common.Address) bool {
	s.state.DeleteAccount(addr)
	return true
}

// HasSuicided는 계정이 자살(삭제) 상태인지 확인합니다.
func (s *StateAdapter) HasSuicided(addr common.Address) bool {
	account := s.state.GetAccount(addr)
	return account == nil // nil인 경우 삭제된 것으로 간주
}

// Exist는 계정이 존재하는지 확인합니다.
func (s *StateAdapter) Exist(addr common.Address) bool {
	return s.state.GetAccount(addr) != nil
}

// Empty는 계정이 비어 있는지 확인합니다.
func (s *StateAdapter) Empty(addr common.Address) bool {
	account := s.state.GetAccount(addr)
	return account == nil || 
		(account.Balance.Sign() == 0 && 
		account.Nonce == 0 && 
		(account.Code == nil || len(account.Code) == 0))
}

// RevertToSnapshot은 지정된 스냅샷으로 상태를 되돌립니다.
func (s *StateAdapter) RevertToSnapshot(id int) {
	// 실제 구현에서는 스냅샷 관리 필요
}

// Snapshot은 현재 상태의 스냅샷을 만들고 스냅샷 ID를 반환합니다.
func (s *StateAdapter) Snapshot() int {
	// 실제 구현에서는 스냅샷 관리 필요
	return 0
}

// AddRefund는 환불 카운터를 증가시킵니다.
func (s *StateAdapter) AddRefund(gas uint64) {
	// 실제 구현에서는 가스 환불 관리 필요
}

// GetRefund는 환불 카운터를 반환합니다.
func (s *StateAdapter) GetRefund() uint64 {
	// 실제 구현에서는 가스 환불 관리 필요
	return 0
}

// AddLog는 로그를 추가합니다.
func (s *StateAdapter) AddLog(log *vm.Log) {
	// 실제 구현에서는 로그 저장 필요
}

// CanTransfer는 주소가 금액을 송금할 수 있는지 확인합니다.
func CanTransfer(db StateDB, addr common.Address, amount *big.Int) bool {
	return db.GetBalance(addr).Cmp(amount) >= 0
}

// Transfer는 한 주소에서 다른 주소로 금액을 송금합니다.
func Transfer(db StateDB, sender, recipient common.Address, amount *big.Int) {
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}

// ExecuteTx는 트랜잭션을 실행합니다.
func ExecuteTx(state *State, tx *Transaction) ([]byte, uint64, error) {
	// 상태 어댑터 생성
	stateDB := NewStateAdapter(state)
	
	// 기본 가스 사용량
	intrinsicGas := uint64(21000) // 기본 트랜잭션 가스 비용
	
	// 발신자 주소 가져오기
	from, err := tx.From()
	if err != nil {
		return nil, 0, err
	}
	
	// 논스 검증
	if tx.Nonce() != stateDB.GetNonce(from) {
		return nil, 0, errors.New("invalid nonce")
	}
	
	// 가스 제한이 내재적 가스보다 작은지 확인
	if tx.Gas() < intrinsicGas {
		return nil, 0, ErrOutOfGas
	}
	
	// 발신자가 충분한 자금과 가스를 가지고 있는지 확인
	cost := new(big.Int).Mul(new(big.Int).SetUint64(tx.Gas()), tx.GasPrice())
	cost.Add(cost, tx.Value())
	if !CanTransfer(stateDB, from, cost) {
		return nil, 0, errors.New("insufficient funds for gas * price + value")
	}
	
	// 발신자의 가스 비용 차감
	stateDB.SubBalance(from, cost)
	
	// 남은 가스 초기화
	gasRemaining := tx.Gas()
	gasRemaining -= intrinsicGas
	
	// EVM 컨텍스트 생성
	context := NewEVMContext(
		from,
		common.Address{}, // coinbase (block producer)
		new(big.Int).SetInt64(0), // block number
		new(big.Int).SetInt64(0), // time
		new(big.Int).SetUint64(tx.Gas()), // gas limit
		tx.GasPrice(),
	)
	
	// EVM 설정
	config := params.MainnetChainConfig
	logConfig := vm.LogConfig{}
	
	// EVM 인스턴스 생성
	evm := vm.NewEVM(vm.BlockContext{}, vm.TxContext{}, stateDB, config, logConfig)
	
	// 초기 발신자 논스 증가
	stateDB.SetNonce(from, stateDB.GetNonce(from)+1)
	
	var (
		ret       []byte
		leftOverGas uint64
		contractAddr common.Address
		vmErr     error
	)
	
	// 계약 생성인지 메시지 호출인지 확인
	if tx.To() == nil {
		// 계약 생성
		ret, contractAddr, leftOverGas, vmErr = evm.Create(
			vm.AccountRef(from),
			tx.Data(),
			gasRemaining,
			tx.Value(),
		)
	} else {
		// 메시지 호출
		ret, leftOverGas, vmErr = evm.Call(
			vm.AccountRef(from),
			*tx.To(),
			tx.Data(),
			gasRemaining,
			tx.Value(),
		)
	}
	
	// 가스 사용량 계산
	gasUsed := tx.Gas() - leftOverGas
	
	// 상태 커밋
	// 실제 구현에서는 상태 커밋 전에 실패한 트랜잭션을 되돌릴 수 있어야 함
	
	// 환불 가스 금액 계산 및 환불
	refund := new(big.Int).Mul(new(big.Int).SetUint64(leftOverGas), tx.GasPrice())
	stateDB.AddBalance(from, refund)
	
	return ret, gasUsed, vmErr
} 