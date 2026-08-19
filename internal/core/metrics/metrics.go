package metrics

//go:generate mockgen -source=metrics.go -destination mocks/mock_metrics.go -package mocks

type Metrics interface {
	Consumed()
	Processed()
	Requested()
	Response(status int)
}
