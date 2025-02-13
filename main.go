package main

import (
	"context"
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

func (*implExampleServer) ServerStreamOK(
	req *emptypb.Empty,
	stream grpc.ServerStreamingServer[example.ExampleResponse],
) error {
	log.Println("ServerStreamOK open")
	<-stream.Context().Done()
	log.Println("ServerStreamOK close")
	return nil
}

func (*implExampleServer) ServerStreamBroken(
	req *emptypb.Empty,
	stream grpc.ServerStreamingServer[example.ExampleResponse],
) error {
	log.Println("ServerStreamBroken open")
	<-stream.Context().Done()
	log.Println("ServerStreamBroken close")
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

	time.Sleep(time.Second)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"http://localhost:8081/example/v1/ServerStreamOK",
		strings.NewReader("{}"),
	)
	if err != nil {
		log.Println("New request ServerStreamOK:", err)
	}
	go (&http.Client{}).Do(req)

	req, err = http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"http://localhost:8081/example/v1/ServerStreamBroken",
		strings.NewReader("{}"),
	)
	if err != nil {
		log.Println("Request ServerStreamBroken:", err)
	}
	go (&http.Client{}).Do(req)

	time.Sleep(time.Second)
}

func main() {
	go server()
	go gateway()
	go client()
	select {}
}
