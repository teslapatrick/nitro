package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/offchainlabs/nitro/privacy"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	var api privacy.IPrivacyAPI
	api = privacy.NewPrivacyAPI(privacy.NewWrapper(&privacy.PrivacyRPCConfigDefault))
	rpcAPI := []rpc.API{{
		Namespace: "privacy",
		Version:   "1.0",
		Service:   api,
		Public:    true,
	}}

	srv := rpc.NewServer()
	err := node.RegisterApis(rpcAPI, []string{"privacy"}, srv)
	if err != nil {
		log.Fatal(err)
	}
	handler := node.NewHTTPHandlerStack(srv, []string{}, []string{}, nil)
	privacyHandler := handler
	//
	s, addr, err := node.StartHTTPEndpoint("127.0.0.1:8080", rpc.DefaultHTTPTimeouts, privacyHandler)
	if err != nil {
		log.Fatal(err)
	}
	extapiURL := fmt.Sprintf("http://%v/", addr)
	fmt.Println("HTTP endpoint opened", "url", extapiURL)
	defer func() {
		s.Shutdown(context.Background())
	}()

	abortChan := make(chan os.Signal, 1)
	signal.Notify(abortChan, os.Interrupt)

	sig := <-abortChan
	log.Println("Exiting...", "signal", sig)
}

// LoggingMiddleware 是一个用于日志记录的 middleware
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		//lrw := &LoggingResponseWriter{
		//	ResponseWriter: w,
		//	status:         http.StatusOK, // default status
		//}

		lrw := &privacy.PrivacyResponseWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(lrw, r)

		duration := time.Now().Sub(startTime)
		log.Printf("%s %s %d %s", r.Method, r.RequestURI, lrw.Status, duration)
	})
}
