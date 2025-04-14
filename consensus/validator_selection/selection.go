package validator_selection

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
)

// SelectProposer는 주어진 검증자 세트와 블록 높이, 라운드를 기반으로 제안자를 선택합니다.
// 투표력을 기반으로 가중치를 계산하고, 이를 토대로 제안자를 확률적으로 선택합니다.
// 이 알고리즘은 전체 투표력에 비례하여 공정한 확률로 제안자를 선택합니다.
func SelectProposer(validators *types.ValidatorSet, height int64, round int32, logger log.Logger) *types.Validator {
	if validators == nil || validators.Size() == 0 {
		return nil
	}

	// 블록 높이와 라운드를 기반으로 시드 생성
	seed := generateSeed(height, round)

	// 투표력 기반의 누적 가중치 계산
	totalVotingPower := validators.TotalVotingPower()
	weights := make([]*big.Int, validators.Size())
	cumulativeWeights := make([]*big.Int, validators.Size())

	cumulativeWeight := big.NewInt(0)

	for i, val := range validators.Validators {
		// 검증자의 상대적 가중치는 그의 투표력에 비례
		weight := big.NewInt(int64(val.VotingPower))
		weights[i] = weight

		cumulativeWeight = new(big.Int).Add(cumulativeWeight, weight)
		cumulativeWeights[i] = new(big.Int).Set(cumulativeWeight)
	}

	// 선택할 랜덤 포인트 계산 (0 ~ 총 투표력)
	randomPoint := new(big.Int).Mod(new(big.Int).SetBytes(seed), cumulativeWeight)

	// 이진 검색을 통해 선택된 검증자 찾기
	selectedIdx := sort.Search(len(cumulativeWeights), func(i int) bool {
		return cumulativeWeights[i].Cmp(randomPoint) >= 0
	})

	// 로그 기록
	if logger != nil {
		logger.Debug("Selected proposer",
			"height", height,
			"round", round,
			"proposer", validators.Validators[selectedIdx].Address.String(),
			"voting_power", validators.Validators[selectedIdx].VotingPower,
			"total_voting_power", totalVotingPower)
	}

	return validators.Validators[selectedIdx]
}

// GetNextProposerRotation은 라운드 로빈 방식으로 다음 제안자를 선택합니다.
// 단순히 다음 인덱스의 검증자를 선택하는 방식입니다.
func GetNextProposerRotation(validators *types.ValidatorSet, lastProposerAddress []byte) *types.Validator {
	if validators == nil || validators.Size() == 0 {
		return nil
	}

	// 마지막 제안자의 인덱스 찾기
	lastIdx := -1
	for i, val := range validators.Validators {
		if bytes.Equal(val.Address, lastProposerAddress) {
			lastIdx = i
			break
		}
	}

	// 다음 제안자 선택 (순환)
	nextIdx := (lastIdx + 1) % validators.Size()
	return validators.Validators[nextIdx]
}

// generateSeed는 블록 높이와 라운드를 기반으로 결정론적 랜덤 시드를 생성합니다.
func generateSeed(height int64, round int32) []byte {
	heightBytes := make([]byte, 8)
	roundBytes := make([]byte, 4)

	// 높이와 라운드를 바이트로 변환
	for i := 0; i < 8; i++ {
		heightBytes[i] = byte(height >> uint(i*8))
	}

	for i := 0; i < 4; i++ {
		roundBytes[i] = byte(round >> uint(i*8))
	}

	// 높이와 라운드를 결합
	combined := append(heightBytes, roundBytes...)

	// 해시 함수를 통해 시드 생성
	return crypto.Sha256(combined)
}

// WeightedSelection은 Bor 합의 알고리즘을 참고하여 구현한 방식으로,
// 지분 기반 확률적 선택 방식을 구현합니다.
func WeightedSelection(validators *types.ValidatorSet, seed []byte) *types.Validator {
	if validators == nil || validators.Size() == 0 {
		return nil
	}

	// 총 투표력 계산
	totalVotingPower := validators.TotalVotingPower()

	// 시드를 기반으로 랜덤 값 생성
	randomValue := new(big.Int).SetBytes(seed)
	randomValue = new(big.Int).Mod(randomValue, big.NewInt(int64(totalVotingPower)))

	// 누적된 투표력을 기준으로 검증자 선택
	cumulativeVotingPower := int64(0)
	for _, val := range validators.Validators {
		cumulativeVotingPower += int64(val.VotingPower)
		if big.NewInt(cumulativeVotingPower).Cmp(randomValue) > 0 {
			return val
		}
	}

	// 기본값으로 첫 번째 검증자 반환
	return validators.Validators[0]
}

// ValidatorSetDifference는 두 검증자 세트의 차이점을 계산합니다.
// 이 함수는 검증자 세트가 변경되었을 때 추가/제거된 검증자를 식별하는 데 사용됩니다.
func ValidatorSetDifference(oldSet, newSet *types.ValidatorSet) (added, removed []*types.Validator) {
	if oldSet == nil || newSet == nil {
		return nil, nil
	}

	// 추가된 검증자 식별
	for _, newVal := range newSet.Validators {
		found := false
		for _, oldVal := range oldSet.Validators {
			if bytes.Equal(oldVal.Address, newVal.Address) {
				found = true
				break
			}
		}
		if !found {
			added = append(added, newVal)
		}
	}

	// 제거된 검증자 식별
	for _, oldVal := range oldSet.Validators {
		found := false
		for _, newVal := range newSet.Validators {
			if bytes.Equal(oldVal.Address, newVal.Address) {
				found = true
				break
			}
		}
		if !found {
			removed = append(removed, oldVal)
		}
	}

	return added, removed
}
