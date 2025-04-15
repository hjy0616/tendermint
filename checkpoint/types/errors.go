package types

import (
	"fmt"
)

// 체크포인트 오류 코드 정의
const (
	ErrCodeUnknown        uint32 = 0
	ErrCodeNotInitialized uint32 = 1
	ErrCodeInvalidHeight  uint32 = 2
	ErrCodeInvalidData    uint32 = 3
	ErrCodeNotFound       uint32 = 4
	ErrCodeDatabaseError  uint32 = 5
)

// CheckpointError는 체크포인트 관련 오류를 나타냅니다.
type CheckpointError struct {
	Code    uint32
	Message string
}

// NewCheckpointError는 체크포인트 오류를 생성합니다.
func NewCheckpointError(code uint32, message string) *CheckpointError {
	return &CheckpointError{
		Code:    code,
		Message: message,
	}
}

// Error는 오류 메시지를 구현합니다.
func (e *CheckpointError) Error() string {
	return fmt.Sprintf("checkpoint error [%d]: %s", e.Code, e.Message)
}

// IsNotInitializedError는 초기화되지 않았을 때의 오류인지 확인합니다.
func IsNotInitializedError(err error) bool {
	if err == nil {
		return false
	}

	cErr, ok := err.(*CheckpointError)
	if !ok {
		return false
	}

	return cErr.Code == ErrCodeNotInitialized
}

// IsNotFoundError는 체크포인트를 찾을 수 없을 때의 오류인지 확인합니다.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	cErr, ok := err.(*CheckpointError)
	if !ok {
		return false
	}

	return cErr.Code == ErrCodeNotFound
}
