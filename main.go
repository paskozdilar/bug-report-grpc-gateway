package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/paskozdilar/bug-report-grpc-gateway/example"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type implExampleServer struct {
	example.UnsafeExampleServiceServer
}

func (*implExampleServer) UnaryBody(
	ctx context.Context,
	req *emptypb.Empty,
) (*example.ExampleResponse, error) {
	fmt.Println("UnaryBody open")
	<-ctx.Done()
	fmt.Println("UnaryBody close")
	return &example.ExampleResponse{}, nil
}

func (*implExampleServer) UnaryNoBody(
	ctx context.Context,
	req *emptypb.Empty,
) (*example.ExampleResponse, error) {
	fmt.Println("UnaryNoBody open")
	<-ctx.Done()
	fmt.Println("UnaryNoBody close")
	return &example.ExampleResponse{}, nil
}

func (*implExampleServer) ServerStreamBody(
	req *emptypb.Empty,
	stream grpc.ServerStreamingServer[example.ExampleResponse],
) error {
	fmt.Println("ServerStreamBody open")
	<-stream.Context().Done()
	fmt.Println("ServerStreamBody close")
	return nil
}

func (*implExampleServer) ServerStreamNoBody(
	req *emptypb.Empty,
	stream grpc.ServerStreamingServer[example.ExampleResponse],
) error {
	fmt.Println("ServerStreamNoBody open")
	<-stream.Context().Done()
	fmt.Println("ServerStreamNoBody close")
	return nil
}

func server() {
	server := grpc.NewServer()
	example.RegisterExampleServiceServer(server, &implExampleServer{})

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
	if err := example.RegisterExampleServiceHandlerFromEndpoint(
		context.Background(),
		mux,
		"localhost:8080",
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	); err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func client() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		conn, err := net.Dial("tcp", "localhost:8081")
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	fireRequest := func(name string, body io.Reader) error {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			fmt.Sprintf(
				"http://localhost:8081/example/v1/%s",
				name,
			),
			body,
		)
		if err != nil {
			fmt.Printf("New request %s: %v\n", name, err)
		}
		go func() {
			fmt.Printf("requesting: %s\n", name)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Printf("request failed: %s: %v\n", name, err)
				return
			}
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("request failed: %s: %v\n", name, resp.Status)
				return
			}
			fmt.Printf("request success: %s\n", name)
		}()
		return nil
	}

	fmt.Println("> Running invalid requests:")
	fireRequest("UnaryBody", strings.NewReader("{}"+strings.Repeat(".", 511)))
	fireRequest("UnaryNoBody", strings.NewReader("."))
	fireRequest("ServerStreamBody", strings.NewReader("{}"+strings.Repeat(".", 511)))
	fireRequest("ServerStreamNoBody", strings.NewReader("."))
	time.Sleep(time.Second)
	cancel()

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("> Running valid requests:")
	fireRequest("UnaryBody", strings.NewReader("{}"))
	fireRequest("UnaryNoBody", nil)
	fireRequest("ServerStreamBody", strings.NewReader("{}"))
	fireRequest("ServerStreamNoBody", nil)
	time.Sleep(time.Second)
	cancel()
}

func main() {
	go server()
	go gateway()
	go client()
	select {}
}
