package validator_selection

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
)

var (
	// ErrValidatorNotFound 검증자를 찾을 수 없을 때 반환되는 에러
	ErrValidatorNotFound = errors.New("validator not found")

	// ErrInvalidValidatorSet 유효하지 않은 검증자 세트
	ErrInvalidValidatorSet = errors.New("invalid validator set")

	// ErrValidatorSyncTimeout 검증자 동기화 타임아웃
	ErrValidatorSyncTimeout = errors.New("validator sync timeout")
)

// ValidatorSetSyncer는 온체인 및 오프체인 검증자 세트 관리를 담당합니다.
type ValidatorSetSyncer struct {
	mtx               sync.RWMutex
	logger            log.Logger
	onChainValidators *types.ValidatorSet
	tendermintValSet  *types.ValidatorSet

	// 동기화 관련 설정
	syncInterval time.Duration
	syncTimeout  time.Duration

	// 동기화 상태 추적
	lastSyncTime   time.Time
	isSyncing      bool
	syncInProgress sync.WaitGroup
}

// NewValidatorSetSyncer 새로운 검증자 세트 동기화 관리자 생성
func NewValidatorSetSyncer(initialValSet *types.ValidatorSet) *ValidatorSetSyncer {
	return &ValidatorSetSyncer{
		logger:            log.NewNopLogger(),
		onChainValidators: initialValSet.Copy(),
		tendermintValSet:  initialValSet.Copy(),
		syncInterval:      time.Minute * 5,  // 기본 동기화 간격: 5분
		syncTimeout:       time.Second * 30, // 기본 동기화 타임아웃: 30초
		lastSyncTime:      time.Time{},
		isSyncing:         false,
	}
}

// SetLogger 로거 설정
func (vss *ValidatorSetSyncer) SetLogger(logger log.Logger) {
	vss.mtx.Lock()
	defer vss.mtx.Unlock()
	vss.logger = logger
}

// SetSyncInterval 동기화 간격 설정
func (vss *ValidatorSetSyncer) SetSyncInterval(interval time.Duration) {
	vss.mtx.Lock()
	defer vss.mtx.Unlock()
	vss.syncInterval = interval
}

// SetSyncTimeout 동기화 타임아웃 설정
func (vss *ValidatorSetSyncer) SetSyncTimeout(timeout time.Duration) {
	vss.mtx.Lock()
	defer vss.mtx.Unlock()
	vss.syncTimeout = timeout
}

// GetOnChainValidators 온체인 검증자 세트 반환
func (vss *ValidatorSetSyncer) GetOnChainValidators() *types.ValidatorSet {
	vss.mtx.RLock()
	defer vss.mtx.RUnlock()
	return vss.onChainValidators.Copy()
}

// GetTendermintValidators Tendermint 검증자 세트 반환
func (vss *ValidatorSetSyncer) GetTendermintValidators() *types.ValidatorSet {
	vss.mtx.RLock()
	defer vss.mtx.RUnlock()
	return vss.tendermintValSet.Copy()
}

// UpdateOnChainValidators 온체인 검증자 세트 업데이트
func (vss *ValidatorSetSyncer) UpdateOnChainValidators(newValSet *types.ValidatorSet) error {
	if newValSet == nil {
		return ErrInvalidValidatorSet
	}

	vss.mtx.Lock()
	defer vss.mtx.Unlock()

	// 변경사항 로깅
	added, removed, updated := ValidatorSetDifference(vss.onChainValidators, newValSet)

	// 로그 출력
	if len(added) > 0 || len(removed) > 0 || len(updated) > 0 {
		vss.logger.Info("Updating on-chain validator set",
			"added", len(added),
			"removed", len(removed),
			"updated", len(updated),
			"total", newValSet.Size())

		for _, val := range added {
			vss.logger.Info("Validator added to on-chain set",
				"address", fmt.Sprintf("%X", val.Address),
				"voting_power", val.VotingPower)
		}

		for _, val := range removed {
			vss.logger.Info("Validator removed from on-chain set",
				"address", fmt.Sprintf("%X", val.Address),
				"voting_power", val.VotingPower)
		}

		for _, val := range updated {
			vss.logger.Info("Validator updated in on-chain set",
				"address", fmt.Sprintf("%X", val.Address),
				"voting_power", val.VotingPower)
		}
	}

	// 새 검증자 세트로 교체
	vss.onChainValidators = newValSet.Copy()
	return nil
}

