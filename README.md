# go-wget (http client wrapper)

go-wget is a http client package to make use of Go built-in net/http

### Usage

This package is fully go-getable. So just type `go get github.com/rosbit/go-wget` to install.

```go
package main

import (
	"github.com/rosbit/go-wget"
	"fmt"
)

func main() {
	params := map[string]interface{}{
		"a": "b",
		"c": 1,
	}
	headers := map[string]string{
		"X-Param": "x value",
	}

	status, content, resp, err := wget.Wget("http://yourname.com/path/to/url", "get", params, headers)
	/*
	// POST as request method
	status, content, resp, err := wget.Wget("http://yourname.com/path/to/url", "post", params, headers)
	// post body as a JSON 
	status, content, resp, err := wget.PostJson("http://yourname.com/path/to/url", "", params, headers)
	// post body as a JSON, even the method is GET
	status, content, resp, err := wget.PostJson("http://yourname.com/path/to/url", "GET", params, headers)
	// request method is GET, request params as a FORM body
	status, content, resp, err := wget.GetUsingBodyParams("http://yourname.com/path/to/url", params, headers)
	*/
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Printf("status: %d\n", status)
	fmt.Printf("reponse content: %s\n", string(content))
	respHeaders := wget.GetHeaders(resp)
	for k, v := range respHeaders {
		fmt.Printf("%s: %s\n", k, v)
	}
}
```

### Usage as fs
```go
package main

import (
	"github.com/rosbit/go-wget"
	"io"
	"os"
	"fmt"
)

func main() {
	// GET
	fp := wget.Get("http://httpbin.org/get")
	defer fp.Close()
	io.Copy(os.Stdout, fp)

	// POST JSON
	fp2 := wget.Post("http://httpbin.org/post", &wget.Args{Params: map[string]interface{}{"a": "b", "c": 1}, JsonCall: true})
	defer fp2.Close()
	io.Copy(os.Stdout, fp2)

	// with helper
	status, body, err := wget.FsCall("http://httpbin.org/post", "POST", &wget.Args{Params: map[string]interface{}{"a": "b", "c": 1}, JsonCall: true})
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	fmt.Printf("status: %d\n", status)
	if body != nil {
		defer body.Close()
		io.Copy(os.Stdout, body)
	}
}
```

### Usage with multi-baseurl
```go
    multiBase, err := NewBaseUrl(BaseItem("http://192.168.0.241:8088"), BaseItem("http://httpbin.org"))
    if err != nil {
         // err
    }
    status, body, _, err := multiBase.HttpCall("/post", http.MethodPost, params, headers)
    multiBase.JsonCall("/post", http.MethodPost, params, headers)
    PostJson("/post", http.MethodPost, params, headers, Options{MultiBase:multiBase})
```

### Status

The package is not fully tested, so be careful.

### Contribution

Pull requests are welcome! Also, if you want to discuss something send a pull request with proposal and changes.
__Convention:__ fork the repository and make changes on your fork in a feature branch.
