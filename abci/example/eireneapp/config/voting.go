package config

import (
	"math/big"

	"github.com/zenanetwork/go-zenanet/common"
)

// VotingPowerCalculator는 검증자의 투표력을 계산하는 인터페이스입니다.
type VotingPowerCalculator interface {
	// CalculateVotingPower는 주어진 검증자의 투표력을 계산합니다.
	CalculateVotingPower(validatorAddress common.Address, stake *big.Int) int64
}

// StakeBasedVotingPower는 스테이크 양에 따른 투표력 계산기입니다.
type StakeBasedVotingPower struct {
	// 최소 스테이크 양 (이 이하는 검증자로 참여 불가)
	MinStake *big.Int
	
	// 스테이크 1단위당 투표력 (스테이크 양 / ConversionRate)
	ConversionRate *big.Int
	
	// 최대 투표력 (한 검증자가 가질 수 있는 최대 투표력)
	MaxVotingPower int64
}

// NewStakeBasedVotingPower는 새로운 스테이크 기반 투표력 계산기를 생성합니다.
func NewStakeBasedVotingPower() *StakeBasedVotingPower {
	return &StakeBasedVotingPower{
		MinStake:       big.NewInt(1e18),            // 1 토큰 (소수점 18자리)
		ConversionRate: big.NewInt(1e16),            // 100 토큰당 1 투표력
		MaxVotingPower: 100,                         // 최대 100 투표력
	}
}

// CalculateVotingPower는 스테이크 양에 따라 투표력을 계산합니다.
func (calc *StakeBasedVotingPower) CalculateVotingPower(
	validatorAddress common.Address, 
	stake *big.Int,
) int64 {
	// 최소 스테이크 검사
	if stake.Cmp(calc.MinStake) < 0 {
		return 0
	}
	
	// 투표력 계산: stake / ConversionRate
	votingPower := new(big.Int).Div(stake, calc.ConversionRate)
	
	// int64 범위 체크
	if !votingPower.IsInt64() {
		return calc.MaxVotingPower
	}
	
	// 최대 투표력 제한
	power := votingPower.Int64()
	if power > calc.MaxVotingPower {
		return calc.MaxVotingPower
	}
	
	return power
}

// VotingWeightAdjuster는 투표 과정에서 투표 가중치를 조정하는 인터페이스입니다.
type VotingWeightAdjuster interface {
	// AdjustVotingWeight는 투표 과정과 라운드에 따라 투표력을 조정합니다.
	AdjustVotingWeight(validatorAddress common.Address, basePower int64, voteType string, round int32) int64
}

// StandardVotingAdjuster는 기본 투표 가중치 조정기입니다.
type StandardVotingAdjuster struct {
	// 라운드에 따른 감소율 (%)
	RoundPenalty int
	
	// 장기 검증자 보너스 (%)
	LongevityBonus int
	
	// 검증자 성과 (%)
	PerformanceScore map[common.Address]int
}

// NewStandardVotingAdjuster는 새로운 표준 투표 가중치 조정기를 생성합니다.
func NewStandardVotingAdjuster() *StandardVotingAdjuster {
	return &StandardVotingAdjuster{
		RoundPenalty:     5,                       // 라운드당 5% 감소
		LongevityBonus:   10,                      // 장기 검증자 10% 보너스
		PerformanceScore: make(map[common.Address]int), // 검증자 성과 맵
	}
}

// AdjustVotingWeight는 투표 과정과 라운드에 따라 투표력을 조정합니다.
func (adj *StandardVotingAdjuster) AdjustVotingWeight(
	validatorAddress common.Address, 
	basePower int64, 
	voteType string,
	round int32,
) int64 {
	// 기본 투표력
	power := basePower
	
	// 라운드 패널티: 높은 라운드에서는 투표력 감소
	// 합의를 빠르게 도달하도록 유도
	roundPenalty := int64(adj.RoundPenalty * int(round))
	
	// 검증자 성과 보너스
	performanceBonus := int64(0)
	if score, ok := adj.PerformanceScore[validatorAddress]; ok {
		performanceBonus = int64(score)
	}
	
	// 최종 조정된 투표력 계산
	// 최소 투표력은 1
	adjustedPower := power - roundPenalty + performanceBonus
	if adjustedPower < 1 {
		adjustedPower = 1
	}
	
	return adjustedPower
}

// VotingPowerManager는 검증자의 투표력을 관리하는 메인 컴포넌트입니다.
type VotingPowerManager struct {
	Calculator VotingPowerCalculator
	Adjuster   VotingWeightAdjuster
}

// NewVotingPowerManager는 새로운 투표력 관리자를 생성합니다.
func NewVotingPowerManager() *VotingPowerManager {
	return &VotingPowerManager{
		Calculator: NewStakeBasedVotingPower(),
		Adjuster:   NewStandardVotingAdjuster(),
	}
}

// GetVotingPower는 검증자의 기본 투표력을 계산합니다.
func (vpm *VotingPowerManager) GetVotingPower(
	validatorAddress common.Address, 
	stake *big.Int,
) int64 {
	return vpm.Calculator.CalculateVotingPower(validatorAddress, stake)
}

// GetAdjustedVotingPower는 투표 과정과 라운드에 따라 조정된 투표력을 계산합니다.
func (vpm *VotingPowerManager) GetAdjustedVotingPower(
	validatorAddress common.Address, 
	stake *big.Int, 
	voteType string, 
	round int32,
) int64 {
	// 기본 투표력 계산
	basePower := vpm.GetVotingPower(validatorAddress, stake)
	
	// 투표 조정 적용
	adjustedPower := vpm.Adjuster.AdjustVotingWeight(validatorAddress, basePower, voteType, round)
	
	return adjustedPower
} 