/**
 * http client implementation
 * Rosbit Xu <me@rosbit.cn>
 * Jan. 8, 2018
 */
package wget

import (
	"fmt"
	"net/http"
	"strings"
	"bytes"
	"net/url"
	"io/ioutil"
	"time"
	"encoding/json"
	"os"
	"io"
	"crypto/tls"
	"crypto/x509"
)

type Request struct {
	client  *http.Client
	options *Options
}

type Options struct {
	Timeout           int  // timeout in seconds to wait while connect/send/recv-ing
	DontReadRespBody bool  // if it is true, it's your resposibility to get body from http.Response.Body
}

type HttpFunc func(string,string,interface{},map[string]string,...Options)(int,[]byte,*http.Response,error)

const (
	connect_timeout = 5    // default seconds to wait while trying to connect
)

func NewRequest(connectTimeout int, options ...Options) *Request {
	timeout, option := getOptions(connectTimeout, options...)
	return &Request{client: &http.Client{Timeout: time.Duration(timeout)*time.Second}, options: option}
}

func NewHttpsRequest(connectTimeout int, options ...Options) *Request {
	timeout, option := getOptions(connectTimeout, options...)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return &Request{client: &http.Client{Transport: transport, Timeout: time.Duration(timeout)*time.Second}, options: option}
}

func NewHttpsRequestWithCerts(connectTimeout int, certPemFile, keyPemFile string, options ...Options) (*Request, error) {
	timeout, option := getOptions(connectTimeout, options...)
	cert, err := tls.LoadX509KeyPair(certPemFile, keyPemFile)
	if err != nil {
		return nil, err
	}
	certBytes, err := ioutil.ReadFile(certPemFile)
	if err != nil {
		return nil, err
	}
	clientCertPool := x509.NewCertPool()
	if !clientCertPool.AppendCertsFromPEM(certBytes) {
		return nil, fmt.Errorf("Failed to AppendCertsFromPEM")
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:            clientCertPool,
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		},
	}
	return &Request{client: &http.Client{Transport: transport, Timeout: time.Duration(timeout)*time.Second}, options: option}, nil
}

func getOptions(connectTimeout int, options ...Options) (int, *Options) {
	var option *Options
	if len(options) > 0 {
		option = &options[0]
		if option.Timeout <= 0 {
			if connectTimeout < 0 {
				connectTimeout = connect_timeout
			}
		} else {
			connectTimeout = option.Timeout
		}
	} else if connectTimeout <= 0 {
		connectTimeout = connect_timeout
	}

	return connectTimeout, option
}

func newRequest(url string, connectTimeout int, options ...Options) *Request {
	if strings.Index(url, "https://") == 0 {
		return NewHttpsRequest(connectTimeout, options...)
	} else {
		return NewRequest(connectTimeout, options...)
	}
}

func Wget(url, method string, params interface{}, header map[string]string, options ...Options) (status int, content []byte, resp *http.Response, err error) {
	return newRequest(url, 0, options...).Run(url, method, params, header)
}

func PostJson(url, method string, params interface{}, header map[string]string, options ...Options) (status int, content []byte, resp *http.Response, err error) {
	return newRequest(url, 0, options...).PostJson(url, method, params, header)
}

func GetUsingBodyParams(url string, params interface{}, header map[string]string, options ...Options) (status int, content []byte, resp *http.Response, err error) {
	return newRequest(url, 0, options...).GetUsingBodyParams(url, params, header)
}

func GetStatus(resp *http.Response) (int, string) {
	return resp.StatusCode, resp.Status
}

func GetHeaders(resp *http.Response) map[string]string {
	res := make(map[string]string, len(resp.Header))
	for k, v := range resp.Header {
		if v == nil || len(v) == 0 {
			res[k] = ""
		} else {
			res[k] = v[0]
		}
	}
	return res
}

func GetLastModified(resp *http.Response) (time.Time, error) {
	if resp == nil {
		return time.Time{}, fmt.Errorf("no response given")
	}
	if lastModified, ok := resp.Header["Last-Modified"]; ok {
		return time.Parse(time.RFC1123, lastModified[0])
	}
	return time.Time{}, fmt.Errorf("no response header Last-Modified")
}

func isHttpUrl(rawurl string) bool {
	return (strings.Index(rawurl, "http://") == 0) || (strings.Index(rawurl, "https://") == 0)
}

