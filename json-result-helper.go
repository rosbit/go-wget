package wget

import (
	"encoding/json"
	// "io"
	// "os"
)

type FnCallJ func(url string, method string, params interface{}, headers map[string]string, res interface{}) (status int, err error)

func HttpCallJ(url string, method string, postData interface{}, headers map[string]string, res interface{}) (int, error) {
	return callWgetJ(url, method, postData, headers, Wget, res)
}

func JsonCallJ(url string, method string, jsonData interface{}, headers map[string]string, res interface{}) (int, error) {
	return callWgetJ(url, method, jsonData, headers, PostJson, res)
}

func callWgetJ(url string, method string, postData interface{}, headers map[string]string, fnCall HttpFunc, res interface{}) (int, error) {
	status, _, resp, err := fnCall(url, method, postData, headers, Options{DontReadRespBody:true})
	if err != nil || resp.Body == nil {
		return status, err
	}
	defer resp.Body.Close()

	return status, json.NewDecoder(resp.Body).Decode(res)

	/*
	io.WriteString(os.Stdout, "body: ")
	r := io.TeeReader(resp.Body, os.Stdout)
	defer io.WriteString(os.Stdout, "\n")

	return status, json.NewDecoder(r).Decode(res)
	*/
}
