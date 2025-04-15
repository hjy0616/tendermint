package validator_selection

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/types"
)

// 테스트용 검증자 생성 헬퍼 함수
func createValidatorSet(n int) *types.ValidatorSet {
	validators := make([]*types.Validator, n)
	totalVotingPower := int64(0)

	for i := 0; i < n; i++ {
		pubKey := ed25519.GenPrivKey().PubKey()
		votingPower := int64(rand.Intn(100) + 1) // 1~100 사이 투표력
		validators[i] = types.NewValidator(pubKey, votingPower)
		totalVotingPower += votingPower
	}

	return types.NewValidatorSet(validators)
}

func TestSelectProposer(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	valSet := createValidatorSet(10)

	// 동일한 높이와 라운드에서는 항상 동일한 제안자가 선택되어야 함
	proposer1 := SelectProposer(valSet, 1, 0)
	proposer2 := SelectProposer(valSet, 1, 0)
	require.Equal(t, proposer1, proposer2, "같은 조건에서 다른 제안자 선택됨")

	// 다른 높이에서는 다른 제안자가 선택될 수 있음
	for i := int64(1); i < 10; i++ {
		for j := int32(0); j < 3; j++ {
			proposer := SelectProposer(valSet, i, j)
			require.GreaterOrEqual(t, proposer, 0, "유효하지 않은 제안자 인덱스")
			require.Less(t, proposer, valSet.Size(), "유효하지 않은 제안자 인덱스")
		}
	}

	// 빈 검증자 세트 테스트
	emptyValSet := types.NewValidatorSet([]*types.Validator{})
	proposer := SelectProposer(emptyValSet, 1, 0)
	require.Equal(t, -1, proposer, "빈 검증자 세트에서 -1이 아닌 값 반환")

	// nil 검증자 세트 테스트
	proposer = SelectProposer(nil, 1, 0)
	require.Equal(t, -1, proposer, "nil 검증자 세트에서 -1이 아닌 값 반환")
}

func TestGetNextProposerRotation(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	valSet := createValidatorSet(5)

	// 순환 확인
	curr := 0
	for i := 0; i < valSet.Size()*2; i++ {
		next := GetNextProposerRotation(valSet, curr)
		require.Equal(t, (curr+1)%valSet.Size(), next, "잘못된 순환 로직")
		curr = next
	}

	// 경계 조건 테스트
	next := GetNextProposerRotation(valSet, valSet.Size()-1)
	require.Equal(t, 0, next, "마지막 인덱스에서 0으로 순환되지 않음")

	// 유효하지 않은 입력 테스트
	next = GetNextProposerRotation(valSet, -1)
	require.GreaterOrEqual(t, next, 0, "유효하지 않은 입력에서 유효한 인덱스 반환되지 않음")
	require.Less(t, next, valSet.Size(), "유효하지 않은 입력에서 유효하지 않은 인덱스 반환")

	// 빈 검증자 세트 테스트
	emptyValSet := types.NewValidatorSet([]*types.Validator{})
	next = GetNextProposerRotation(emptyValSet, 0)
	require.Equal(t, -1, next, "빈 검증자 세트에서 -1이 아닌 값 반환")
}

func TestWeightedSelection(t *testing.T) {
	// 고정된 시드로 테스트 결과 일관성 보장
	rand.Seed(42)

	// 투표력이 다른 검증자 세트 생성
	validators := make([]*types.Validator, 3)
	validators[0] = types.NewValidator(ed25519.GenPrivKey().PubKey(), 10) // 낮은 투표력
	validators[1] = types.NewValidator(ed25519.GenPrivKey().PubKey(), 50) // 중간 투표력
	validators[2] = types.NewValidator(ed25519.GenPrivKey().PubKey(), 40) // 높은 투표력
	valSet := types.NewValidatorSet(validators)

	// 여러 블록 높이에서 선택 테스트
	selections := make(map[int]int)
	iterations := 1000

	for i := int64(0); i < int64(iterations); i++ {
		selected := WeightedSelection(valSet, i)
		require.GreaterOrEqual(t, selected, 0, "유효하지 않은 제안자 인덱스")
		require.Less(t, selected, valSet.Size(), "유효하지 않은 제안자 인덱스")
		selections[selected]++
	}

	// 투표력에 비례하여 선택 빈도 확인
	// 정확한 비율은 아니지만 대략적인 경향 확인
	require.Greater(t, selections[1], selections[0], "투표력이 높은 검증자가 더 자주 선택되어야 함")
	require.Greater(t, selections[2], selections[0], "투표력이 높은 검증자가 더 자주 선택되어야 함")

	t.Logf("Selection distribution: %v", selections)

	// 총 합계 확인
	total := 0
	for _, count := range selections {
		total += count
	}
	require.Equal(t, iterations, total, "선택 횟수 합계가 반복 횟수와 일치해야 함")
}

