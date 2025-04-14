package types

import (
	"bufio"
	"encoding/binary"
	"io"

	"github.com/gogo/protobuf/proto"
)

const (
	maxMsgSize = 104857600 // 100MB
)

// WriteMessage writes a varint length-delimited protobuf message.
func WriteMessage(msg proto.Message, w io.Writer) error {
	bz, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return encodeByteSlice(w, bz)
}

// ReadMessage reads a varint length-delimited protobuf message.
func ReadMessage(r io.Reader, msg proto.Message) error {
	return readProtoMsg(r, msg, maxMsgSize)
}

func readProtoMsg(r io.Reader, msg proto.Message, maxSize int) error {
	// binary.ReadVarint takes an io.ByteReader, eg. a bufio.Reader
	reader, ok := r.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}
	length64, err := binary.ReadVarint(reader)
	if err != nil {
		return err
	}
	length := int(length64)
	if length < 0 || length > maxSize {
		return io.ErrShortBuffer
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return err
	}
	return proto.Unmarshal(buf, msg)
}

//-----------------------------------------------------------------------
// NOTE: we copied wire.EncodeByteSlice from go-wire rather than keep
// go-wire as a dep

func encodeByteSlice(w io.Writer, bz []byte) (err error) {
	err = encodeVarint(w, int64(len(bz)))
	if err != nil {
		return
	}
	_, err = w.Write(bz)
	return
}

func encodeVarint(w io.Writer, i int64) (err error) {
	var buf [10]byte
	n := binary.PutVarint(buf[:], i)
	_, err = w.Write(buf[0:n])
	return
}

//----------------------------------------

func ToRequestEcho(message string) *Request {
	return &Request{
		Value: &Request_Echo{&RequestEcho{Message: message}},
	}
}

func ToRequestFlush() *Request {
	return &Request{
		Value: &Request_Flush{&RequestFlush{}},
	}
}

func ToRequestInfo(req RequestInfo) *Request {
	return &Request{
		Value: &Request_Info{&req},
	}
}

func ToRequestSetOption(req RequestSetOption) *Request {
	return &Request{
		Value: &Request_SetOption{&req},
	}
}

func ToRequestDeliverTx(req RequestDeliverTx) *Request {
	return &Request{
		Value: &Request_DeliverTx{&req},
	}
}

func ToRequestCheckTx(req RequestCheckTx) *Request {
	return &Request{
		Value: &Request_CheckTx{&req},
	}
}

func ToRequestCommit() *Request {
	return &Request{
		Value: &Request_Commit{&RequestCommit{}},
	}
}

func ToRequestQuery(req RequestQuery) *Request {
	return &Request{
		Value: &Request_Query{&req},
	}
}

func ToRequestInitChain(req RequestInitChain) *Request {
	return &Request{
		Value: &Request_InitChain{&req},
	}
}

func ToRequestBeginBlock(req RequestBeginBlock) *Request {
	return &Request{
		Value: &Request_BeginBlock{&req},
	}
}

func ToRequestEndBlock(req RequestEndBlock) *Request {
	return &Request{
		Value: &Request_EndBlock{&req},
	}
}

func ToRequestListSnapshots(req RequestListSnapshots) *Request {
	return &Request{
		Value: &Request_ListSnapshots{&req},
	}
}

func ToRequestOfferSnapshot(req RequestOfferSnapshot) *Request {
	return &Request{
		Value: &Request_OfferSnapshot{&req},
	}
}

func ToRequestLoadSnapshotChunk(req RequestLoadSnapshotChunk) *Request {
	return &Request{
		Value: &Request_LoadSnapshotChunk{&req},
	}
}

func ToRequestApplySnapshotChunk(req RequestApplySnapshotChunk) *Request {
	return &Request{
		Value: &Request_ApplySnapshotChunk{&req},
	}
}

//----------------------------------------

func ToResponseException(errStr string) *Response {
	return &Response{
		Value: &Response_Exception{&ResponseException{Error: errStr}},
	}
}

func ToResponseEcho(message string) *Response {
	return &Response{
		Value: &Response_Echo{&ResponseEcho{Message: message}},
	}
}

func ToResponseFlush() *Response {
	return &Response{
		Value: &Response_Flush{&ResponseFlush{}},
	}
}

func ToResponseInfo(res ResponseInfo) *Response {
	return &Response{
		Value: &Response_Info{&res},
	}
}

func ToResponseSetOption(res ResponseSetOption) *Response {
	return &Response{
		Value: &Response_SetOption{&res},
	}
}

func ToResponseDeliverTx(res ResponseDeliverTx) *Response {
	return &Response{
		Value: &Response_DeliverTx{&res},
	}
}

func ToResponseCheckTx(res ResponseCheckTx) *Response {
	return &Response{
		Value: &Response_CheckTx{&res},
	}
}

func ToResponseCommit(res ResponseCommit) *Response {
	return &Response{
		Value: &Response_Commit{&res},
	}
}

func ToResponseQuery(res ResponseQuery) *Response {
	return &Response{
		Value: &Response_Query{&res},
	}
}

