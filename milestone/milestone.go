package milestone

import (
	"context"
	"sync"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/milestone/types"
	dbm "github.com/tendermint/tm-db"
)

var (
	// defaultService는 전역 마일스톤 서비스 인스턴스입니다.
	defaultService     MilestoneService
	defaultServiceLock sync.RWMutex
	defaultServiceOnce sync.Once
)

// InitMilestoneService는 기본 마일스톤 서비스를 초기화합니다.
func InitMilestoneService(db dbm.DB, logger log.Logger) {
	defaultServiceOnce.Do(func() {
		defaultServiceLock.Lock()
		defer defaultServiceLock.Unlock()

		defaultService = NewMilestoneService(
			db,
			WithLogger(logger),
		)

		// 서비스 시작
		if err := defaultService.Start(); err != nil {
			logger.Error("Failed to start milestone service", "error", err)
		}
	})
}

// GetMilestoneService는 기본 마일스톤 서비스를 반환합니다.
func GetMilestoneService() MilestoneService {
	defaultServiceLock.RLock()
	defer defaultServiceLock.RUnlock()
	return defaultService
}

// AddMilestone은 마일스톤을 추가합니다.
func AddMilestone(ctx context.Context, milestone *types.Milestone) error {
	if defaultService == nil {
		return ErrMilestoneNotFound
	}
	return defaultService.AddMilestone(ctx, milestone)
}

// GetMilestone은 ID로 마일스톤을 조회합니다.
func GetMilestone(ctx context.Context, id uint64) (*types.Milestone, error) {
	if defaultService == nil {
		return nil, ErrMilestoneNotFound
	}
	return defaultService.GetMilestone(ctx, id)
}

// GetMilestones는 ID 범위로 마일스톤 목록을 조회합니다.
func GetMilestones(ctx context.Context, fromID, toID uint64) ([]types.Milestone, error) {
	if defaultService == nil {
		return nil, ErrMilestoneNotFound
	}
	return defaultService.GetMilestones(ctx, fromID, toID)
}

// GetNoAckMilestones는 확인되지 않은 마일스톤 목록을 조회합니다.
func GetNoAckMilestones(ctx context.Context) ([]types.Milestone, error) {
	if defaultService == nil {
		return nil, ErrMilestoneNotFound
	}
	return defaultService.GetNoAckMilestones(ctx)
}

// AckMilestone은 마일스톤을 확인 처리합니다.
func AckMilestone(ctx context.Context, id uint64) error {
	if defaultService == nil {
		return ErrMilestoneNotFound
	}
	return defaultService.AckMilestone(ctx, id)
}

// GetMilestoneCount는 마일스톤 수를 조회합니다.
func GetMilestoneCount(ctx context.Context) (uint64, error) {
	if defaultService == nil {
		return 0, ErrMilestoneNotFound
	}
	return defaultService.GetMilestoneCount(ctx)
}

// ShutdownMilestoneService는 마일스톤 서비스를 종료합니다.
func ShutdownMilestoneService() error {
	defaultServiceLock.Lock()
	defer defaultServiceLock.Unlock()

	if defaultService != nil {
		err := defaultService.Stop()
		defaultService = nil
		return err
	}
	return nil
}
