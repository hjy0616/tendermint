package eireneapp

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/tendermint/tendermint/abci/example/eireneapp/validators"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"
)

// EireneApp는 Ethereum 호환 ABCI 애플리케이션입니다.
// 이 구현은 Tendermint와 Ethereum 간의 브릿지 역할을 합니다.
type EireneApp struct {
	types.BaseApplication

	// 애플리케이션 상태
	db           tmdb.DB                // 상태 저장용 데이터베이스
	state        *AppState              // 애플리케이션 상태
	valManager   *validators.Manager    // 검증자 관리자
	currentBlock *BlockInfo             // 현재 처리 중인 블록 정보

	// 트랜잭션 처리
	txIndex    int64                    // 현재 블록 내 트랜잭션 인덱스
	txResults  []*TxResult              // 현재 블록의 트랜잭션 결과
	txMap      map[string]bool          // 트랜잭션 중복 검사용

	// 동시성 제어
	mtx        sync.Mutex               // 상태 액세스 뮤텍스

	// 로깅
	logger     log.Logger               // 로거
}

// NewEireneApp는 새로운 EireneApp 인스턴스를 생성합니다.
func NewEireneApp(db tmdb.DB) *EireneApp {
	state := loadState(db)
	if state == nil {
		state = &AppState{
			Height:     0,
			AppHash:    make([]byte, 32), // 초기 해시는 0으로 채워진 32바이트
			Validators: []Validator{},
			StateRoot:  make([]byte, 32), // 초기 상태 루트
		}
	}

	logger := log.NewNopLogger()
	valManager := validators.NewManager()

	return &EireneApp{
		db:         db,
		state:      state,
		valManager: valManager,
		logger:     logger,
		txMap:      make(map[string]bool),
	}
}

// SetLogger는 애플리케이션의 로거를 설정합니다.
func (app *EireneApp) SetLogger(l log.Logger) {
	app.logger = l
	app.valManager.SetLogger(l)
}

// 상태 로드
func loadState(db tmdb.DB) *AppState {
	stateBytes, err := db.Get([]byte("state"))
	if err != nil || stateBytes == nil {
		return nil
	}

	state := &AppState{}
	err = cdc.UnmarshalJSON(stateBytes, state)
	if err != nil {
		panic(err)
	}

	return state
}

// 상태 저장
func saveState(db tmdb.DB, state *AppState) {
	stateBytes, err := cdc.MarshalJSON(state)
	if err != nil {
		panic(err)
	}

	err = db.Set([]byte("state"), stateBytes)
	if err != nil {
		panic(err)
	}
}

// Info는 현재 애플리케이션 상태에 대한 정보를 반환합니다.
func (app *EireneApp) Info(req types.RequestInfo) types.ResponseInfo {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return types.ResponseInfo{
		Data:             fmt.Sprintf("{\"height\":%d,\"state_root\":\"%X\"}", app.state.Height, app.state.StateRoot),
		Version:          "v1.0.0",
		AppVersion:       1,
		LastBlockHeight:  app.state.Height,
		LastBlockAppHash: app.state.AppHash,
	}
}

// InitChain은 체인 초기화 시 호출됩니다.
func (app *EireneApp) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	app.logger.Info("InitChain", "chain_id", req.ChainId, "validators", len(req.Validators))

	// 초기 검증자 설정
	app.valManager.InitValidators(req.Validators)

	// genesis 상태 처리
	if len(req.AppStateBytes) > 0 {
		var genesisState map[string]json.RawMessage
		err := json.Unmarshal(req.AppStateBytes, &genesisState)
		if err != nil {
			app.logger.Error("Failed to parse genesis state", "error", err)
		} else {
			// 여기서 genesis 상태를 처리 (예: 초기 계정, 컨트랙트 등)
			app.logger.Info("Genesis state parsed", "keys", len(genesisState))
			
			// 초기 상태 루트 계산 (이더리움 호환)
			app.state.StateRoot = calculateStateRoot(genesisState)
		}
	}

	// 상태 초기화
	app.state.Validators = convertValidators(req.Validators)
	saveState(app.db, app.state)

	return types.ResponseInitChain{
		ConsensusParams: req.ConsensusParams,
		Validators:      req.Validators,
	}
}

// BeginBlock은 새 블록 처리가 시작될 때 호출됩니다.
func (app *EireneApp) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	app.logger.Info("BeginBlock", "height", req.Header.Height)

	// 현재 블록 정보 설정
	app.currentBlock = &BlockInfo{
		Height:     req.Header.Height,
		Time:       req.Header.Time.Unix(),
		Hash:       req.Hash,
		ProposerID: req.Header.ProposerAddress,
		TxCount:    0,
	}

	// 트랜잭션 처리 초기화
	app.txIndex = 0
	app.txResults = make([]*TxResult, 0)
	app.txMap = make(map[string]bool)

	// 바이잔틴 검증자 처리
	if len(req.ByzantineValidators) > 0 {
		app.logger.Info("Byzantine validators detected", "count", len(req.ByzantineValidators))
		for _, v := range req.ByzantineValidators {
			app.logger.Info("Byzantine validator", "address", fmt.Sprintf("%X", v.Validator.Address))
			// 여기서 슬래싱 등의 처리 가능
		}
	}

	return types.ResponseBeginBlock{}
}

