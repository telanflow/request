package request

import (
	"encoding/xml"
	"html"
	"io/ioutil"
	"net/http"
	"encoding/json"
)

type Response struct {
	Status        string // e.g. "200 OK"
	StatusCode    int    // e.g. 200
	Proto         string // e.g. "HTTP/1.0"
	ProtoMajor    int    // e.g. 1
	ProtoMinor    int    // e.g. 0
	Body          []byte
	ContentLength int64
	Request       *http.Request
	Response	  *http.Response
}

func NewResponse(resp *http.Response) (*Response, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{
		Status:        resp.Status,
		StatusCode:    resp.StatusCode,
		Proto:         resp.Proto,
		ProtoMajor:    resp.ProtoMajor,
		ProtoMinor:    resp.ProtoMinor,
		Body:          body,
		ContentLength: resp.ContentLength,
		Request:       resp.Request,
		Response:      resp,
	}, nil
}

func (r *Response) GetHeader(key string) string {
	return r.Response.Header.Get(key)
}

func (r *Response) ContextType() string {
	return r.Response.Header.Get("Context-Type")
}

func (r *Response) Json(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

func (r *Response) Html() string {
	return html.UnescapeString(string(r.Body))
}

func (r *Response) Xml(v interface{}) error {
	return xml.Unmarshal(r.Body, v)
}

func (r *Response) String() string {
	return string(r.Body)
}

func (r *Response) Bytes() []byte {
	return r.Body
}
