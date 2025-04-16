// +build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tendermint/tendermint/checkpoint"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/store"
	tmdb "github.com/tendermint/tm-db"
)

func main() {
	// 커맨드 라인 인자 파싱
	var (
		homeDir    = flag.String("home", filepath.Join(os.Getenv("HOME"), ".tendermint"), "Tendermint 홈 디렉토리")
		dbBackend  = flag.String("db", "goleveldb", "데이터베이스 백엔드 (goleveldb, cleveldb, boltdb, rocksdb, badgerdb)")
		createFlag = flag.Bool("create", false, "체크포인트 생성")
		height     = flag.Uint64("height", 0, "체크포인트 생성 높이 (--create와 함께 사용)")
		getFlag    = flag.Bool("get", false, "체크포인트 조회")
		number     = flag.Uint64("number", 0, "조회할 체크포인트 번호 (--get와 함께 사용)")
		listFlag   = flag.Bool("list", false, "모든 체크포인트 나열")
	)
	flag.Parse()

	// 로거 초기화
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger = log.NewFilter(logger, log.AllowInfo())

	// 홈 디렉토리 설정
	cfg := config.DefaultConfig()
	cfg.SetRoot(*homeDir)

	// 데이터베이스 초기화
	blockStoreDB, err := tmdb.NewDB("blockstore", tmdb.BackendType(*dbBackend), cfg.DBDir())
	if err != nil {
		logger.Error("Failed to create blockstore DB", "err", err)
		os.Exit(1)
	}
	defer blockStoreDB.Close()

	stateDB, err := tmdb.NewDB("state", tmdb.BackendType(*dbBackend), cfg.DBDir())
	if err != nil {
		logger.Error("Failed to create state DB", "err", err)
		os.Exit(1)
	}
	defer stateDB.Close()

	checkpointDB, err := tmdb.NewDB("checkpoint", tmdb.BackendType(*dbBackend), cfg.DBDir())
	if err != nil {
		logger.Error("Failed to create checkpoint DB", "err", err)
		os.Exit(1)
	}
	defer checkpointDB.Close()

	// 블록 저장소 및 상태 저장소 초기화
	blockStore := store.NewBlockStore(blockStoreDB)
	stateStore := state.NewStore(stateDB, state.StoreOptions{})

	// 체크포인트 매니저 초기화
	manager := checkpoint.NewManager(
		checkpointDB,
		blockStore,
		stateStore,
		&checkpoint.Config{
			CheckpointInterval: 1000,
		},
	)
	manager.SetLogger(logger)

	// 명령 실행
	if *createFlag {
		// 체크포인트 생성 (height 필요)
		if *height == 0 {
			logger.Error("Height is required for checkpoint creation")
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cp, err := manager.CreateCheckpoint(ctx, *height)
		if err != nil {
			logger.Error("Failed to create checkpoint", "err", err)
			os.Exit(1)
		}

		logger.Info("Checkpoint created successfully",
			"number", cp.Number,
			"height", cp.Data.Height,
			"block_hash", fmt.Sprintf("%X", cp.Data.BlockHash),
			"state_root", fmt.Sprintf("%X", cp.Data.StateRootHash),
		)
	} else if *getFlag {
		// 체크포인트 조회 (number 필요)
		if *number == 0 {
			// 최신 체크포인트 조회
			cp, err := manager.GetLatestCheckpoint()
			if err != nil {
				if err == checkpoint.ErrCheckpointNotFound {
					logger.Info("No checkpoints found")
					os.Exit(0)
				}
				logger.Error("Failed to get latest checkpoint", "err", err)
				os.Exit(1)
			}

			logger.Info("Latest checkpoint",
				"number", cp.Number,
				"height", cp.Data.Height,
				"block_hash", fmt.Sprintf("%X", cp.Data.BlockHash),
				"state_root", fmt.Sprintf("%X", cp.Data.StateRootHash),
				"app_hash", fmt.Sprintf("%X", cp.Data.AppHash),
				"timestamp", time.Unix(cp.Data.Timestamp, 0),
			)
		} else {
			// 특정 번호의 체크포인트 조회
			cp, err := manager.GetCheckpoint(*number)
			if err != nil {
				if err == checkpoint.ErrCheckpointNotFound {
					logger.Info("Checkpoint not found", "number", *number)
					os.Exit(0)
				}
				logger.Error("Failed to get checkpoint", "number", *number, "err", err)
				os.Exit(1)
			}

			logger.Info("Checkpoint",
				"number", cp.Number,
				"height", cp.Data.Height,
				"block_hash", fmt.Sprintf("%X", cp.Data.BlockHash),
				"state_root", fmt.Sprintf("%X", cp.Data.StateRootHash),
				"app_hash", fmt.Sprintf("%X", cp.Data.AppHash),
				"timestamp", time.Unix(cp.Data.Timestamp, 0),
			)
		}
	} else if *listFlag {
		// 체크포인트 개수 조회
		count, err := manager.GetCheckpointCount()
		if err != nil {
			logger.Error("Failed to get checkpoint count", "err", err)
			os.Exit(1)
		}

		logger.Info("Total checkpoints", "count", count)

		// 모든 체크포인트 나열
		if count > 0 {
			for i := uint64(1); i <= count; i++ {
				cp, err := manager.GetCheckpoint(i)
				if err != nil {
					logger.Error("Failed to get checkpoint", "number", i, "err", err)
					continue
				}

				logger.Info("Checkpoint",
					"number", cp.Number,
					"height", cp.Data.Height,
					"block_hash", fmt.Sprintf("%.8X...", cp.Data.BlockHash[:4]),
					"timestamp", time.Unix(cp.Data.Timestamp, 0),
				)
			}
		}
	} else {
		// 도움말 표시
		fmt.Println("Usage:")
		flag.PrintDefaults()
	}
} 