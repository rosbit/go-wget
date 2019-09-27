package wget

import (
	"fmt"
	"testing"
	"net/http"
	"strings"
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

func Test_Reader(t *testing.T) {
	print_result(PostJson("http://httpbin.org/post", http.MethodPost, strings.NewReader(`{"a":"b","c":"d"}`), headers))
	fmt.Printf("------------ done for PostJson io.Reader with POST -------------\n")
}

func Test_httpBuildParmas(t *testing.T) {
	s := strings.NewReader(`{"a":"b","c":"d"}`)
	if _, err := buildHttpParams(s); err != nil {
		fmt.Printf("----failed to buildHttpParams: %v\n", err)
	} else {
		fmt.Printf("----buildHttpParmas ok\n")
	}
}
