package checkpoint

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"

	dbm "github.com/tendermint/tm-db"

	"github.com/tendermint/tendermint/checkpoint/types"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	// DB 네임스페이스
	dbNamespace = "checkpoint"
	// 최신 체크포인트 키
	latestCheckpointKey = "latest_checkpoint"
)

// Store는 체크포인트를 영구 저장하기 위한 저장소입니다.
type Store struct {
	logger log.Logger
	db     dbm.DB
	mtx    sync.RWMutex
}

// NewStore는 새로운 체크포인트 저장소를 생성합니다.
func NewStore(logger log.Logger, dataDir string) (*Store, error) {
	dbPath := filepath.Join(dataDir, "data", "checkpoint.db")
	db, err := dbm.NewGoLevelDB(dbNamespace, dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkpoint database: %v", err)
	}

	return &Store{
		logger: logger,
		db:     db,
	}, nil
}

// SaveCheckpoint는 체크포인트를 저장소에 저장합니다.
func (s *Store) SaveCheckpoint(checkpoint *types.Checkpoint) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// 체크포인트를 JSON으로 직렬화
	data, err := checkpoint.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %v", err)
	}

	// 체크포인트 ID로 저장
	key := []byte(fmt.Sprintf("checkpoint_%d", checkpoint.Height))
	if err := s.db.Set(key, data); err != nil {
		return fmt.Errorf("failed to save checkpoint to db: %v", err)
	}

	// 최신 체크포인트 업데이트
	latestBytes, err := s.db.Get([]byte(latestCheckpointKey))
	if err != nil {
		return fmt.Errorf("failed to get latest checkpoint: %v", err)
	}

	if latestBytes == nil || len(latestBytes) == 0 {
		// 최초 체크포인트
		s.setLatestCheckpoint(checkpoint.Height)
		return nil
	}

	var latestHeight int64
	if err := json.Unmarshal(latestBytes, &latestHeight); err != nil {
		return fmt.Errorf("failed to unmarshal latest checkpoint height: %v", err)
	}

	if checkpoint.Height > latestHeight {
		s.setLatestCheckpoint(checkpoint.Height)
	}

	return nil
}

// GetCheckpoint는 지정된 높이의 체크포인트를 반환합니다.
func (s *Store) GetCheckpoint(height int64) (*types.Checkpoint, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	key := []byte(fmt.Sprintf("checkpoint_%d", height))
	data, err := s.db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint from db: %v", err)
	}

	if data == nil || len(data) == 0 {
		return nil, nil
	}

	checkpoint, err := types.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %v", err)
	}

	return checkpoint, nil
}

// GetLatestCheckpoint는 가장 최신의 체크포인트를 반환합니다.
func (s *Store) GetLatestCheckpoint() (*types.Checkpoint, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	latestBytes, err := s.db.Get([]byte(latestCheckpointKey))
	if err != nil {
		return nil, fmt.Errorf("failed to get latest checkpoint: %v", err)
	}

	if latestBytes == nil || len(latestBytes) == 0 {
		return nil, nil
	}

	var latestHeight int64
	if err := json.Unmarshal(latestBytes, &latestHeight); err != nil {
		return nil, fmt.Errorf("failed to unmarshal latest checkpoint height: %v", err)
	}

	return s.GetCheckpoint(latestHeight)
}

// setLatestCheckpoint는 최신 체크포인트 높이를 설정합니다.
func (s *Store) setLatestCheckpoint(height int64) error {
	data, err := json.Marshal(height)
	if err != nil {
		return fmt.Errorf("failed to marshal height: %v", err)
	}

	if err := s.db.Set([]byte(latestCheckpointKey), data); err != nil {
		return fmt.Errorf("failed to set latest checkpoint height: %v", err)
	}

	return nil
}

// Close는 저장소를 닫습니다.
func (s *Store) Close() error {
	return s.db.Close()
}
