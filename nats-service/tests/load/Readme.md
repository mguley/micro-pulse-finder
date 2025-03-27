
---
1. Run the application and expose the metrics to Prometheus.
```bash
make run/nats-grpc-server
```

---
2. Run load tests, thus, we are stressing the application.
```bash
make run/load-tests
```

---
3. Collect data.

```bash
go tool pprof -alloc_objects http://localhost:50555/debug/pprof/heap?seconds=60
```

```bash
go tool pprof -alloc_space http://localhost:50555/debug/pprof/heap?seconds=60
```