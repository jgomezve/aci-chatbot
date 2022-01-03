package mocks

import (
	"net/http"
	"time"
)

type MockClient struct {
	DoFunc  func(req *http.Request) (*http.Response, error)
	Timeout time.Duration
}

var (
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}
