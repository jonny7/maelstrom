package gateway

type Orchestrator interface {
	HealthHandler() bool
}
