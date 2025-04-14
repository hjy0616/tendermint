package validators

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

// MockEthereumClient는 테스트용 이더리움 클라이언트 모킹 구현체입니다.
type MockEthereumClient struct {
	validators     []*ValidatorInfo
	logger         log.Logger
	errorMode      bool
	delayDuration  time.Duration
	contractStates map[string][]byte
}

// NewMockEthereumClient는 모킹된 이더리움 클라이언트를 생성합니다.
func NewMockEthereumClient() *MockEthereumClient {
	return &MockEthereumClient{
		validators:     make([]*ValidatorInfo, 0),
		logger:         log.NewNopLogger(),
		contractStates: make(map[string][]byte),
	}
}

// SetLogger는 로거를 설정합니다.
func (m *MockEthereumClient) SetLogger(logger log.Logger) {
	m.logger = logger
}

// SetErrorMode는 에러를 발생시키는 모드를 설정합니다.
func (m *MockEthereumClient) SetErrorMode(enabled bool) {
	m.errorMode = enabled
}

// SetDelay는 네트워크 지연을 시뮬레이션하기 위한 지연 시간을 설정합니다.
func (m *MockEthereumClient) SetDelay(delay time.Duration) {
	m.delayDuration = delay
}

// AddValidator는 테스트용 검증자를 추가합니다.
func (m *MockEthereumClient) AddValidator(address string, pubKey []byte, votingPower *big.Int, commission uint32) {
	m.validators = append(m.validators, &ValidatorInfo{
		Address:     address,
		PubKey:      pubKey,
		VotingPower: votingPower,
		Commission:  commission,
		Since:       uint64(time.Now().Unix()),
	})
}

// RemoveValidator는 테스트용 검증자를 제거합니다.
func (m *MockEthereumClient) RemoveValidator(address string) {
	var newValidators []*ValidatorInfo
	for _, v := range m.validators {
		if v.Address != address {
			newValidators = append(newValidators, v)
		}
	}
	m.validators = newValidators
}

// GetValidatorsFromContract는 컨트랙트에서 검증자 정보를 가져오는 것을 시뮬레이션합니다.
func (m *MockEthereumClient) GetValidatorsFromContract(ctx context.Context, contractAddress string) ([]*ValidatorInfo, error) {
	// 네트워크 지연 시뮬레이션
	if m.delayDuration > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(m.delayDuration):
			// 지연 후 계속
		}
	}

	// 에러 모드면 에러 반환
	if m.errorMode {
		return nil, ErrMockEthereumClient
	}

	// 검증자 복사본 반환
	result := make([]*ValidatorInfo, len(m.validators))
	for i, v := range m.validators {
		result[i] = &ValidatorInfo{
			Address:     v.Address,
			PubKey:      v.PubKey,
			VotingPower: new(big.Int).Set(v.VotingPower),
			Commission:  v.Commission,
			Since:       v.Since,
		}
	}

	m.logger.Debug("Mock 이더리움 클라이언트가 검증자 반환", "contractAddress", contractAddress, "count", len(result))
	return result, nil
}

// SubmitValidatorUpdates는 검증자 업데이트를 컨트랙트에 제출하는 것을 시뮬레이션합니다.
func (m *MockEthereumClient) SubmitValidatorUpdates(ctx context.Context, updates []types.ValidatorUpdate, contractAddress string) error {
	// 네트워크 지연 시뮬레이션
	if m.delayDuration > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(m.delayDuration):
			// 지연 후 계속
		}
	}

	// 에러 모드면 에러 반환
	if m.errorMode {
		return ErrMockEthereumClient
	}

	m.logger.Debug("Mock 이더리움 클라이언트가 검증자 업데이트 처리", "contractAddress", contractAddress, "updates", len(updates))
	
	// 검증자 업데이트 처리 로직
	for _, update := range updates {
		pubKey := update.PubKey.GetEd25519()
		power := update.Power
		
		// 기존 검증자 찾기
		found := false
		for i, v := range m.validators {
			if string(v.PubKey) == string(pubKey) {
				if power == 0 {
					// 검증자 제거
					m.validators = append(m.validators[:i], m.validators[i+1:]...)
				} else {
					// 검증자 업데이트
					v.VotingPower = big.NewInt(power * 10000) // 투표력을 스테이크 금액으로 변환
				}
				found = true
				break
			}
		}
		
		// 새 검증자 추가
		if !found && power > 0 {
			m.validators = append(m.validators, &ValidatorInfo{
				Address:     "0x" + string(pubKey[:10]), // 주소는 pubKey에서 유도 (단순 데모용)
				PubKey:      pubKey,
				VotingPower: big.NewInt(power * 10000), // 투표력을 스테이크 금액으로 변환
				Commission:  100, // 기본 커미션 1%
				Since:       uint64(time.Now().Unix()),
			})
		}
	}
	
	return nil
}

// StoreContractState는 컨트랙트 상태를 저장하는 것을 시뮬레이션합니다.
func (m *MockEthereumClient) StoreContractState(contractAddress string, key string, value []byte) {
	stateKey := contractAddress + ":" + key
	m.contractStates[stateKey] = value
}

// GetContractState는 컨트랙트 상태를 가져오는 것을 시뮬레이션합니다.
func (m *MockEthereumClient) GetContractState(contractAddress string, key string) []byte {
	stateKey := contractAddress + ":" + key
	return m.contractStates[stateKey]
}

// ErrMockEthereumClient는 모킹된 에러입니다.
var ErrMockEthereumClient = errors.New("mock ethereum client 오류") 