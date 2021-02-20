// +build go1.16

package wget

import (
	"testing"
	"io"
	"os"
	"fmt"
	"io/fs"
)

func TestFSGet(t *testing.T) {
	fp := Get("http://httpbin.org/get")
	fs_output(fp)
	fmt.Printf("\n---- done to TestFSGet() ---\n\n")
}

func TestFSJson(t *testing.T) {
	fp := Post("http://httpbin.org/post", &Args{Params: map[string]interface{}{"a": "b", "c": 1}, WithJson: true})
	fs_output(fp)
	fmt.Printf("\n---- done to TestFSJson() ---\n\n")
}

func fs_output(fp fs.File) {
	defer fp.Close()
	io.Copy(os.Stdout, fp)
}
