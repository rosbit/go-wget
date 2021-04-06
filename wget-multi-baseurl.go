package wget

import (
	wr "github.com/mroth/weightedrand"
	// "path"
	"fmt"
	"io"
	"time"
	"net/http"
	"math/rand"
)

type baseItem struct {
	baseUrl string
	weight  uint
	lastAccessTime int64
}

func BaseItem(baseUrl string, weight ...uint) baseItem {
	getWeight := func() uint {
		if len(weight)>0 {
			return weight[0]
		}
		return 0
	}

	return baseItem {
		baseUrl: baseUrl,
		weight: getWeight(),
		lastAccessTime: time.Now().Unix(),
	}
}

type BaseUrl struct {
	baseItems []baseItem
	chooser *wr.Chooser
	rd *rand.Rand
	lastOKIndex int
}

func NewBaseUrl(baseItem ...baseItem) (b *BaseUrl, err error) {
	if len(baseItem) == 0 {
		err = fmt.Errorf("no items")
		return
	}

	b = &BaseUrl{
		baseItems: baseItem,
	}

	if err = b.caclWeights(); err != nil {
		return
	}

	b.createRandChooser()
	b.lastOKIndex = -1
	return
}

func (b *BaseUrl) HttpCall(uri, method string, params interface{}, header map[string]string, options ...Options) (status int, content []byte, resp *http.Response, err error) {
	if isHttpUrl(uri) {
		return newRequest(uri, 0, options...).Run(uri, method, params, header)
	}

	var paramsReader io.ReadSeeker
	if uri, method, paramsReader, header, err = adjustHttpArgs(uri, method, params, header); err != nil {
		return
	}

	return b.run(uri, method, paramsReader, header, options...)
}

func (b *BaseUrl) JsonCall(uri, method string, params interface{}, header map[string]string, options ...Options) (status int, content []byte, resp *http.Response, err error) {
	if isHttpUrl(uri) {
		return newRequest(uri, 0, options...).PostJson(uri, method, params, header)
	}

	var paramsReader io.ReadSeeker
	if method, paramsReader, header, err = adjustJsonArgs(method, params, header); err != nil {
		return
	}

	return b.run(uri, method, paramsReader, header, options...)
}

func (b *BaseUrl) GetWithBody(uri string, params interface{}, header map[string]string, options ...Options) (status int, content []byte, resp *http.Response, err error) {
	if isHttpUrl(uri) {
		return newRequest(uri, 0, options...).GetUsingBodyParams(uri, params, header)
	}

	var paramsReader io.ReadSeeker
	if _, _, paramsReader, header, err = adjustHttpArgs(uri, http.MethodPost, params, header); err != nil {
		return
	}

	return b.run(uri, http.MethodGet, paramsReader, header, options...)
}

func (b *BaseUrl) run(uri, method string, paramsReader io.ReadSeeker, header map[string]string, options ...Options) (status int, content []byte, resp *http.Response, err error) {
	startIdx := b.pick()
	for i:=startIdx; i<len(b.baseItems); i++ {
		url := fmt.Sprintf("%s%s", b.baseItems[i].baseUrl, uri)
		if paramsReader != nil {
			paramsReader.Seek(0, io.SeekStart)
		}
		status, content, resp, err = newRequest(url, 0, options...).run(url, method, paramsReader, header)
		if err == nil {
			return
		}
	}
	for i:=0; i<startIdx; i++ {
		url := fmt.Sprintf("%s%s", b.baseItems[i].baseUrl, uri)
		if paramsReader != nil {
			paramsReader.Seek(0, io.SeekStart)
		}
		status, content, resp, err = newRequest(url, 0, options...).run(url, method, paramsReader, header)
		if err == nil {
			return
		}
	}

	return
}

func (b *BaseUrl) pick() int {
	return b.chooser.PickSource(b.rd).(int)
}

func (b *BaseUrl) caclWeights() error {
	if !isHttpUrl(b.baseItems[0].baseUrl) {
		return fmt.Errorf("prefix of base URL %s is not http or https", b.baseItems[0].baseUrl)
	}
	allNoWeight := (b.baseItems[0].weight == 0)
	c := len(b.baseItems)

	for i:=1; i<c; i++ {
		bi := b.baseItems[i]
		if !isHttpUrl(bi.baseUrl) {
			return fmt.Errorf("prefix of base URL %s is not http or https", bi.baseUrl)
		}
		if bi.weight > 0 {
			if allNoWeight {
				return fmt.Errorf("weights before item #%d expected", i)
			}
		} else {
			if !allNoWeight {
				return fmt.Errorf("weight for item #%d(%s) expected", i, bi.baseUrl)
			}
		}
	}

	if allNoWeight {
		for i, _ := range b.baseItems {
			bi := &b.baseItems[i]
			bi.weight = 20 // any number greater than 0 is ok
		}
	}
	return nil
}

func (b *BaseUrl) createRandChooser() {
	choices := make([]wr.Choice, len(b.baseItems))
	for i, bi := range b.baseItems {
		choices[i].Item = i
		choices[i].Weight = bi.weight
	}

	b.rd = rand.New(rand.NewSource(time.Now().UnixNano()))
	b.chooser, _ = wr.NewChooser(choices...)
	fmt.Printf("b: %#v\n", b)
}
