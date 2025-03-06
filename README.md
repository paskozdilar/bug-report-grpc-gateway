# bug-report-grpc-gateway

This is a bug report for the grpc-gateway project.

~Relevant issue: https://github.com/grpc-ecosystem/grpc-gateway/issues/5236~

**UPDATE: The HTTP hanging issue has since been fixed. This repository now
demonstrates an issue caused by this fix, related to WebSocket connections.**

To reproduce it yourself, run:

```bash
git clone https://github.com/paskozdilar/bug-report-grpc-gateway.git
cd bug-report-grpc-gateway
go mod tidy
go run main.go
```

Output:

```
2025/03/06 15:00:10 httpclient
2025/03/06 15:00:10 ServerStreamBroken open
2025/03/06 15:00:10 ServerStreamOK open
2025/03/06 15:00:11 httpclient end
2025/03/06 15:00:11 ServerStreamOK close
2025/03/06 15:00:11 ServerStreamBroken close
2025/03/06 15:00:12 wsclient
2025/03/06 15:00:12 ServerStreamOK open
2025/03/06 15:00:13 wsclient end
2025/03/06 15:00:13 ERROR: Failed to notify error to client: io: read/write on closed pipe
2025/03/06 15:00:13 ERROR: Failed to write response: io: read/write on closed pipe
2025/03/06 15:00:13 ServerStreamOK close
```
