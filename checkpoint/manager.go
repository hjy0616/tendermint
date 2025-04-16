package checkpoint

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/store"
	tmdb "github.com/tendermint/tm-db"
)

const (
	// DefaultCheckpointInterval은 기본 체크포인트 생성 간격입니다 (블록 수).
	DefaultCheckpointInterval = 1000

	// 키 접두사
	checkpointPrefix    = "cp_"
	checkpointCountKey  = "cp_count"
	latestCheckpointKey = "cp_latest"
)

var (
	// ErrCheckpointNotFound는 체크포인트를 찾을 수 없을 때 반환됩니다.
	ErrCheckpointNotFound = errors.New("checkpoint not found")

	// ErrInvalidCheckpoint는 체크포인트가 유효하지 않을 때 반환됩니다.
	ErrInvalidCheckpoint = errors.New("invalid checkpoint")

	// ErrInvalidSignature는 체크포인트 서명이 유효하지 않을 때 반환됩니다.
	ErrInvalidSignature = errors.New("invalid checkpoint signature")
)

// DefaultManager는 Manager 인터페이스의 기본 구현입니다.
type DefaultManager struct {
	config Config
	db     tmdb.DB
	store  *store.BlockStore
	state  state.Store
	logger log.Logger
	mtx    sync.RWMutex
}

// NewManager는 새로운 체크포인트 매니저를 생성합니다.
func NewManager(db tmdb.DB, blockStore *store.BlockStore, stateStore state.Store, config *Config) Manager {
	if config == nil {
		config = &Config{
			CheckpointInterval: DefaultCheckpointInterval,
		}
	}

	return &DefaultManager{
		config: *config,
		db:     db,
		store:  blockStore,
		state:  stateStore,
		logger: log.NewNopLogger(),
	}
}

// SetLogger는 로거를 설정합니다.
func (m *DefaultManager) SetLogger(logger log.Logger) {
	m.logger = logger
}

// CreateCheckpoint는 지정된 블록 높이에서 새 체크포인트를 생성합니다.
func (m *DefaultManager) CreateCheckpoint(ctx context.Context, height uint64) (*Checkpoint, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.logger.Info("Creating checkpoint", "height", height)

	// 블록 및 상태 정보 가져오기
	block := m.store.LoadBlock(int64(height))
	if block == nil {
		return nil, fmt.Errorf("block at height %d not found", height)
	}

	// Load() 메서드 사용
	state, err := m.state.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	validatorSet, err := m.state.LoadValidators(int64(height))
	if err != nil {
		return nil, fmt.Errorf("failed to load validators: %w", err)
	}

	// 체크포인트 번호 생성
	count, err := m.GetCheckpointCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint count: %w", err)
	}

	checkpointNumber := count + 1

	// 체크포인트 데이터 생성
	data := CheckpointData{
		Height:        height,
		BlockHash:     block.Hash(),
		StateRootHash: state.AppHash,
		AppHash:       block.AppHash,
		ValidatorHash: validatorSet.Hash(),
		Timestamp:     time.Now().Unix(),
	}

	// 서명 추가 (구성에 개인키가 있는 경우)
	if m.config.PrivateKey != nil {
		// 데이터 직렬화 (서명을 위해)
		dataBytes := m.serializeCheckpointData(data)

		// 서명 생성
		signature, err := m.config.PrivateKey.Sign(dataBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to sign checkpoint: %w", err)
		}

		data.Signature = signature
		data.SignerAddress = m.config.PrivateKey.PubKey().Address()
	}

	// 체크포인트 생성
	checkpoint := &Checkpoint{
		Data:        data,
		Number:      checkpointNumber,
		BlockHeight: height,
	}

	// 저장
	if err := m.saveCheckpoint(checkpoint); err != nil {
		return nil, fmt.Errorf("failed to save checkpoint: %w", err)
	}

	// 카운트 업데이트
	if err := m.updateCheckpointCount(checkpointNumber); err != nil {
		return nil, fmt.Errorf("failed to update checkpoint count: %w", err)
	}

	// 최신 체크포인트 업데이트
	if err := m.updateLatestCheckpoint(checkpointNumber); err != nil {
		return nil, fmt.Errorf("failed to update latest checkpoint: %w", err)
	}

	m.logger.Info("Checkpoint created", "number", checkpointNumber, "height", height)

	return checkpoint, nil
}

