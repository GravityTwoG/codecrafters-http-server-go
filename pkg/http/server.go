package http

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

func CreateServer(handleFunc func(req *HTTPContext)) {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	wg := new(sync.WaitGroup)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			break
		}

		wg.Add(1)
		go func() {
			req, err := readRequest(conn)
			if err != nil {
				fmt.Println("Error reading request: ", err.Error())
				return
			}
			fmt.Printf("Method: \"%s\" Path: \"%s\"\n", req.Method, req.Path)
			for key, value := range req.Headers {
				fmt.Printf("Header: %s: %s\n", key, value)
			}

			ctx := &HTTPContext{conn: conn, Req: req}
			handleFunc(ctx)
			defer conn.Close()
			defer wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println("Server closed connection")
}

type HTTPContext struct {
	conn net.Conn
	Req  *HTTPRequest
}

func (ctx *HTTPContext) Respond(res *HTTPResponse) error {
	return ctx.writeResponse(res)
}

func (ctx *HTTPContext) writeResponse(res *HTTPResponse) error {
	res.Version = "HTTP/1.1"

	fmt.Printf("Responding with HTTP/1.1 %d %s\n", res.StatusCode, res.StatusText)
	fmt.Printf("Resp Headers: %+v\n", res.Headers)
	fmt.Printf("Resp Body: %s\n", res.Body)

	encoding, acceptsEncoding := ctx.Req.Headers["Accept-Encoding"]
	if acceptsEncoding && strings.Contains(encoding, "gzip") {
		res.Headers["Content-Encoding"] = "gzip"
	}

	_, err := ctx.conn.Write([]byte(res.Version + " " + fmt.Sprintf("%d", res.StatusCode) + " " + res.StatusText + "\r\n"))
	if err != nil {
		return err
	}

	_, hasContentLength := res.Headers["Content-Length"]
	if len(res.Body) > 0 && !hasContentLength {
		res.Headers["Content-Length"] = fmt.Sprintf("%d", len(res.Body))
	}

	for key, value := range res.Headers {
		_, err = ctx.conn.Write([]byte(key + ": " + value + "\r\n"))
		if err != nil {
			return err
		}
	}

	_, err = ctx.conn.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	_, err = ctx.conn.Write([]byte(res.Body))
	return err
}

type HTTPRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
}

type HTTPResponse struct {
	Version    string
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       []byte
}
