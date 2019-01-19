package request

import (
	"context"
	"net"
	"net/http"
	"time"
)

// Request Interface
type RequestIF interface {
	Get(oUrl string, oParams... interface{}) (*Response, error)
	Post(oUrl string, oParams... interface{}) (*Response, error)
	PostForm(oUrl string, oParams... interface{}) (*Response, error)
	Put(oUrl string, oParams... interface{}) (*Response, error)
	Head(oUrl string, oParams... interface{}) (*Response, error)
	Options(oUrl string, oParams... interface{}) (*Response, error)
	Delete(oUrl string, oParams... interface{}) (*Response, error)
	Suck(oMethod, oUrl string, oParams... interface{}) (*Response, error)
	Download(oUrl string, toFile string) error

	SetHost(host string) RequestIF
	GetHost(host string) string
	SetProxy(proxy string) RequestIF
	SetRedirectTimes(t int) RequestIF
	SetRedirectHandler(handler redirectHandler) RequestIF
	SetInsecureSkipVerify(s bool) RequestIF
	SetCookieJar(jar http.CookieJar) RequestIF
	DialContext(fn func(ctx context.Context, network, addr string) (net.Conn, error)) RequestIF

	SetReferer(referer string) RequestIF
	SetCharset(charset string) RequestIF
	SetUserAgent(ua string) RequestIF
	SetHeader(key, val string) RequestIF
	GetHeader(key string) string
	SetHeaders(header http.Header) RequestIF
	AddHeaders(header http.Header) RequestIF
	SetTimeout(t time.Duration) RequestIF
	SetDialTimeout(t time.Duration) RequestIF
	SetTLSTimeout(t time.Duration) RequestIF
	SetResponseHeaderTimeout(t time.Duration) RequestIF

	Reset() RequestIF
	ExecTime() time.Duration
}

// New request interface
func New() RequestIF {
	return NewRequest()
}