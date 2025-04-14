package validators

import (
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

// StakeInfo는 검증자 스테이킹 정보를 나타냅니다.
type StakeInfo struct {
	PubKey       []byte    // 검증자 공개키
	Address      string    // 검증자 주소
	Amount       *big.Int  // 스테이킹된 금액
	StakedAt     time.Time // 스테이킹 시간
	LastRewarded time.Time // 마지막 보상 시간
	Commission   uint32    // 커미션 비율 (1/10000 단위, 예: 100은 1%)
}

// Delegation은 위임 정보를 나타냅니다.
type Delegation struct {
	Delegator string   // 위임자 주소
	Validator string   // 검증자 주소
	Amount    *big.Int // 위임 금액
	Since     time.Time // 위임 시작 시간
}

// StakingManager는 검증자 스테이킹을 관리합니다.
type StakingManager struct {
	stakes            map[string]*StakeInfo      // 검증자 주소로 인덱싱된 스테이킹 정보
	delegations       map[string][]*Delegation   // 검증자 주소로 인덱싱된 위임 정보
	delegatorStakes   map[string][]*Delegation   // 위임자 주소로 인덱싱된 위임 정보
	minStakeAmount    *big.Int                   // 최소 스테이킹 금액
	maxValidators     int                         // 최대 검증자 수
	powerConversion   uint64                      // 투표력 변환 비율 (스테이크 금액 -> 투표력)
	validatorManager  *Manager                    // 검증자 관리자 참조
	cumulativeRewards map[string]*big.Int         // 검증자별 누적 보상
	annual_reward     *big.Int                    // 연간 보상 금액 (퍼센트)
	mtx               sync.RWMutex                // 동시성 제어
	logger            log.Logger                  // 로거
}

// NewStakingManager는 새 스테이킹 관리자를 생성합니다.
func NewStakingManager(validatorManager *Manager, minStake *big.Int, maxValidators int) *StakingManager {
	return &StakingManager{
		stakes:            make(map[string]*StakeInfo),
		delegations:       make(map[string][]*Delegation),
		delegatorStakes:   make(map[string][]*Delegation),
		minStakeAmount:    minStake,
		maxValidators:     maxValidators,
		powerConversion:   10000, // 10000 단위당 1 투표력
		validatorManager:  validatorManager,
		cumulativeRewards: make(map[string]*big.Int),
		annual_reward:     big.NewInt(5), // 5% 연간 보상
		logger:            log.NewNopLogger(),
	}
}

// SetLogger는 로거를 설정합니다.
func (sm *StakingManager) SetLogger(logger log.Logger) {
	sm.mtx.Lock()
	defer sm.mtx.Unlock()
	sm.logger = logger
}

// CalculateVotingPower는 스테이킹 금액에 기반한 투표력을 계산합니다.
func (sm *StakingManager) CalculateVotingPower(stakeAmount *big.Int) int64 {
	// 최소 스테이크 확인
	if stakeAmount.Cmp(sm.minStakeAmount) < 0 {
		return 0
	}

	// 스테이크 금액을 투표력으로 변환 (지정된 변환 비율로 나눔)
	power := new(big.Int).Div(stakeAmount, big.NewInt(int64(sm.powerConversion)))
	
	// 상한선 설정 (필요시)
	maxPower := big.NewInt(1000000) // 최대 투표력
	if power.Cmp(maxPower) > 0 {
		power = maxPower
	}

	return power.Int64()
}

// Stake는 검증자의 스테이킹을 처리합니다.
func (sm *StakingManager) Stake(pubKey []byte, address string, amount *big.Int, commission uint32) error {
	sm.mtx.Lock()
	defer sm.mtx.Unlock()

	// 최소 스테이크 확인
	if amount.Cmp(sm.minStakeAmount) < 0 {
		return fmt.Errorf("스테이크 금액이 최소 요구사항(%s)보다 적습니다", sm.minStakeAmount.String())
	}

	// 커미션 비율 검증 (0%~100%)
	if commission > 10000 {
		return fmt.Errorf("커미션 비율은 0에서 10000(100%%) 사이여야 합니다")
	}

	now := time.Now()
	info := &StakeInfo{
		PubKey:       pubKey,
		Address:      address,
		Amount:       new(big.Int).Set(amount),
		StakedAt:     now,
		LastRewarded: now,
		Commission:   commission,
	}

	// 기존 스테이크 업데이트 또는 새 스테이크 추가
	if existing, ok := sm.stakes[address]; ok {
		// 기존 금액에 추가
		existing.Amount = new(big.Int).Add(existing.Amount, amount)
		// 커미션 업데이트
		existing.Commission = commission
		sm.logger.Info("스테이크 업데이트", "주소", address, "총액", existing.Amount, "커미션", commission)
	} else {
		sm.stakes[address] = info
		sm.cumulativeRewards[address] = big.NewInt(0)
		sm.logger.Info("새 스테이커 추가", "주소", address, "금액", amount, "커미션", commission)
	}

	// 투표력 계산 및 검증자 업데이트
	votingPower := sm.CalculateVotingPower(sm.GetTotalStake(address))

	// Tendermint의 ValidatorUpdate 형식으로 변환
	update := types.Ed25519ValidatorUpdate(pubKey, votingPower)
	sm.validatorManager.QueueUpdate(update)

	return nil
}

// Unstake는 검증자의 스테이크 회수를 처리합니다.
func (sm *StakingManager) Unstake(address string, amount *big.Int) error {
	sm.mtx.Lock()
	defer sm.mtx.Unlock()

	stake, ok := sm.stakes[address]
	if !ok {
		return fmt.Errorf("주소 %s에 대한 스테이크를 찾을 수 없습니다", address)
	}

	// 언스테이크 금액이 전체 스테이크 금액보다 작거나 같은지 확인
	if amount.Cmp(stake.Amount) > 0 {
		return fmt.Errorf("언스테이크 금액(%s)이 스테이킹된 금액(%s)보다 큽니다", amount.String(), stake.Amount.String())
	}

	// 스테이크 금액 감소
	stake.Amount = new(big.Int).Sub(stake.Amount, amount)

	// 스테이크가 최소 요구사항 미만이면 검증자 제거
	if stake.Amount.Cmp(sm.minStakeAmount) < 0 {
		delete(sm.stakes, address)
		sm.logger.Info("스테이커 제거", "주소", address, "이유", "최소 스테이크 금액 미만")

		// 검증자 업데이트 큐에 추가 (투표력 0으로 설정해 제거)
		update := types.Ed25519ValidatorUpdate(stake.PubKey, 0)
		sm.validatorManager.QueueUpdate(update)
	} else {
		// 투표력 재계산
		votingPower := sm.CalculateVotingPower(sm.GetTotalStake(address))
		sm.logger.Info("스테이크 감소", "주소", address, "남은 금액", stake.Amount, "새 투표력", votingPower)

		// 검증자 업데이트 큐에 추가
		update := types.Ed25519ValidatorUpdate(stake.PubKey, votingPower)
		sm.validatorManager.QueueUpdate(update)
	}

	return nil
}

// Delegate는 위임을 처리합니다.
func (sm *StakingManager) Delegate(delegator string, validator string, amount *big.Int) error {
	sm.mtx.Lock()
	defer sm.mtx.Unlock()

	// 검증자가 존재하는지 확인
	stake, ok := sm.stakes[validator]
	if !ok {
		return fmt.Errorf("주소 %s에 대한 검증자를 찾을 수 없습니다", validator)
	}

	// 새 위임 생성
	delegation := &Delegation{
		Delegator: delegator,
		Validator: validator,
		Amount:    new(big.Int).Set(amount),
		Since:     time.Now(),
	}

	// 위임 정보 업데이트
	sm.delegations[validator] = append(sm.delegations[validator], delegation)
	sm.delegatorStakes[delegator] = append(sm.delegatorStakes[delegator], delegation)

	sm.logger.Info("새 위임", "위임자", delegator, "검증자", validator, "금액", amount)

	// 투표력 재계산
	votingPower := sm.CalculateVotingPower(sm.GetTotalStake(validator))

	// 검증자 업데이트 큐에 추가
	update := types.Ed25519ValidatorUpdate(stake.PubKey, votingPower)
	sm.validatorManager.QueueUpdate(update)

	return nil
}

// Undelegate는 위임 철회를 처리합니다.
func (sm *StakingManager) Undelegate(delegator string, validator string, amount *big.Int) error {
	sm.mtx.Lock()
	defer sm.mtx.Unlock()

	// 검증자가 존재하는지 확인
	stake, ok := sm.stakes[validator]
	if !ok {
		return fmt.Errorf("주소 %s에 대한 검증자를 찾을 수 없습니다", validator)
	}

	// 위임자의 위임 내역 찾기
	delegations := sm.delegations[validator]
	var totalDelegated *big.Int = big.NewInt(0)
	for _, d := range delegations {
		if d.Delegator == delegator {
			totalDelegated = new(big.Int).Add(totalDelegated, d.Amount)
		}
	}

	// 철회 금액이 위임된 금액보다 작거나 같은지 확인
	if amount.Cmp(totalDelegated) > 0 {
		return fmt.Errorf("언델리게이트 금액(%s)이 위임된 금액(%s)보다 큽니다", amount.String(), totalDelegated.String())
	}

	// 위임 정보 업데이트
	remainingAmount := new(big.Int).Set(amount)
	var newDelegations []*Delegation

	for _, d := range delegations {
		if d.Delegator == delegator && remainingAmount.Sign() > 0 {
			if remainingAmount.Cmp(d.Amount) >= 0 {
				// 위임을 완전히 제거
				remainingAmount = new(big.Int).Sub(remainingAmount, d.Amount)
			} else {
				// 위임 금액 일부 감소
				d.Amount = new(big.Int).Sub(d.Amount, remainingAmount)
				newDelegations = append(newDelegations, d)
				remainingAmount = big.NewInt(0)
			}
		} else {
			newDelegations = append(newDelegations, d)
		}
	}

	sm.delegations[validator] = newDelegations

	// 위임자의 위임 내역도 업데이트
	var newDelegatorStakes []*Delegation
	for _, d := range sm.delegatorStakes[delegator] {
		if d.Validator != validator {
			newDelegatorStakes = append(newDelegatorStakes, d)
		}
	}
	sm.delegatorStakes[delegator] = newDelegatorStakes

	sm.logger.Info("위임 철회", "위임자", delegator, "검증자", validator, "금액", amount)

	// 투표력 재계산
	votingPower := sm.CalculateVotingPower(sm.GetTotalStake(validator))

	// 검증자 업데이트 큐에 추가
	update := types.Ed25519ValidatorUpdate(stake.PubKey, votingPower)
	sm.validatorManager.QueueUpdate(update)

	return nil
}

// GetTotalStake는 검증자의 총 스테이크(자신+위임 받은 금액)를 반환합니다.
func (sm *StakingManager) GetTotalStake(validatorAddress string) *big.Int {
	stake, ok := sm.stakes[validatorAddress]
	if !ok {
		return big.NewInt(0)
	}

	total := new(big.Int).Set(stake.Amount)

	// 위임 받은 금액 추가
	for _, delegation := range sm.delegations[validatorAddress] {
		total = new(big.Int).Add(total, delegation.Amount)
	}

	return total
}

// GetTopValidators는 스테이크 금액 기준 상위 N개 검증자를 반환합니다.
func (sm *StakingManager) GetTopValidators(n int) []*StakeInfo {
	sm.mtx.RLock()
	defer sm.mtx.RUnlock()

	// 모든 검증자를 슬라이스로 변환
	validators := make([]*StakeInfo, 0, len(sm.stakes))
	for _, v := range sm.stakes {
		validators = append(validators, v)
	}

	// 총 스테이크 금액 기준으로 정렬
	sort.Slice(validators, func(i, j int) bool {
		totalI := sm.GetTotalStake(validators[i].Address)
		totalJ := sm.GetTotalStake(validators[j].Address)
		return totalI.Cmp(totalJ) > 0 // 내림차순
	})

	// 요청된 수 또는 최대 가능한 수 반환
	if n > len(validators) {
		n = len(validators)
	}
	return validators[:n]
}

// DistributeRewards는 블록 생산 보상을 분배합니다.
func (sm *StakingManager) DistributeRewards(blockHeight int64, rewards *big.Int) {
	sm.mtx.Lock()
	defer sm.mtx.Unlock()

	// 활성 검증자가 없으면 보상 패스
	if len(sm.stakes) == 0 {
		return
	}

	// 각 검증자의 투표력 비율로 보상 분배
	totalPower := sm.validatorManager.GetTotalPower()
	if totalPower <= 0 {
		return
	}

	sm.logger.Info("블록 보상 분배", "블록", blockHeight, "총보상", rewards)

	// 각 검증자별 보상 계산
	validators := sm.validatorManager.GetValidators()
	for _, validator := range validators {
		validatorAddress := ""
		// validator.PubKey로 주소 찾기
		for addr, stake := range sm.stakes {
			if string(stake.PubKey) == string(validator.PubKey) {
				validatorAddress = addr
				break
			}
		}

		if validatorAddress == "" || validator.Power <= 0 {
			continue
		}

		// 투표력 비율에 따른 보상 계산
		proportion := new(big.Float).Quo(
			new(big.Float).SetInt64(validator.Power),
			new(big.Float).SetInt64(totalPower),
		)
		
		validatorReward := new(big.Float).Mul(new(big.Float).SetInt(rewards), proportion)
		
		// big.Float -> big.Int 변환
		validatorRewardInt, _ := validatorReward.Int(nil)
		
		// 누적 보상에 추가
		if _, ok := sm.cumulativeRewards[validatorAddress]; !ok {
			sm.cumulativeRewards[validatorAddress] = big.NewInt(0)
		}
		sm.cumulativeRewards[validatorAddress] = new(big.Int).Add(
			sm.cumulativeRewards[validatorAddress],
			validatorRewardInt,
		)
		
		sm.logger.Debug("검증자 보상 분배", 
			"주소", validatorAddress, 
			"투표력", validator.Power,
			"비율", proportion, 
			"보상", validatorRewardInt,
			"누적보상", sm.cumulativeRewards[validatorAddress])
	}
}

// ClaimRewards는 검증자가 자신의 보상을 청구하는 것을 처리합니다.
func (sm *StakingManager) ClaimRewards(validatorAddress string) (*big.Int, error) {
	sm.mtx.Lock()
	defer sm.mtx.Unlock()

	// 검증자 존재 확인
	if _, ok := sm.stakes[validatorAddress]; !ok {
		return nil, fmt.Errorf("주소 %s에 대한 검증자를 찾을 수 없습니다", validatorAddress)
	}

	// 보상 청구
	rewards, ok := sm.cumulativeRewards[validatorAddress]
	if !ok || rewards.Sign() <= 0 {
		return big.NewInt(0), nil
	}

	claimAmount := new(big.Int).Set(rewards)
	sm.cumulativeRewards[validatorAddress] = big.NewInt(0)
	
	sm.logger.Info("보상 청구", "검증자", validatorAddress, "금액", claimAmount)
	
	return claimAmount, nil
}

// GetValidatorInfo는 검증자 정보를 반환합니다.
func (sm *StakingManager) GetValidatorInfo(address string) (*StakeInfo, error) {
	sm.mtx.RLock()
	defer sm.mtx.RUnlock()

	stake, ok := sm.stakes[address]
	if !ok {
		return nil, fmt.Errorf("주소 %s에 대한 검증자를 찾을 수 없습니다", address)
	}

	return stake, nil
}

// UpdateValidatorSet은 활성 검증자 집합을 업데이트합니다.
func (sm *StakingManager) UpdateValidatorSet() {
	sm.mtx.Lock()
	defer sm.mtx.Unlock()

	// 투표력 기준으로 정렬된 검증자 목록 가져오기
	topValidators := sm.GetTopValidators(sm.maxValidators)
	
	// 현재 활성 검증자 목록과 비교하여 변경된 사항만 업데이트 큐에 추가
	for _, validator := range topValidators {
		power := sm.CalculateVotingPower(sm.GetTotalStake(validator.Address))
		currentPower := sm.validatorManager.GetValidatorPower(validator.PubKey)
		
		if power != currentPower {
			update := types.Ed25519ValidatorUpdate(validator.PubKey, power)
			sm.validatorManager.QueueUpdate(update)
			
			sm.logger.Info("검증자 업데이트", 
				"주소", validator.Address, 
				"이전투표력", currentPower, 
				"새투표력", power)
		}
	}
} 