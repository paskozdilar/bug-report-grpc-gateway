package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	stsv1 "github.com/paskozdilar/bug-report-grpc-gateway/gen/sts/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type implExampleServer struct {
	stsv1.UnsafeSTSServiceServer
}

func (*implExampleServer) Exchange(
	ctx context.Context,
	req *stsv1.ExchangeRequest,
) (*stsv1.ExchangeResponse, error) {
	log.Println("Exchange open")
	go func() {
		<-ctx.Done()
		log.Println("Exchange close")
	}()
	return &stsv1.ExchangeResponse{}, nil
}

func server() {
	server := grpc.NewServer()
	stsv1.RegisterSTSServiceServer(server, &implExampleServer{})

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err := server.Serve(l); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func gateway() {
	mux := runtime.NewServeMux()
	if err := stsv1.RegisterSTSServiceHandlerFromEndpoint(
		context.Background(),
		mux,
		"localhost:8080",
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	); err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}
	mux.HandlePath("POST", "/sts/exchange", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		r.ParseForm()
		audience := r.FormValue("audience")
		grantType := r.FormValue("grant_type")
		subjectToken := r.FormValue("subject_token")
		fmt.Println("Audience:", audience)
		fmt.Println("Grand type:", grantType)
		fmt.Println("Subject token:", subjectToken)
	})

	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("failed to serve gateway: %v", err)
	}
}

func httpclient() {
	time.Sleep(500 * time.Millisecond)
	log.Println("httpclient starting requests")

	{
		formData := url.Values{
			"audience":      {"https://example.com"},
			"grant_type":    {"urn:ietf:params:oauth:grant-type:token-exchange"},
			"subject_token": {"eyJ..."},
		}
		req, _ := http.NewRequest(http.MethodPost, "http://localhost:8081/sts/exchange", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		_, _ = (&http.Client{}).Do(req)
	}

	time.Sleep(1 * time.Second)
	log.Println("httpclient finished")
}

func main() {
	go server()
	go gateway()
	httpclient()
}
