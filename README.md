# gohttp

Recreation of the HTTP protocol in Go.

This library provides an API for dealing with HTTP requests, by parsing each request by its raw text, sent over TCP.

Usage example:
```go
server := NewServer("localhost", 8080)
err := server.Run()

if err != nil {
	log.Fatal(err)
}
```

This is a server that does nothing; but you can add endpoints too:
```go
server := NewServer("localhost", 8080)

server.Get("/", func(req *HttpRequest, res *HttpResponse) {
	res.SetContent("<h1>Hello!</h1>")
})

err := server.Run()

if err != nil {
	log.Fatal(err)
}
```

You can also add a handler for dealing with resources that do not exist (i.e. 404 status code):
```go
server := NewServer("localhost", 8080)

server.DefineNotFoundHandler(func(req *HttpRequest, res *HttpResponse) {
	res.Write("<h1>Not Found</h1>")
})

server.Get("/", func(req *HttpRequest, res *HttpResponse) {
	res.SetContent("<h1>Hello!</h1>")
})

err := server.Run()

if err != nil {
	log.Fatal(err)
}
```

> Did you notice the methods `Write` and `SetContent`?  <br />
>
> `Write` appends to the existing `content`, while `SetContent` overrides it.

The response content is a string, which means you can read a file and set the response content to the contents of the file too.

All fields are private; the `HttpRequest` struct provides getters to read the request's details. <br />
The `HttpResponse` struct provides both getters and setters.

This API handles requests synchronously, which means it is single-threaded.
