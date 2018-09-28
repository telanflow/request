package request

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type redirectHandler func(req *http.Request, via []*http.Request) error

type Request struct {
	Header        http.Header     // 请求头
	Host		  string		  // 自定义http协议域名
	url           string          // 请求地址
	method        string          // 请求方式
	params        io.Reader       // 请求参数
	redirect      redirectHandler // 自定义重定向
	redirectTimes int             // 重定向次数  默认5次
	client        *http.Client    // 客户端
	transport	  *http.Transport

	execTime	  time.Duration	  // 执行时间
}

func New() *Request {
	return &Request{
		Header:        make(http.Header),
		redirectTimes: 5,
		client:        &http.Client{},
		transport:	   &http.Transport{},
	}
}

// Get请求
func (r *Request) Get(oUrl string, oParams interface{}) (*Response, error) {
	return r.Suck(http.MethodGet, oUrl, oParams)
}

// Post请求
func (r *Request) Post(oUrl string, oParams interface{}) (*Response, error) {
	return r.Suck(http.MethodPost, oUrl, oParams)
}

// Post表单请求
func (r *Request) PostForm(oUrl string, oParams interface{}) (*Response, error) {
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r.Suck(http.MethodPost, oUrl, oParams)
}

// Put请求
func (r *Request) Put(oUrl string, oParams interface{}) (*Response, error) {
	return r.Suck(http.MethodPut, oUrl, oParams)
}

// Head请求
func (r *Request) Head(oUrl string, oParams interface{}) (*Response, error) {
	return r.Suck(http.MethodHead, oUrl, oParams)
}

// Http下载文件
func (r *Request) Download(oUrl string, toFile string) error {
	resp, err := r.Suck(http.MethodGet, oUrl, nil)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(toFile, resp.Body, 0644)
}

// 设置重定向次数
func (r *Request) SetRedirectTimes(t int) *Request {
	r.redirectTimes = t
	return r
}

// 自定义重定向Handler
func (r *Request) SetRedirectHandler(handler redirectHandler) *Request {
	r.redirect = handler
	return r
}

// 设置代理地址
// proxy	http://127.0.0.1:8081
func (r *Request) SetProxy(proxy string) *Request {
	u, err := url.Parse(proxy)
	if err != nil {
		panic(err.Error())
	}
	r.transport.Proxy = http.ProxyURL(u)
	return r
}

// 指定创建TCP连接的拨号函数
func (r *Request) DialContext(fn func(ctx context.Context, network, addr string) (net.Conn, error)) *Request {
	r.transport.DialContext = fn
	return r
}

// 自定义请求服务器 IP:PORT - 废弃
// address	127.0.0.1:80、127.0.0.1:443
func (r *Request) _cdn(address string) *Request {
	d := new(net.Dialer)
	r.transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return d.DialContext(ctx, "tcp", address)
	}
	return r
}

// 设置域名
// GET /index HTTP/1.1
// Host: 域名
// ....
func (r *Request) SetHost(host string) *Request {
	r.Host = host
	return r
}

// SSL 不校验服务器证书
func (r *Request) SetInsecureSkipVerify(s bool) *Request {
	r.transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: s,
	}
	return r
}
// 设置CookieJar
func (r *Request) SetCookieJar(jar http.CookieJar) *Request {
	r.client.Jar = jar
	return r
}

// 设置Referer
func (r *Request) SetReferer(referer string) *Request {
	r.Header.Set("Referer", referer)
	return r
}

// 设置字符集
func (r *Request) SetCharset(charset string) *Request {
	r.Header.Set("Accept-Charset", charset)
	return r
}

// 设置UserAgent
func (r *Request) SetUserAgent(ua string) *Request {
	r.Header.Set("User-Agent", ua)
	return r
}

// 设置请求头
func (r *Request) SetHeader(key, val string) *Request {
	r.Header.Set(key, val)
	return r
}

// 获取请求头
func (r *Request) GetHeader(key string) string {
	return r.Header.Get(key)
}

// 设置请求头
func (r *Request) SetHeaders(header http.Header) *Request {
	r.Header = header
	return r
}

// 获取上一次请求执行时间
func (r *Request) GetExecTime() time.Duration {
	return r.execTime
}

// 新增请求头
func (r *Request) AddHeaders(header http.Header) *Request {
	for key, val := range header {
		for _, v := range val {
			r.Header.Set(key, v)
		}
	}
	return r
}

// 设置超时限制
func (r *Request) SetTimeout(t time.Duration) *Request {
	r.client.Timeout = t
	return r
}

// 嘬取数据
// oMethod	请求类型
// oUrl		目标地址
// oParams	请求参数
func (r *Request) Suck(oMethod, oUrl string, oParams interface{}) (*Response, error) {
	r.url = oUrl
	r.method = oMethod
	r.params = buildParams(oParams)
	return r.transmission()
}

// 重置默认值
func (r *Request) Reset() *Request {
	r.transport = &http.Transport{}
	r.Header = make(http.Header)
	r.redirect = nil
	r.redirectTimes = 5
	return r
}

// 数据接口传输层
func (r *Request) transmission() (*Response, error) {
	if r.params != nil {
		if r.method == http.MethodGet {
			parsedURL, err := url.Parse(r.url)
			if err != nil {
				return nil, err
			}

			str, _ := ioutil.ReadAll(r.params)
			r.url = addQueryParams(parsedURL, string(str))
			r.params = nil
		}
	}

	req, err := http.NewRequest(r.method, r.url, r.params)
	if err != nil {
		return nil, err
	}

	// Host设置
	if r.Host != "" {
		req.Host = r.Host
	}

	// 请求头设置
	if r.Header != nil && len(r.Header) > 0 {
		req.Header = cloneHeader(r.Header)
	}

	// 网络设置
	r.client.Transport = r.transport

	// 重定向
	r.client.CheckRedirect = r.checkRedirect()

	// Do
	startTime := time.Now()
	resp, err := r.client.Do(req)
	r.execTime = time.Now().Sub(startTime)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return NewResponse(resp)
}

// 默认重定向Handler
func (r *Request) checkRedirect() redirectHandler {
	fn := func(req *http.Request, via []*http.Request) error {
		if len(via) >= r.redirectTimes {
			return errors.New("stopped after " + strconv.Itoa(r.redirectTimes) + " redirects")
		}
		return nil
	}

	if r.redirect != nil {
		fn = r.redirect
	}

	return fn
}

func buildParams(oParams interface{}) (params io.Reader) {

	switch t := oParams.(type) {
	case string:
		params = strings.NewReader(t)

	case []string:
		var buffer bytes.Buffer
		for i := 0; i < len(t); i++ {
			buffer.WriteString(t[i] + "&")
		}
		params = bytes.NewReader(bytes.Trim(buffer.Bytes(), "&"))

	case []byte:
		params = bytes.NewReader(t)

	case map[string]string:
		val := make(url.Values)
		for key, value := range t {
			val.Set(key, value)
		}
		params = strings.NewReader(val.Encode())

	case url.Values:
		params = strings.NewReader(t.Encode())

	case io.Reader:
		params = t

	default:
		params = nil
	}
	return
}

func addQueryParams(parsedURL *url.URL, parsedQuery string) string {
	return strings.Join([]string{strings.Replace(parsedURL.String(), "?"+parsedURL.RawQuery, "", -1), parsedQuery}, "?")
}

func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}