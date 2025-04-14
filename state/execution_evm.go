package state

import (
	"fmt"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
)

// DefaultEVMExecutor는 기본 EVM 실행 엔진 구현체입니다.
type DefaultEVMExecutor struct {
	// EVM 실행에 필요한 설정
	GasLimit uint64
}

// NewDefaultEVMExecutor는 새로운 DefaultEVMExecutor 인스턴스를 생성합니다.
func NewDefaultEVMExecutor(gasLimit uint64) *DefaultEVMExecutor {
	return &DefaultEVMExecutor{
		GasLimit: gasLimit,
	}
}

// Execute는 주어진 트랜잭션을 EVM 환경에서 실행합니다.
func (e *DefaultEVMExecutor) Execute(ctx EVMContext, stateDB EVMStateDB, tx []byte) ([]byte, uint64, error) {
	// 여기서는 실제 EVM 구현이 필요합니다.
	// 현재는 기본 인터페이스만 제공하고, 나중에 실제 EVM 라이브러리와 연동할 수 있습니다.

	// 가상의 EVM 실행 결과
	result := []byte("evm execution result")
	gasUsed := uint64(21000) // 기본 트랜잭션 비용

	return result, gasUsed, nil
}

// ProcessEVMTx는 EVM 트랜잭션을 처리하고 ABCI 응답을 반환합니다.
func ProcessEVMTx(executor EVMExecutor, ctx EVMContext, stateDB EVMStateDB, tx []byte) (*abci.ResponseDeliverTx, error) {
	// 트랜잭션 실행
	result, gasUsed, err := executor.Execute(ctx, stateDB, tx)
	if err != nil {
		return &abci.ResponseDeliverTx{
			Code:      1,
			Log:       fmt.Sprintf("EVM execution failed: %v", err),
			GasWanted: int64(executor.(*DefaultEVMExecutor).GasLimit),
			GasUsed:   int64(gasUsed),
		}, err
	}

	// 성공 응답 생성
	return &abci.ResponseDeliverTx{
		Code:      0,
		Data:      result,
		Log:       "EVM execution succeeded",
		GasWanted: int64(executor.(*DefaultEVMExecutor).GasLimit),
		GasUsed:   int64(gasUsed),
		Events: []abci.Event{
			{
				Type: "ethereum_tx",
				Attributes: []abci.EventAttribute{
					{Key: []byte("block_number"), Value: []byte(fmt.Sprintf("%d", ctx.BlockNumber))},
					{Key: []byte("block_time"), Value: []byte(ctx.BlockTime.Format(time.RFC3339))},
					{Key: []byte("tx_hash"), Value: ctx.TransactionHash},
				},
			},
		},
	}, nil
}

// DefaultEVMStateDB는 EVM 상태 접근을 위한 간단한 구현체입니다.
type DefaultEVMStateDB struct {
	accounts  map[string]EVMAccount
	storage   map[string]map[string][]byte
	logs      []*EVMLog
	preimages map[string][]byte
	snapshots []Snapshot
	stateRoot []byte
}

// EVMAccount는 EVM 계정 정보를 저장합니다.
type EVMAccount struct {
	Balance  []byte
	Nonce    uint64
	Code     []byte
	CodeHash []byte
}

// Snapshot은 상태 스냅샷 정보를 저장합니다.
type Snapshot struct {
	Accounts  map[string]EVMAccount
	Storage   map[string]map[string][]byte
	LogLength int
}

// NewDefaultEVMStateDB는 새로운 DefaultEVMStateDB 인스턴스를 생성합니다.
func NewDefaultEVMStateDB() *DefaultEVMStateDB {
	return &DefaultEVMStateDB{
		accounts:  make(map[string]EVMAccount),
		storage:   make(map[string]map[string][]byte),
		logs:      make([]*EVMLog, 0),
		preimages: make(map[string][]byte),
		snapshots: make([]Snapshot, 0),
	}
}

// CreateAccount는 새로운 계정을 생성합니다.
func (db *DefaultEVMStateDB) CreateAccount(address []byte) {
	addrStr := string(address)
	db.accounts[addrStr] = EVMAccount{
		Balance:  []byte{0},
		Nonce:    0,
		Code:     []byte{},
		CodeHash: []byte{},
	}
	db.storage[addrStr] = make(map[string][]byte)
}

// GetBalance는 계정의 잔액을 조회합니다.
func (db *DefaultEVMStateDB) GetBalance(address []byte) []byte {
	addrStr := string(address)
	if account, exists := db.accounts[addrStr]; exists {
		return account.Balance
	}
	return []byte{0}
}

// SetBalance는 계정의 잔액을 설정합니다.
func (db *DefaultEVMStateDB) SetBalance(address []byte, amount []byte) {
	addrStr := string(address)
	if account, exists := db.accounts[addrStr]; exists {
		account.Balance = amount
		db.accounts[addrStr] = account
	}
}