// CheckTx는 트랜잭션을 메모리 풀에 추가하기 전에 검증합니다.
func (app *EireneApp) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
	app.logger.Debug("CheckTx", "tx_size", len(req.Tx))

	// 트랜잭션 해시 계산
	txHash := calculateTxHash(req.Tx)
	txHashStr := string(txHash)

	// 중복 트랜잭션 검사
	app.mtx.Lock()
	if app.txMap[txHashStr] {
		app.mtx.Unlock()
		return types.ResponseCheckTx{
			Code: 2,
			Log:  "duplicate transaction",
		}
	}
	app.mtx.Unlock()

	// 간단한 형식 검증
	if len(req.Tx) < 2 {
		return types.ResponseCheckTx{
			Code: 1,
			Log:  "tx too short",
		}
	}

	// 트랜잭션 파싱 및 기본 검증
	if isValidTx(req.Tx) {
		// 유효한 트랜잭션은 우선순위 계산 (예: gas price 기반)
		priority := calculateTxPriority(req.Tx)
		
		return types.ResponseCheckTx{
			Code:      0,
			Log:       "tx accepted",
			Priority:  priority,
			GasWanted: estimateGasWanted(req.Tx),
		}
	}

	return types.ResponseCheckTx{
		Code: 1,
		Log:  "invalid transaction format",
	}
}

// DeliverTx는 트랜잭션을 실행하고 상태를 변경합니다.
func (app *EireneApp) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	txHash := calculateTxHash(req.Tx)
	txHashStr := string(txHash)

	app.logger.Debug("DeliverTx", "tx_hash", fmt.Sprintf("%X", txHash), "index", app.txIndex)
	app.txIndex++
	app.currentBlock.TxCount++

	// 중복 트랜잭션 검사
	if app.txMap[txHashStr] {
		return types.ResponseDeliverTx{
			Code: 2,
			Log:  "duplicate transaction",
		}
	}
	app.txMap[txHashStr] = true

	// 트랜잭션 실행
	txData, status, gasUsed, events, err := executeTx(req.Tx, app.state)
	
	if err != nil {
		app.logger.Error("Failed to execute transaction", "error", err, "tx_hash", fmt.Sprintf("%X", txHash))
		result := &TxResult{
			Hash:   txHash,
			Status: 0, // 실패
			Data:   []byte{},
			Events: []types.Event{
				{
					Type: "eirene_tx_error",
					Attributes: []types.EventAttribute{
						{Key: []byte("tx_index"), Value: []byte(fmt.Sprintf("%d", app.txIndex-1)), Index: true},
						{Key: []byte("error"), Value: []byte(err.Error()), Index: true},
					},
				},
			},
			Gas:    gasUsed,
		}
		app.txResults = append(app.txResults, result)

		return types.ResponseDeliverTx{
			Code:   1,
			Log:    err.Error(),
			Events: result.Events,
			GasUsed: int64(gasUsed),
		}
	}

	// 성공적인 트랜잭션 처리
	result := &TxResult{
		Hash:   txHash,
		Status: status,
		Data:   txData,
		Events: events,
		Gas:    gasUsed,
	}
	app.txResults = append(app.txResults, result)

	return types.ResponseDeliverTx{
		Code:      0,
		Data:      txData,
		Events:    events,
		GasUsed:   int64(gasUsed),
		GasWanted: int64(estimateGasWanted(req.Tx)),
	}
}

// EndBlock은 블록 처리가 끝날 때 호출됩니다.
func (app *EireneApp) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	app.logger.Info("EndBlock", "height", req.Height, "num_txs", app.currentBlock.TxCount)

	// 블록 보상 분배 (구현 필요)
	distributeBlockRewards(app.state, app.currentBlock, app.txResults)

	// 검증자 업데이트 계산
	validatorUpdates := app.valManager.GetPendingUpdates()

	// 검증자 변경 사항 기록
	if len(validatorUpdates) > 0 {
		app.logger.Info("Validator updates", "count", len(validatorUpdates))
		app.state.Validators = updateValidators(app.state.Validators, validatorUpdates)
	}

	// 블록 이벤트 생성
	events := []types.Event{
		{
			Type: "eirene_block_complete",
			Attributes: []types.EventAttribute{
				{Key: []byte("height"), Value: []byte(fmt.Sprintf("%d", req.Height)), Index: true},
				{Key: []byte("num_txs"), Value: []byte(fmt.Sprintf("%d", app.currentBlock.TxCount)), Index: true},
				{Key: []byte("proposer"), Value: []byte(fmt.Sprintf("%X", app.currentBlock.ProposerID)), Index: true},
			},
		},
	}

	return types.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
		Events:           events,
	}
}