// UpdateTendermintValidators Tendermint 검증자 세트 업데이트
func (vss *ValidatorSetSyncer) UpdateTendermintValidators(newValSet *types.ValidatorSet) error {
	if newValSet == nil {
		return ErrInvalidValidatorSet
	}

	vss.mtx.Lock()
	defer vss.mtx.Unlock()

	// 변경사항 로깅
	added, removed, updated := ValidatorSetDifference(vss.tendermintValSet, newValSet)

	// 로그 출력
	if len(added) > 0 || len(removed) > 0 || len(updated) > 0 {
		vss.logger.Info("Updating Tendermint validator set",
			"added", len(added),
			"removed", len(removed),
			"updated", len(updated),
			"total", newValSet.Size())

		for _, val := range added {
			vss.logger.Info("Validator added to Tendermint set",
				"address", fmt.Sprintf("%X", val.Address),
				"voting_power", val.VotingPower)
		}

		for _, val := range removed {
			vss.logger.Info("Validator removed from Tendermint set",
				"address", fmt.Sprintf("%X", val.Address),
				"voting_power", val.VotingPower)
		}

		for _, val := range updated {
			vss.logger.Info("Validator updated in Tendermint set",
				"address", fmt.Sprintf("%X", val.Address),
				"voting_power", val.VotingPower)
		}
	}

	// 새 검증자 세트로 교체
	vss.tendermintValSet = newValSet.Copy()
	return nil
}

// SyncValidatorSets 온체인과 Tendermint 검증자 세트 동기화
func (vss *ValidatorSetSyncer) SyncValidatorSets(ctx context.Context) error {
	vss.mtx.Lock()

	// 이미 동기화 중인지 확인
	if vss.isSyncing {
		vss.mtx.Unlock()
		return nil
	}

	// 마지막 동기화 시간 확인
	if time.Since(vss.lastSyncTime) < vss.syncInterval {
		vss.mtx.Unlock()
		return nil
	}

	vss.isSyncing = true
	vss.syncInProgress.Add(1)
	vss.mtx.Unlock()

	defer func() {
		vss.mtx.Lock()
		vss.isSyncing = false
		vss.lastSyncTime = time.Now()
		vss.mtx.Unlock()
		vss.syncInProgress.Done()
	}()

	// 컨텍스트에 타임아웃 추가
	ctx, cancel := context.WithTimeout(ctx, vss.syncTimeout)
	defer cancel()

	// 온체인 검증자 세트를 Tendermint 검증자 세트로 동기화
	// 실제 구현에서는 여기서 네트워크 호출 또는 상태 DB 접근이 발생할 수 있음
	vss.mtx.Lock()
	tendermintValSet := vss.tendermintValSet.Copy()
	onChainValSet := vss.onChainValidators.Copy()
	vss.mtx.Unlock()

	// 동기화가 필요한지 확인
	needsSync := false
	added, removed, updated := ValidatorSetDifference(tendermintValSet, onChainValSet)
	if len(added) > 0 || len(removed) > 0 || len(updated) > 0 {
		needsSync = true
	}

	if !needsSync {
		vss.logger.Debug("Validator sets already in sync, no update needed")
		return nil
	}

	// 온체인 검증자 세트를 Tendermint 검증자 세트로 적용
	err := vss.UpdateTendermintValidators(onChainValSet)
	if err != nil {
		vss.logger.Error("Failed to sync validator sets", "error", err)
		return err
	}

	vss.logger.Info("Successfully synchronized validator sets",
		"total_validators", onChainValSet.Size())

	return nil
}

