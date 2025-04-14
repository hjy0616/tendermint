package ethereum

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"
	"sync"

	"github.com/zenanetwork/go-zenanet/common"
)

// State는 이더리움 상태 스냅샷을 나타냅니다.
type State struct {
	accounts map[common.Address]*Account // 계정 상태 저장
	stateRoot []byte                     // 상태 루트 해시
	height    int64                      // 현재 블록 높이
	mu        sync.RWMutex               // 동시성 제어를 위한 뮤텍스
}

// Account는 이더리움 계정 상태를 나타냅니다.
type Account struct {
	Nonce    uint64            // 계정 논스
	Balance  *big.Int          // 계정 잔액
	Root     []byte            // 스토리지 루트 해시
	CodeHash []byte            // 계정 코드 해시
	Code     []byte            // 계정 코드
	Storage  map[string][]byte // 계정 스토리지
	Deleted  bool              // 삭제된 계정 여부
}

// NewState는 새 상태 인스턴스를 생성합니다.
func NewState() *State {
	return &State{
		accounts: make(map[common.Address]*Account),
		height:   0,
	}
}

// Copy는 상태의 딥 카피를 생성합니다.
func (s *State) Copy() *State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	newState := NewState()
	newState.height = s.height
	newState.stateRoot = s.stateRoot

	for addr, account := range s.accounts {
		newState.accounts[addr] = &Account{
			Nonce:    account.Nonce,
			Balance:  new(big.Int).Set(account.Balance),
			Root:     account.Root,
			CodeHash: account.CodeHash,
			Code:     account.Code,
			Storage:  make(map[string][]byte),
			Deleted:  account.Deleted,
		}

		for key, value := range account.Storage {
			newState.accounts[addr].Storage[key] = value
		}
	}

	return newState
}

// GetAccount은 주어진 주소의 계정을 반환합니다.
func (s *State) GetAccount(addr common.Address) *Account {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if account, ok := s.accounts[addr]; ok && !account.Deleted {
		return account
	}
	return nil
}

// CreateAccount는 새 계정을 생성합니다.
func (s *State) CreateAccount(addr common.Address) *Account {
	s.mu.Lock()
	defer s.mu.Unlock()

	account := &Account{
		Nonce:    0,
		Balance:  new(big.Int),
		Storage:  make(map[string][]byte),
		Deleted:  false,
	}
	s.accounts[addr] = account
	return account
}

// AddBalance은 계정 잔액을 증가시킵니다.
func (s *State) AddBalance(addr common.Address, amount *big.Int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, ok := s.accounts[addr]
	if !ok || account.Deleted {
		account = &Account{
			Nonce:    0,
			Balance:  new(big.Int),
			Storage:  make(map[string][]byte),
			Deleted:  false,
		}
		s.accounts[addr] = account
	}

	account.Balance = new(big.Int).Add(account.Balance, amount)
}

// SubBalance은 계정 잔액을 감소시킵니다.
func (s *State) SubBalance(addr common.Address, amount *big.Int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, ok := s.accounts[addr]
	if !ok || account.Deleted {
		return fmt.Errorf("account not found: %s", addr.Hex())
	}

	if account.Balance.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient funds: address %v has %v but tried to send %v",
			addr.Hex(), account.Balance, amount)
	}

	account.Balance = new(big.Int).Sub(account.Balance, amount)
	return nil
}

// GetBalance는 계정 잔액을 반환합니다.
func (s *State) GetBalance(addr common.Address) *big.Int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, ok := s.accounts[addr]
	if !ok || account.Deleted {
		return new(big.Int)
	}
	return new(big.Int).Set(account.Balance)
}

// GetNonce는 계정의 논스를 반환합니다.
func (s *State) GetNonce(addr common.Address) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, ok := s.accounts[addr]
	if !ok || account.Deleted {
		return 0
	}
	return account.Nonce
}

