package validators

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

// ValidatorSyncConfig는 검증자 동기화 설정을 정의합니다.
type ValidatorSyncConfig struct {
	SyncInterval         time.Duration // 동기화 간격
	MaxRetries           int           // 최대 재시도 횟수
	RetryDelay           time.Duration // 재시도 지연 시간
	ValidatorSetContract string        // 검증자 세트 계약 주소
}

// DefaultValidatorSyncConfig는 기본 동기화 설정을 반환합니다.
func DefaultValidatorSyncConfig() *ValidatorSyncConfig {
	return &ValidatorSyncConfig{
		SyncInterval:   time.Minute * 2,
		MaxRetries:     5,
		RetryDelay:     time.Second * 10,
		ValidatorSetContract: "", // 개발자는 실제 계약 주소로 설정해야 함
	}
}

// ValidatorSync는 검증자 세트 동기화를 관리합니다.
type ValidatorSync struct {
	config              *ValidatorSyncConfig
	manager             *Manager
	stakingManager      *StakingManager
	logger              log.Logger
	blockHeight         int64
	syncMutex           sync.Mutex
	ctx                 context.Context
	cancel              context.CancelFunc
	isRunning           bool
	lastSyncTime        time.Time
	lastSuccessfulSync  time.Time
	syncErrors          []error
	ethereumClient      EthereumClient // 인터페이스 참조
}

// EthereumClient는 이더리움과의 통신을 위한 인터페이스를 정의합니다.
type EthereumClient interface {
	GetValidatorsFromContract(ctx context.Context, contractAddress string) ([]*ValidatorInfo, error)
	SubmitValidatorUpdates(ctx context.Context, updates []types.ValidatorUpdate, contractAddress string) error
}

// ValidatorInfo는 이더리움 계약에서 가져온 검증자 정보를 나타냅니다.
type ValidatorInfo struct {
	Address     string   `json:"address"`
	PubKey      []byte   `json:"pubKey"`
	VotingPower *big.Int `json:"votingPower"`
	Commission  uint32   `json:"commission"`
	Since       uint64   `json:"since"`
}

// NewValidatorSync는 새 검증자 동기화 인스턴스를 생성합니다.
func NewValidatorSync(config *ValidatorSyncConfig, manager *Manager, stakingManager *StakingManager, client EthereumClient) *ValidatorSync {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ValidatorSync{
		config:         config,
		manager:        manager,
		stakingManager: stakingManager,
		ctx:            ctx,
		cancel:         cancel,
		logger:         log.NewNopLogger(),
		ethereumClient: client,
	}
}

// SetLogger는 로거를 설정합니다.
func (vs *ValidatorSync) SetLogger(logger log.Logger) {
	vs.logger = logger
}

// Start는 검증자 동기화 프로세스를 시작합니다.
func (vs *ValidatorSync) Start() error {
	vs.syncMutex.Lock()
	defer vs.syncMutex.Unlock()
	
	if vs.isRunning {
		return errors.New("검증자 동기화가 이미 실행 중입니다")
	}
	
	vs.isRunning = true
	
	go vs.syncLoop()
	
	vs.logger.Info("검증자 동기화 시작됨")
	return nil
}

// Stop은 검증자 동기화 프로세스를 중지합니다.
func (vs *ValidatorSync) Stop() error {
	vs.syncMutex.Lock()
	defer vs.syncMutex.Unlock()
	
	if !vs.isRunning {
		return errors.New("검증자 동기화가 실행 중이 아닙니다")
	}
	
	vs.cancel()
	vs.isRunning = false
	
	vs.logger.Info("검증자 동기화 중지됨")
	return nil
}

// UpdateBlockHeight는 현재 블록 높이를 업데이트합니다.
func (vs *ValidatorSync) UpdateBlockHeight(height int64) {
	vs.syncMutex.Lock()
	defer vs.syncMutex.Unlock()
	
	vs.blockHeight = height
}

