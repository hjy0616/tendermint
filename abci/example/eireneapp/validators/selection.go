package validators

import (
	"math/big"
	"math/rand"
	"sort"
	"time"

	"github.com/tendermint/tendermint/libs/log"
)

// SelectionSystem은 PoS 기반 검증자 선택 시스템을 관리합니다.
type SelectionSystem struct {
	stakingManager *StakingManager  // 스테이킹 관리자 참조
	logger         log.Logger       // 로거
	randSource     *rand.Rand       // 랜덤 소스 (결정론적으로 사용 가능)
}

// NewSelectionSystem은 새로운 검증자 선택 시스템을 생성합니다.
func NewSelectionSystem(stakingManager *StakingManager) *SelectionSystem {
	source := rand.NewSource(time.Now().UnixNano())
	randSource := rand.New(source)
	
	return &SelectionSystem{
		stakingManager: stakingManager,
		logger:         log.NewNopLogger(),
		randSource:     randSource,
	}
}

// SetLogger는 로거를 설정합니다.
func (ss *SelectionSystem) SetLogger(logger log.Logger) {
	ss.logger = logger
}

// SetRandomSeed는 결정론적 랜덤을 위한 시드를 설정합니다.
func (ss *SelectionSystem) SetRandomSeed(seed int64) {
	source := rand.NewSource(seed)
	ss.randSource = rand.New(source)
}

// ProposerWeightedSelection은 가중치 기반 블록 제안자를 선택합니다.
// 투표력에 비례하여 제안자가 선택될 확률이 높아집니다.
func (ss *SelectionSystem) ProposerWeightedSelection() (*StakeInfo, error) {
	validators := ss.stakingManager.GetTopValidators(ss.stakingManager.maxValidators)
	if len(validators) == 0 {
		return nil, nil // 검증자가 없음
	}
	
	// 총 투표력 계산
	totalPower := int64(0)
	for _, validator := range validators {
		stake := ss.stakingManager.GetTotalStake(validator.Address)
		power := ss.stakingManager.CalculateVotingPower(stake)
		totalPower += power
	}
	
	if totalPower <= 0 {
		// 총 투표력이 0이면 무작위 선택
		return validators[ss.randSource.Intn(len(validators))], nil
	}
	
	// 투표력에 비례하여 선택 확률 설정
	target := ss.randSource.Int63n(totalPower)
	cumulativePower := int64(0)
	
	for _, validator := range validators {
		stake := ss.stakingManager.GetTotalStake(validator.Address)
		power := ss.stakingManager.CalculateVotingPower(stake)
		cumulativePower += power
		
		if cumulativePower > target {
			return validator, nil
		}
	}
	
	// 마지막 검증자 반환 (위의 루프에서 선택되지 않은 경우)
	return validators[len(validators)-1], nil
}

// BlockProducerSelection은 다음 블록 생성자를 결정합니다.
// 블록 높이와 현재 시간을 고려하여 결정론적 선택을 수행합니다.
func (ss *SelectionSystem) BlockProducerSelection(height int64, timestamp time.Time) (*StakeInfo, error) {
	// 결정론적 시드 생성 (블록 높이와 타임스탬프 기반)
	seed := height + timestamp.Unix()
	ss.SetRandomSeed(seed)
	
	return ss.ProposerWeightedSelection()
}

// CommitteeSelection은 특정 높이에서의 위원회(검증자 집합)를 선택합니다.
// 위원회 크기는 기본적으로 총 검증자 수의 2/3 이상입니다.
func (ss *SelectionSystem) CommitteeSelection(height int64, minCommitteeSize int) ([]*StakeInfo, error) {
	validators := ss.stakingManager.GetTopValidators(ss.stakingManager.maxValidators)
	if len(validators) == 0 {
		return nil, nil // 검증자가 없음
	}
	
	// 위원회 크기 결정 (최소 크기 이상, 최대는 모든 검증자)
	committeeSize := len(validators) * 2 / 3
	if committeeSize < minCommitteeSize {
		committeeSize = minCommitteeSize
	}
	if committeeSize > len(validators) {
		committeeSize = len(validators)
	}
	
	// 결정론적 선택을 위한 시드 설정
	ss.SetRandomSeed(height)
	
	// Fisher-Yates 셔플 알고리즘으로 검증자 목록 섞기
	for i := len(validators) - 1; i > 0; i-- {
		j := ss.randSource.Intn(i + 1)
		validators[i], validators[j] = validators[j], validators[i]
	}
	
	// 섞인 목록에서 위원회 크기만큼 선택
	committee := validators[:committeeSize]
	
	// 투표력 기준으로 다시 정렬 (선택 후 정렬)
	sort.Slice(committee, func(i, j int) bool {
		powerI := ss.stakingManager.CalculateVotingPower(ss.stakingManager.GetTotalStake(committee[i].Address))
		powerJ := ss.stakingManager.CalculateVotingPower(ss.stakingManager.GetTotalStake(committee[j].Address))
		return powerI > powerJ
	})
	
	return committee, nil
}

