package http

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type FakeHTTP struct{}

// NewFakeHTTP returns a mock HTTP client, which returns random HTTP responses. This is useful for testing metrics and charting
func NewFakeHTTP() FakeHTTP {
	return FakeHTTP{}
}

var statuses = []int{200, 204, 404, 500}

func (f FakeHTTP) Do(_ *http.Request) (*http.Response, error) {
	d := rand.Intn(800-600) + 600
	ms, err := time.ParseDuration(fmt.Sprintf("%dms", d))
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(ms)
	r := rand.Intn(4 - 0)
	return &http.Response{StatusCode: statuses[r]}, nil
}
