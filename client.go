package request

import (
	"net/http"
	"sync"
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// Http client pool
var CtPool = sync.Pool {
	New: func() interface{} {
		return &http.Client{}
	},
}