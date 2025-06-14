package pkg

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/valyala/fasthttp"
)

type ProxyServer struct {
	allowedHosts map[string]bool
}

func NewProxyServer(allowedHosts string) *ProxyServer {
	allowed := map[string]bool{}
	for host := range strings.SplitSeq(allowedHosts, ",") {
		allowed[strings.TrimSpace(host)] = true
	}

	fmt.Println("Allowed hosts:", allowed)

	return &ProxyServer{
		allowedHosts: allowed,
	}
}

func (p *ProxyServer) Run(ctx context.Context, ln net.Listener) error {
	server := &fasthttp.Server{
		Handler:            p.handleRequest,
		MaxConnsPerIP:      10000,
		MaxRequestsPerConn: 10000,
	}

	go func() {
		<-ctx.Done()
		server.Shutdown()
	}()

	return server.Serve(ln)
}

func (p *ProxyServer) handleRequest(ctx *fasthttp.RequestCtx) {
	if !p.allowedHosts[string(ctx.Host())] {
		ctx.Response.SetStatusCode(http.StatusForbidden)
		return
	}

	if string(ctx.Method()) == http.MethodConnect {
		p.handleTunneling(ctx)
	} else {
		p.handleHTTP(ctx)
	}
}

func (p *ProxyServer) handleTunneling(ctx *fasthttp.RequestCtx) {
	destinationConn, err := net.Dial("tcp", string(ctx.Host()))
	if err != nil {
		ctx.Error("Failed to connect to destination", http.StatusServiceUnavailable)
		return
	}

	ctx.SetStatusCode(http.StatusOK)
	ctx.Response.SetBodyRaw(nil)
	ctx.Hijack(func(clientConn net.Conn) {
		defer clientConn.Close()
		defer destinationConn.Close()

		go func() {
			_, _ = io.Copy(destinationConn, clientConn)
		}()
		_, _ = io.Copy(clientConn, destinationConn)
	})
}

func (p *ProxyServer) handleHTTP(ctx *fasthttp.RequestCtx) {
	client := &fasthttp.Client{
		Dial: fasthttp.Dial,
	}

	if err := client.Do(&ctx.Request, &ctx.Response); err != nil {
		ctx.Error("Failed to process request: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
