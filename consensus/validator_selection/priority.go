package validator_selection

import (
	"bytes"
	"sort"

	"github.com/tendermint/tendermint/types"
)

// PriorityBasedProposer는 검증자 우선순위를 기반으로 제안자를 선택합니다.
// 높은 우선순위를 가진 검증자가 제안자로 선택되고, 선택 후 해당 검증자의 우선순위는 감소합니다.
func PriorityBasedProposer(validators *types.ValidatorSet) int {
	if validators == nil || validators.Size() == 0 {
		return -1
	}

	// 우선순위에 따라 정렬된 인덱스 배열 생성
	indices := make([]int, validators.Size())
	for i := 0; i < validators.Size(); i++ {
		indices[i] = i
	}

	// 우선순위 기준으로 내림차순 정렬
	sort.Slice(indices, func(i, j int) bool {
		return validators.Validators[indices[i]].ProposerPriority > validators.Validators[indices[j]].ProposerPriority
	})

	// 가장 높은 우선순위를 가진 검증자 선택
	return indices[0]
}

// UpdateProposerPriorities는 제안자 선택 후 우선순위를 업데이트합니다.
// proposerIndex: 선택된 제안자 인덱스
// validators: 검증자 집합
func UpdateProposerPriorities(proposerIndex int, validators *types.ValidatorSet) {
	if validators == nil || validators.Size() == 0 || proposerIndex < 0 || proposerIndex >= validators.Size() {
		return
	}

	// 총 투표력 계산
	totalVotingPower := int64(0)
	for i := 0; i < validators.Size(); i++ {
		totalVotingPower += validators.Validators[i].VotingPower
	}

	// 모든 검증자의 우선순위 증가 (투표력에 비례)
	for i := 0; i < validators.Size(); i++ {
		validators.Validators[i].ProposerPriority += validators.Validators[i].VotingPower
	}

	// 선택된 제안자의 우선순위 감소 (총 투표력만큼)
	validators.Validators[proposerIndex].ProposerPriority -= totalVotingPower

	// 우선순위 정규화 (큰 수가 되는 것 방지)
	normalizeProposerPriorities(validators)
}

// normalizeProposerPriorities는 우선순위 값이 너무 커지지 않도록 정규화합니다.
func normalizeProposerPriorities(validators *types.ValidatorSet) {
	if validators == nil || validators.Size() == 0 {
		return
	}

	// 평균 우선순위 계산
	sum := int64(0)
	for i := 0; i < validators.Size(); i++ {
		sum += validators.Validators[i].ProposerPriority
	}
	avg := sum / int64(validators.Size())

	// 모든 우선순위에서 평균값 차감
	for i := 0; i < validators.Size(); i++ {
		validators.Validators[i].ProposerPriority -= avg
	}
}

// CalculateNextProposer는 다음 블록의 제안자를 계산합니다.
// height: 현재 블록 높이
// validators: 검증자 집합
// 반환값: 다음 제안자 인덱스
func CalculateNextProposer(height int64, validators *types.ValidatorSet) int {
	if validators == nil || validators.Size() == 0 {
		return -1
	}

	// 현재 제안자 인덱스 찾기
	currentProposerIndex := PriorityBasedProposer(validators)

	// 우선순위 업데이트
	UpdateProposerPriorities(currentProposerIndex, validators)

	// 업데이트된 우선순위로 다음 제안자 선택
	return PriorityBasedProposer(validators)
}

// GetValidatorByAddress는 주소로 검증자를 찾습니다.
// validators: 검증자 집합
// address: 검색할 검증자 주소
// 반환값: 검증자 인덱스, 존재 여부
func GetValidatorByAddress(validators *types.ValidatorSet, address []byte) (int, bool) {
	if validators == nil || validators.Size() == 0 {
		return -1, false
	}

	for i := 0; i < validators.Size(); i++ {
		// bytes.Equal을 사용하여 주소 비교
		if bytes.Equal(validators.Validators[i].Address, address) {
			return i, true
		}
	}

	return -1, false
}

// RotateProposer는 투표력과 우선순위를 모두 고려하여 제안자를 순환합니다.
// validators: 검증자 집합
// height: 현재 블록 높이
// 반환값: 다음 제안자 인덱스
func RotateProposer(validators *types.ValidatorSet, height int64) int {
	if validators == nil || validators.Size() == 0 {
		return -1
	}

	// 제안자 우선순위 재계산
	for i := 0; i < validators.Size(); i++ {
		power := validators.Validators[i].VotingPower
		validators.Validators[i].ProposerPriority += power
	}

	// 블록 높이를 고려한 추가 조정 (일정 주기로 재설정)
	if height%100 == 0 {
		// 주기적으로 우선순위 재설정하여 모든 검증자에게 기회 부여
		for i := 0; i < validators.Size(); i++ {
			validators.Validators[i].ProposerPriority = validators.Validators[i].VotingPower
		}
	}

	// 우선순위 기반 제안자 선택
	proposerIndex := PriorityBasedProposer(validators)

	// 선택된 제안자의 우선순위 감소
	if proposerIndex >= 0 {
		totalPower := int64(0)
		for i := 0; i < validators.Size(); i++ {
			totalPower += validators.Validators[i].VotingPower
		}
		validators.Validators[proposerIndex].ProposerPriority -= totalPower
	}

	return proposerIndex
}