// Commit은 현재 상태를 커밋하고 애플리케이션 해시를 반환합니다.
func (app *EireneApp) Commit() types.ResponseCommit {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	app.logger.Info("Commit", "height", app.state.Height+1)

	// 상태 루트 계산 (이더리움 호환)
	stateRoot := calculateStateRootFromResults(app.state, app.txResults)
	app.state.StateRoot = stateRoot

	// 높이 증가
	app.state.Height++

	// 애플리케이션 해시 계산
	app.state.AppHash = calculateAppHash(app.state)

	// 상태 저장
	saveState(app.db, app.state)

	// 추가 데이터 저장 (트랜잭션 결과, 블록 정보 등)
	storeBlockData(app.db, app.currentBlock, app.txResults)

	return types.ResponseCommit{
		Data: app.state.AppHash,
	}
}

// Query는 애플리케이션 상태를 쿼리합니다.
func (app *EireneApp) Query(req types.RequestQuery) types.ResponseQuery {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	app.logger.Debug("Query", "path", req.Path, "data", string(req.Data))

	// 쿼리 경로에 따라 다른 처리
	switch req.Path {
	case "state":
		stateBytes, err := cdc.MarshalJSON(app.state)
		if err != nil {
			return types.ResponseQuery{
				Code: 1,
				Log:  err.Error(),
			}
		}
		return types.ResponseQuery{
			Code:  0,
			Value: stateBytes,
		}
	case "validators":
		validators, err := cdc.MarshalJSON(app.state.Validators)
		if err != nil {
			return types.ResponseQuery{
				Code: 1,
				Log:  err.Error(),
			}
		}
		return types.ResponseQuery{
			Code:  0,
			Value: validators,
		}
	case "tx":
		// 트랜잭션 조회
		if len(req.Data) == 0 {
			return types.ResponseQuery{
				Code: 1,
				Log:  "tx hash required",
			}
		}
		
		txInfo := getTxInfo(app.db, req.Data)
		if txInfo == nil {
			return types.ResponseQuery{
				Code: 1,
				Log:  "tx not found",
			}
		}
		
		txInfoBytes, err := cdc.MarshalJSON(txInfo)
		if err != nil {
			return types.ResponseQuery{
				Code: 1,
				Log:  err.Error(),
			}
		}
		
		return types.ResponseQuery{
			Code:  0,
			Value: txInfoBytes,
		}
	case "account":
		// 계정 정보 조회
		if len(req.Data) == 0 {
			return types.ResponseQuery{
				Code: 1,
				Log:  "account address required",
			}
		}
		
		accountInfo := getAccountInfo(app.db, req.Data)
		if accountInfo == nil {
			return types.ResponseQuery{
				Code: 1,
				Log:  "account not found",
			}
		}
		
		accountBytes, err := cdc.MarshalJSON(accountInfo)
		if err != nil {
			return types.ResponseQuery{
				Code: 1,
				Log:  err.Error(),
			}
		}
		
		return types.ResponseQuery{
			Code:  0,
			Value: accountBytes,
		}
	default:
		return types.ResponseQuery{
			Code: 1,
			Log:  "unknown query path",
		}
	}
}

// 헬퍼 함수들

// 애플리케이션 해시 계산
func calculateAppHash(state *AppState) []byte {
	// 상태의 각 요소를 결합하여 해시 계산
	heightBytes := []byte(fmt.Sprintf("%d", state.Height))
	validatorsBytes, _ := cdc.MarshalJSON(state.Validators)
	
	// SHA-256 해시 계산
	hash := sha256.New()
	hash.Write(heightBytes)
	hash.Write(state.StateRoot)
	hash.Write(validatorsBytes)
	
	return hash.Sum(nil)
}

// 트랜잭션 해시 계산
func calculateTxHash(tx []byte) []byte {
	hash := sha256.New()
	hash.Write(tx)
	return hash.Sum(nil)
}

// 트랜잭션 유효성 검사
func isValidTx(tx []byte) bool {
	// 실제 구현에서는 이더리움 트랜잭션 형식 검증
	return len(tx) >= 2
}

// 트랜잭션 우선순위 계산
func calculateTxPriority(tx []byte) int64 {
	// 실제 구현에서는 gas price 기반 우선순위 계산
	return 1
}

// 예상 가스 사용량 계산
func estimateGasWanted(tx []byte) int64 {
	// 실제 구현에서는 트랜잭션 유형에 따른 가스 계산
	return 21000 // 기본 이더리움 트랜잭션 가스
}

