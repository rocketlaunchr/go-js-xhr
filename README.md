# js/xhr

Package xhr provides GopherJS bindings for the XMLHttpRequest API.

## Install

    go get github.com/rocketlaunchr/gopherjs-xhr

## Usage

### node.js

Install npm package `xhr2`

```javascript 
File: imports.inc.js

global.XMLHttpRequest = require('xhr2'); 
```

### node.js and browser

```go
import (
	"context"

	xhr "github.com/rocketlaunchr/gopherjs-xhr"
	"github.com/gopherjs/gopherjs/js"
	"github.com/rocketlaunchr/react/forks/encoding/json"
)

req := xhr.NewRequest("POST", reqURL)
req.ResponseType = xhr.Text // Returns response as string
req.SetRequestHeader("Content-Type", xhr.ApplicationForm)


postBody := NewParams(js.M{"setting": 4})

err := req.Send(context.Background(), postBody.String())
if err != nil {
	// Could not connect to internet???
	// Unfortunately XMLHttpRequest does not provide nuanced reasons.
	return
}

if !req.IsStatus2xx() {
	// Something went wrong
	return
}

// Unmarshal json response here using encoding/json. Otherwise set req.ResponseType = "json".
err = json.Unmarshal(req.ResponseBytes(), &sb)
```


## Documentation

For documentation, see http://godoc.org/github.com/rocketlaunchr/gopherjs-xhr
