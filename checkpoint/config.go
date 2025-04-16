package checkpoint

// Config는 체크포인트 시스템의 설정을 담는 구조체입니다.
type Config struct {
	// 체크포인트 활성화 여부
	Enabled bool `mapstructure:"enabled"`

	// 체크포인트 생성 간격 (블록 수)
	Interval int64 `mapstructure:"interval"`

	// 체크포인트 저장 경로
	DataDir string `mapstructure:"data_dir"`

	// 체크포인트 보관 일수
	RetentionDays int `mapstructure:"retention_days"`

	// 체크포인트 수 제한
	MaxCount int `mapstructure:"max_count"`
}

// DefaultConfig는 기본 체크포인트 설정을 반환합니다.
func DefaultConfig() Config {
	return Config{
		Enabled:       true,
		Interval:      100,  // 100블록마다 체크포인트 생성
		DataDir:       "",   // 기본 데이터 디렉토리 사용
		RetentionDays: 7,    // 7일 동안 보관
		MaxCount:      1000, // 최대 1000개 체크포인트 보관
	}
}

// ValidateBasic은 기본적인 설정 유효성을 검사합니다.
func (c Config) ValidateBasic() error {
	// 체크포인트가 비활성화된 경우 다른 설정은 검사하지 않음
	if !c.Enabled {
		return nil
	}

	// 체크포인트 간격은 최소 1 이상이어야 함
	if c.Interval < 1 {
		c.Interval = 1
	}

	// 최대 체크포인트 수는 최소 10 이상이어야 함
	if c.MaxCount < 10 {
		c.MaxCount = 10
	}

	return nil
}
