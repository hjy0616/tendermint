package config

import (
	"path/filepath"

	tmconfig "github.com/tendermint/tendermint/config"
)

// EireneConfig는 Eirene 애플리케이션의 모든 설정을 포함합니다.
type EireneConfig struct {
	// 텐더민트 기본 설정
	TmConfig *tmconfig.Config
	
	// 투표 메커니즘 관리자
	VotingManager *VotingPowerManager
	
	// 추가 설정
	ChainID      string
	NetworkID    uint64
	IsTestnet    bool
	DataDir      string
	KeyStoreDir  string
	LogLevel     string
	RPCEnabled   bool
	RPCPort      int
	WSEnabled    bool
	WSPort       int
	MetricsEnabled bool
}

// DefaultEireneConfig는 기본 Eirene 설정을 반환합니다.
func DefaultEireneConfig() *EireneConfig {
	// 텐더민트 기본 설정
	tmConfig := tmconfig.DefaultConfig()
	
	// Eirene에 맞게 조정
	tmConfig = AdjustForEirene(tmConfig)
	
	return &EireneConfig{
		TmConfig:       tmConfig,
		VotingManager:  NewVotingPowerManager(),
		ChainID:        "eirene_1",
		NetworkID:      1,
		IsTestnet:      false,
		DataDir:        "data",
		KeyStoreDir:    "keystore",
		LogLevel:       "info",
		RPCEnabled:     true,
		RPCPort:        8545,
		WSEnabled:      true,
		WSPort:         8546,
		MetricsEnabled: true,
	}
}

// TestnetConfig는 테스트넷 환경을 위한 Eirene 설정을 반환합니다.
func TestnetConfig() *EireneConfig {
	config := DefaultEireneConfig()
	config.ChainID = "eirene_testnet"
	config.NetworkID = 3
	config.IsTestnet = true
	
	// 테스트넷용 텐더민트 설정 조정
	config.TmConfig.Consensus = TestEireneConsensusConfig().ConsensusConfig
	
	return config
}

// DevConfig는 개발 환경을 위한 Eirene 설정을 반환합니다.
func DevConfig() *EireneConfig {
	config := DefaultEireneConfig()
	config.ChainID = "eirene_dev"
	config.NetworkID = 2018
	config.IsTestnet = true
	
	// 개발용 빠른 블록 생성
	config.TmConfig = DevelopmentConfig()
	
	return config
}

// EnsureDirs는 필요한 디렉토리가 존재하는지 확인합니다.
func (cfg *EireneConfig) EnsureDirs() error {
	dirs := []string{
		cfg.DataDir,
		cfg.KeyStoreDir,
		filepath.Join(cfg.DataDir, "tendermint"),
		filepath.Join(cfg.DataDir, "ethereum"),
	}
	
	for _, dir := range dirs {
		if err := ensureDir(dir); err != nil {
			return err
		}
	}
	
	return nil
}

// ensureDir은 디렉토리가 존재하는지 확인하고, 없으면 생성합니다.
func ensureDir(dir string) error {
	// 간단한 디렉토리 생성 로직
	// 실제 구현에서는 os.MkdirAll 같은 함수를 사용
	return nil
}

// SetupEireneNode는 Eirene 노드의 설정을 초기화합니다.
func SetupEireneNode(homeDir string) (*EireneConfig, error) {
	config := DefaultEireneConfig()
	
	// 노드 홈 디렉토리 설정
	config.TmConfig.SetRoot(homeDir)
	config.DataDir = filepath.Join(homeDir, "data")
	config.KeyStoreDir = filepath.Join(homeDir, "keystore")
	
	// 디렉토리 생성
	if err := config.EnsureDirs(); err != nil {
		return nil, err
	}
	
	return config, nil
}

// CreateGenesisConfig는 제네시스 설정을 생성합니다.
func CreateGenesisConfig(chainID string, validators []string) error {
	// 제네시스 설정 로직 구현
	// validators는 초기 검증자 주소 목록
	return nil
} 