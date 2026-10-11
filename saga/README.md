### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
mkdir -p bin
go build -o bin/saga-worker saga/worker/main.go
./bin/saga-worker
```
3) Run the following command to start the example
```
go run saga/start/main.go
```

Based on https://github.com/temporalio/money-transfer-project-template-go
