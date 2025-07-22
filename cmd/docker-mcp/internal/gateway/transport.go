package gateway

import (
	"context"
	"io"
	"net"
	"net/http"

	"github.com/mark3labs/mcp-go/server"
)

func (g *Gateway) startStdioServer(ctx context.Context, mcpServer *server.MCPServer, stdin io.Reader, stdout io.Writer) error {
	return server.NewStdioServer(mcpServer).Listen(ctx, stdin, stdout)
}

func (g *Gateway) startSseServer(ctx context.Context, mcpServer *server.MCPServer, ln net.Listener) error {
	mux := http.NewServeMux()
	sseServer := server.NewSSEServer(mcpServer)
	mux.Handle("/sse", sseServer.SSEHandler())
	mux.Handle("/message", sseServer.MessageHandler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/sse", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		if g.health.IsHealthy() {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	})
	httpServer := &http.Server{
		Handler: mux,
	}
	go func() {
		<-ctx.Done()
		ln.Close()
	}()
	return httpServer.Serve(ln)
}

func (g *Gateway) startStreamingServer(ctx context.Context, mcpServer *server.MCPServer, ln net.Listener) error {
	mux := http.NewServeMux()
	mux.Handle("/mcp", server.NewStreamableHTTPServer(mcpServer))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/mcp", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		if g.health.IsHealthy() {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	})
	httpServer := &http.Server{
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		ln.Close()
	}()
	return httpServer.Serve(ln)
}