// WaitForSync 동기화 완료 대기
func (vss *ValidatorSetSyncer) WaitForSync() {
	vss.syncInProgress.Wait()
}

// IsProposer 주어진 주소가 현재 제안자인지 확인
func (vss *ValidatorSetSyncer) IsProposer(address []byte) bool {
	vss.mtx.RLock()
	defer vss.mtx.RUnlock()

	if vss.tendermintValSet == nil || vss.tendermintValSet.Size() == 0 {
		return false
	}

	proposerIdx := PriorityBasedProposer(vss.tendermintValSet)
	if proposerIdx < 0 || proposerIdx >= vss.tendermintValSet.Size() {
		return false
	}

	proposer := vss.tendermintValSet.Validators[proposerIdx]
	return proposer != nil && len(address) > 0 && len(proposer.Address) > 0 &&
		bytes.Equal(proposer.Address, address)
}

// IsValidator 주어진 주소가 검증자인지 확인
func (vss *ValidatorSetSyncer) IsValidator(address []byte) bool {
	vss.mtx.RLock()
	defer vss.mtx.RUnlock()

	if vss.tendermintValSet == nil || vss.tendermintValSet.Size() == 0 {
		return false
	}

	idx, found := GetValidatorByAddress(vss.tendermintValSet, address)
	return found && idx >= 0
}

// GetVotingPower 주어진 주소의 투표력 반환
func (vss *ValidatorSetSyncer) GetVotingPower(address []byte) int64 {
	vss.mtx.RLock()
	defer vss.mtx.RUnlock()

	if vss.tendermintValSet == nil || vss.tendermintValSet.Size() == 0 {
		return 0
	}

	idx, found := GetValidatorByAddress(vss.tendermintValSet, address)
	if !found || idx < 0 {
		return 0
	}

	return vss.tendermintValSet.Validators[idx].VotingPower
}

// GetValidatorSecurityInfo 검증자 세트 보안 정보 반환
// 네트워크 분산도, 평균 투표력, 총 투표력 등 분석 정보 제공
func (vss *ValidatorSetSyncer) GetValidatorSecurityInfo() map[string]interface{} {
	vss.mtx.RLock()
	defer vss.mtx.RUnlock()

	if vss.tendermintValSet == nil || vss.tendermintValSet.Size() == 0 {
		return map[string]interface{}{
			"validators_count": 0,
			"total_power":      0,
			"avg_power":        0,
			"max_power":        0,
			"min_power":        0,
			"power_variance":   0,
		}
	}

	totalPower := int64(0)
	maxPower := int64(0)
	minPower := int64(-1)

	for i := 0; i < vss.tendermintValSet.Size(); i++ {
		power := vss.tendermintValSet.Validators[i].VotingPower
		totalPower += power

		if power > maxPower {
			maxPower = power
		}

		if minPower == -1 || power < minPower {
			minPower = power
		}
	}

	avgPower := float64(0)
	if vss.tendermintValSet.Size() > 0 {
		avgPower = float64(totalPower) / float64(vss.tendermintValSet.Size())
	}

	// 분산 계산
	variance := float64(0)
	for i := 0; i < vss.tendermintValSet.Size(); i++ {
		power := float64(vss.tendermintValSet.Validators[i].VotingPower)
		diff := power - avgPower
		variance += diff * diff
	}

	if vss.tendermintValSet.Size() > 0 {
		variance /= float64(vss.tendermintValSet.Size())
	}

	return map[string]interface{}{
		"validators_count": vss.tendermintValSet.Size(),
		"total_power":      totalPower,
		"avg_power":        avgPower,
		"max_power":        maxPower,
		"min_power":        minPower,
		"power_variance":   variance,
		"last_sync_time":   vss.lastSyncTime,
	}
}
