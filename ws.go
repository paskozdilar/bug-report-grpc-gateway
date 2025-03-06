package main

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Transparent WebSocket wrapper over ordinary HTTP handler.
func NewWrapper(h http.Handler) http.Handler {
	return &wrapper{
		h: h,
		u: websocket.Upgrader{},
	}
}

type wrapper struct {
	h http.Handler
	u websocket.Upgrader
}

func (w *wrapper) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	if strings.ToLower(r.Header.Get("Upgrade")) != "websocket" {
		w.h.ServeHTTP(wr, r)
		return
	}
	conn, err := w.u.Upgrade(wr, r, nil)
	if err != nil {
		http.Error(wr, "Error upgrading to websocket", http.StatusInternalServerError)
		return
	}
	w.ServeWS(conn, r)
}

// Forward WebSocket connection to the embedded HTTP handler.
func (w *wrapper) ServeWS(ws *websocket.Conn, r *http.Request) {
	// WebSocket connection / HTTP request pipes for passing data around
	wsReader, wsWriter := io.Pipe()
	httpReader, httpWriter := io.Pipe()

	// Create mock *http.Request and http.ResponseWriter for executing
	// ServeHTTP
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	url := r.URL.String()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, eofPipeReader{httpReader})
	if err != nil {
		return
	}
	fwdr := newResponseForwarder(wsWriter, r.Header)

	// Forward data from pipes into websocket
	go func() {
		socketRead(httpWriter, ws)
		r.Body.Close()
		wsWriter.Close()
		cancel()
	}()
	go func() {
		socketWrite(ws, wsReader)
		ws.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(5*time.Second))
		ws.Close()
	}()

	// Execute HTTP request
	w.h.ServeHTTP(fwdr, r)
	wsReader.Close()
}

// listen to websocket messages and write them to writer
func socketRead(w io.Writer, ws *websocket.Conn) {
	buf := bufio.NewWriter(w)
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			return
		}

		buf.Write(message)
		buf.WriteRune('\n')
		if err := buf.Flush(); err != nil {
			return
		}
	}
}

// listen to reader message and write them to websocket
func socketWrite(ws *websocket.Conn, r io.Reader) {
	buf := bufio.NewReader(r)
	for {
		s, err := buf.ReadString('\n')
		if err != nil {
			return
		}

		if err := ws.WriteMessage(websocket.TextMessage, []byte(s)); err != nil {
			return
		}
	}
}

// Implementation of http.ResponseWriter interface that redirects all response
// data to embedded *io.PipeWriter.
//
// Useful for passing custom writers to http.Handler.ServeHTTP() method.
type responseForwarder struct {
	*io.PipeWriter
	h http.Header
}

func newResponseForwarder(w *io.PipeWriter, h http.Header) http.ResponseWriter {
	return &responseForwarder{w, h}
}

func (rf *responseForwarder) Header() http.Header {
	return rf.h
}

func (rf *responseForwarder) WriteHeader(int) {
}

func (rf *responseForwarder) Flush() {
}

type eofPipeReader struct {
	*io.PipeReader
}

func (r eofPipeReader) Read(p []byte) (int, error) {
	n, err := r.PipeReader.Read(p)
	if err == io.ErrClosedPipe {
		err = io.EOF
	}
	return n, err
}
