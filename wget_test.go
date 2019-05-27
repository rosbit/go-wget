package wget

import (
	"fmt"
	"testing"
	"net/http"
)

var (
	params = map[string]interface{}{
		"a": "b",
		"c": 1,
	}

	headers = map[string]string{
		"X-Param": "x value",
	}
)

func print_result(status int, content []byte, resp *http.Response, err error) {
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Printf("status: %d\n", status)
	fmt.Printf("reponse content: %s\n", string(content))
	respHeaders := GetHeaders(resp)
	for k, v := range respHeaders {
		fmt.Printf("%s: %s\n", k, v)
	}
}

func wget_test(url string, method string) {
	print_result(Wget(url, method, params, headers))
	fmt.Printf("------------ done for Wget %s with %s -------------\n", url, method)
}

func json_test(url string, method string) {
	print_result(PostJson(url, method, params, headers))
	fmt.Printf("------------ done for PostJson %s with %s -------------\n", url, method)
}

func Test_Wget(t *testing.T) {
	wget_test("http://httpbin.org/get",  http.MethodGet)
	wget_test("http://httpbin.org/post", http.MethodPost)
}

func Test_PostJson(t *testing.T) {
	json_test("http://httpbin.org/get",  http.MethodGet)
	json_test("http://httpbin.org/post", http.MethodPost)
}
