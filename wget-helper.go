package wget

import (
	"io"
)

type FnCall func(url string, method string, params interface{}, headers map[string]string) (status int, body io.ReadCloser, err error)

func HttpCall(url string, method string, postData interface{}, headers map[string]string) (int, io.ReadCloser, error) {
	return callWget(url, method, postData, headers, Wget)
}

func JsonCall(url string, method string, jsonData interface{}, headers map[string]string) (int, io.ReadCloser, error) {
	return callWget(url, method, jsonData, headers, PostJson)
}

func callWget(url string, method string, postData interface{}, headers map[string]string, fnCall HttpFunc) (int, io.ReadCloser, error) {
	status, _, resp, err := fnCall(url, method, postData, headers, Options{DontReadRespBody:true})
	if err != nil {
		return status, nil, err
	}
	return status, resp.Body, nil
}
