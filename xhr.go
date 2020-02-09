// Package xhr provides GopherJS bindings for the XMLHttpRequest API.
//
// This package provides two ways of using XHR directly. The first one
// is via the Request type and the NewRequest function. This way, one
// can specify all desired details of the request's behaviour
// (timeout, response format). It also allows access to response
// details such as the status code. Furthermore, using this way is
// required if one wants to abort in-flight requests or if one wants
// to register additional event listeners.
//
//   req := xhr.NewRequest("GET", "/endpoint")
//   req.ResponseType = "document"
//   err := req.Send(ctx, nil)
//   if err != nil { handle_error() }
//   // req.Response will contain a JavaScript Document element that can
//   // for example be used with the js/dom package.
//
//
// The other way is via the package function Send, which is a helper
// that internally constructs a Request and assigns sane defaults to
// it. It's the easiest way of doing an XHR request that should just
// return unprocessed data.
//
//     data, err := xhr.Send(ctx, "POST", "/endpoint", []byte("payload here"))
//     if err != nil { handle_error() }
//     console.Log("Retrieved data", data)
//
//
// If you don't need to/want to deal with the underlying details of
// XHR, you may also just use the net/http.DefaultTransport, which
// GopherJS replaces with an XHR-enabled version, making this package
// useless most of the time.
package xhr

import (
	"errors"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/rocketlaunchr/react/forks/context"
	"honnef.co/go/js/util"
)

// The possible values of Request.ReadyState.
const (
	// Open has not been called yet
	Unsent = iota
	// Send has not been called yet
	Opened
	HeadersReceived
	Loading
	Done
)

// The possible values of Request.ResponseType
const (
	ArrayBuffer = "arraybuffer"
	Blob        = "blob"
	Document    = "document"
	JSON        = "json"
	Text        = "text"
)

const (
	// ApplicationForm is a common "Content-Type" used by forms.
	ApplicationForm = "application/x-www-form-urlencoded"
	// ApplicationJSON is a common "Content-Type" used when making a POST request.
	ApplicationJSON = "application/json"
	// TextPlain is a common "Content-Type".
	TextPlain = "text/plain"
)

// MultipartFormData is a common "Content-Type" when transferring files in a POST request.
var MultipartFormData = func(boundary string) string {
	return "multipart/form-data;boundary=\"" + boundary + "\""
}

// Request wraps XMLHttpRequest objects. New instances have to be
// created with NewRequest. Each instance may only be used for a
// single request.
//
// To create a request that behaves in the same way as the top-level
// Send function with regard to handling binary data, use the
// following:
//
//   req := xhr.NewRequest("POST", "http://example.com")
//   req.ResponseType = xhr.ArrayBuffer
//   req.Send(ctx, []byte("data"))
//   b := js.Global.Get("Uint8Array").New(req.Response).Interface().([]byte)
type Request struct {
	*js.Object
	util.EventTarget
	ReadyState      int        `js:"readyState"`
	Response        *js.Object `js:"response"`
	ResponseText    string     `js:"responseText"`
	ResponseType    string     `js:"responseType"`
	ResponseXML     *js.Object `js:"responseXML"`
	Status          int        `js:"status"`
	StatusText      string     `js:"statusText"`
	WithCredentials bool       `js:"withCredentials"`

	alreadySent bool // Indicate that send has been called
}

// Upload wraps XMLHttpRequestUpload objects.
type Upload struct {
	*js.Object
	util.EventTarget
}

// Upload returns the XMLHttpRequestUpload object associated with the
// request. It can be used to register events for tracking the
// progress of uploads.
func (r *Request) Upload() *Upload {
	o := r.Get("upload")
	return &Upload{o, util.EventTarget{Object: o}}
}

// ErrFailure is the error returned by Send when it failed for a
// reason other than abortion or timeouts.
//
// The specific reason for the error is unknown because the XHR API
// does not provide us with any information. One common reason is
// network failure.
var ErrFailure = errors.New("send failed")

