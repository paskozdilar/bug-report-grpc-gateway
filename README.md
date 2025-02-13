# bug-report-grpc-gateway

This is a bug report for the grpc-gateway project.

To reproduce it yourself, run:

```bash
git clone https://github.com/paskozdilar/bug-report-grpc-gateway.git
cd bug-report-grpc-gateway
go mod tidy
go run main.go
```

Output:

```
ServerStreamOK open
ServerStreamOK close
ServerStreamBroken open
```

## Cause

The issue only occurs when using `google.protobuf.Empty` as request in a
ServerStreamMethod, and `body` is not set in `google.http.api` annotation:

```proto3
rpc ServerStreamBroken (google.protobuf.Empty) returns (stream ExampleResponse) {
  option (google.api.http) = {
    post: "/example/v1/ServerStreamBroken";
  };
}
```
