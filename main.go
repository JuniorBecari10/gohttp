package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	CRLF = "\r\n"
	DoubleCRLF = CRLF + CRLF
 
	GetMethod = "GET"
	HeadMethod = "HEAD"
	PostMethod = "POST"
	PutMethod = "PUT"
	DeleteMethod = "DELETE"
	ConnectMethod = "CONNECT"
	OptionsMethod = "OPTIONS"
	TraceMethod = "TRACE"
	PatchMethod = "PATCH"
)

type Server struct {
	address string
	port int
}

type Request struct {
	method string
	path string
	version string

	headers map[string]string
	body string
}

type Response struct {
	status string
	contentType string

	content string
}

// ---

func NewServer(address string, port int) Server {
	return Server{
		address,
		port,
	}
}

func NewResponse(status, contentType, content string) Response {
	return Response{
		status,
		contentType,
		content,
	}
}

// ---

func (s *Server) Run() error {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.address, s.port))

	if err != nil {
		return err
	}

	for {
		conn, err := listen.Accept()

		if err != nil {
			return err
		}

		buffer := make([]byte, 2048)
		_, err = conn.Read(buffer)

		if err != nil {
			return err
		}

		request := strings.Trim(string(buffer), "\x00")
		req, err := parseRequest(request)

		if err != nil {
			return err
		}

		fmt.Println(req)

		res := NewResponse("200 OK", "text/html", "<h1>Hello!</h1>")
		conn.Write([]byte(res.construct()))
		conn.Close()
	}
}

// ---

func parseRequest(request string) (Request, error) {
	req := Request{ headers: make(map[string]string) }
	lines := strings.Split(request, CRLF)

	for i, line := range lines {
		if i == 0 {
			tokens := strings.Split(line, " ")

			if len(tokens) != 3 {
				return Request{}, fmt.Errorf("not enough tokens in request header; need 3, got %d: '%s'", len(tokens), line)
			}

			req.method = tokens[0]
			req.path = tokens[1]
			
			versionTks := strings.Split(tokens[2], "/")

			if len(versionTks) != 2 {
				return Request{}, fmt.Errorf("not enough tokens in HTTP version header; need 2, got %d: '%s'", len(tokens), line)
			}

			req.version = versionTks[1]
		} else {
			if line == "" {
				break
			}

			// split by ':' or ': '?
			headers := strings.Split(line, ": ")
	
			if len(headers) != 2 {
				return Request{}, fmt.Errorf("not enough tokens in header; need 2, got %d: '%s'", len(headers), line)
			}

			req.headers[headers[0]] = headers[1]
		}
	}

	sections := strings.Split(strings.TrimSpace(request), DoubleCRLF)

	if len(sections) >= 2 {
		req.body = strings.TrimSpace(sections[1])
	}

	// if not, there's no body

	return req, nil
}

func (r *Response) construct() string {
	return fmt.Sprintf(
		"HTTP/1.1 %s%sContent-Type: %s%sContent-Length: %d%s%s",
		r.status,
		CRLF,
		r.contentType,
		CRLF,
		len(r.content),
		DoubleCRLF,
		r.content,
	)
}

// ---

func main() {
	server := NewServer("localhost", 8080)

	fmt.Println("Listening on port 8080")
	err := server.Run()

	if err != nil {
		log.Fatal(err)
	}
}
