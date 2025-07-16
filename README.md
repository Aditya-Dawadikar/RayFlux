# RayFlux – High-Throughput Pub/Sub Over WebSockets

RayFlux is a **real-time, stateless, and horizontally scalable publish-subscribe system** built for event-driven applications. It leverages WebSockets and Kubernetes to deliver low-latency broadcasting at scale.

---

## Features

- **High Throughput**: Achieves 5000+ RPS under 100ms average latency in constrained environments.
- **Stateless Nodes**: All FluxNodes and FluxBalancers are completely stateless, enabling elastic scaling.
- **WebSocket-First**: Native push-based architecture, no polling required.
- **Dual Buffering**: Optimized for near-real-time streaming with best-effort consistency.
- **S3-based Persistence**: Message snapshots for recovery and offline delivery.

---

## Architecture

- `FluxBalancer`: Stateless load balancer with consistent hashing for sticky routing.
- `FluxNode`: Publishes messages to all subscribers of a topic via WebSocket fanout.
- `Locust`: Used for load generation and benchmarking.
- `K8s`: Horizontal scaling via HPA with optional resource constraints.

---

## Benchmark Highlights

- **Tested with 200 concurrent users, 10 Topics**
- **FluxNode and FluxBalancer Pod Count: 5**
- **Average Latency: 86.91ms**
- **95%ile Latency: 160ms**
- **99%ile Latency: 210ms**
- **RPS: 5974**
- **0% failure**

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

Create Secrets
```
kubectl create secret generic flux-secrets \
--from-literal=AWS_ACCESS_KEY_ID=<AWS_ACCESS_KEY> \
--from-literal=AWS_SECRET_ACCESS_KEY=<AWS_SECRET_ACCESS_KEY>
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