func TestValidatorSetDifference(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// 초기 검증자 세트 생성
	validators1 := make([]*types.Validator, 5)
	for i := 0; i < 5; i++ {
		pubKey := ed25519.GenPrivKey().PubKey()
		validators1[i] = types.NewValidator(pubKey, int64(i+1)*10)
	}
	valSet1 := types.NewValidatorSet(validators1)

	// 두 번째 검증자 세트: 일부 동일, 일부 변경, 일부 제거, 일부 추가
	validators2 := make([]*types.Validator, 6)
	// 처음 2개는 동일
	validators2[0] = validators1[0]
	validators2[1] = validators1[1]
	// 다음 2개는 투표력만 변경
	validators2[2] = types.NewValidator(validators1[2].PubKey, validators1[2].VotingPower+5)
	validators2[3] = types.NewValidator(validators1[3].PubKey, validators1[3].VotingPower-5)
	// 마지막 하나는 제거 (validators1[4])
	// 새로운 검증자 2개 추가
	validators2[4] = types.NewValidator(ed25519.GenPrivKey().PubKey(), 100)
	validators2[5] = types.NewValidator(ed25519.GenPrivKey().PubKey(), 110)

	valSet2 := types.NewValidatorSet(validators2)

	// 차이 계산
	added, removed, updated := ValidatorSetDifference(valSet1, valSet2)

	// 검증
	require.Equal(t, 2, len(added), "새로 추가된 검증자 수가 일치해야 함")
	require.Equal(t, 1, len(removed), "제거된 검증자 수가 일치해야 함")
	require.Equal(t, 2, len(updated), "업데이트된 검증자 수가 일치해야 함")

	// 세부 검증
	require.Equal(t, validators2[4].Address, added[0].Address, "추가된 검증자 주소 불일치")
	require.Equal(t, validators2[5].Address, added[1].Address, "추가된 검증자 주소 불일치")
	require.Equal(t, validators1[4].Address, removed[0].Address, "제거된 검증자 주소 불일치")

	// nil 테스트
	added, removed, updated = ValidatorSetDifference(nil, valSet2)
	require.Nil(t, added, "nil 입력에서 nil이 아닌 결과 반환")
	require.Nil(t, removed, "nil 입력에서 nil이 아닌 결과 반환")
	require.Nil(t, updated, "nil 입력에서 nil이 아닌 결과 반환")
}

func TestCalculateProposerPriority(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// 테스트용 검증자 생성
	validators := make([]*types.Validator, 3)
	validators[0] = types.NewValidator(ed25519.GenPrivKey().PubKey(), 10)
	validators[1] = types.NewValidator(ed25519.GenPrivKey().PubKey(), 20)
	validators[2] = types.NewValidator(ed25519.GenPrivKey().PubKey(), 30)

	// 초기 우선순위 모두 0으로 설정
	for i := range validators {
		validators[i].ProposerPriority = 0
	}

	// 우선순위 계산
	updatedVals := CalculateProposerPriority(validators)

	// 투표력이 높을수록 우선순위가 더 높아야 함
	require.Greater(t, updatedVals[2].ProposerPriority, updatedVals[1].ProposerPriority,
		"투표력이 높은 검증자의 우선순위가 더 높아야 함")
	require.Greater(t, updatedVals[1].ProposerPriority, updatedVals[0].ProposerPriority,
		"투표력이 높은 검증자의 우선순위가 더 높아야 함")

	// 총합이 거의 0에 가까운지 확인 (정규화 확인)
	sum := updatedVals[0].ProposerPriority + updatedVals[1].ProposerPriority + updatedVals[2].ProposerPriority
	require.LessOrEqual(t, sum, int64(3), "우선순위 합이 거의 0에 가까워야 함")
	require.GreaterOrEqual(t, sum, int64(-3), "우선순위 합이 거의 0에 가까워야 함")
}
