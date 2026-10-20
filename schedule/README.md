This sample demonstrates how to setup a schedule to run a workflow

Steps to run this sample:
1) Run a [Temporal service](../README.md#setup).
2) Run
```
mkdir -p bin
tgo build -o bin/schedule-worker schedule/worker/main.go
./bin/schedule-worker
```
to start worker for schedule workflow.
3) Run
```
tgo run schedule/starter/main.go
```
to start a schedule to run a workflow every second.
