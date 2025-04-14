package validators

import (
	"fmt"
	"sync"

	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

// Validator는 검증자 정보를 나타냅니다.
type Validator struct {
	PubKey []byte // 검증자 공개키
	Power  int64  // 검증자 투표력
}

// Manager는 검증자 세트를 관리합니다.
type Manager struct {
	validatorsByPubKey map[string]*Validator // 공개키로 인덱싱된 검증자
	totalPower         int64                 // 총 투표력
	pendingUpdates     []types.ValidatorUpdate // 대기 중인 검증자 업데이트
	mtx                sync.RWMutex          // 동시성 제어
	logger             log.Logger            // 로거
}

// NewManager는 새 검증자 관리자를 생성합니다.
func NewManager() *Manager {
	return &Manager{
		validatorsByPubKey: make(map[string]*Validator),
		totalPower:         0,
		pendingUpdates:     []types.ValidatorUpdate{},
		logger:             log.NewNopLogger(),
	}
}

// SetLogger는 로거를 설정합니다.
func (m *Manager) SetLogger(logger log.Logger) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.logger = logger
}

// InitValidators는 초기 검증자 세트를 설정합니다.
func (m *Manager) InitValidators(validators []types.ValidatorUpdate) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	// 기존 검증자 초기화
	m.validatorsByPubKey = make(map[string]*Validator)
	m.totalPower = 0

	// 새 검증자 추가
	for _, val := range validators {
		pubKeyStr := fmt.Sprintf("%X", val.PubKey.GetEd25519())
		validator := &Validator{
			PubKey: val.PubKey.GetEd25519(),
			Power:  val.Power,
		}
		m.validatorsByPubKey[pubKeyStr] = validator
		m.totalPower += val.Power

		m.logger.Info("Initialized validator", "pubKey", pubKeyStr, "power", val.Power)
	}

	m.logger.Info("Validators initialized", "count", len(validators), "totalPower", m.totalPower)
}

// AddValidator는 검증자를 추가합니다.
func (m *Manager) AddValidator(val types.ValidatorUpdate) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	pubKeyBytes := val.PubKey.GetEd25519()
	pubKeyStr := fmt.Sprintf("%X", pubKeyBytes)

	// 이미 존재하는 검증자 업데이트
	if existing, ok := m.validatorsByPubKey[pubKeyStr]; ok {
		// 투표력 0은 제거를 의미
		if val.Power == 0 {
			delete(m.validatorsByPubKey, pubKeyStr)
			m.totalPower -= existing.Power
			m.logger.Info("Removed validator", "pubKey", pubKeyStr)
		} else {
			// 투표력 변경
			m.totalPower = m.totalPower - existing.Power + val.Power
			existing.Power = val.Power
			m.logger.Info("Updated validator", "pubKey", pubKeyStr, "power", val.Power)
		}
	} else if val.Power > 0 {
		// 새 검증자 추가
		validator := &Validator{
			PubKey: pubKeyBytes,
			Power:  val.Power,
		}
		m.validatorsByPubKey[pubKeyStr] = validator
		m.totalPower += val.Power
		m.logger.Info("Added validator", "pubKey", pubKeyStr, "power", val.Power)
	}
}

// QueueUpdate는 검증자 업데이트를 대기열에 추가합니다.
func (m *Manager) QueueUpdate(val types.ValidatorUpdate) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.pendingUpdates = append(m.pendingUpdates, val)
	m.logger.Debug("Queued validator update", "pubKey", fmt.Sprintf("%X", val.PubKey.GetEd25519()), "power", val.Power)
}

// GetPendingUpdates는 대기 중인 검증자 업데이트를 반환하고 대기열을 비웁니다.
func (m *Manager) GetPendingUpdates() []types.ValidatorUpdate {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	updates := m.pendingUpdates
	m.pendingUpdates = []types.ValidatorUpdate{}
	return updates
}

// GetValidatorCount는 현재 검증자 수를 반환합니다.
func (m *Manager) GetValidatorCount() int {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	return len(m.validatorsByPubKey)
}

// GetTotalPower는 총 투표력을 반환합니다.
func (m *Manager) GetTotalPower() int64 {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	return m.totalPower
}

// GetValidators는 현재 검증자 목록을 반환합니다.
func (m *Manager) GetValidators() []*Validator {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	validators := make([]*Validator, 0, len(m.validatorsByPubKey))
	for _, v := range m.validatorsByPubKey {
		validators = append(validators, v)
	}

	return validators
}

// IsValidator는 주어진 공개키가 검증자인지 확인합니다.
func (m *Manager) IsValidator(pubKey []byte) bool {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	pubKeyStr := fmt.Sprintf("%X", pubKey)
	_, exists := m.validatorsByPubKey[pubKeyStr]
	return exists
}

// GetValidatorPower는 주어진 공개키의 검증자 투표력을 반환합니다.
func (m *Manager) GetValidatorPower(pubKey []byte) int64 {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	pubKeyStr := fmt.Sprintf("%X", pubKey)
	if val, exists := m.validatorsByPubKey[pubKeyStr]; exists {
		return val.Power
	}
	return 0
} 