// 트랜잭션 실행
func executeTx(tx []byte, state *AppState) ([]byte, uint32, uint64, []types.Event, error) {
	// 실제 구현에서는 EVM 실행 로직 구현
	// 간단한 예시
	data := []byte("executed")
	status := uint32(1) // 성공
	gasUsed := uint64(21000)
	events := []types.Event{
		{
			Type: "eirene_tx",
			Attributes: []types.EventAttribute{
				{Key: []byte("success"), Value: []byte("true"), Index: true},
			},
		},
	}
	
	return data, status, gasUsed, events, nil
}

// 블록 보상 분배
func distributeBlockRewards(state *AppState, block *BlockInfo, txResults []*TxResult) {
	// 실제 구현에서는 검증자 보상 로직 구현
}

// 초기 상태 루트 계산
func calculateStateRoot(genesisState map[string]json.RawMessage) []byte {
	// 실제 구현에서는 이더리움 스타일 상태 트리의 루트 계산
	// 간단한 예시
	hash := sha256.New()
	for k, v := range genesisState {
		hash.Write([]byte(k))
		hash.Write(v)
	}
	return hash.Sum(nil)
}

// 트랜잭션 결과로부터 상태 루트 계산
func calculateStateRootFromResults(state *AppState, txResults []*TxResult) []byte {
	// 실제 구현에서는 트랜잭션 실행 후 이더리움 상태 트리의 루트 계산
	// 간단한 예시
	hash := sha256.New()
	hash.Write(state.StateRoot)
	for _, result := range txResults {
		hash.Write(result.Hash)
		hash.Write(result.Data)
	}
	return hash.Sum(nil)
}

// 블록 데이터 저장
func storeBlockData(db tmdb.DB, block *BlockInfo, txResults []*TxResult) {
	// 실제 구현에서는 블록 및 트랜잭션 데이터 저장
	blockKey := fmt.Sprintf("block:%d", block.Height)
	blockBytes, _ := cdc.MarshalJSON(block)
	db.Set([]byte(blockKey), blockBytes)
	
	// 트랜잭션 저장
	for i, result := range txResults {
		txKey := fmt.Sprintf("tx:%X", result.Hash)
		txBytes, _ := cdc.MarshalJSON(result)
		db.Set([]byte(txKey), txBytes)
		
		// 블록-트랜잭션 인덱스
		idxKey := fmt.Sprintf("block_tx:%d:%d", block.Height, i)
		db.Set([]byte(idxKey), result.Hash)
	}
}

// 트랜잭션 정보 조회
func getTxInfo(db tmdb.DB, txHash []byte) interface{} {
	txKey := fmt.Sprintf("tx:%X", txHash)
	txBytes, err := db.Get([]byte(txKey))
	if err != nil || txBytes == nil {
		return nil
	}
	
	var txInfo TxResult
	err = cdc.UnmarshalJSON(txBytes, &txInfo)
	if err != nil {
		return nil
	}
	
	return &txInfo
}

// 계정 정보 조회
func getAccountInfo(db tmdb.DB, address []byte) interface{} {
	// 실제 구현에서는 이더리움 계정 정보 조회
	// 간단한 예시
	return nil
}

// Validator 타입을 내부 형식으로 변환
func convertValidators(updates []types.ValidatorUpdate) []Validator {
	validators := make([]Validator, len(updates))
	for i, update := range updates {
		validators[i] = Validator{
			PubKey: update.PubKey.GetEd25519(),
			Power:  update.Power,
		}
	}
	return validators
}

// 검증자 목록 업데이트
func updateValidators(current []Validator, updates []types.ValidatorUpdate) []Validator {
	// 기존 검증자 맵 생성
	validatorMap := make(map[string]Validator)
	for _, v := range current {
		key := string(v.PubKey)
		validatorMap[key] = v
	}
	
	// 업데이트 적용
	for _, update := range updates {
		pubKey := update.PubKey.GetEd25519()
		key := string(pubKey)
		
		if update.Power == 0 {
			// 검증자 제거
			delete(validatorMap, key)
		} else {
			// 검증자 추가/업데이트
			validatorMap[key] = Validator{
				PubKey: pubKey,
				Power:  update.Power,
			}
		}
	}
	
	// 맵을 슬라이스로 변환
	newValidators := make([]Validator, 0, len(validatorMap))
	for _, v := range validatorMap {
		newValidators = append(newValidators, v)
	}
	
	return newValidators
}

// JSON 코덱 구현
type jsonCodec struct{}

func (jc *jsonCodec) MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jc *jsonCodec) UnmarshalJSON(bz []byte, ptr interface{}) error {
	return json.Unmarshal(bz, ptr)
}

// Codec 초기화
var cdc = makeCodec()

func makeCodec() Codec {
	return &jsonCodec{}
} 