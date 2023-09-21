package api

import "net/http"

type API interface {
	CreateSubscriber() http.HandlerFunc
}
