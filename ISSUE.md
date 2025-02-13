### What happened?

Since v2.26.2 (PR #5240), generated gateway code includes `io.Copy(io.Discard, req.Body)` for endpoints without a `body:` annotation. This drains the HTTP request body before downstream handlers can read it.

For endpoints that use `application/x-www-form-urlencoded` POST data with a custom handler (registered via `ServeMux.HandlePath` or similar), `req.ParseForm()` finds an empty body because it was already discarded. All form fields are silently lost.

### How to reproduce

1. Define a gRPC service with an HTTP binding that uses POST but has no `body:` annotation:

```protobuf
rpc Exchange(ExchangeRequest) returns (ExchangeResponse) {
  option (google.api.http) = {
    post: "/sts/exchange"
  };
}
```

2. Generate the gateway code with protoc-gen-grpc-gateway >= v2.26.2

3. Register a custom handler that reads form data:

```go
mux.HandlePath("POST", "/sts/exchange", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
    r.ParseForm()
    audience := r.FormValue("audience")       // empty!
    grantType := r.FormValue("grant_type")    // empty!
    subjectToken := r.FormValue("subject_token") // empty!
})
```

4. Send a form-encoded POST:

```bash
curl -X POST https://example.com/sts/exchange \
  -d "grant_type=urn:ietf:params:oauth:grant-type:token-exchange" \
  -d "audience=https://example.com" \
  -d "subject_token=eyJ..."
```

All form values are empty because the generated code drained the body.

### What did you expect?

Endpoints without a `body:` annotation that receive POST requests should not have their body discarded. The body drain should only apply when `body: ""` is explicitly set (indicating the endpoint intentionally ignores the body).

### Versions

- protoc-gen-grpc-gateway: v2.26.2+ (works on v2.22.0)
- Introduced by: #5240

### Related issues

- #5326 — v2.26.2 breaks grpc-websocket-proxy (same root cause)
- #5667 — Panic from request drain when req.Body is nil

### Suggested fix

Only emit `io.Copy(io.Discard, req.Body)` when the HTTP binding explicitly sets `body: ""`, not when the `body` field is absent. Absent `body` should preserve the existing pre-v2.26.2 behavior of leaving the body untouched for custom handlers.
