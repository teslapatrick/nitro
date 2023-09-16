package privacy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

// var db ICacheService
var currentWrapper *PrivacyWrapper

const EmptyHash string = ""

const EmptyAddress string = ""

type PrivacyWrapper struct {
	config *PrivacyConfig
	cache  ICacheService
}

func NewWrapper(config *PrivacyConfig) *PrivacyWrapper {
	// todo change cache service
	db, _ := NewBigCache(BigCacheConfigDefault)
	wrapper := &PrivacyWrapper{
		config: config,
		cache:  db,
	}
	// rewrite current wrapper
	currentWrapper = wrapper
	return wrapper
}

func CurrentWrapper() *PrivacyWrapper {
	return currentWrapper
}

type hashFunc func(...[]byte) common.Hash

type PrivacyResponseWriter struct {
	http.ResponseWriter
	buf      bytes.Buffer
	done     bool
	hash     hashFunc
	token    string
	hasToken bool
}

// WriteHeader implements http.ResponseWriter.WriteHeader
func (pw *PrivacyResponseWriter) WriteHeader(status int) {
	pw.ResponseWriter.WriteHeader(status)
}

// Write writes the data to the ResponseWriter.
func (pw *PrivacyResponseWriter) Write(b []byte) (int, error) {
	if pw.done {
		return 0, nil
	}
	return pw.buf.Write(b)
}

// RpcResponseMiddleware is a middleware that regenerate the data to the response.
func RpcResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if not enabled privacy, or have jwt authentication
		if !currentWrapper.config.Enable {
			next.ServeHTTP(w, r)
			return
		}

		startTime := time.Now()

		pw := &PrivacyResponseWriter{
			ResponseWriter: w,
			buf:            bytes.Buffer{},
			hasToken:       false,
			hash:           crypto.Keccak256Hash,
		}
		// check bearer token first
		pw.token, pw.hasToken = containsTokenHeader(r)
		var responseData []byte

		d, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		r.Body.Close()

		// rewrite the request body, weired to do it here
		r.Body = io.NopCloser(bytes.NewBuffer(d))

		// serve the request
		next.ServeHTTP(pw, r)

		// get the response data
		responseDataOri := pw.buf.Bytes()
		// unmarshal request message
		var reqMessage JsonrpcMessage
		_ = json.Unmarshal(d, &reqMessage)

		switch reqMessage.Method {
		case methodGetBalance():
			modifyBalanceMessage(&responseData, &responseDataOri, pw, &reqMessage)

		case methodGetTrasnaction():
			modifyTxMessage(&responseData, &responseDataOri, pw)

		case methodGetTransactionCount():
			modifyTxCountMessage(&responseData, &responseDataOri)

		case methodGetTrasnactionReceipt():
			modifyTxReceiptMessage(&responseData, &responseDataOri)

		case methodGetBlockByHash():
			modifyBlockByHashMessage(&responseData, &responseDataOri)

		case methodGetBlockByNumber():
			modifyBlockByNumberMessage(&responseData, &responseDataOri)

		default:
			responseData = responseDataOri
		}

		_, _ = pw.Write(responseData)
		// use gzip writer
		_, _ = w.Write(responseData)

		log.Trace("Privacy API Serve", "method", reqMessage.Method, "time", time.Since(startTime))
	})
}

func containsTokenHeader(r *http.Request) (strToken string, has bool) {
	strToken = r.Header.Get("X-ASN-Privacy-Token")
	has = len(strToken) > 0
	return
}

func methodGetBalance() string {
	return "eth_getBalance"
}

func methodGetTrasnaction() string {
	return "eth_getTransactionByHash"
}

func methodGetTrasnactionReceipt() string {
	return "eth_getTransactionReceipt"
}

func methodGetTransactionCount() string {
	return "eth_getTransactionCount"
}

func methodGetBlockByNumber() string {
	return "eth_getBlockByNumber"
}

func methodGetBlockByHash() string {
	return "eth_getBlockByHash"
}

func errorMessage(code int, message string) *jsonError {
	data := &jsonError{
		Code:    code,
		Message: message,
		Data:    nil,
	}
	return data
}

func modifyBalanceMessage(new *[]byte, ori *[]byte, pw *PrivacyResponseWriter, reqMessage *JsonrpcMessage) {
	var params []string
	_ = json.Unmarshal(reqMessage.Params, &params)
	token, _ := CurrentWrapper().cache.Get(context.Background(), common.HexToAddress(params[0]).String())
	// truly authorized
	log.Info("RpcResponseMiddleware", "token", pw.token, "hasToken", pw.hasToken, "cache token", string(token))
	if pw.hasToken && string(token) == pw.token {
		*new = *ori
		return
	}

	j := JsonrpcMessage{
		ID:      reqMessage.ID,
		Version: reqMessage.Version,
		Error:   errorMessage(-32802, "unauthorized to get balance"),
	}
	*new, _ = json.Marshal(j)
	return
}

// modifyTxMessage modifies the `data` params of response data of the eth_getTransaction method
func modifyTxMessage(new *[]byte, ori *[]byte, pw *PrivacyResponseWriter) {

	var resMessage JsonrpcMessage
	err := json.Unmarshal(*ori, &resMessage)
	if err != nil || resMessage.Error != nil {
		*new = *ori
		return
	}
	var tx RPCTransaction
	_ = json.Unmarshal(resMessage.Result, &tx)

	// if input is empty, return
	if tx.Input.String() == "0x" {
		*new = *ori
		return
	}

	// get token
	tokenFrom, _ := CurrentWrapper().cache.Get(context.Background(), tx.From.String())

	var tokenTo []byte
	if tx.To == nil {
		tokenTo = []byte(EmptyAddress)
	} else {
		tokenTo, _ = CurrentWrapper().cache.Get(context.Background(), tx.To.String())
	}

	// truly authorized
	if pw.hasToken && (string(tokenFrom) == (*pw).token || string(tokenTo) == pw.token) {
		*new = *ori
		return
	}

	// if not authorized, regenerate the response data
	input := pw.hash(tx.Input)
	tx.Input = input.Bytes()
	d, _ := json.Marshal(tx)

	*new, _ = json.Marshal(&JsonrpcMessage{
		ID:      resMessage.ID,
		Version: resMessage.Version,
		Result:  d,
	})
	return
}

func modifyTxCountMessage(new *[]byte, ori *[]byte) {
	*new = *ori
	return
}

func modifyTxReceiptMessage(new *[]byte, ori *[]byte) {
	*new = *ori
	return
}

func modifyBlockByHashMessage(new *[]byte, ori *[]byte) {
	*new = *ori
	var block types.Block
	_ = json.Unmarshal(*ori, &block)
	//block := types.NewBlock(block.Header(), block.Transactions(), block.Uncles(), block.Receipts())
	return
}

func modifyBlockByNumberMessage(new *[]byte, ori *[]byte) {
	*new = *ori
	return
}

func modifyBlock(b *Block) Block {
	return Block{}
}
