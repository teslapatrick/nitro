package privacy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

// var db ICacheService
var currentWrapper *PrivacyWrapper

const EmptyInput string = "0x"
const EmptyHashOrAddress string = ""

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
		if currentWrapper == nil || !currentWrapper.config.Enable {
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
		_ = r.Body.Close()

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

		case methodGetTransaction():
			modifyTxMessage(&responseData, &responseDataOri, pw)

		case methodGetTransactionCount():
			modifyTxCountMessage(&responseData, &responseDataOri, pw, &reqMessage)

		case methodGetTransactionReceipt():
			modifyTxReceiptMessage(&responseData, &responseDataOri)

		case methodGetBlockByHash():
			modifyBlockByHashMessage(&responseData, &responseDataOri, pw, &reqMessage)

		case methodGetBlockByNumber():
			modifyBlockByNumberMessage(&responseData, &responseDataOri, pw, &reqMessage)

		default:
			responseData = responseDataOri
		}

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

func methodGetTransaction() string {
	return "eth_getTransactionByHash"
}

func methodGetTransactionReceipt() string {
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

func parseAddressFromReq(reqMessage *JsonrpcMessage) (string, error) {
	var params []string
	err := json.Unmarshal(reqMessage.Params, &params)
	if err != nil {
		return "", err
	}
	// make sure the address is well formatted
	return common.HexToAddress(params[0]).String(), nil
}

func parseHashAndFlagFromReq(reqMessage *JsonrpcMessage) (string, bool) {
	var params []interface{}
	_ = json.Unmarshal(reqMessage.Params, &params)
	return params[0].(string), params[1].(bool)
}

func (pw *PrivacyResponseWriter) authorized(token []byte) bool {
	return pw.hasToken && string(token) == pw.token
}

func getAddressToken(addr string) ([]byte, error) {
	return CurrentWrapper().cache.Get(context.Background(), addr)
}

func modifyBalanceMessage(new *[]byte, ori *[]byte, pw *PrivacyResponseWriter, reqMessage *JsonrpcMessage) {
	addr, _ := parseAddressFromReq(reqMessage)
	token, _ := getAddressToken(addr)
	// truly authorized
	log.Info("RpcResponseMiddleware", "token", pw.token, "hasToken", pw.hasToken, "cache token", string(token))
	if pw.authorized(token) {
		*new = *ori
	} else {
		*new, _ = json.Marshal(JsonrpcMessage{
			ID:      reqMessage.ID,
			Version: reqMessage.Version,
			Error:   errorMessage(-32802, "unauthorized to get balance"),
		})
	}
}

func txWithInputHash(hash hashFunc, tx *RPCTransaction) {
	(*tx).Input = hash(tx.Input).Bytes()
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
	if tx.Input.String() == EmptyInput {
		*new = *ori
		return
	}

	// get token, `From` and `To`
	tokenFrom, _ := getAddressToken(tx.From.String())
	var tokenTo []byte
	if tx.To == nil {
		tokenTo = []byte(EmptyHashOrAddress)
	} else {
		tokenTo, _ = getAddressToken(tx.To.String())
	}

	// check authorization
	if pw.authorized(tokenFrom) || pw.authorized(tokenTo) {
		*new = *ori
	} else {
		// if not authorized, regenerate the response data
		txWithInputHash(pw.hash, &tx)
		d, _ := json.Marshal(tx)
		*new, _ = json.Marshal(&JsonrpcMessage{
			ID:      resMessage.ID,
			Version: resMessage.Version,
			Result:  d,
		})
	}
}

func modifyTxCountMessage(new *[]byte, ori *[]byte, pw *PrivacyResponseWriter, reqMessage *JsonrpcMessage) {
	addr, _ := parseAddressFromReq(reqMessage)
	token, _ := getAddressToken(addr)
	// truly authorized
	if pw.authorized(token) {
		*new = *ori
	} else {
		*new, _ = json.Marshal(JsonrpcMessage{
			ID:      reqMessage.ID,
			Version: reqMessage.Version,
			Error:   errorMessage(-32803, "unauthorized to get transaction count"),
		})
	}
}

func modifyTxReceiptMessage(new *[]byte, ori *[]byte) {
	*new = *ori
}

func modifyBlockByHashMessage(new *[]byte, ori *[]byte, pw *PrivacyResponseWriter, reqMessage *JsonrpcMessage) {
	var resMessage JsonrpcMessage
	err := json.Unmarshal(*ori, &resMessage)
	if err != nil || resMessage.Error != nil {
		*new = *ori
		return
	}

	var block RPCBlock
	_ = json.Unmarshal(resMessage.Result, &block)

	_, fullTx := parseHashAndFlagFromReq(reqMessage)
	txs := make(Transactions, 0, len(block.Transactions))
	if !fullTx {
		*new = *ori
	} else {
		for _, tx := range block.Transactions {
			// if input is empty, return
			if tx.Input.String() == EmptyInput {
				txs = append(txs, tx)
				continue
			}
			// get token, `From` and `To`
			tokenFrom, _ := getAddressToken(tx.From.String())
			var tokenTo []byte
			if tx.To == nil {
				tokenTo = []byte(EmptyHashOrAddress)
			} else {
				tokenTo, _ = getAddressToken(tx.To.String())
			}

			// check authorization
			if !pw.authorized(tokenFrom) && !pw.authorized(tokenTo) {
				// if not authorized, regenerate the response data
				txWithInputHash(pw.hash, &tx)
			}
			txs = append(txs, tx)
		}
		block.Transactions = txs
		d, _ := json.Marshal(block)
		*new, _ = json.Marshal(&JsonrpcMessage{
			ID:      resMessage.ID,
			Version: resMessage.Version,
			Result:  d,
		})
	}
}

func modifyBlockByNumberMessage(new *[]byte, ori *[]byte, pw *PrivacyResponseWriter, reqMessage *JsonrpcMessage) {
	modifyBlockByHashMessage(new, ori, pw, reqMessage)
}
