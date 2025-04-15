package checkpoint

import (
	"encoding/binary"
	"fmt"
	"sync"

	dbm "github.com/tendermint/tm-db"

	"github.com/tendermint/tendermint/checkpoint/types"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	// 체크포인트 저장소에서 사용하는 키 접두사
	checkpointPrefix   = "checkpoint:"       // 체크포인트 데이터
	latestHeightKey    = "checkpoint:latest" // 가장 최근 체크포인트 높이
	checkpointCountKey = "checkpoint:count"  // 체크포인트 카운트
)

// DBStore는 DB 기반 체크포인트 저장소입니다.
type DBStore struct {
	db     dbm.DB
	logger log.Logger
	mtx    sync.RWMutex
}

// NewDBStore는 새 DB 체크포인트 저장소를 생성합니다.
func NewDBStore(db dbm.DB, logger log.Logger) *DBStore {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	return &DBStore{
		db:     db,
		logger: logger,
	}
}

// checkpointKey는 체크포인트 높이에 대한 DB 키를 생성합니다.
func checkpointKey(height int64) []byte {
	key := make([]byte, len(checkpointPrefix)+8)
	copy(key, checkpointPrefix)
	binary.BigEndian.PutUint64(key[len(checkpointPrefix):], uint64(height))
	return key
}

// SaveCheckpoint는 새 체크포인트를 저장합니다.
func (s *DBStore) SaveCheckpoint(checkpoint *types.Checkpoint) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	height := checkpoint.Height
	if height <= 0 {
		return fmt.Errorf("invalid checkpoint height: %d", height)
	}

	// 체크포인트 데이터 직렬화
	bz, err := cdc.MarshalBinaryBare(checkpoint)
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	batch := s.db.NewBatch()
	defer batch.Close()

	// 체크포인트 저장
	batch.Set(checkpointKey(height), bz)

	// 최신 높이 업데이트
	latestHeight, err := s.getLatestHeight()
	if err != nil || latestHeight < height {
		latestHeightBz := make([]byte, 8)
		binary.BigEndian.PutUint64(latestHeightBz, uint64(height))
		batch.Set([]byte(latestHeightKey), latestHeightBz)
	}

	// 카운트 업데이트
	count, err := s.GetCheckpointCount()
	if err != nil {
		count = 0
	}
	countBz := make([]byte, 8)
	binary.BigEndian.PutUint64(countBz, uint64(count+1))
	batch.Set([]byte(checkpointCountKey), countBz)

	// 트랜잭션 커밋
	if err := batch.Write(); err != nil {
		return fmt.Errorf("failed to write batch: %w", err)
	}

	s.logger.Info("Saved checkpoint", "height", height)
	return nil
}

// getLatestHeight는 가장 최근 체크포인트 높이를 반환합니다.
func (s *DBStore) getLatestHeight() (int64, error) {
	bz, err := s.db.Get([]byte(latestHeightKey))
	if err != nil {
		return 0, err
	}
	if len(bz) == 0 {
		return 0, nil
	}
	return int64(binary.BigEndian.Uint64(bz)), nil
}

// GetCheckpoint는 지정된 높이의 체크포인트를 조회합니다.
func (s *DBStore) GetCheckpoint(height int64) (*types.Checkpoint, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	bz, err := s.db.Get(checkpointKey(height))
	if err != nil {
		return nil, err
	}
	if len(bz) == 0 {
		return nil, fmt.Errorf("checkpoint at height %d not found", height)
	}

	var checkpoint types.Checkpoint
	if err := cdc.UnmarshalBinaryBare(bz, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// GetLatestCheckpoint는 가장 최근 체크포인트를 반환합니다.
func (s *DBStore) GetLatestCheckpoint() (*types.Checkpoint, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	height, err := s.getLatestHeight()
	if err != nil {
		return nil, err
	}
	if height == 0 {
		return nil, fmt.Errorf("no checkpoints found")
	}

	return s.GetCheckpoint(height)
}

// GetCheckpointCount는 저장된 체크포인트 수를 반환합니다.
func (s *DBStore) GetCheckpointCount() (int64, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	bz, err := s.db.Get([]byte(checkpointCountKey))
	if err != nil {
		return 0, err
	}
	if len(bz) == 0 {
		return 0, nil
	}
	return int64(binary.BigEndian.Uint64(bz)), nil
}

// PruneCheckpoints는 특정 높이 이전의 체크포인트를 제거합니다.
func (s *DBStore) PruneCheckpoints(heightThreshold int64) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// 모든 체크포인트 조회
	it, err := s.db.Iterator([]byte(checkpointPrefix), nil)
	if err != nil {
		return err
	}
	defer it.Close()

	batch := s.db.NewBatch()
	defer batch.Close()

	pruned := int64(0)
	for ; it.Valid(); it.Next() {
		key := it.Key()
		if len(key) < len(checkpointPrefix)+8 {
			continue
		}
		height := int64(binary.BigEndian.Uint64(key[len(checkpointPrefix):]))
		if height < heightThreshold {
			batch.Delete(key)
			pruned++
		}
	}

	if pruned > 0 {
		// 카운트 업데이트
		count, err := s.GetCheckpointCount()
		if err != nil {
			count = 0
		}
		countBz := make([]byte, 8)
		binary.BigEndian.PutUint64(countBz, uint64(count-pruned))
		batch.Set([]byte(checkpointCountKey), countBz)

		// 최신 높이 확인 및 업데이트
		latestHeight, err := s.getLatestHeight()
		if err == nil && latestHeight < heightThreshold {
			// 새 최신 높이 찾기
			it, err := s.db.ReverseIterator([]byte(checkpointPrefix), nil)
			if err == nil {
				defer it.Close()
				if it.Valid() {
					key := it.Key()
					if len(key) >= len(checkpointPrefix)+8 {
						newLatestHeight := int64(binary.BigEndian.Uint64(key[len(checkpointPrefix):]))
						latestHeightBz := make([]byte, 8)
						binary.BigEndian.PutUint64(latestHeightBz, uint64(newLatestHeight))
						batch.Set([]byte(latestHeightKey), latestHeightBz)
					}
				}
			}
		}

		// 트랜잭션 커밋
		if err := batch.Write(); err != nil {
			return fmt.Errorf("failed to write batch: %w", err)
		}
	}

	s.logger.Info("Pruned checkpoints", "count", pruned, "threshold", heightThreshold)
	return nil
}
