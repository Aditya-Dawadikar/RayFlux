# RayFlux – High-Throughput Pub/Sub Over WebSockets

RayFlux is a **real-time, stateless, and horizontally scalable publish-subscribe system** built for event-driven applications. It leverages WebSockets and Kubernetes to deliver low-latency broadcasting at scale.

---

## Features

- **High Throughput**: Achieves 1600+ RPS under 100ms latency in constrained environments.
- **Stateless Nodes**: All FluxNodes and FluxBalancers are completely stateless, enabling elastic scaling.
- **WebSocket-First**: Native push-based architecture, no polling required.
- **Dual Buffering**: Optimized for near-real-time streaming with best-effort consistency.
- **S3-based Persistence (Planned)**: Message snapshots for recovery and offline delivery.
- **FluxManager (Planned)**: Control plane for managing clients, topics, and failover logic.

---

## Architecture

- `FluxBalancer`: Stateless load balancer with consistent hashing for sticky routing.
- `FluxNode`: Publishes messages to all subscribers of a topic via WebSocket fanout.
- `Locust`: Used for load generation and benchmarking.
- `K8s`: Horizontal scaling via HPA with optional resource constraints.

---

## Benchmark Highlights

- **Tested with 100–500 concurrent users**
- **0% failure up to 300 subscribers**
- **Autoscaling FluxNode & FluxBalancer pods (min 2, max 5)**
- **Real-time WebSocket delivery with event-based metrics**

> Resource Constraints Used:
```yaml
resources:
  requests:
    cpu: "100m"
    memory: "128Mi"
  limits:
    cpu: "500m"
    memory: "256Mi"
```

## Project Structure
```
RayFlux/
├── flux_node/           # WebSocket subscriber management and fanout
├── flux_balancer/       # Custom load balancer with sticky routing
├── benchmark/           # Locust-based benchmarking suite
├── deployments/         # Kubernetes YAMLs (Deployments, Services, HPA)
└── README.md
```

## Deployment
### Dev Deployment

Start Minikube
```
minikube start
```

Start FluxBalancer
```
cd flux_balancer
skaffold dev
```

Start FluxNode
```
cd flux_node
skaffold dev
```

FluxBalancer will be made available using a Kubernetes Service, to access the URL, run the following command

```
minikube service flux-balancer --url
```
