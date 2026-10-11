This sample demonstrates how to setup a schedule to run a workflow

Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run
```
mkdir -p bin
go build -o bin/schedule-worker schedule/worker/main.go
./bin/schedule-worker
```
to start worker for schedule workflow.
3) Run
```
go run schedule/starter/main.go
```
to start a schedule to run a workflow every second.