// GetCheckpoint는 지정된 체크포인트 번호로 체크포인트를 조회합니다.
func (m *DefaultManager) GetCheckpoint(number uint64) (*Checkpoint, error) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	m.logger.Debug("Getting checkpoint", "number", number)

	key := []byte(fmt.Sprintf("%s%d", checkpointPrefix, number))
	data, err := m.db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint from db: %w", err)
	}

	if data == nil {
		return nil, ErrCheckpointNotFound
	}

	// JSON 언마샬링 사용
	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// GetLatestCheckpoint는 가장 최근 체크포인트를 조회합니다.
func (m *DefaultManager) GetLatestCheckpoint() (*Checkpoint, error) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	m.logger.Debug("Getting latest checkpoint")

	data, err := m.db.Get([]byte(latestCheckpointKey))
	if err != nil {
		return nil, fmt.Errorf("failed to get latest checkpoint number: %w", err)
	}

	if data == nil {
		return nil, ErrCheckpointNotFound
	}

	var number uint64
	number = binary.BigEndian.Uint64(data)

	return m.GetCheckpoint(number)
}

// GetCheckpointCount는 전체 체크포인트 수를 반환합니다.
func (m *DefaultManager) GetCheckpointCount() (uint64, error) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	data, err := m.db.Get([]byte(checkpointCountKey))
	if err != nil {
		return 0, fmt.Errorf("failed to get checkpoint count: %w", err)
	}

	if data == nil {
		return 0, nil
	}

	return binary.BigEndian.Uint64(data), nil
}

// VerifyCheckpoint는 체크포인트의 유효성을 검증합니다.
func (m *DefaultManager) VerifyCheckpoint(checkpoint *Checkpoint) error {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	m.logger.Debug("Verifying checkpoint", "number", checkpoint.Number)

	// 체크포인트 데이터 검증
	if checkpoint.Data.Height == 0 || len(checkpoint.Data.BlockHash) == 0 || len(checkpoint.Data.StateRootHash) == 0 {
		return ErrInvalidCheckpoint
	}

	// 서명 검증 (구성에 공개키가 있는 경우)
	if m.config.PublicKey != nil && len(checkpoint.Data.Signature) > 0 {
		dataBytes := m.serializeCheckpointData(checkpoint.Data)

		// 서명 검증
		if !m.config.PublicKey.VerifySignature(dataBytes, checkpoint.Data.Signature) {
			return ErrInvalidSignature
		}
	}

	return nil
}

// saveCheckpoint는 체크포인트를 데이터베이스에 저장합니다.
func (m *DefaultManager) saveCheckpoint(checkpoint *Checkpoint) error {
	key := []byte(fmt.Sprintf("%s%d", checkpointPrefix, checkpoint.Number))

	// JSON 마샬링 사용
	data, err := json.Marshal(checkpoint)
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	return m.db.Set(key, data)
}

// updateCheckpointCount는 체크포인트 카운트를 업데이트합니다.
func (m *DefaultManager) updateCheckpointCount(count uint64) error {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, count)

	return m.db.Set([]byte(checkpointCountKey), data)
}

// updateLatestCheckpoint는 최신 체크포인트 번호를 업데이트합니다.
func (m *DefaultManager) updateLatestCheckpoint(number uint64) error {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, number)

	return m.db.Set([]byte(latestCheckpointKey), data)
}

// serializeCheckpointData는 체크포인트 데이터를 직렬화합니다 (서명용).
func (m *DefaultManager) serializeCheckpointData(data CheckpointData) []byte {
	// 체크포인트 서명을 위한 데이터는 다음과 같이 구성됩니다:
	// - 높이 (8바이트)
	// - 블록 해시
	// - 상태 루트 해시
	// - 애플리케이션 해시
	// - 검증자 해시
	// - 타임스탬프 (8바이트)

	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, data.Height)

	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(data.Timestamp))

	serialized := append(heightBytes, data.BlockHash...)
	serialized = append(serialized, data.StateRootHash...)
	serialized = append(serialized, data.AppHash...)
	serialized = append(serialized, data.ValidatorHash...)
	serialized = append(serialized, timestampBytes...)

	return serialized
}