// syncLoop는 주기적으로 검증자 세트를 동기화하는 고루틴입니다.
func (vs *ValidatorSync) syncLoop() {
	ticker := time.NewTicker(vs.config.SyncInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-vs.ctx.Done():
			return
		case <-ticker.C:
			if err := vs.SyncValidators(); err != nil {
				vs.logger.Error("검증자 동기화 실패", "error", err)
			}
		}
	}
}

// SyncValidators는 검증자 세트를 동기화합니다.
func (vs *ValidatorSync) SyncValidators() error {
	vs.syncMutex.Lock()
	defer vs.syncMutex.Unlock()
	
	vs.logger.Info("검증자 동기화 시작")
	vs.lastSyncTime = time.Now()
	
	// 이더리움 계약에서 검증자 가져오기
	var validators []*ValidatorInfo
	var err error
	
	for retry := 0; retry < vs.config.MaxRetries; retry++ {
		validators, err = vs.ethereumClient.GetValidatorsFromContract(vs.ctx, vs.config.ValidatorSetContract)
		if err == nil {
			break
		}
		
		vs.logger.Error("이더리움 계약에서 검증자 가져오기 실패", "retry", retry+1, "error", err)
		
		// 마지막 시도가 아니면 지연 후 재시도
		if retry < vs.config.MaxRetries-1 {
			select {
			case <-vs.ctx.Done():
				return vs.ctx.Err()
			case <-time.After(vs.config.RetryDelay):
				// 재시도
			}
		}
	}
	
	if err != nil {
		vs.syncErrors = append(vs.syncErrors, err)
		if len(vs.syncErrors) > 10 {
			vs.syncErrors = vs.syncErrors[1:]
		}
		return err
	}
	
	// 검증자 정보 로깅
	vs.logger.Info("이더리움 검증자 데이터 수신됨", "count", len(validators))
	
	// 현재 검증자와 비교 및 업데이트
	for _, validator := range validators {
		// 스테이킹된 금액으로 이더리움과 동기화
		err = vs.stakingManager.Stake(validator.PubKey, validator.Address, validator.VotingPower, validator.Commission)
		if err != nil {
			vs.logger.Error("검증자 스테이킹 업데이트 실패", "address", validator.Address, "error", err)
			continue
		}
	}
	
	// 현재 텐더민트 검증자 목록과 이더리움 검증자 목록 간의 차이 확인
	// 제거된 검증자 처리
	currentValidators := vs.stakingManager.GetTopValidators(1000) // 충분히 큰 숫자로 모든 검증자 가져오기
	
	// 주소 맵 생성
	ethereumAddresses := make(map[string]bool)
	for _, val := range validators {
		ethereumAddresses[val.Address] = true
	}
	
	// 이더리움에 없는 검증자 제거
	for _, currentVal := range currentValidators {
		if !ethereumAddresses[currentVal.Address] {
			vs.logger.Info("이더리움 계약에 없는 검증자 제거", "address", currentVal.Address)
			err = vs.stakingManager.Unstake(currentVal.Address, currentVal.Amount)
			if err != nil {
				vs.logger.Error("검증자 언스테이킹 실패", "address", currentVal.Address, "error", err)
			}
		}
	}
	
	// 동기화 완료 후 업데이트된 검증자 세트 전송
	pendingUpdates := vs.manager.GetPendingUpdates()
	if len(pendingUpdates) > 0 {
		vs.logger.Info("텐더민트 검증자 업데이트 제출", "count", len(pendingUpdates))
		
		// 이더리움 계약에 업데이트 제출
		err = vs.ethereumClient.SubmitValidatorUpdates(vs.ctx, pendingUpdates, vs.config.ValidatorSetContract)
		if err != nil {
			vs.logger.Error("검증자 업데이트 제출 실패", "error", err)
			return err
		}
	}
	
	vs.lastSuccessfulSync = time.Now()
	vs.logger.Info("검증자 동기화 완료", "duration", time.Since(vs.lastSyncTime))
	
	return nil
}

