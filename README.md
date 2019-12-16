# js/xhr

Package xhr provides GopherJS bindings for the XMLHttpRequest API.

## Install

    go get github.com/rocketlaunchr/gopherjs-xhr

## Usage

### node.js

Install npm package `xhr2`

```javascript 
File: imports.inc.js

// polyfill to allow net/http to work with node.js
global.XMLHttpRequest = require('xhr2'); 
```

### node.js and browser

```go
import (
	"context"
	"strings"

	xhr github.com/rocketlaunchr/gopherjs-xhr
)

req := xhr.NewRequest("POST", reqURL)
req.ResponseType = xhr.Text // Returns response as string
req.SetRequestHeader("Content-Type", "application/x-www-form-urlencoded")


postBody := NewParams()
postBody.Append("setting", 4)

err := req.Send(context.Background(), postBody.String())
if err != nil {
	if strings.Contains(err.Error(), "net/http: fetch() failed") {
		// Could not connect to internet
		return
	}

	// Another type of error
	return
}

if !req.Status2xx() {
	// Something went wrong
	return
}

data := []byte(req.ResponseText)

// Unmarshal json response here using encoding/json. Otherwise set req.ResponseType = "json".
```


## Documentation

For documentation, see http://godoc.org/github.com/rocketlaunchr/gopherjs-xhr
