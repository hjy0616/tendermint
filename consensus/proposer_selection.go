package consensus

import (
	"bytes"

	"github.com/tendermint/tendermint/consensus/validator_selection"
	"github.com/tendermint/tendermint/types"
)

// ProposerSelectionMode는 제안자 선택 모드를 정의합니다.
type ProposerSelectionMode int

const (
	// DefaultProposerSelection은 기본 Tendermint 제안자 선택 알고리즘을 사용합니다.
	DefaultProposerSelection ProposerSelectionMode = iota

	// PriorityBasedSelection은 우선순위 기반 제안자 선택 알고리즘을 사용합니다.
	PriorityBasedSelection

	// WeightedSelection은 가중치 기반 제안자 선택 알고리즘을 사용합니다.
	WeightedSelection

	// RotationSelection은 순환 방식 제안자 선택 알고리즘을 사용합니다.
	RotationSelection
)

// SelectProposer는 현재 설정된 모드에 따라 제안자를 선택합니다.
// validators: 검증자 집합
// height: 현재 블록 높이
// round: 현재 라운드
// mode: 제안자 선택 모드
// lastProposerIndex: 마지막 제안자 인덱스 (RotationSelection 모드에서 사용)
// 반환값: 선택된 제안자 인덱스
func SelectProposer(
	validators *types.ValidatorSet,
	height int64,
	round int32,
	mode ProposerSelectionMode,
	lastProposerIndex int,
) int {
	if validators == nil || validators.Size() == 0 {
		return -1
	}

	switch mode {
	case PriorityBasedSelection:
		return validator_selection.PriorityBasedProposer(validators)

	case WeightedSelection:
		return validator_selection.WeightedSelection(validators, height)

	case RotationSelection:
		return validator_selection.GetNextProposerRotation(validators, lastProposerIndex)

	default:
		// 기본 Tendermint 제안자 선택 로직 사용
		return validator_selection.SelectProposer(validators, height, round)
	}
}

// UpdateProposerPriorities는 제안자 선택 후 우선순위를 업데이트합니다.
// proposerIndex: 선택된 제안자 인덱스
// validators: 검증자 집합
func UpdateProposerPriorities(proposerIndex int, validators *types.ValidatorSet) {
	validator_selection.UpdateProposerPriorities(proposerIndex, validators)
}

// IsProposer는 주어진 주소가 현재 블록 제안자인지 확인합니다.
// address: 확인할 주소
// validators: 검증자 집합
// proposerIndex: 현재 제안자 인덱스
// 반환값: 제안자 여부
func IsProposer(address []byte, validators *types.ValidatorSet, proposerIndex int) bool {
	if validators == nil || validators.Size() == 0 || proposerIndex < 0 || proposerIndex >= validators.Size() {
		return false
	}

	proposer := validators.Validators[proposerIndex]
	return proposer != nil && len(address) > 0 && len(proposer.Address) > 0 &&
		bytes.Equal(proposer.Address, address)
}

// GetProposerByHeight는 특정 블록 높이의 제안자를 결정합니다.
// validators: 검증자 집합
// height: 블록 높이
// round: 현재 라운드
// mode: 제안자 선택 모드
// lastProposerIndex: 마지막 제안자 인덱스
// 반환값: 제안자 인덱스, 제안자 주소
func GetProposerByHeight(
	validators *types.ValidatorSet,
	height int64,
	round int32,
	mode ProposerSelectionMode,
	lastProposerIndex int,
) (int, []byte) {
	proposerIndex := SelectProposer(validators, height, round, mode, lastProposerIndex)
	if proposerIndex < 0 || proposerIndex >= validators.Size() {
		return -1, nil
	}

	return proposerIndex, validators.Validators[proposerIndex].Address
}

// CalculateNextProposerRotation은 순환 방식으로 다음 제안자를 계산합니다.
// validators: 검증자 집합
// lastProposerIndex: 마지막 제안자 인덱스
// 반환값: 다음 제안자 인덱스
func CalculateNextProposerRotation(validators *types.ValidatorSet, lastProposerIndex int) int {
	return validator_selection.GetNextProposerRotation(validators, lastProposerIndex)
}