// GetSyncStatus는 동기화 상태 정보를 반환합니다.
func (vs *ValidatorSync) GetSyncStatus() map[string]interface{} {
	vs.syncMutex.Lock()
	defer vs.syncMutex.Unlock()
	
	status := map[string]interface{}{
		"is_running":           vs.isRunning,
		"block_height":         vs.blockHeight,
		"last_sync_time":       vs.lastSyncTime,
		"last_successful_sync": vs.lastSuccessfulSync,
		"sync_interval":        vs.config.SyncInterval.String(),
		"validator_count":      vs.manager.GetValidatorCount(),
	}
	
	// 최근 에러가 있으면 추가
	if len(vs.syncErrors) > 0 {
		recentErrors := make([]string, 0, len(vs.syncErrors))
		for _, err := range vs.syncErrors {
			recentErrors = append(recentErrors, err.Error())
		}
		status["recent_errors"] = recentErrors
	}
	
	return status
}

// ExportValidatorsToJSON은 현재 검증자 세트를 JSON 형식으로 내보냅니다.
func (vs *ValidatorSync) ExportValidatorsToJSON() ([]byte, error) {
	vs.syncMutex.Lock()
	defer vs.syncMutex.Unlock()
	
	validators := vs.stakingManager.GetTopValidators(1000)
	
	type ExportedValidator struct {
		Address     string   `json:"address"`
		PubKey      string   `json:"pubKey"`
		Power       int64    `json:"power"`
		StakeAmount string   `json:"stakeAmount"`
		Commission  uint32   `json:"commission"`
	}
	
	exportedVals := make([]ExportedValidator, len(validators))
	
	for i, val := range validators {
		power := vs.stakingManager.CalculateVotingPower(vs.stakingManager.GetTotalStake(val.Address))
		exportedVals[i] = ExportedValidator{
			Address:     val.Address,
			PubKey:      fmt.Sprintf("%X", val.PubKey),
			Power:       power,
			StakeAmount: vs.stakingManager.GetTotalStake(val.Address).String(),
			Commission:  val.Commission,
		}
	}
	
	return json.Marshal(exportedVals)
}

// ImportValidatorsFromJSON은 JSON 형식에서 검증자 세트를 가져옵니다.
func (vs *ValidatorSync) ImportValidatorsFromJSON(data []byte) error {
	vs.syncMutex.Lock()
	defer vs.syncMutex.Unlock()
	
	type ImportedValidator struct {
		Address     string   `json:"address"`
		PubKey      string   `json:"pubKey"`
		StakeAmount string   `json:"stakeAmount"`
		Commission  uint32   `json:"commission"`
	}
	
	var importedVals []ImportedValidator
	
	if err := json.Unmarshal(data, &importedVals); err != nil {
		return err
	}
	
	for _, val := range importedVals {
		// 16진수 문자열에서 바이트 배열로 변환
		pubKeyBytes, err := hex.DecodeString(val.PubKey)
		if err != nil {
			vs.logger.Error("PubKey 디코딩 실패", "pubKey", val.PubKey, "error", err)
			continue
		}
		
		// 스테이크 금액 파싱
		stakeAmount := new(big.Int)
		stakeAmount, ok := stakeAmount.SetString(val.StakeAmount, 10)
		if !ok {
			vs.logger.Error("스테이크 금액 파싱 실패", "amount", val.StakeAmount)
			continue
		}
		
		// 검증자 스테이킹
		err = vs.stakingManager.Stake(pubKeyBytes, val.Address, stakeAmount, val.Commission)
		if err != nil {
			vs.logger.Error("검증자 가져오기 실패", "address", val.Address, "error", err)
			continue
		}
		
		vs.logger.Info("검증자 가져옴", "address", val.Address, "stake", stakeAmount)
	}
	
	vs.logger.Info("검증자 가져오기 완료", "count", len(importedVals))
	
	return nil
} 