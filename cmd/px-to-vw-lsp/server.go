package main

import (
	"context"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"io"
	"os"
)

func StartServer(logger *zap.Logger) {
	stream := &readWriteCloser{os.Stdin, os.Stdout}
	conn := jsonrpc2.NewConn(jsonrpc2.NewStream(stream))

	handler, ctx, err := NewHandler(
		context.Background(),
		protocol.ServerDispatcher(conn, logger),
		logger,
	)
	if err != nil {
		logger.Sugar().Fatalf("init handler error: %v", err)
	}

	conn.Go(ctx, protocol.ServerHandler(handler, jsonrpc2.MethodNotFoundHandler))
	<-conn.Done()
}

type readWriteCloser struct {
	reader io.ReadCloser
	writer io.WriteCloser
}

func (r *readWriteCloser) Read(b []byte) (int, error)  { return r.reader.Read(b) }
func (r *readWriteCloser) Write(b []byte) (int, error) { return r.writer.Write(b) }
func (r *readWriteCloser) Close() error                { return multierr.Append(r.reader.Close(), r.writer.Close()) }
