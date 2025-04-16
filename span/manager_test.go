package span

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
)

func TestSpanManager(t *testing.T) {
	logger := log.NewNopLogger()
	manager := NewManager(logger)

	// 검증자 샘플 생성
	validators := types.ValidatorSet{
		Validators: []*types.Validator{
			types.NewValidator(nil, 10),
			types.NewValidator(nil, 20),
			types.NewValidator(nil, 30),
		},
	}

	// 테스트: 스팬 생성
	t.Run("CreateSpan", func(t *testing.T) {
		span, err := manager.CreateSpan(1, 100, 200, validators)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), span.ID)
		assert.Equal(t, uint64(100), span.StartBlock)
		assert.Equal(t, uint64(200), span.EndBlock)
		assert.Equal(t, 3, len(span.Validators.Validators))
	})

	// 테스트: 중복 스팬 ID 처리
	t.Run("DuplicateSpanID", func(t *testing.T) {
		_, err := manager.CreateSpan(1, 300, 400, validators)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	// 테스트: 블록 번호 범위 충돌
	t.Run("BlockRangeConflict", func(t *testing.T) {
		_, err := manager.CreateSpan(2, 150, 250, validators)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "conflict")
	})

	// 테스트: ID로 스팬 조회
	t.Run("GetSpanByID", func(t *testing.T) {
		span, err := manager.GetSpanByID(1)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), span.ID)

		_, err = manager.GetSpanByID(999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	// 테스트: 블록 번호로 스팬 조회
	t.Run("GetSpanByBlockNumber", func(t *testing.T) {
		span, err := manager.GetSpanByBlockNumber(150)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), span.ID)

		_, err = manager.GetSpanByBlockNumber(999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no span found")
	})

	// 테스트: 스팬 활성화
	t.Run("ActivateSpan", func(t *testing.T) {
		// 먼저 새 스팬 생성
		_, err := manager.CreateSpan(2, 300, 400, validators)
		require.NoError(t, err)

		// 활성화
		err = manager.ActivateSpan(2)
		require.NoError(t, err)

		// 현재 활성 스팬 확인
		currentSpan := manager.GetCurrentSpan()
		assert.Equal(t, uint64(2), currentSpan.ID)
	})

	// 테스트: 존재하지 않는 스팬 활성화 시도
	t.Run("ActivateNonExistingSpan", func(t *testing.T) {
		err := manager.ActivateSpan(999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	// 테스트: 스팬 활성 상태 확인
	t.Run("SpanActiveStatus", func(t *testing.T) {
		// 새 스팬 생성 (미래 시작)
		futureStart := time.Now().Add(1 * time.Hour)
		span := NewSpan(3, 500, 600, validators, futureStart)
		status := getSpanStatus(span)
		assert.Equal(t, "pending", status)

		// 활성 스팬
		activeSpan := NewSpan(4, 700, 800, validators, time.Now().Add(-1*time.Minute))
		status = getSpanStatus(activeSpan)
		assert.Equal(t, "active", status)

		// 완료된 스팬
		completedSpan := NewSpan(5, 900, 1000, validators, time.Now().Add(-2*time.Hour))
		completedSpan.CompleteSpan(time.Now().Add(-1 * time.Hour))
		status = getSpanStatus(completedSpan)
		assert.Equal(t, "completed", status)
	})

	// 테스트: JSON 내보내기/가져오기
	t.Run("ExportImport", func(t *testing.T) {
		// 현재 상태 내보내기
		data, err := manager.Export()
		require.NoError(t, err)

		// 새 관리자 생성 및 상태 가져오기
		newManager := NewManager(logger)
		err = newManager.Import(data)
		require.NoError(t, err)

		// 데이터 비교
		span1, _ := manager.GetSpanByID(1)
		span2, _ := newManager.GetSpanByID(1)
		assert.Equal(t, span1.ID, span2.ID)
		assert.Equal(t, span1.StartBlock, span2.StartBlock)
		assert.Equal(t, span1.EndBlock, span2.EndBlock)
	})
}
