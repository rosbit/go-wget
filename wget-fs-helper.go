// remove go1.16 dependency, build go1.16

package wget

import (
	"io"
	"encoding/json"
)

func FsCall(url string, method string, options ...*Args) (status int, body io.ReadCloser, err error) {
	fp := wget_fs(url, method, options...)
	fi, e := fp.Stat()
	if e != nil {
		err = e
		return
	}
	result := fi.Sys().(*Result)
	status, body = result.Status, result.Resp.Body
	return
}

func FsCallAndParseJSON(url string, method string, res interface{}, options ...*Args) (status int, err error) {
	var body io.ReadCloser
	status, body, err = FsCall(url, method, options...)
	if err != nil || body == nil {
		return
	}
	defer body.Close()

	if len(options) == 0 || options[0] == nil || options[0].Logger == nil {
		err = json.NewDecoder(body).Decode(res)
		return
	}

	w := options[0].Logger
	io.WriteString(w, "body: ")
	r := io.TeeReader(body, w)
	defer io.WriteString(w, "\n")
	return status, json.NewDecoder(r).Decode(res)
}