// SetNonce는 계정의 논스를 설정합니다.
func (s *State) SetNonce(addr common.Address, nonce uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, ok := s.accounts[addr]
	if !ok || account.Deleted {
		account = &Account{
			Nonce:    0,
			Balance:  new(big.Int),
			Storage:  make(map[string][]byte),
			Deleted:  false,
		}
		s.accounts[addr] = account
	}

	account.Nonce = nonce
}

// GetCode는 계정 코드를 반환합니다.
func (s *State) GetCode(addr common.Address) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, ok := s.accounts[addr]
	if !ok || account.Deleted {
		return nil
	}
	return account.Code
}

// SetCode는 계정 코드를 설정합니다.
func (s *State) SetCode(addr common.Address, code []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, ok := s.accounts[addr]
	if !ok || account.Deleted {
		account = &Account{
			Nonce:    0,
			Balance:  new(big.Int),
			Storage:  make(map[string][]byte),
			Deleted:  false,
		}
		s.accounts[addr] = account
	}

	account.Code = code
	// TODO: 실제 구현에서는 코드 해시를 계산해야 합니다.
	account.CodeHash = []byte("codehash")
}

// GetState는 계정 스토리지에서 키에 대한 값을 반환합니다.
func (s *State) GetState(addr common.Address, key string) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, ok := s.accounts[addr]
	if !ok || account.Deleted {
		return nil
	}

	if value, ok := account.Storage[key]; ok {
		return value
	}
	return nil
}

// SetState는 계정 스토리지의 키에 값을 설정합니다.
func (s *State) SetState(addr common.Address, key string, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, ok := s.accounts[addr]
	if !ok || account.Deleted {
		account = &Account{
			Nonce:    0,
			Balance:  new(big.Int),
			Storage:  make(map[string][]byte),
			Deleted:  false,
		}
		s.accounts[addr] = account
	}

	if value == nil || len(value) == 0 {
		delete(account.Storage, key)
	} else {
		account.Storage[key] = value
	}
}

// DeleteAccount는 계정을 삭제합니다.
func (s *State) DeleteAccount(addr common.Address) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if account, ok := s.accounts[addr]; ok {
		account.Deleted = true
	}
}

// Commit은 현재 상태를 커밋하고 상태 루트를 계산합니다.
func (s *State) Commit(height int64) []byte {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 계정을 정렬하여 결정적인 상태 루트를 보장합니다.
	addresses := make([]common.Address, 0, len(s.accounts))
	for addr, account := range s.accounts {
		if !account.Deleted {
			addresses = append(addresses, addr)
		} else {
			delete(s.accounts, addr)
		}
	}

	// 주소 정렬
	sort.Slice(addresses, func(i, j int) bool {
		return bytes.Compare(addresses[i][:], addresses[j][:]) < 0
	})

	// 간단한 해시 계산 (실제 구현에서는 머클 트리 사용)
	buffer := new(bytes.Buffer)
	for _, addr := range addresses {
		account := s.accounts[addr]
		
		// 계정 데이터 직렬화
		buffer.Write(addr[:])
		buffer.WriteString(fmt.Sprintf("%d", account.Nonce))
		buffer.WriteString(account.Balance.String())
		buffer.Write(account.CodeHash)
		
		// 스토리지 키 정렬
		storageKeys := make([]string, 0, len(account.Storage))
		for key := range account.Storage {
			storageKeys = append(storageKeys, key)
		}
		sort.Strings(storageKeys)
		
		// 스토리지 데이터 직렬화
		for _, key := range storageKeys {
			buffer.WriteString(key)
			buffer.Write(account.Storage[key])
		}
	}

	// TODO: 실제 구현에서는 keccak256 해시 함수 사용
	s.stateRoot = []byte(fmt.Sprintf("stateroot-%d", height))
	s.height = height
	
	return s.stateRoot
}

// GetStateRoot는 현재 상태 루트를 반환합니다.
func (s *State) GetStateRoot() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.stateRoot
}

// GetHeight는 현재 상태 높이를 반환합니다.
func (s *State) GetHeight() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.height
} 