// GetVotingPowerDistribution은 현재 검증자 집합의 투표력 분포를 반환합니다.
// 투표력의 집중도(허핀달-허쉬만 지수 등)를 계산하는 데 사용될 수 있습니다.
func (ss *SelectionSystem) GetVotingPowerDistribution() map[string]float64 {
	validators := ss.stakingManager.GetTopValidators(ss.stakingManager.maxValidators)
	if len(validators) == 0 {
		return nil
	}
	
	// 총 투표력 계산
	totalPower := int64(0)
	validatorPowers := make(map[string]int64)
	
	for _, validator := range validators {
		stake := ss.stakingManager.GetTotalStake(validator.Address)
		power := ss.stakingManager.CalculateVotingPower(stake)
		validatorPowers[validator.Address] = power
		totalPower += power
	}
	
	// 비율 계산
	distribution := make(map[string]float64)
	for addr, power := range validatorPowers {
		if totalPower > 0 {
			distribution[addr] = float64(power) / float64(totalPower)
		} else {
			distribution[addr] = 0
		}
	}
	
	return distribution
}

// CalculateHerfindahlHirschmanIndex는 투표력 집중도를 계산합니다.
// 낮은 값은 분산된 네트워크를, 높은 값은 집중된 네트워크를 나타냅니다.
func (ss *SelectionSystem) CalculateHerfindahlHirschmanIndex() float64 {
	distribution := ss.GetVotingPowerDistribution()
	if len(distribution) == 0 {
		return 0
	}
	
	var hhi float64
	for _, ratio := range distribution {
		// 비율의 제곱을 합산
		hhi += ratio * ratio
	}
	
	return hhi
}

// IsNetworkCentralized는 네트워크 분산화 상태를 평가합니다.
// HHI > 0.25는 일반적으로 집중된 네트워크로 간주됩니다.
func (ss *SelectionSystem) IsNetworkCentralized() bool {
	hhi := ss.CalculateHerfindahlHirschmanIndex()
	return hhi > 0.25
}

// GetActiveValidatorsCount는 활성 검증자 수를 반환합니다.
func (ss *SelectionSystem) GetActiveValidatorsCount() int {
	validators := ss.stakingManager.GetTopValidators(ss.stakingManager.maxValidators)
	return len(validators)
}

// CalculateNetworkSecurity는 네트워크 보안 지수를 계산합니다.
// 검증자 수, 분산도, 총 스테이킹 금액 등을 고려합니다.
func (ss *SelectionSystem) CalculateNetworkSecurity() float64 {
	validators := ss.stakingManager.GetTopValidators(ss.stakingManager.maxValidators)
	if len(validators) == 0 {
		return 0
	}
	
	// 검증자 수 점수 (최대 검증자 수 대비 현재 검증자 수 비율)
	validatorCountRatio := float64(len(validators)) / float64(ss.stakingManager.maxValidators)
	if validatorCountRatio > 1 {
		validatorCountRatio = 1
	}
	
	// 탈중앙화 점수 (1 - HHI)
	decentralizationScore := 1 - ss.CalculateHerfindahlHirschmanIndex()
	
	// 총 스테이킹 금액
	totalStake := new(big.Int)
	for _, validator := range validators {
		totalStake.Add(totalStake, ss.stakingManager.GetTotalStake(validator.Address))
	}
	
	// 최소 기대 스테이킹 금액 (각 검증자가 최소 금액만 스테이킹한 경우)
	minExpectedStake := new(big.Int).Mul(
		ss.stakingManager.minStakeAmount,
		big.NewInt(int64(ss.stakingManager.maxValidators)),
	)
	
	// 스테이킹 금액 점수
	var stakeRatio float64 = 0.5 // 기본값
	if minExpectedStake.Sign() > 0 {
		stakeRatioFloat := new(big.Float).Quo(
			new(big.Float).SetInt(totalStake),
			new(big.Float).SetInt(minExpectedStake),
		)
		stakeRatioFloat64, _ := stakeRatioFloat.Float64()
		if stakeRatioFloat64 > 1 {
			stakeRatio = 1
		} else {
			stakeRatio = stakeRatioFloat64
		}
	}
	
	// 보안 지수 계산 (각 요소에 가중치 적용)
	securityIndex := (0.4 * validatorCountRatio) + (0.4 * decentralizationScore) + (0.2 * stakeRatio)
	
	return securityIndex
} 