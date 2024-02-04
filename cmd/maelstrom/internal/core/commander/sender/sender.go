package sender

import "net/http"

//go:generate mockgen -source=sender.go -destination mocks/mock_sender.go -package mocks

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}
