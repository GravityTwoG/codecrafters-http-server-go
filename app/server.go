package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/GravityTwoG/http-server-starter-go/pkg/http"
)

func main() {
	workdir := flag.String("directory", "~", "absolute path to the working directory")
	flag.Parse()
	fmt.Println("Workdir:", *workdir)

	fileServer := createFileServer(*workdir)

	router := func(ctx *http.HTTPContext) {
		req := ctx.Req

		switch {
		case req.Path == "/":
			index(ctx)
			return

		case strings.Contains(req.Path, "/echo/"):
			echo(ctx)
			return

		case strings.Contains(req.Path, "/user-agent"):
			userAgent(ctx)
			return

		case strings.Contains(req.Path, "/files/"):
			if req.Method == "GET" {
				fileServer.serveFiles(ctx)
			} else if req.Method == "POST" {
				fileServer.createFile(ctx)
			}
			return
		}

		res := http.HTTPResponse{
			StatusCode: 404,
			StatusText: "Not Found",
			Headers:    map[string]string{},
			Body:       []byte(""),
		}
		ctx.Respond(&res)
	}

	http.CreateServer(router)
}

func index(ctx *http.HTTPContext) {
	res := http.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Headers:    map[string]string{},
		Body:       []byte(""),
	}
	ctx.Respond(&res)
}

func echo(ctx *http.HTTPContext) {
	message := ctx.Req.Path[len("/echo/"):]

	var headers map[string]string = make(map[string]string)
	headers["Content-Type"] = "text/plain"

	encoding, acceptsEncoding := ctx.Req.Headers["Accept-Encoding"]
	if acceptsEncoding && encoding == "gzip" {
		headers["Content-Encoding"] = "gzip"
	}

	res := http.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Headers:    headers,
		Body:       []byte(message),
	}
	ctx.Respond(&res)
}

func userAgent(ctx *http.HTTPContext) {
	res := http.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Headers: map[string]string{
			"Content-Type": "text/plain",
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
			StatusCode: 404,
			StatusText: "Not Found",
			Headers:    map[string]string{},
			Body:       []byte{},
		}
		ctx.Respond(&res)
		return
	}
	res := http.HTTPResponse{
		StatusCode: 200,
		StatusText: "OK",
		Headers: map[string]string{
			"Content-Type": "application/octet-stream",
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
			StatusCode: 500,
			StatusText: "Internal Server Error",
			Headers:    map[string]string{},
			Body:       []byte(""),
		}
		ctx.Respond(&res)
	}

	res := http.HTTPResponse{
		StatusCode: 201,
		StatusText: "OK",
		Headers:    map[string]string{},
		Body:       []byte(""),
	}
	ctx.Respond(&res)
}
