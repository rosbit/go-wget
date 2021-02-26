package wget

import (
	"encoding/json"
	"io"
)

type FnCallJ func(url string, method string, params interface{}, headers map[string]string, res interface{}, debugWriter ...io.Writer) (status int, err error)

func HttpCallJ(url string, method string, postData interface{}, headers map[string]string, res interface{}, debugWriter ...io.Writer) (int, error) {
	return callWgetJ(url, method, postData, headers, Wget, res, debugWriter...)
}

func JsonCallJ(url string, method string, jsonData interface{}, headers map[string]string, res interface{}, debugWriter ...io.Writer) (int, error) {
	return callWgetJ(url, method, jsonData, headers, PostJson, res, debugWriter...)
}

func callWgetJ(url string, method string, postData interface{}, headers map[string]string, fnCall HttpFunc, res interface{}, debugWriter ...io.Writer) (int, error) {
	status, _, resp, err := fnCall(url, method, postData, headers, Options{DontReadRespBody:true})
	if err != nil || resp.Body == nil {
		return status, err
	}
	defer resp.Body.Close()

	if len(debugWriter) == 0 || debugWriter[0] == nil {
		return status, json.NewDecoder(resp.Body).Decode(res)
	}

	w := debugWriter[0]
	io.WriteString(w, "body: ")
	r := io.TeeReader(resp.Body, w)
	defer io.WriteString(w, "\n")
	return status, json.NewDecoder(r).Decode(res)
}
