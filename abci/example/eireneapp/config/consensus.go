package config

import (
	"time"

	tmconfig "github.com/tendermint/tendermint/config"
)

// EireneConsensusConfig 이더리움 호환성을 위한 텐더민트 합의 파라미터 설정
type EireneConsensusConfig struct {
	*tmconfig.ConsensusConfig
}

// DefaultEireneConsensusConfig는 이더리움 호환성을 위한 기본 합의 파라미터 설정을 반환합니다.
func DefaultEireneConsensusConfig() *EireneConsensusConfig {
	// 기본 텐더민트 합의 설정 가져오기
	tmCfg := tmconfig.DefaultConsensusConfig()

	// 이더리움 호환성을 위한 설정 조정
	// 이더리움은 보통 ~15초 블록 시간을 가짐
	tmCfg.TimeoutPropose = 3000 * time.Millisecond      // 3초
	tmCfg.TimeoutProposeDelta = 500 * time.Millisecond  // 라운드당 0.5초 증가
	tmCfg.TimeoutPrevote = 1000 * time.Millisecond      // 1초
	tmCfg.TimeoutPrevoteDelta = 500 * time.Millisecond  // 라운드당 0.5초 증가
	tmCfg.TimeoutPrecommit = 1000 * time.Millisecond    // 1초
	tmCfg.TimeoutPrecommitDelta = 500 * time.Millisecond // 라운드당 0.5초 증가
	tmCfg.TimeoutCommit = 5000 * time.Millisecond       // 5초 (총 블록 생성 시간 ~10-15초)

	// 빈 블록 생성 전략
	// 이더리움은 트랜잭션이 없어도 일정 간격으로 블록 생성
	tmCfg.CreateEmptyBlocks = true
	tmCfg.CreateEmptyBlocksInterval = 15 * time.Second // 최대 15초마다 빈 블록 생성

	// 피어 통신 최적화
	tmCfg.PeerGossipSleepDuration = 50 * time.Millisecond      // 더 빠른 가십 전파
	tmCfg.PeerQueryMaj23SleepDuration = 1000 * time.Millisecond // 더 빠른 합의 쿼리

	return &EireneConsensusConfig{
		ConsensusConfig: tmCfg,
	}
}

// TestEireneConsensusConfig는 테스트 환경을 위한 합의 파라미터 설정을 반환합니다.
func TestEireneConsensusConfig() *EireneConsensusConfig {
	tmCfg := tmconfig.TestConsensusConfig()
	
	// 테스트 환경에서는 빠른 합의를 위해 시간 단축
	tmCfg.TimeoutCommit = 500 * time.Millisecond
	tmCfg.CreateEmptyBlocks = true
	tmCfg.CreateEmptyBlocksInterval = 1 * time.Second
	
	return &EireneConsensusConfig{
		ConsensusConfig: tmCfg,
	}
}

// AdjustForEirene는 텐더민트 합의 파라미터를 이더리움 호환성에 맞게 조정합니다.
func AdjustForEirene(cfg *tmconfig.Config) *tmconfig.Config {
	// 합의 파라미터 조정
	eireneCfg := DefaultEireneConsensusConfig()
	cfg.Consensus = eireneCfg.ConsensusConfig
	
	// 멤풀 설정 조정
	// 이더리움은 가스 가격 기반 우선순위 사용
	cfg.Mempool.Version = tmconfig.MempoolV1 // 우선순위 기반 멤풀 사용
	cfg.Mempool.Size = 5000                  // 멤풀 크기 증가
	
	// 최대 트랜잭션 크기 설정 (이더리움 트랜잭션 수용)
	cfg.Mempool.MaxTxBytes = 1024 * 1024  // 1MB
	cfg.Mempool.MaxTxsBytes = 1024 * 1024 * 1024 // 1GB
	
	// TTL 설정
	cfg.Mempool.TTLDuration = 60 * time.Minute
	cfg.Mempool.TTLNumBlocks = 60 // 약 15-20분
	
	// P2P 설정 조정
	cfg.P2P.MaxNumInboundPeers = 50
	cfg.P2P.MaxNumOutboundPeers = 20
	
	// 블록 쿼리를 위한 RPC 설정
	cfg.RPC.MaxSubscriptionsPerClient = 10
	cfg.RPC.TimeoutBroadcastTxCommit = 30 * time.Second
	
	return cfg
}

// DevelopmentConfig는 개발 환경에 적합한 설정을 반환합니다.
func DevelopmentConfig() *tmconfig.Config {
	cfg := tmconfig.DefaultConfig()
	eireneCfg := DefaultEireneConsensusConfig()
	
	// 개발 환경에서는 더 빠른 블록 생성
	eireneCfg.TimeoutCommit = 1 * time.Second
	eireneCfg.CreateEmptyBlocksInterval = 5 * time.Second
	
	cfg.Consensus = eireneCfg.ConsensusConfig
	return cfg
}

// ProductionConfig는 프로덕션 환경에 적합한 설정을 반환합니다.
func ProductionConfig() *tmconfig.Config {
	cfg := tmconfig.DefaultConfig()
	return AdjustForEirene(cfg)
} 