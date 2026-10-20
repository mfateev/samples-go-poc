### Steps to run this sample:
1) Run a [Temporal service](../README.md#setup).
2) Run the following command to start the worker
```
mkdir -p bin
tgo build -o bin/metrics-worker metrics/worker/main.go
./bin/metrics-worker
```
3) Run the following command to start the example
```
tgo run metrics/starter/main.go
```
4) Check metrics at http://localhost:9090/metrics (this is where the Prometheus agent scrapes it from).
