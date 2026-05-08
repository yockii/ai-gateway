package router

import (
	"io"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// NewFastHTTPToHTTPHandler converts a net/http handler to a fasthttp handler
func NewFastHTTPToHTTPHandler(handler http.Handler) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Use fasthttpadaptor to convert the handler
		fasthttpHandler := fasthttpadaptor.NewFastHTTPHandler(handler)
		fasthttpHandler(c.RequestCtx())
		return nil
	}
}

// responseWriterAdapter adapts fasthttp.Response to http.ResponseWriter
type responseWriterAdapter struct {
	response *fasthttp.Response
	header   http.Header
	written  bool
}

func newResponseWriterAdapter(response *fasthttp.Response) *responseWriterAdapter {
	return &responseWriterAdapter{
		response: response,
		header:   make(http.Header),
	}
}

func (a *responseWriterAdapter) Header() http.Header {
	return a.header
}

func (a *responseWriterAdapter) Write(b []byte) (int, error) {
	a.written = true
	// Copy headers from http.Header to fasthttp.Response.Header
	for k, vv := range a.header {
		for _, v := range vv {
			a.response.Header.Set(k, v)
		}
	}
	return a.response.BodyWriter().Write(b)
}

func (a *responseWriterAdapter) WriteHeader(statusCode int) {
	a.written = true
	a.response.SetStatusCode(statusCode)
	// Copy headers from http.Header to fasthttp.Response.Header
	for k, vv := range a.header {
		for _, v := range vv {
			a.response.Header.Set(k, v)
		}
	}
}

// requestAdapter adapts fasthttp.Request to http.Request
type requestAdapter struct {
	request *fasthttp.Request
}

func newRequestAdapter(request *fasthttp.Request) *requestAdapter {
	return &requestAdapter{request: request}
}

func (a *requestAdapter) Method() string {
	return string(a.request.Header.Method())
}

func (a *requestAdapter) URL() string {
	return string(a.request.URI().FullURI())
}

func (a *requestAdapter) Proto() string {
	return "HTTP/1.1"
}

func (a *requestAdapter) ProtoMajor() int {
	return 1
}

func (a *requestAdapter) ProtoMinor() int {
	return 1
}

func (a *requestAdapter) Header() http.Header {
	header := make(http.Header)
	a.request.Header.VisitAll(func(key, value []byte) {
		header.Add(string(key), string(value))
	})
	return header
}

func (a *requestAdapter) Body() io.ReadCloser {
	return io.NopCloser(a.request.BodyStream())
}

// ServeHTTPAdapter converts a net/http handler to work with Fiber
func ServeHTTPAdapter(handler http.Handler) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Create a response writer adapter
		rw := newResponseWriterAdapter(c.Response())

		// Create a request adapter
		req, err := http.NewRequest(
			string(c.Method()),
			string(c.Request().RequestURI()),
			c.Request().BodyStream(),
		)
		if err != nil {
			return err
		}

		// Copy headers
		c.Request().Header.VisitAll(func(key, value []byte) {
			req.Header.Add(string(key), string(value))
		})

		// Serve the HTTP request
		handler.ServeHTTP(rw, req)

		return nil
	}
}
