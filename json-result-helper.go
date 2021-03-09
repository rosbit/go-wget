package wget

import (
	"encoding/json"
	"io"
)

type FnCallJ func(url string, method string, params interface{}, headers map[string]string, res interface{}, options ...Options) (status int, err error)

func HttpCallJ(url string, method string, postData interface{}, headers map[string]string, res interface{}, options ...Options) (int, error) {
	return callWgetJ(url, method, postData, headers, Wget, res, options...)
}

func JsonCallJ(url string, method string, jsonData interface{}, headers map[string]string, res interface{}, options ...Options) (int, error) {
	return callWgetJ(url, method, jsonData, headers, PostJson, res, options...)
}

func callWgetJ(url string, method string, postData interface{}, headers map[string]string, fnCall HttpFunc, res interface{}, options ...Options) (int, error) {
	var op *Options
	if len(options) > 0 {
		op = &options[0]
		op.DontReadRespBody = true
	} else {
		op = &Options{DontReadRespBody:true}
	}

	status, _, resp, err := fnCall(url, method, postData, headers, *op)
	if err != nil || resp.Body == nil {
		return status, err
	}
	defer resp.Body.Close()

	if op.DebugWriter == nil {
		return status, json.NewDecoder(resp.Body).Decode(res)
	}

	w := op.DebugWriter
	io.WriteString(w, "body: ")
	r := io.TeeReader(resp.Body, w)
	defer io.WriteString(w, "\n")
	return status, json.NewDecoder(r).Decode(res)
}