func ModTime(rawurl string) (time.Time, error) {
	if isHttpUrl(rawurl) {
		_, _, resp, err := newRequest(rawurl, 0).Run(rawurl, http.MethodHead, nil, nil)
		if err != nil {
			return time.Time{}, err
		}
		return GetLastModified(resp)
	} else {
		st, e := os.Stat(rawurl)
		if e != nil {
			return time.Time{}, e
		}
		return st.ModTime(), nil
	}
}

func (wget *Request) Run(url, method string, params interface{}, header map[string]string) (status int, content []byte, resp *http.Response, err error) {
	return wget.run(url, method, params, header, false)
}

func (wget *Request) PostJson(url, method string, params interface{}, header map[string]string) (status int, content []byte, resp *http.Response, err error) {
	if method == "" {
		method = http.MethodPost
	}

	j, e := buildJsonParams(params)
	if e != nil {
		status, err = http.StatusBadRequest, e
		return
	}

	if header == nil {
		header = make(map[string]string, 1)
	}
	header["Content-Type"] = "application/json"
	return wget.run(url, method, j, header, true)
}

func (wget *Request) GetUsingBodyParams(url string, params interface{}, header map[string]string) (status int, content []byte, resp *http.Response, err error) {
	return wget.run(url, "GET", params, header, true)
}

func (wget *Request) run(url, method string, params interface{}, header map[string]string, withGetBody bool) (int, []byte, *http.Response, error) {
	var req *http.Request
	param, err := buildHttpParams(params)
	if err != nil {
		return http.StatusBadRequest, nil, nil, err
	}

	if method == "" {
		method = http.MethodGet
	} else {
		method = strings.ToUpper(method)
	}
	if param == nil {
		if req, err = http.NewRequest(method, url, nil); err != nil {
			return http.StatusBadRequest, nil, nil, err
		}
	} else {
		setForm := true
		switch method {
		case http.MethodGet, http.MethodHead:
			if !withGetBody {
				setForm = false
				p, _ := ioutil.ReadAll(param)
				url = fmt.Sprintf("%s?%s", url, string(p))
				params = nil
			}
			fallthrough
		case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
			if req, err = http.NewRequest(method, url, param); err != nil {
				return http.StatusBadRequest, nil, nil, err
			}
			if setForm {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		default:
			return http.StatusMethodNotAllowed, nil, nil, fmt.Errorf("method %s not supported", method)
		}
	}
	if header != nil {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}

	resp, err := wget.client.Do(req)
	if err != nil {
		return http.StatusInternalServerError, nil, nil, err
	}

	if wget.options != nil && wget.options.DontReadRespBody {
		return resp.StatusCode, nil, resp, nil
	}

	defer resp.Body.Close()

	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return resp.StatusCode, nil, nil, err
	} else {
		return resp.StatusCode, body, resp, nil
	}
}

func buildHttpParams(params interface{}) (io.Reader, error) {
	if params == nil {
		return nil, nil
	}
	if r, ok := params.(io.Reader); ok {
		return r, nil
	}
	switch params.(type) {
	case []byte:
		return bytes.NewReader(params.([]byte)), nil
	case string:
		return strings.NewReader(params.(string)), nil
	case map[string]interface{}:
		m,_ := params.(map[string]interface{})
		u := url.Values{}
		for k, v := range m {
			u.Set(k, fmt.Sprintf("%v", v))
		}
		return strings.NewReader(u.Encode()), nil
	case map[string]string:
		m,_ := params.(map[string]string)
		u := url.Values{}
		for k, v := range m {
			u.Set(k, v)
		}
		return strings.NewReader(u.Encode()), nil
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		bool:
		return strings.NewReader(fmt.Sprintf("%v", params)), nil
	default:
		return nil, fmt.Errorf("unknown type to build http params")
	}
}

func buildJsonParams(params interface{}) (io.Reader, error) {
	if params == nil {
		return nil, fmt.Errorf("no params to build json")
	}

	if j, ok := params.(io.Reader); ok {
		return j, nil
	}

	jr, jw := io.Pipe()
	go func() {
		enc := json.NewEncoder(jw)
		if err := enc.Encode(params); err != nil {
			jw.CloseWithError(err)
		} else {
			jw.Close()
		}
	}()
	return jr, nil
}
