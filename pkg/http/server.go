package http

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
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
	err := writeResponse(ctx.conn, res)
	if err != nil {
		return err
	}
	return nil
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

func readRequest(conn net.Conn) (*HTTPRequest, error) {
	reader := bufio.NewReader(conn)

	// Read start line
	startLineStr, err := reader.ReadString('\r')
	if err != nil {
		return &HTTPRequest{}, err
	}
	_, err = reader.ReadByte()
	if err != nil {
		return &HTTPRequest{}, err
	}

	// Parse start line
	startLine, err := parseStartLine(startLineStr)
	if err != nil {
		return &HTTPRequest{}, err
	}

	headers, err := parseHeaders(reader)
	if err != nil {
		return &HTTPRequest{}, err
	}

	body := []byte{}
	if startLine.Method == "POST" {
		body, err = parseBody(headers, reader)
		if err != nil {
			return &HTTPRequest{}, err
		}
		headers["Content-Length"] = fmt.Sprintf("%d", len(body))
	}

	return &HTTPRequest{
		Method:  startLine.Method,
		Path:    startLine.Path,
		Headers: headers,
		Body:    body,
	}, nil
}

type HTTPStartLine struct {
	Method  string
	Path    string
	Version string
}

func parseStartLine(startLine string) (*HTTPStartLine, error) {

	var method string
	for i := 0; i < len(startLine); i++ {
		if startLine[i] == ' ' {
			method = startLine[0:i]
			break
		}
	}

	var version string
	for i := len(startLine) - 1; i >= 0; i-- {
		if startLine[i] == ' ' {
			version = startLine[i+1:]
			break
		}
	}

	path := startLine[len(method)+1 : len(startLine)-len(version)-1]

	return &HTTPStartLine{
		Method:  method,
		Path:    path,
		Version: version,
	}, nil
}

func parseHeaders(reader *bufio.Reader) (map[string]string, error) {
	headers := map[string]string{}
	for {
		line, err := reader.ReadString('\r')
		if err != nil {
			return headers, err
		}
		ret, err := reader.ReadByte()
		if err != nil {
			return headers, err
		}
		if ret != '\n' {
			return headers, fmt.Errorf("invalid header: %s", line)
		}
		line = strings.Trim(line, "\r\n")

		if line == "\r" || line == "\n" || line == "" {
			break
		}
		keyValue := strings.Split(line, ": ")
		if len(keyValue) != 2 {
			return headers, fmt.Errorf("invalid header: %s", line)
		}
		headers[keyValue[0]] = keyValue[1]
	}
	return headers, nil
}

func parseBody(headers map[string]string, reader *bufio.Reader) ([]byte, error) {
	var body []byte
	contentLength, err := strconv.Atoi(headers["Content-Length"])
	if err != nil {
		return body, err
	}
	body = make([]byte, contentLength)
	n, err := reader.Read(body)
	if err != nil {
		return nil, err
	}
	if n != contentLength {
		return nil, fmt.Errorf("invalid body length: %d", n)
	}

	return body, nil
}

func writeResponse(conn net.Conn, response *HTTPResponse) error {
	_, err := conn.Write([]byte(response.Version + " " + fmt.Sprintf("%d", response.StatusCode) + " " + response.StatusText + "\r\n"))
	if err != nil {
		return err
	}
	for key, value := range response.Headers {
		_, err = conn.Write([]byte(key + ": " + value + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err = conn.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte(response.Body))
	return err
}
