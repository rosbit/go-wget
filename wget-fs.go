// remove go1.16 dependency build go1.16

package wget

import (
	// "io/fs"
	"io"
	"os"
	"time"
	"path"
	"net/url"
	"net/http"
)

// arguments for HTTP request
type Args struct {
	Params interface{}
	Headers map[string]string
	Timeout int
	JsonCall bool
	Logger io.Writer
}

// result of HTTP response, returned by FileInfo.Sys()
type Result struct {
	Status int
	Resp *http.Response
	Err error
}

func HttpRequest(url string, method string, options ...*Args) *File /*fs.File*/ {
	return wget_fs(url, method, options...)
}

func Get(url string, options ...*Args) *File /*fs.File*/ {
	return wget_fs(url, http.MethodGet, options...)
}

func Post(url string, options ...*Args) *File /*fs.File*/ {
	return wget_fs(url, http.MethodPost, options...)
}

func Put(url string, options ...*Args) *File /*fs.File*/ {
	return wget_fs(url, http.MethodPut, options...)
}

func Delete(url string, options ...*Args) *File /*fs.File*/ {
	return wget_fs(url, http.MethodPut, options...)
}

func Head(url string, options ...*Args) *File /*fs.File*/ {
	return wget_fs(url, http.MethodHead, options...)
}

func wget_fs(url string, method string, options ...*Args) *File /*fs.File*/ {
	f := &File{
		method: method,
		url: url,
	}
	f.setOptions(options)
	return f
}

/*
// ---- implementation of fs.FS ----
type wfs_t struct {
}

var (
	wfs = &wfs_t{}
)

func (wfs *wfs_t) Open(name string) (fs.File, error) {
	return Get(name)
}*/

// ---- implementation of fs.File ----
type File struct {
	method string
	url string
	jsonCall bool
	params interface{}
	headers map[string]string
	timeout int

	Result
}

func (f *File) Stat() (*FileInfo /*fs.FileInfo*/, error) {
	f.run()
	if f.Err != nil {
		return nil, f.Err
	}
	return &FileInfo{f: f}, nil
}

func (f *File) Read(p []byte) (int, error) {
	f.run()
	if f.Err != nil {
		return 0, f.Err
	}
	if f.Resp.Body == nil {
		return 0, os.ErrNotExist /*fs.ErrNotExist*/
	}
	return f.Resp.Body.Read(p)
}

func (f *File) Close() error {
	f.run()
	if f.Err != nil {
		return f.Err
	}
	if f.Resp.Body == nil {
		return os.ErrNotExist /*fs.ErrNotExist*/
	}
	return f.Resp.Body.Close()
}

func (f *File) setOptions(options []*Args) {
	if len(options) == 0 || options[0] == nil {
		return
	}

	option := options[0]
	f.params = option.Params
	f.headers = option.Headers
	f.timeout = option.Timeout
	f.jsonCall = option.JsonCall
}

func (f *File) run() {
	if f.Status > 0 {
		return
	}

	var call HttpFunc
	if f.jsonCall {
		call = PostJson
	} else {
		call = Wget
	}
	f.Status, _, f.Resp, f.Err = call(f.url, f.method, f.params, f.headers, Options{Timeout: f.timeout, DontReadRespBody: true})
}

// ---- implementation of fs.FileInfo ----
type FileInfo struct {
	f *File
	u *url.URL
	e error
}

// base name of the file
func (fi *FileInfo) Name() string {
	fi.parse()
	if fi.e != nil {
		return ""
	}
	if len(fi.u.Path) == 0 {
		return ""
	}
	return path.Base(fi.u.Path)
}

// length in bytes for regular files; system-dependent for others
func (fi *FileInfo) Size() int64 {
	return fi.f.Resp.ContentLength
}

// file mode bits
/*
func (fi *FileInfo) Mode() fs.FileMode {
	return fs.ModeSocket
}*/

// modification time
func (fi *FileInfo) ModTime() time.Time {
	t, _ := GetLastModified(fi.f.Resp)
	return t
}

// abbreviation for Mode().IsDir()
func (fi *FileInfo) IsDir() bool {
	return false
}

// underlying data source (can return nil)
func (fi *FileInfo) Sys() interface{} {
	return &fi.f.Result
}

func (fi *FileInfo) parse() {
	if fi.e != nil || fi.u != nil {
		return
	}
	fi.u, fi.e = url.Parse(fi.f.url)
}
