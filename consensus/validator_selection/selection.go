package validator_selection

import (
	"sort"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
)

// SelectionSystem은 검증자 선택 시스템을 관리합니다.
type SelectionSystem struct {
	logger log.Logger
}

// NewSelectionSystem은 새로운 검증자 선택 시스템을 생성합니다.
func NewSelectionSystem() *SelectionSystem {
	return &SelectionSystem{
		logger: log.NewNopLogger(),
	}
}

// SetLogger는 로거를 설정합니다.
func (ss *SelectionSystem) SetLogger(logger log.Logger) {
	ss.logger = logger
}

// SelectProposer는 투표력 기반의 제안자 선택 알고리즘을 구현합니다.
// validators: 검증자 집합
// height: 현재 블록 높이
// round: 현재 라운드
// 반환값: 선택된 제안자 인덱스
func SelectProposer(validators *types.ValidatorSet, height int64, round int32) int {
	if validators == nil || validators.Size() == 0 {
		return -1
	}

	// 블록 높이와 라운드를 기준으로 결정론적 선택
	seed := height + int64(round)
	proposerIndex := seed % int64(validators.Size())

	return int(proposerIndex)
}

// GetNextProposerRotation은 순환 방식으로 다음 제안자를 선택합니다.
// validators: 검증자 집합
// lastProposerIndex: 마지막 제안자 인덱스
// 반환값: 다음 제안자 인덱스
func GetNextProposerRotation(validators *types.ValidatorSet, lastProposerIndex int) int {
	if validators == nil || validators.Size() == 0 {
		return -1
	}

	nextIndex := (lastProposerIndex + 1) % validators.Size()
	return nextIndex
}

// WeightedSelection은 Bor 합의 알고리즘을 참고하여 가중치 기반 제안자 선택을 구현합니다.
// validators: 검증자 집합
// height: 현재 블록 높이
// 반환값: 선택된 제안자 인덱스
func WeightedSelection(validators *types.ValidatorSet, height int64) int {
	if validators == nil || validators.Size() == 0 {
		return -1
	}

	// 총 투표력 계산
	totalVotingPower := int64(0)
	for i := 0; i < validators.Size(); i++ {
		val := validators.Validators[i]
		totalVotingPower += val.VotingPower
	}

	if totalVotingPower <= 0 {
		return 0
	}

	// 블록 높이를 시드로 사용하여 가중치 기반 선택
	// 실제 구현에서는 더 복잡한 랜덤 생성 알고리즘 사용 가능
	seed := height % totalVotingPower

	// 가중치 기반 선택
	cumulativePower := int64(0)
	for i := 0; i < validators.Size(); i++ {
		val := validators.Validators[i]
		cumulativePower += val.VotingPower

		if cumulativePower > seed {
			return i
		}
	}

	// 기본값으로 첫 번째 검증자 반환
	return 0
}

// ValidatorSetDifference는 두 검증자 세트의 차이를 계산합니다.
// oldSet: 이전 검증자 집합
// newSet: 새 검증자 집합
// 반환값: 추가된 검증자, 제거된 검증자, 업데이트된 검증자
func ValidatorSetDifference(
	oldSet *types.ValidatorSet,
	newSet *types.ValidatorSet,
) (added, removed, updated []*types.Validator) {
	if oldSet == nil || newSet == nil {
		return nil, nil, nil
	}

	// 기존 검증자 맵 생성
	oldValMap := make(map[string]*types.Validator)
	for _, val := range oldSet.Validators {
		oldValMap[string(val.Address)] = val
	}

	// 신규 검증자 맵 생성
	newValMap := make(map[string]*types.Validator)
	for _, val := range newSet.Validators {
		newValMap[string(val.Address)] = val

		// 기존에 없던 검증자는 추가 목록에 포함
		if _, exists := oldValMap[string(val.Address)]; !exists {
			added = append(added, val)
		} else if oldValMap[string(val.Address)].VotingPower != val.VotingPower {
			// 투표력이 변경된 검증자는 업데이트 목록에 포함
			updated = append(updated, val)
		}
	}

	// 삭제된 검증자 확인
	for _, val := range oldSet.Validators {
		if _, exists := newValMap[string(val.Address)]; !exists {
			removed = append(removed, val)
		}
	}

	return added, removed, updated
}

// CalculateProposerPriority는 검증자의 우선순위를 재계산합니다.
// validators: 검증자 배열
// 반환값: 우선순위가 업데이트된 검증자 배열
func CalculateProposerPriority(validators []*types.Validator) []*types.Validator {
	if len(validators) == 0 {
		return validators
	}

	// 총 투표력 계산
	totalVotingPower := int64(0)
	for _, val := range validators {
		totalVotingPower += val.VotingPower
	}

	// 검증자별 우선순위 계산
	for i, val := range validators {
		// 우선순위 조정: 투표력에 비례하여 우선순위 증가
		validators[i].ProposerPriority += val.VotingPower

		// 균형을 위해 총 투표력으로 나눈 평균값을 차감
		validators[i].ProposerPriority -= totalVotingPower / int64(len(validators))
	}

	return validators
}

// GetTopByVotingPower는 투표력 기준으로 상위 N개 검증자를 반환합니다.
// validators: 검증자 배열
// n: 반환할 검증자 수
// 반환값: 투표력 기준 상위 n개 검증자
func GetTopByVotingPower(validators []*types.Validator, n int) []*types.Validator {
	if len(validators) <= n {
		return validators
	}

	// 복사본 생성
	sortedVals := make([]*types.Validator, len(validators))
	copy(sortedVals, validators)

	// 투표력 기준 내림차순 정렬
	sort.Slice(sortedVals, func(i, j int) bool {
		return sortedVals[i].VotingPower > sortedVals[j].VotingPower
	})

	return sortedVals[:n]
}
