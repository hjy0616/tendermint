package checkpoint

import (
	"sync"

	"github.com/tendermint/tendermint/checkpoint/types"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	// 전역 체크포인트 싱글톤 인스턴스
	instance *Service
	once     sync.Once
)

// Service는 체크포인트 관련 기능을 제공하는 서비스입니다.
type Service struct {
	logger log.Logger
	mtx    sync.RWMutex

	// 체크포인트 저장소
	checkpoints map[int64]*types.Checkpoint
}

// GetService는 체크포인트 서비스의 싱글톤 인스턴스를 반환합니다.
func GetService(logger log.Logger) *Service {
	once.Do(func() {
		instance = &Service{
			logger:      logger,
			checkpoints: make(map[int64]*types.Checkpoint),
		}
	})
	return instance
}

// SaveCheckpoint는 새로운 체크포인트를 저장합니다.
func (s *Service) SaveCheckpoint(checkpoint *types.Checkpoint) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.logger.Info("Saving checkpoint", "height", checkpoint.Height)
	s.checkpoints[checkpoint.Height] = checkpoint

	return nil
}

// GetCheckpoint는 지정된 높이의 체크포인트를 반환합니다.
func (s *Service) GetCheckpoint(height int64) (*types.Checkpoint, bool) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	checkpoint, exists := s.checkpoints[height]
	return checkpoint, exists
}

// GetLatestCheckpoint는 가장 최신의 체크포인트를 반환합니다.
func (s *Service) GetLatestCheckpoint() (*types.Checkpoint, bool) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	var latestHeight int64 = -1
	var latestCheckpoint *types.Checkpoint

	for height, checkpoint := range s.checkpoints {
		if height > latestHeight {
			latestHeight = height
			latestCheckpoint = checkpoint
		}
	}

	if latestHeight == -1 {
		return nil, false
	}

	return latestCheckpoint, true
}

// GetCheckpointCount는 현재 저장된 체크포인트의 수를 반환합니다.
func (s *Service) GetCheckpointCount() int64 {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return int64(len(s.checkpoints))
}
