package metrics

//go:generate mockgen -source=metrics.go -destination mock_metrics.go -package mocks

type Metrics interface {
	Increment()
}
