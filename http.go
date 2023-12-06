package gohttp

import (
	"fmt"
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

type HttpCallback func(*HttpRequest, *HttpResponse)

type HttpServer struct {
	address string
	port uint16

	requests map[HttpPath]HttpCallback
	notFoundHandler HttpCallback
}

type HttpRequest struct {
	method string
	path string
	version string

	headers map[string]string
	body string
}

type HttpResponse struct {
	status string
	contentType string

	content string
}

type HttpPath struct {
	path string
	method string
}

// --- Constructors ---

func NewServer(address string, port uint16) HttpServer {
	return HttpServer{
		address,
		port,
		map[HttpPath]HttpCallback {},
		nil,
	}
}

// Private, the user doesn't need to use it
func newResponse(status, contentType, content string) HttpResponse {
	return HttpResponse{
		status,
		contentType,
		content,
	}
}

// --- Server: Requests ---

func (s *HttpServer) Get(path string, callback HttpCallback) {
	s.handleRequest(path, GetMethod, callback)
}

func (s *HttpServer) Post(path string, callback HttpCallback) {
	s.handleRequest(path, PostMethod, callback)
}

func (s *HttpServer) Put(path string, callback HttpCallback) {
	s.handleRequest(path, PutMethod, callback)
}

func (s *HttpServer) Delete(path string, callback HttpCallback) {
	s.handleRequest(path, DeleteMethod, callback)
}

func (s *HttpServer) Connect(path string, callback HttpCallback) {
	s.handleRequest(path, ConnectMethod, callback)
}

func (s *HttpServer) Options(path string, callback HttpCallback) {
	s.handleRequest(path, OptionsMethod, callback)
}

func (s *HttpServer) Trace(path string, callback HttpCallback) {
	s.handleRequest(path, TraceMethod, callback)
}

func (s *HttpServer) Patch(path string, callback HttpCallback) {
	s.handleRequest(path, PatchMethod, callback)
}

// --- Server: Public ---

func (s *HttpServer) DefineNotFoundHandler(callback HttpCallback) {
	s.notFoundHandler = callback
}

func (s *HttpServer) Run() error {
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

		reqSearch := HttpPath {
			path: req.path,
			method: req.method,
		}

		res := newResponse("404 Not Found", "text/html", "")
		fn, ok := s.requests[reqSearch]

		if !ok {
			if s.notFoundHandler != nil {
				res = newResponse("200 OK", "text/html", "")
				s.notFoundHandler(&req, &res)
			}

			conn.Write([]byte(res.construct()))
			conn.Close()

			continue
		}

		res = newResponse("200 OK", "text/html", "")
		fn(&req, &res)
		
		conn.Write([]byte(res.construct()))
		conn.Close()
	}
}

// --- 

func (s *HttpServer) handleRequest(path, method string, callback HttpCallback) {
	req := HttpPath {
		path,
		method,
	}

	s.requests[req] = callback
}

// --- Request: Getters

func (r *HttpRequest) GetMethod() string {
	return r.method
}

func (r *HttpRequest) GetPath() string {
	return r.path
}

func (r *HttpRequest) GetVersion() string {
	return r.version
}

func (r *HttpRequest) GetHeaders() map[string]string {
	return r.headers
}

func (r *HttpRequest) GetBody() string {
	return r.body
}

// --- Response: Getters and Setters ---

func (r *HttpResponse) Write(s string) {
	r.content += s
}

// ---

func (r *HttpResponse) GetContent() string {
	return r.content
}

func (r *HttpResponse) SetContent(s string) {
	r.content = s
}

// ---

func (r *HttpResponse) GetContentType() string {
	return r.contentType
}

func (r *HttpResponse) SetContentType(s string) {
	r.contentType = s
}

// ---

func (r *HttpResponse) GetStatus() string {
	return r.status
}

func (r *HttpResponse) SetStatus(s string) {
	r.status = s
}

// ---

func (r *HttpResponse) construct() string {
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

// --- Util ---

func parseRequest(request string) (HttpRequest, error) {
	req := HttpRequest{ headers: make(map[string]string) }
	lines := strings.Split(request, CRLF)

	for i, line := range lines {
		if i == 0 {
			tokens := strings.Split(line, " ")

			if len(tokens) != 3 {
				return HttpRequest{}, fmt.Errorf("not enough tokens in request header; need 3, got %d: '%s'", len(tokens), line)
			}

			req.method = tokens[0]
			req.path = processPath(tokens[1])
			
			versionTks := strings.Split(tokens[2], "/")

			if len(versionTks) != 2 {
				return HttpRequest{}, fmt.Errorf("not enough tokens in HTTP version header; need 2, got %d: '%s'", len(tokens), line)
			}

			req.version = versionTks[1]
		} else {
			if line == "" {
				break
			}

			// split by ':' or ': '?
			headers := strings.Split(line, ": ")
	
			if len(headers) != 2 {
				return HttpRequest{}, fmt.Errorf("not enough tokens in header; need 2, got %d: '%s'", len(headers), line)
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

func processPath(path string) string {
	path = strings.Trim(path, "/")
	path = "/" + path

	return path
}
