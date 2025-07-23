package gateway

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/server"
)

func (g *Gateway) startStdioServer(ctx context.Context, mcpServer *server.MCPServer, stdin io.Reader, stdout io.Writer) error {
	return server.NewStdioServer(mcpServer).Listen(ctx, stdin, stdout)
}

func (g *Gateway) startSseServer(ctx context.Context, mcpServer *server.MCPServer, ln net.Listener) error {
	mux := http.NewServeMux()
	sseServer := server.NewSSEServer(mcpServer)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/sse", http.StatusTemporaryRedirect)
	})
	mux.Handle("/sse", sseServer.SSEHandler())
	mux.Handle("/message", sseServer.MessageHandler())
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
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/mcp", http.StatusTemporaryRedirect)
	})
	mux.Handle("/mcp", server.NewStreamableHTTPServer(mcpServer))
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

func (g *Gateway) startCentralStreamingServer(ctx context.Context, newMCPServer func() *server.MCPServer, ln net.Listener, configuration Configuration) error {
	var lock sync.Mutex
	handlersPerSelectionOfServers := map[string]*server.StreamableHTTPServer{}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/mcp", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		serverNames := r.Header.Get("x-mcp-servers")
		if len(serverNames) == 0 {
			_, _ = w.Write([]byte("No server names provided in the request header 'x-mcp-servers'"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		lock.Lock()
		handler := handlersPerSelectionOfServers[serverNames]
		if handler == nil {
			mcpServer := newMCPServer()
			if err := g.reloadConfiguration(ctx, mcpServer, configuration, parseServerNames(serverNames)); err != nil {
				lock.Unlock()
				_, _ = w.Write([]byte("Failed to reload configuration"))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			handler = server.NewStreamableHTTPServer(mcpServer)
			handlersPerSelectionOfServers[serverNames] = handler
		}
		lock.Unlock()

		handler.ServeHTTP(w, r)
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
	g.health.SetHealthy()
	return httpServer.Serve(ln)
}

func parseServerNames(serverNames string) []string {
	var names []string
	for _, name := range strings.Split(serverNames, ",") {
		names = append(names, strings.TrimSpace(name))
	}
	return names
}
