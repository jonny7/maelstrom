package service

//
//import (
//	"fmt"
//
//	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
//	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics"
//)
//
//type engine string
//
//const (
//	Prometheus engine = "prometheus"
//	Nop        engine = "nop"
//)
//
//func NewMetrics(e engine, logger logging.Logger) (metrics.Metrics, error) {
//	switch e {
//	case Prometheus:
//		return metrics.NewNop()
//	case Nop:
//		return metrics.NewNop()
//	default:
//		return nil, fmt.Errorf("unrecognized consumer type provider, expected either kafka or nop, but receieved: %s", t)
//	}
//}
