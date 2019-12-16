package xhr

import (
	"github.com/gopherjs/gopherjs/js"
)

// Params represents a URLSearchParams object
type Params struct {
	*js.Object
}

// NewParams returns a new URLSearchParams object.
func NewParams(kv ...js.M) *Params {
	o := js.Global.Get("URLSearchParams").New()
	p := &Params{Object: o}
	if len(kv) > 0 {
		p.Append(kv[0])
	}
	return p
}

// Appends a specified key/value pair as a new search parameter.
func (p *Params) Append(kv js.M) {
	for name, val := range kv {
		p.Call("append", name, val)
	}
}

// String returns a string containing a query string suitable for use in a URL.
func (p *Params) String() string {
	return p.Call("toString").String()
}