// NewRequest creates a new XMLHttpRequest object, which may be used
// for a single request.
func NewRequest(method, url string) *Request {
	o := js.Global.Get("XMLHttpRequest").New()
	r := &Request{Object: o, EventTarget: util.EventTarget{Object: o}}
	r.Call("open", method, url, true)
	return r
}

// ResponseHeaders returns all response headers.
func (r *Request) ResponseHeaders() string {
	return r.Call("getAllResponseHeaders").String()
}

// ResponseHeader returns the value of the specified header.
func (r *Request) ResponseHeader(name string) string {
	value := r.Call("getResponseHeader", name)
	if value == nil {
		return ""
	}
	return value.String()
}

// OverrideMimeType overrides the MIME type returned by the server.
func (r *Request) OverrideMimeType(mimetype string) {
	r.Call("overrideMimeType", mimetype)
}

// ResponseBytes returns the ResponseText as a slice of bytes.
func (r *Request) ResponseBytes() []byte {
	return []byte(r.ResponseText)
}

// IsStatus2xx returns true if the request returned a 2xx status code.
func (r *Request) IsStatus2xx() bool {
	if r.Status < 200 || r.Status > 299 {
		return false
	}
	return true
}

// IsStatus4xx returns true if the request returned a 4xx status code.
func (r *Request) IsStatus4xx() bool {
	if r.Status < 400 || r.Status > 499 {
		return false
	}
	return true
}

// IsStatus5xx returns true if the request returned a 5xx status code.
func (r *Request) IsStatus5xx() bool {
	if r.Status < 500 || r.Status > 599 {
		return false
	}
	return true
}

// Send sends the request that was prepared with Open. The data
// argument is optional and can either be a string or []byte payload,
// or a *js.Object containing an ArrayBufferView, Blob, Document or
// Formdata.
//
// Send will block until a response was received or an error occured.
//
// Only errors of the network layer are treated as errors. HTTP status
// codes 4xx and 5xx are not treated as errors. In order to check
// status codes, use the Request's Status field.
func (r *Request) Send(ctx context.Context, data interface{}) error {

	if r.alreadySent {
		panic("must not use a Request for multiple requests")
	}

	if dt, ok := ctx.Deadline(); ok {
		diff := time.Until(dt) / time.Millisecond
		if diff != 0 {
			r.Set("timeout", diff)
		}
	}

	errChan := make(chan error)
	returnedChan := make(chan struct{}) // Used to indicate that this function has returned

	defer func() {
		r.alreadySent = true
		returnedChan <- struct{}{}
	}()

	go func() {
		select {
		case <-ctx.Done():
			r.Call("abort")
			errChan <- ctx.Err()
		case <-returnedChan:
		}
	}()

	r.AddEventListener("load", false, func(*js.Object) {
		go func() { errChan <- nil }()
	})
	r.AddEventListener("error", false, func(*js.Object) {
		go func() { errChan <- ErrFailure }()
	})
	r.AddEventListener("timeout", false, func(*js.Object) {
		go func() { errChan <- context.DeadlineExceeded }()
	})

	r.Call("send", data)

	return <-errChan
}

// SetRequestHeader sets a header of the request.
func (r *Request) SetRequestHeader(header, value string) {
	r.Call("setRequestHeader", header, value)
}

// Send constructs a new Request and sends it. The response, if any,
// is interpreted as binary data and returned as is.
//
// For more control over the request, as well as the option to send
// types other than []byte, construct a Request yourself.
//
// Only errors of the network layer are treated as errors. HTTP status
// codes 4xx and 5xx are not treated as errors. In order to check
// status codes, use NewRequest instead.
func Send(ctx context.Context, method, url string, data []byte) ([]byte, error) {
	xhr := NewRequest(method, url)
	xhr.ResponseType = ArrayBuffer
	err := xhr.Send(ctx, data)
	if err != nil {
		return nil, err
	}
	return js.Global.Get("Uint8Array").New(xhr.Response).Interface().([]byte), nil
}
