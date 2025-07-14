# RayFlux Benchmark Suite

This folder contains all performance benchmarking scripts and configuration files used to evaluate the scalability and latency characteristics of **RayFlux**, a real-time publish-subscribe system built on Kubernetes.

---

## 🔧 Folder Structure

```
benchmark/
├── locustfile.py # Locust user classes for publisher and subscriber
├── requirements.txt # Python dependencies for Locust and WebSocket
├── README.md # This file
```

---

## Running Benchmarks

### 1. Install Dependencies
```
pip install -r requirements.txt
```

### 2. Launch Locust

```
locust -f locustfile.py
```


Then open http://localhost:8089

## Creating Benchmarks

`locustfile.py` supports 2 users classes - Publisher and Subscriber. Publisher is based on Locust's HTTPUser and Subscriber is based on Websockets. User distribution is 80-20 (20 Publishers, 80 Subscribers) and 10 topics.

Configure the total user count, the host (URL where FluxBalancer can be reached) and duration of the stress test.

RayFlux has Horizontal Pod Autoscaling enabled, feel free to change the minimum and maximum container counts, to stress load with auto-scaling in action.