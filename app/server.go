package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/pkg/http"
)

func main() {
	workdir := flag.String("directory", "~", "absolute path to the working directory")

	flag.Parse()
	fmt.Println("Workdir:", *workdir)

	fileServer := createFileServer(*workdir)

	handleFunc := func(ctx *http.HTTPContext) {
		// Write HTTP response
		req := ctx.Req
		if req.Path == "/" {
			index(ctx)
			return
		} else if strings.Contains(req.Path, "/echo/") {
			echo(ctx)
			return
		} else if strings.Contains(req.Path, "/user-agent") {
			userAgent(ctx)
			return
		} else if strings.Contains(req.Path, "/files/") {
			if req.Method == "GET" {
				fileServer.serveFiles(ctx)
				return
			}
			if req.Method == "POST" {
				fileServer.createFile(ctx)
				return
			}
		}

		res := http.HTTPResponse{
			Version:    "HTTP/1.1",
			StatusCode: 404,
			StatusText: "Not Found",
			Headers:    map[string]string{},
			Body:       []byte(""),
		}
		ctx.Respond(&res)
	}

	http.CreateServer(handleFunc)
}

func index(ctx *http.HTTPContext) {
	res := http.HTTPResponse{
		Version:    "HTTP/1.1",
		StatusCode: 200,
		StatusText: "OK",
		Headers:    map[string]string{},
		Body:       []byte(""),
	}
	ctx.Respond(&res)
}

func echo(ctx *http.HTTPContext) {
	message := ctx.Req.Path[len("/echo/"):]

	res := http.HTTPResponse{
		Version:    "HTTP/1.1",
		StatusCode: 200,
		StatusText: "OK",
		Headers: map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(message)),
		},
		Body: []byte(message),
	}
	ctx.Respond(&res)
}

func userAgent(ctx *http.HTTPContext) {
	res := http.HTTPResponse{
		Version:    "HTTP/1.1",
		StatusCode: 200,
		StatusText: "OK",
		Headers: map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(ctx.Req.Headers["User-Agent"])),
		},
		Body: []byte(ctx.Req.Headers["User-Agent"]),
	}
	ctx.Respond(&res)
}

type FileServer struct {
	workdir string
}

func createFileServer(workdir string) *FileServer {
	return &FileServer{workdir: workdir}
}

func (f *FileServer) serveFiles(ctx *http.HTTPContext) {
	filename := ctx.Req.Path[len("/files/"):]

	data, err := os.ReadFile(path.Join(f.workdir, filename))
	if err != nil {
		res := http.HTTPResponse{
			Version:    "HTTP/1.1",
			StatusCode: 404,
			StatusText: "Not Found",
			Headers:    map[string]string{},
			Body:       []byte(""),
		}
		ctx.Respond(&res)
	}
	res := http.HTTPResponse{
		Version:    "HTTP/1.1",
		StatusCode: 200,
		StatusText: "OK",
		Headers: map[string]string{
			"Content-Type":   "application/octet-stream",
			"Content-Length": fmt.Sprintf("%d", len(data)),
		},
		Body: data,
	}
	ctx.Respond(&res)
}

func (f *FileServer) createFile(ctx *http.HTTPContext) {
	filename := ctx.Req.Path[len("/files/"):]

	err := os.WriteFile(path.Join(f.workdir, filename), []byte(ctx.Req.Body), 0644)
	if err != nil {
		res := http.HTTPResponse{
			Version:    "HTTP/1.1",
			StatusCode: 500,
			StatusText: "Internal Server Error",
			Headers:    map[string]string{},
			Body:       []byte(""),
		}
		ctx.Respond(&res)
	}

	res := http.HTTPResponse{
		Version:    "HTTP/1.1",
		StatusCode: 201,
		StatusText: "OK",
		Headers:    map[string]string{},
		Body:       []byte(""),
	}
	ctx.Respond(&res)
}
