# Maelstrom

A distributed load-testing tool: nodes cluster over serf, consume or generate events, and fire them as HTTP load at a target.

## Running stand-alone

The only required setting outside k8s is `SEED_NODES` — its default points at the k8s seed node's DNS name, which won't resolve locally. Seed the node with itself:

```sh
SEED_NODES=127.0.0.1:7946 go run ./cmd/maelstrom
```

The k8s client warnings on startup are expected outside a cluster — scaling falls back to a no-op fake. The event generator is currently the built-in fake (see `cmd/maelstrom/main.go`), so no Kafka is needed.

| endpoint            | address                       |
|---------------------|-------------------------------|
| UI                  | http://localhost:4000         |
| API                 | http://localhost:3000/api     |
| API health          | http://localhost:3000/api/healthz |
| Prometheus metrics  | http://localhost:2112         |
| serf / RPC          | 7946 / 7373                   |

## Configuration

All configuration is via env vars.

| var | default | |
|-----|---------|---|
| `SEED_NODES` | `maelstrom-0.maelstrom-svc.default.svc.cluster.local:7946` | serf addresses to join; set to own address for a single node |
| `SERF_SERVICE` | `$HOSTNAME:$SERF_PORT` | serf bind host:port |
| `SERF_PORT` | `7946` | |
| `RPC_PORT` | `7373` | |
| `NODE_NAME` | `$HOSTNAME` | unique node name in the cluster |
| `API_HOST` / `API_PORT` | `0.0.0.0` / `3000` | |
| `UI_HOST` / `UI_PORT` | `0.0.0.0` / `4000` | |
| `UI_PROMETHEUS_ADDR` | `http://localhost:9090` | where the UI reads chart metrics from |
| `UI_PROMETHEUS_USER` / `UI_PROMETHEUS_PASS` | `admin` / `admin` | |
| `K8S_NAMESPACE` | `default` | namespace of the StatefulSet to scale |
| `K8S_STATEFULSET` | `maelstrom` | StatefulSet name to scale |
| `KAFKA_CONSUMER_BROKERS`, `KAFKA_CONSUMER_TOPICS` (required), `KAFKA_CONSUMER_GROUP`, `KAFKA_USER`, `KAFKA_PASSWORD`, `KAFKA_ENABLE_TLS` | | only used when the real Kafka generator is wired in `main.go` |

## Code generation

The OpenAPI spec lives at `maelstrom.yaml`; the generated chi server and types are used by the API adapter only.

```sh
oapi-codegen -generate chi-server -package maelstrom maelstrom.yaml > cmd/maelstrom/internal/maelstrom/maelstrom.gen.go
oapi-codegen -generate types -package maelstrom maelstrom.yaml > cmd/maelstrom/internal/maelstrom/types.gen.go
```

Mocks are generated with mockgen via `go generate ./...`.
