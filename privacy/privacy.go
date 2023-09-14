package privacy

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
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
	hasToken bool
	address  string
	token    string
	hash     hashFunc
	Status   int
}

// WriteHeader implements http.ResponseWriter.WriteHeader
func (pw *PrivacyResponseWriter) WriteHeader(code int) {
	pw.ResponseWriter.WriteHeader(code)
	pw.Status = code
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
		// if not enabled privacy
		if !currentWrapper.config.Enable {
			next.ServeHTTP(w, r)
			return
		}

		// truly enabled privacy
		startTime := time.Now()
		pw := &PrivacyResponseWriter{
			ResponseWriter: w,
			buf:            bytes.Buffer{},
			hasToken:       false,
			hash:           crypto.Keccak256Hash, // todo
		}

		var writer io.Writer = pw
		var responseData []byte

		d, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		r.Body.Close()

		// rewrite the request body, weired to do it here
		r.Body = io.NopCloser(bytes.NewBuffer(d))

		// check if the header requires the gzip encoding
		if containsGzipHeader(r) {
			r.Header.Del("Accept-Encoding")
			pw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
			// new gzip writer
			writer = gzip.NewWriter(pw.ResponseWriter)
			defer writer.(*gzip.Writer).Close()
		}
		next.ServeHTTP(pw, r)
		// get the response data
		responseDataOri := pw.buf.Bytes()

		var reqMessage JsonrpcMessage
		_ = json.Unmarshal(d, &reqMessage)

		// check bearer token first
		pw.token, pw.hasToken = containsTokenHeader(r)
		switch reqMessage.Method {
		case methodGetBalance():
			var params []string
			_ = json.Unmarshal(reqMessage.Params, &params)
			token, _ := CurrentWrapper().cache.Get(context.Background(), params[0])
			// truly authorized
			if pw.hasToken && string(token) == pw.token {
				responseData = responseDataOri
				break
			}

			j := JsonrpcMessage{
				ID:      reqMessage.ID,
				Version: reqMessage.Version,
				Error:   errorMessage(-32802, "unauthorized to get balance"),
			}
			responseData, _ = json.Marshal(j)

		case methodGetTrasnaction():

			_ = modifyTxMessage(&responseData, &responseDataOri, pw)

		case methodGetTransactionCount():
			_ = modifyTxCountMessage(&responseData, &responseDataOri)
			break
		case methodGetTrasnactionReceipt():
			err := modifyTxReceiptMessage(&responseData, &responseDataOri)
			if err != nil {
				// responseData = errorMessage(-32805, "cannot get transaction receipt")
				break
			}

		case methodGetBlockByHash():
			err := modifyBlockByHashMessage(&responseData, &responseDataOri)
			if err != nil {
				// responseData = errorMessage(-32806, "cannot get block")
				break
			}

		case methodGetBlockByNumber():
			err := modifyBlockByNumberMessage(&responseData, &responseDataOri)
			if err != nil {
				// responseData = errorMessage(-32807, "cannot get block")
				break
			}

		default:
			responseData = responseDataOri
		}
		//}

		writer.Write(responseData)

		log.Info("Privacy API Serve", "method", reqMessage.Method, "time", time.Since(startTime))
	})
}

func containsGzipHeader(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
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

// modifyTxMessage modifies the `data` params of response data of the eth_getTransaction method
func modifyTxMessage(new *[]byte, ori *[]byte, pw *PrivacyResponseWriter) error {

	var resMessage JsonrpcMessage
	err := json.Unmarshal(*ori, &resMessage)
	if err != nil || resMessage.Error != nil {
		*new = *ori
		return nil
	}
	var tx RPCTransaction
	_ = json.Unmarshal(resMessage.Result, &tx)

	// if input is empty, return
	if tx.Input.String() == "0x" {
		*new = *ori
		return nil
	}

	// get token
	tokenFrom, _ := CurrentWrapper().cache.Get(context.Background(), tx.From.String())
	var tokenTo []byte
	if tx.To == nil {
		tokenTo = []byte(EmptyAddress)
	} else {
		tokenTo, _ = CurrentWrapper().cache.Get(context.Background(), tx.To.String())
	}

	log.Trace("modifyTxMessage", "tokenFrom", string(tokenFrom), "tokenTo", string(tokenTo))

	// truly authorized
	if pw.hasToken && (string(tokenFrom) == pw.token || string(tokenTo) == pw.token) {
		*new = *ori
		return nil
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
	return nil
}

func modifyTxCountMessage(new *[]byte, ori *[]byte) error {
	*new = *ori
	return nil
}

func modifyTxReceiptMessage(new *[]byte, ori *[]byte) error {
	*new = *ori
	return nil
}

