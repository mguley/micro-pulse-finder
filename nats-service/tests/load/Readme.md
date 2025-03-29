##### Running the Application

1. Run the application and expose the metrics to Prometheus.
```bash
make run/nats-grpc-server
```

2. Run load tests, thus, we are stressing the application.
```bash
make run/load-tests
```
```bash
make run/load-subscribe
```

##### Memory Profiling and Performance Testing

###### Initial Profiling

1. Collect baseline memory allocation data
```bash
# Collect allocation objects count
go tool pprof -alloc_objects http://localhost:50555/debug/pprof/heap?seconds=60

# Collect allocation space (in bytes)
go tool pprof -alloc_space http://localhost:50555/debug/pprof/heap?seconds=60
```

2. Save the profile data for comparison
```bash
# Download the heap profile directly
curl -s http://localhost:50555/debug/pprof/heap?seconds=60 > before.prof
```

###### After Code Optimization
1. After implementing optimizations, run the application and load tests again.
2. Collect new memory allocation data

```bash
# Download the heap profile after optimization
curl -s http://localhost:50555/debug/pprof/heap?seconds=60 > after.prof
```

###### Comparing Results

Compare the before and after profiles to quantify the improvement

```bash
# Compare the two profiles to see improvements
go tool pprof -alloc_space -base before.prof after.prof
```

In the pprof interactive session, you can:
- Use `top` to see the biggest changes
- Use `list Publish` to focus on our specific handler
- Use `web` to create a visual graph
- Use `diff` to see a text-based difference report