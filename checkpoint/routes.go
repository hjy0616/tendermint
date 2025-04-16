package checkpoint

import (
	"encoding/json"
	"net/http"

	"github.com/tendermint/tendermint/libs/log"
)

// RegisterHTTPHandlers는 HTTP 핸들러 함수들을 등록합니다.
func RegisterHTTPHandlers(
	mux *http.ServeMux,
	handler *RPCHandler,
	logger log.Logger,
) {
	if logger != nil {
		handler.SetLogger(logger)
	}

	// 체크포인트 조회 (특정 번호)
	mux.HandleFunc("/checkpoint", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		number := r.URL.Query().Get("number")
		if number == "" {
			http.Error(w, "Missing parameter 'number'", http.StatusBadRequest)
			return
		}

		result, err := handler.FetchCheckpoint(nil, number)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondWithJSON(w, result)
	})

	// 최신 체크포인트 조회
	mux.HandleFunc("/checkpoint/latest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		result, err := handler.FetchLatestCheckpoint(nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondWithJSON(w, result)
	})

	// 체크포인트 개수 조회
	mux.HandleFunc("/checkpoint/count", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		result, err := handler.FetchCheckpointCount(nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondWithJSON(w, result)
	})

	// 체크포인트 생성
	mux.HandleFunc("/checkpoint/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		height := r.URL.Query().Get("height")
		if height == "" {
			http.Error(w, "Missing parameter 'height'", http.StatusBadRequest)
			return
		}

		result, err := handler.CreateCheckpoint(nil, height)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondWithJSON(w, result)
	})
}

// JSON 응답 전송 헬퍼 함수
func respondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
