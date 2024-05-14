package http

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func readRequest(conn net.Conn) (*HTTPRequest, error) {
	reader := bufio.NewReader(conn)

	// Parse start line
	startLine, err := parseStartLine(reader)
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

func parseStartLine(reader *bufio.Reader) (*HTTPStartLine, error) {
	startLine, err := reader.ReadString('\r')
	if err != nil {
		return nil, err
	}
	startLine = strings.Trim(startLine, "\r")
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
		if line == "\n\r" {
			reader.ReadByte()
			break
		}
		line = strings.Trim(line, "\n")
		line = strings.Trim(line, "\r")
		keyValue := strings.Split(line, ": ")
		if len(keyValue) != 2 {
			return headers, fmt.Errorf("invalid header: %d: %s", len(line), line)
		}
		headers[keyValue[0]] = keyValue[1]
	}
	return headers, nil
}

func parseBody(headers map[string]string, reader *bufio.Reader) ([]byte, error) {
	var body []byte
	contentLength, err := strconv.Atoi(headers["Content-Length"])
	if err != nil {
		return nil, err
	}
	fmt.Printf("Content-Length: %d\n", contentLength)
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
