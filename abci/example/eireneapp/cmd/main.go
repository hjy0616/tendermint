package main

import (
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/tendermint/tendermint/abci/example/eireneapp"
	"github.com/tendermint/tendermint/abci/server"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"
)

var (
	socketAddr string
	dbDir      string
	logLevel   string
)

func init() {
	flag.StringVar(&socketAddr, "socket", "tcp://127.0.0.1:26658", "ABCI server listen address")
	flag.StringVar(&dbDir, "db", "./data", "데이터베이스 디렉토리")
	flag.StringVar(&logLevel, "log", "info", "로그 레벨 (debug, info, error)")
	flag.Parse()
}

func main() {
	// 로거 초기화
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	
	// 로그 레벨 수동 설정
	switch logLevel {
	case "debug":
		logger = log.NewFilter(logger, log.AllowDebug())
	case "info":
		logger = log.NewFilter(logger, log.AllowInfo())
	case "error":
		logger = log.NewFilter(logger, log.AllowError())
	default:
		logger = log.NewFilter(logger, log.AllowInfo())
	}

	// 디렉토리 생성
	err := os.MkdirAll(dbDir, 0700)
	if err != nil {
		logger.Error("데이터베이스 디렉토리 생성 실패", "error", err)
		os.Exit(1)
	}

	// 데이터베이스 초기화
	db, err := tmdb.NewGoLevelDB("eirene", dbDir)
	if err != nil {
		logger.Error("데이터베이스 초기화 실패", "error", err)
		os.Exit(1)
	}

	// 애플리케이션 생성
	app := eireneapp.NewEireneApp(db)
	app.SetLogger(logger.With("module", "eirene"))

	// ABCI 서버 생성
	srv, err := server.NewServer(socketAddr, "socket", app)
	if err != nil {
		logger.Error("서버 생성 실패", "error", err)
		os.Exit(1)
	}

	// 정상 종료 처리
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	// 서버 시작
	err = srv.Start()
	if err != nil {
		logger.Error("서버 시작 실패", "error", err)
		os.Exit(1)
	}

	// 시작 로그
	logger.Info("ABCI 서버 시작됨", "address", socketAddr)
	absDbDir, _ := filepath.Abs(dbDir)
	logger.Info("데이터베이스", "path", absDbDir)

	// 종료 신호 대기
	<-done
	logger.Info("종료 신호 수신, 서버 중지 중...")

	// 서버 종료
	err = srv.Stop()
	if err != nil {
		logger.Error("서버 종료 중 오류 발생", "error", err)
		os.Exit(1)
	}

	logger.Info("서버가 정상적으로 종료되었습니다.")
} 