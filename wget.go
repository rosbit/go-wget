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
	"io/ioutil"
	"time"
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
	DebugWriter io.Writer
	MultiBase  *BaseUrl
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
	if !isHttpUrl(url) && len(options) > 0 && options[0].MultiBase != nil {
		return options[0].MultiBase.HttpCall(url, method, params, header, options...)
	}
	return newRequest(url, 0, options...).Run(url, method, params, header)
}

func PostJson(url, method string, params interface{}, header map[string]string, options ...Options) (status int, content []byte, resp *http.Response, err error) {
	if !isHttpUrl(url) && len(options) > 0 && options[0].MultiBase != nil {
		return options[0].MultiBase.JsonCall(url, method, params, header, options...)
	}
	return newRequest(url, 0, options...).PostJson(url, method, params, header)
}

func GetUsingBodyParams(url string, params interface{}, header map[string]string, options ...Options) (status int, content []byte, resp *http.Response, err error) {
	if !isHttpUrl(url) && len(options) > 0 && options[0].MultiBase != nil {
		return options[0].MultiBase.GetWithBody(url, params, header, options...)
	}
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
	var paramsReader io.ReadSeeker
	if url, method, paramsReader, header, err = adjustHttpArgs(url, method, params, header); err != nil {
		return
	}
	return wget.run(url, method, paramsReader, header)
}

func (wget *Request) PostJson(url, method string, params interface{}, header map[string]string) (status int, content []byte, resp *http.Response, err error) {
	var paramsReader io.ReadSeeker
	if method, paramsReader, header, err = adjustJsonArgs(method, params, header); err != nil {
		return
	}
	return wget.run(url, method, paramsReader, header)
}

func (wget *Request) GetUsingBodyParams(url string, params interface{}, header map[string]string) (status int, content []byte, resp *http.Response, err error) {
	var paramsReader io.ReadSeeker
	if _, _, paramsReader, header, err = adjustHttpArgs(url, http.MethodPost, params, header); err != nil {
		return
	}
	return wget.run(url, http.MethodGet, paramsReader, header)
}

func (wget *Request) run(url, method string, params io.Reader, header map[string]string) (int, []byte, *http.Response, error) {
	var req *http.Request
	var err error
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		if req, err = http.NewRequest(method, url, params); err != nil {
			return http.StatusBadRequest, nil, nil, err
		}
	default:
		return http.StatusMethodNotAllowed, nil, nil, fmt.Errorf("method %s not supported", method)
	}

	if len(header) > 0 {
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
