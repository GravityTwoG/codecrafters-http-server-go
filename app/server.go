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

	handleFunc := func(ctx *http.HTTPContext) {
		// Write HTTP response
		req := ctx.Req
		if req.Path == "/" {
			res := http.HTTPResponse{
				Version:    "HTTP/1.1",
				StatusCode: 200,
				StatusText: "OK",
				Headers:    map[string]string{},
				Body:       []byte(""),
			}
			ctx.Respond(&res)
			return
		} else if strings.Contains(req.Path, "/echo/") {
			message := req.Path[len("/echo/"):]

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
			return
		} else if strings.Contains(req.Path, "/user-agent") {
			res := http.HTTPResponse{
				Version:    "HTTP/1.1",
				StatusCode: 200,
				StatusText: "OK",
				Headers: map[string]string{
					"Content-Type":   "text/plain",
					"Content-Length": fmt.Sprintf("%d", len(req.Headers["User-Agent"])-1),
				},
				Body: []byte(req.Headers["User-Agent"]),
			}
			ctx.Respond(&res)
			return
		} else if strings.Contains(req.Path, "/files/") {
			filename := req.Path[len("/files/"):]

			data, err := os.ReadFile(path.Join(*workdir, filename))
			if err != nil {
				res := http.HTTPResponse{
					Version:    "HTTP/1.1",
					StatusCode: 404,
					StatusText: "Not Found",
					Headers:    map[string]string{},
					Body:       []byte(""),
				}
				ctx.Respond(&res)
				return
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
			return
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
