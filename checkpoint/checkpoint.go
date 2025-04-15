package checkpoint

import (
	"os"
	"sync"
	"time"

	dbm "github.com/tendermint/tm-db"

	"github.com/tendermint/tendermint/checkpoint/types"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	// 전역 체크포인트 저장소
	globalStore types.Store
	// 락 객체
	mtx sync.RWMutex
	// 로거
	logger log.Logger
	// 초기화 여부
	initialized bool
)

// Initialize는 체크포인트 모듈을 초기화합니다.
// 반드시 애플리케이션 시작 시에 호출해야 합니다.
func Initialize(db dbm.DB, log log.Logger) {
	mtx.Lock()
	defer mtx.Unlock()

	if initialized {
		return
	}

	if log == nil {
		log = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	}

	logger = log.With("module", "checkpoint")
	globalStore = NewDBStore(db, logger)
	initialized = true

	logger.Info("Checkpoint module initialized")
}

// IsInitialized는 체크포인트 모듈의 초기화 여부를 반환합니다.
func IsInitialized() bool {
	mtx.RLock()
	defer mtx.RUnlock()
	return initialized
}

// SetCheckpointInterval은 체크포인트 생성 간격을 설정합니다.
func SetCheckpointInterval(interval int64) {
	if interval > 0 {
		types.SetCheckpointInterval(interval)
		logger.Info("Checkpoint interval set", "blocks", interval)
	}
}

// SaveCheckpoint는 새 체크포인트를 저장합니다.
func SaveCheckpoint(checkpoint *types.Checkpoint) error {
	mtx.RLock()
	defer mtx.RUnlock()

	if !initialized {
		return ErrNotInitialized()
	}

	return globalStore.SaveCheckpoint(checkpoint)
}

// GetCheckpoint는 지정된 높이의 체크포인트를 조회합니다.
func GetCheckpoint(height int64) (*types.Checkpoint, error) {
	mtx.RLock()
	defer mtx.RUnlock()

	if !initialized {
		return nil, ErrNotInitialized()
	}

	return globalStore.GetCheckpoint(height)
}

// GetLatestCheckpoint는 가장 최근 체크포인트를 반환합니다.
func GetLatestCheckpoint() (*types.Checkpoint, error) {
	mtx.RLock()
	defer mtx.RUnlock()

	if !initialized {
		return nil, ErrNotInitialized()
	}

	return globalStore.GetLatestCheckpoint()
}

// GetCheckpointCount는 저장된 체크포인트 수를 반환합니다.
func GetCheckpointCount() (int64, error) {
	mtx.RLock()
	defer mtx.RUnlock()

	if !initialized {
		return 0, ErrNotInitialized()
	}

	return globalStore.GetCheckpointCount()
}

// PruneCheckpoints는 특정 높이 이전의 체크포인트를 제거합니다.
func PruneCheckpoints(heightThreshold int64) error {
	mtx.RLock()
	defer mtx.RUnlock()

	if !initialized {
		return ErrNotInitialized()
	}

	return globalStore.PruneCheckpoints(heightThreshold)
}

// CreateCheckpoint는 현재 상태에서 체크포인트를 생성합니다.
// 이 함수는 블록 처리 후 체크포인트 생성 높이에서 호출됩니다.
func CreateCheckpoint(height int64, appHash, rootHash []byte, validatorSetHash, nextValidatorSetHash []byte) (*types.Checkpoint, error) {
	if !types.IsCheckpointHeight(height) {
		return nil, nil // 체크포인트 생성 높이가 아님
	}

	checkpoint := &types.Checkpoint{
		Height:               height,
		AppHash:              appHash,
		RootHash:             rootHash,
		Timestamp:            currentTime(),
		ValidatorSetHash:     validatorSetHash,
		NextValidatorSetHash: nextValidatorSetHash,
	}

	err := SaveCheckpoint(checkpoint)
	if err != nil {
		return nil, err
	}

	logger.Info("Created checkpoint", "height", height, "app_hash", appHash)
	return checkpoint, nil
}

// ErrNotInitialized는 모듈이 초기화되지 않았을 때 발생하는 오류입니다.
func ErrNotInitialized() error {
	return types.NewCheckpointError(types.ErrCodeNotInitialized, "checkpoint module not initialized")
}

// currentTime는 현재 시간을 반환합니다. (테스트 가능)
var currentTime = func() time.Time {
	return time.Now().UTC()
}
