package xhr

import (
	"github.com/gopherjs/gopherjs/js"
)

// Params represents a URLSearchParams object
type Params struct {
	*js.Object
}

// NewParams returns a new URLSearchParams object.
func NewParams() *Params {
	o := js.Global.Get("URLSearchParams").New()
	return &Params{Object: o}
}

// Appends a specified key/value pair as a new search parameter.
func (p *Params) Append(name string, val interface{}) {
	p.Call("append", name, val)
}

// String returns a string containing a query string suitable for use in a URL.
func (p *Params) String() string {
	return p.Call("toString").String()
}
