package validator_selection

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"math/rand"
	"sort"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
)

// ValidatorSetDifference returns the difference between two validator sets
// first return value is validators to add, second return value is validators to remove
func ValidatorSetDifference(oldSet, newSet *types.ValidatorSet) ([]*types.Validator, []*types.Validator) {
	oldValMap := make(map[string]*types.Validator)
	addValList := make([]*types.Validator, 0)
	removeValList := make([]*types.Validator, 0)

	// 기존 검증자 세트를 맵으로 변환
	for _, val := range oldSet.Validators {
		oldValMap[string(val.Address)] = val
	}

	// 새 검증자 세트와 비교하여 추가된 검증자 찾기
	for _, val := range newSet.Validators {
		if _, exists := oldValMap[string(val.Address)]; !exists {
			addValList = append(addValList, val)
		} else {
			// 검증자가 존재하지만 투표력이 변경된 경우 제거 후 추가 목록에 포함
			oldVal := oldValMap[string(val.Address)]
			if oldVal.VotingPower != val.VotingPower {
				removeValList = append(removeValList, oldVal)
				addValList = append(addValList, val)
			}
			delete(oldValMap, string(val.Address))
		}
	}

	// 남은 검증자(삭제된 검증자)를 제거 목록에 추가
	for _, val := range oldValMap {
		removeValList = append(removeValList, val)
	}

	return addValList, removeValList
}

// WeightedSelection 투표력에 기반한 제안자 선택 알고리즘
// 투표력이 높을수록 제안자로 선택될 확률이 높아짐
func WeightedSelection(validators []*types.Validator, seed []byte) *types.Validator {
	if len(validators) == 0 {
		return nil
	}

	// 총 투표력 계산
	totalVotingPower := int64(0)
	for _, validator := range validators {
		totalVotingPower += validator.VotingPower
	}

	if totalVotingPower <= 0 {
		// 모든 투표력이 0이면 첫 번째 검증자를 반환
		return validators[0]
	}

	// 시드로부터 랜덤 값 생성
	hash := sha256.Sum256(seed)
	rnd := new(big.Int).SetBytes(hash[:])
	rnd = new(big.Int).Mod(rnd, big.NewInt(int64(totalVotingPower)))
	target := rnd.Int64()

	// 누적 투표력으로 제안자 선택
	cumVotingPower := int64(0)
	for _, validator := range validators {
		cumVotingPower += validator.VotingPower
		if cumVotingPower > target {
			return validator
		}
	}

	// 혹시라도 선택되지 않았을 경우 마지막 검증자 반환
	return validators[len(validators)-1]
}

// GetNextProposerRotation 순환 방식으로 다음 제안자를 선택
// 라운드 로빈 방식으로 검증자들 사이에서 순환
func GetNextProposerRotation(validators []*types.Validator, lastProposerIdx int) *types.Validator {
	if len(validators) == 0 {
		return nil
	}

	// 다음 인덱스 계산
	nextIdx := (lastProposerIdx + 1) % len(validators)
	return validators[nextIdx]
}

// SelectProposer 현재 높이와 라운드에 따라 제안자를 선택
// 현재 블록 높이와 라운드를 시드로 사용하여 가중치 기반 선택 알고리즘 적용
func SelectProposer(validators []*types.Validator, height int64, round int32, logger log.Logger) *types.Validator {
	if len(validators) == 0 {
		logger.Error("Cannot select proposer from empty validator set")
		return nil
	}

	// 투표력으로 정렬
	sortedVals := make([]*types.Validator, len(validators))
	copy(sortedVals, validators)
	sort.Slice(sortedVals, func(i, j int) bool {
		if sortedVals[i].VotingPower == sortedVals[j].VotingPower {
			// 투표력이 같으면 주소로 정렬
			return bytes.Compare(sortedVals[i].Address, sortedVals[j].Address) < 0
		}
		return sortedVals[i].VotingPower > sortedVals[j].VotingPower
	})

	// 높이와 라운드를 시드로 하여 제안자 선택
	seed := make([]byte, 16)
	binary.BigEndian.PutUint64(seed[:8], uint64(height))
	binary.BigEndian.PutUint32(seed[8:12], uint32(round))

	// 높은 투표력을 가진 상위 검증자들에게 더 높은 확률 부여
	maxPower := sortedVals[0].VotingPower
	if maxPower <= 0 {
		// 모든 검증자의 투표력이 0이면 랜덤하게 선택
		r := rand.New(rand.NewSource(int64(binary.BigEndian.Uint64(seed))))
		return sortedVals[r.Intn(len(sortedVals))]
	}

	return WeightedSelection(sortedVals, seed)
}

// ValidatorSetFromProto Tendermint proto 검증자 세트를 일반 검증자 세트로 변환
func ValidatorSetFromProto(protoValSet *types.ValidatorSet) *types.ValidatorSet {
	if protoValSet == nil {
		return nil
	}

	return protoValSet
}

// IsValidator 주어진 주소가 현재 검증자 세트에 포함되어 있는지 확인
func IsValidator(valSet *types.ValidatorSet, address crypto.Address) bool {
	if valSet == nil {
		return false
	}

	for _, val := range valSet.Validators {
		if bytes.Equal(val.Address, address) {
			return true
		}
	}

	return false
}

// IsProposer 주어진 주소가 현재 제안자인지 확인
func IsProposer(valSet *types.ValidatorSet, address crypto.Address) bool {
	if valSet == nil || valSet.Proposer == nil {
		return false
	}

	return bytes.Equal(valSet.Proposer.Address, address)
}
