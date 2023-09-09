package privacy

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type PrivacyWrapper struct {
	config *PrivacyConfig
	cache  IPrivacyCache
}

func NewWrapper(config *PrivacyConfig) *PrivacyWrapper {
	return &PrivacyWrapper{
		config: config,
		cache:  nil,
	}
}

type PrivacyResponseWriter struct {
	http.ResponseWriter
	buf      bytes.Buffer
	done     bool
	hasToken bool
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

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		startTime := time.Now()
		pw := &PrivacyResponseWriter{
			ResponseWriter: w,
			buf:            bytes.Buffer{},
		}

		var writer io.Writer = pw
		var responseData []byte

		// if not encoding with gzip,
		if containsGzipHeader(r) {
			// del the accept header
			r.Header.Del("Accept-Encoding")
			next.ServeHTTP(pw, r)
			if pw.Status == 0 {
				pw.Status = http.StatusOK
			}

			pw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
			// new gzip writer
			writer = gzip.NewWriter(pw.ResponseWriter)
			defer writer.(*gzip.Writer).Close()
			// marshal data
			responseData, _ = io.ReadAll(&pw.buf)

		} else {
			next.ServeHTTP(pw, r)
			if pw.Status == 0 {
				pw.Status = http.StatusOK
			}
			responseData = pw.buf.Bytes()
		}

		writer.Write(responseData)

		log.Printf("%s %s %d %s", r.Method, r.RequestURI, pw.Status, time.Since(startTime))
	})
}

func containsGzipHeader(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}

//type PrivacyWrapper struct {
