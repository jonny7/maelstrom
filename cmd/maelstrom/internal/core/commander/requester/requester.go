package requester

import "time"

type Event struct {
	Key       []byte
	Value     []byte
	Headers   []Header
	Timestamp time.Time
}

type Header struct {
	Key   string
	Value []byte
}