func ToResponseInitChain(res ResponseInitChain) *Response {
	return &Response{
		Value: &Response_InitChain{&res},
	}
}

func ToResponseBeginBlock(res ResponseBeginBlock) *Response {
	return &Response{
		Value: &Response_BeginBlock{&res},
	}
}

func ToResponseEndBlock(res ResponseEndBlock) *Response {
	return &Response{
		Value: &Response_EndBlock{&res},
	}
}

func ToResponseListSnapshots(res ResponseListSnapshots) *Response {
	return &Response{
		Value: &Response_ListSnapshots{&res},
	}
}

func ToResponseOfferSnapshot(res ResponseOfferSnapshot) *Response {
	return &Response{
		Value: &Response_OfferSnapshot{&res},
	}
}

func ToResponseLoadSnapshotChunk(res ResponseLoadSnapshotChunk) *Response {
	return &Response{
		Value: &Response_LoadSnapshotChunk{&res},
	}
}

func ToResponseApplySnapshotChunk(res ResponseApplySnapshotChunk) *Response {
	return &Response{
		Value: &Response_ApplySnapshotChunk{&res},
	}
}

type RequestCheckTx struct {
	Tx   []byte      `protobuf:"bytes,1,opt,name=tx,proto3" json:"tx,omitempty"`
	Type CheckTxType `protobuf:"varint,2,opt,name=type,proto3,enum=tendermint.abci.CheckTxType" json:"type,omitempty"`
	// Ethereum specific fields
	IsEthereumTx bool   `protobuf:"varint,3,opt,name=is_ethereum_tx,json=isEthereumTx,proto3" json:"is_ethereum_tx,omitempty"`
	From         string `protobuf:"bytes,4,opt,name=from,proto3" json:"from,omitempty"`
	GasPrice     string `protobuf:"bytes,5,opt,name=gas_price,json=gasPrice,proto3" json:"gas_price,omitempty"`
	GasFeeCap    string `protobuf:"bytes,6,opt,name=gas_fee_cap,json=gasFeeCap,proto3" json:"gas_fee_cap,omitempty"`
	GasTipCap    string `protobuf:"bytes,7,opt,name=gas_tip_cap,json=gasTipCap,proto3" json:"gas_tip_cap,omitempty"`
}

type RequestDeliverTx struct {
	Tx []byte `protobuf:"bytes,1,opt,name=tx,proto3" json:"tx,omitempty"`
	// Ethereum specific fields
	IsEthereumTx bool   `protobuf:"varint,2,opt,name=is_ethereum_tx,json=isEthereumTx,proto3" json:"is_ethereum_tx,omitempty"`
	From         string `protobuf:"bytes,3,opt,name=from,proto3" json:"from,omitempty"`
	GasPrice     string `protobuf:"bytes,4,opt,name=gas_price,json=gasPrice,proto3" json:"gas_price,omitempty"`
	GasFeeCap    string `protobuf:"bytes,5,opt,name=gas_fee_cap,json=gasFeeCap,proto3" json:"gas_fee_cap,omitempty"`
	GasTipCap    string `protobuf:"bytes,6,opt,name=gas_tip_cap,json=gasTipCap,proto3" json:"gas_tip_cap,omitempty"`
}

// EthereumTxInfo 이더리움 트랜잭션 관련 정보를 담은 구조체
type EthereumTxInfo struct {
	IsEthereumTx bool   `json:"is_ethereum_tx"`
	From         string `json:"from"`
	GasPrice     string `json:"gas_price"`
	GasFeeCap    string `json:"gas_fee_cap"`
	GasTipCap    string `json:"gas_tip_cap"`
}

// CreateEthereumCheckTx 이더리움 트랜잭션을 위한 CheckTx 요청 생성
func CreateEthereumCheckTx(tx []byte, txType CheckTxType, ethInfo EthereumTxInfo) RequestCheckTx {
	return RequestCheckTx{
		Tx:   tx,
		Type: txType,
	}
}

// CreateEthereumDeliverTx 이더리움 트랜잭션을 위한 DeliverTx 요청 생성
func CreateEthereumDeliverTx(tx []byte) RequestDeliverTx {
	return RequestDeliverTx{
		Tx: tx,
	}
}

// ToRequestCheckTxWithEthInfo 이더리움 트랜잭션 정보를 포함한 CheckTx 요청 생성
func ToRequestCheckTxWithEthInfo(tx []byte, txType CheckTxType, ethInfo EthereumTxInfo) *Request {
	req := CreateEthereumCheckTx(tx, txType, ethInfo)
	return &Request{
		Value: &Request_CheckTx{&req},
	}
}

// ToRequestDeliverTxWithEthInfo 이더리움 트랜잭션 정보를 포함한 DeliverTx 요청 생성
func ToRequestDeliverTxWithEthInfo(tx []byte, ethInfo EthereumTxInfo) *Request {
	req := CreateEthereumDeliverTx(tx)
	return &Request{
		Value: &Request_DeliverTx{&req},
	}
}
