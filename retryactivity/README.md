### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
mkdir -p bin
go build -o bin/retryactivity-worker retryactivity/worker/main.go
./bin/retryactivity-worker
```
3) Run the following command to start the example
```
go run retryactivity/starter/main.go
```
