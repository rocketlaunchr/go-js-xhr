// Package xhr provides GopherJS bindings for the XMLHttpRequest API.
//
// This package provides two ways of using XHR. The first one is via
// the Request type and the NewRequest function. This way, one can
// specify all desired details of the request's behaviour (timeout,
// response format). It also allows access to response details such as
// the status code. Furthermore, using this way is required if one
// wants to abort in-flight requests or if one wants to register
// additional event listeners.
//
//   req := xhr.NewRequest("GET", "/endpoint")
//   req.Timeout = 1000 // one second, in milliseconds
//   req.ResponseType = "document"
//   err := req.Send(nil)
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
//     data, err := xhr.Send("GET", "/endpoint", nil)
//     if err != nil { handle_error() }
//     console.Log("Retrieved data", data)
//
package xhr // import "honnef.co/go/js/xhr"

import (
	"errors"

	"github.com/gopherjs/gopherjs/js"
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

// Request wraps XMLHttpRequest objects. New instances have to be
// created with NewRequest. Each instance may only be used for a
// single request.
type Request struct {
	js.Object
	util.EventTarget
	ReadyState      int       `js:"readyState"`
	Response        js.Object `js:"response"`
	ResponseText    string    `js:"responseText"`
	ResponseType    string    `js:"responseType"`
	ResponseXML     js.Object `js:"responseXML"`
	Status          int       `js:"status"`
	StatusText      string    `js:"statusText"`
	Timeout         int       `js:"timeout"`
	WithCredentials bool      `js:"withCredentials"`

	ch chan error
}

// Upload wraps XMLHttpRequestUpload objects.
type Upload struct {
	js.Object
	util.EventTarget
}

// Upload returns the XMLHttpRequestUpload object associated with the
// request. It can be used to register events for tracking the
// progress of uploads.
func (r *Request) Upload() *Upload {
	o := r.Get("upload")
	return &Upload{o, util.EventTarget{Object: o}}
}

// ErrAborted is the error returned by Send when a request was
// aborted.
var ErrAborted = errors.New("request aborted")

// ErrTimeout is the error returned by Send when a request timed out.
var ErrTimeout = errors.New("request timed out")

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

// Abort will abort the request. The corresponding Send will return
// ErrAborted, unless the request has already succeeded.
func (r *Request) Abort() {
	if r.ch == nil {
		return
	}

	r.Call("abort")
	select {
	case r.ch <- ErrAborted:
	default:
	}
}

// OverrideMimeType overrides the MIME type returned by the server.
func (r *Request) OverrideMimeType(mimetype string) {
	r.Call("overrideMimeType", mimetype)
}

// Send sends the request that was prepared with Open. The data
// argument is optional and can either be a string or []byte payload,
// or a js.Object containing an ArrayBufferView, Blob, Document or
// Formdata. String and []byte arguments will be sent as binary data,
// without any transformations.
//
// Send will block until a response was received or an error occured.
//
// Only errors of the network layer are treated as errors. HTTP status
// codes 4xx and 5xx are not treated as errors. In order to check
// status codes, use the Request's Status field.
func (r *Request) Send(data interface{}) error {
	if r.ch != nil {
		panic("must not use a Request for multiple requests")
	}
	r.ch = make(chan error, 1)
	r.AddEventListener("load", false, func(js.Object) {
		go func() { r.ch <- nil }()
	})
	r.AddEventListener("error", false, func(o js.Object) {
		go func() { r.ch <- ErrFailure }()
	})
	r.AddEventListener("timeout", false, func(js.Object) {
		go func() { r.ch <- ErrTimeout }()
	})

	// Send strings as binary data
	if s, ok := data.(string); ok {
		data = []byte(s)
	}

	r.Call("send", data)
	val := <-r.ch
	return val
}

// SetRequestHeader sets a header of the request.
func (r *Request) SetRequestHeader(header, value string) {
	r.Call("setRequestHeader", header, value)
}

// Send constructs a new Request and sends it. The response, if any,
// is interpreted as binary data and returned as is. For details about
// the data parameter, see the documentation of Request.Send.
//
// Only errors of the network layer are treated as errors. HTTP status
// codes 4xx and 5xx are not treated as errors. In order to check
// status codes, use NewRequest instead.
func Send(method, url string, data interface{}) ([]byte, error) {
	xhr := NewRequest(method, url)
	xhr.ResponseType = ArrayBuffer
	err := xhr.Send(data)
	if err != nil {
		return nil, err
	}
	return js.Global.Get("Uint8Array").New(xhr.Response).Interface().([]byte), nil
}