// GetNonce는 계정의 nonce를 조회합니다.
func (db *DefaultEVMStateDB) GetNonce(address []byte) uint64 {
	addrStr := string(address)
	if account, exists := db.accounts[addrStr]; exists {
		return account.Nonce
	}
	return 0
}

// SetNonce는 계정의 nonce를 설정합니다.
func (db *DefaultEVMStateDB) SetNonce(address []byte, nonce uint64) {
	addrStr := string(address)
	if account, exists := db.accounts[addrStr]; exists {
		account.Nonce = nonce
		db.accounts[addrStr] = account
	}
}

// GetCodeHash는 계정의 코드 해시를 조회합니다.
func (db *DefaultEVMStateDB) GetCodeHash(address []byte) []byte {
	addrStr := string(address)
	if account, exists := db.accounts[addrStr]; exists {
		return account.CodeHash
	}
	return []byte{}
}

// GetCode는 계정의 코드를 조회합니다.
func (db *DefaultEVMStateDB) GetCode(address []byte) []byte {
	addrStr := string(address)
	if account, exists := db.accounts[addrStr]; exists {
		return account.Code
	}
	return []byte{}
}

// SetCode는 계정의 코드를 설정합니다.
func (db *DefaultEVMStateDB) SetCode(address []byte, code []byte) {
	addrStr := string(address)
	if account, exists := db.accounts[addrStr]; exists {
		account.Code = code
		// 실제로는 여기서 코드 해시도 계산해야 합니다.
		account.CodeHash = []byte("code_hash")
		db.accounts[addrStr] = account
	}
}

// GetCodeSize는 계정의 코드 크기를 조회합니다.
func (db *DefaultEVMStateDB) GetCodeSize(address []byte) int {
	addrStr := string(address)
	if account, exists := db.accounts[addrStr]; exists {
		return len(account.Code)
	}
	return 0
}

// GetState는 계정의 스토리지 값을 조회합니다.
func (db *DefaultEVMStateDB) GetState(address []byte, key []byte) []byte {
	addrStr := string(address)
	keyStr := string(key)
	if storage, exists := db.storage[addrStr]; exists {
		if value, exists := storage[keyStr]; exists {
			return value
		}
	}
	return []byte{}
}

// SetState는 계정의 스토리지 값을 설정합니다.
func (db *DefaultEVMStateDB) SetState(address []byte, key []byte, value []byte) {
	addrStr := string(address)
	keyStr := string(key)
	if _, exists := db.storage[addrStr]; !exists {
		db.storage[addrStr] = make(map[string][]byte)
	}
	db.storage[addrStr][keyStr] = value
}

// AddLog는 로그를 추가합니다.
func (db *DefaultEVMStateDB) AddLog(log *EVMLog) {
	db.logs = append(db.logs, log)
}

// AddPreimage는 해시 프리이미지를 추가합니다.
func (db *DefaultEVMStateDB) AddPreimage(hash []byte, preimage []byte) {
	db.preimages[string(hash)] = preimage
}

// Snapshot은 현재 상태의 스냅샷을 생성합니다.
func (db *DefaultEVMStateDB) Snapshot() int {
	// 계정 복사
	accounts := make(map[string]EVMAccount)
	for addr, account := range db.accounts {
		accounts[addr] = account
	}

	// 스토리지 복사
	storage := make(map[string]map[string][]byte)
	for addr, addrStorage := range db.storage {
		storage[addr] = make(map[string][]byte)
		for key, value := range addrStorage {
			storage[addr][key] = value
		}
	}

	// 스냅샷 저장
	snapshot := Snapshot{
		Accounts:  accounts,
		Storage:   storage,
		LogLength: len(db.logs),
	}
	db.snapshots = append(db.snapshots, snapshot)

	return len(db.snapshots) - 1
}

// RevertToSnapshot은 지정된 스냅샷으로 상태를 되돌립니다.
func (db *DefaultEVMStateDB) RevertToSnapshot(index int) {
	if index < 0 || index >= len(db.snapshots) {
		return
	}

	snapshot := db.snapshots[index]

	// 계정 복원
	db.accounts = snapshot.Accounts

	// 스토리지 복원
	db.storage = snapshot.Storage

	// 로그 복원
	db.logs = db.logs[:snapshot.LogLength]

	// 스냅샷 목록 업데이트
	db.snapshots = db.snapshots[:index]
}

// Commit은 상태 변경을 커밋하고 새로운 상태 루트를 반환합니다.
func (db *DefaultEVMStateDB) Commit() ([]byte, error) {
	// 실제 구현에서는 머클 트리 등을 사용하여 상태 루트를 계산해야 합니다.
	// 여기서는 간단한 구현만 제공합니다.
	db.stateRoot = []byte("state_root")

	// 스냅샷 정리
	db.snapshots = nil

	return db.stateRoot, nil
}
