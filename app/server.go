package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Run TCP server
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	req, err := readRequest(conn)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		os.Exit(1)
	}

	fmt.Printf("Method: \"%s\" Path: \"%s\"\n", req.Method, req.Path)

	// Write HTTP response
	if strings.Contains(req.Path, "/echo") {
		message := req.Path[len("/echo/"):]

		res := HTTPResponse{
			Version:    "HTTP/1.1",
			StatusCode: 200,
			StatusText: "OK",
			Headers: map[string]string{
				"Content-Type":   "text/plain",
				"Content-Length": fmt.Sprintf("%d", len(message)),
			},
			Body: message,
		}
		writeResponse(conn, res)
	} else {
		res := HTTPResponse{
			Version:    "HTTP/1.1",
			StatusCode: 200,
			StatusText: "OK",
			Headers:    map[string]string{},
			Body:       "",
		}
		writeResponse(conn, res)
	}

	err = conn.Close()
	if err != nil {
		fmt.Println("Error closing connection: ", err.Error())
		os.Exit(1)
	}

	fmt.Println("Server closed connection")
}

type HTTPRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
}

type HTTPResponse struct {
	Version    string
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       string
}

func readRequest(conn net.Conn) (*HTTPRequest, error) {
	reader := bufio.NewReader(conn)

	// Read start line
	startLineStr, err := reader.ReadString('\r')
	if err != nil {
		return &HTTPRequest{}, err
	}

	// Parse start line
	startLine, err := parseStartLine(startLineStr)
	if err != nil {
		return &HTTPRequest{}, err
	}

	return &HTTPRequest{
		Method:  startLine.Method,
		Path:    startLine.Path,
		Headers: map[string]string{},
		Body:    "",
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

func writeResponse(conn net.Conn, response HTTPResponse) {
	conn.Write([]byte(response.Version + " " + fmt.Sprintf("%d", response.StatusCode) + " " + response.StatusText + "\r\n"))
	for key, value := range response.Headers {
		conn.Write([]byte(key + ": " + value + "\r\n"))
	}
	conn.Write([]byte("\r\n"))
	conn.Write([]byte(response.Body))
}
