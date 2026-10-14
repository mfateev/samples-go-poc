# Invoking Activities by Name

The purpose of this sample is to demonstrate invocation of workflows and activities using name
rather than strongly typed function.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
mkdir -p bin
tgo build -o bin/dynamic-worker dynamic/worker/main.go
./bin/dynamic-worker
```
3) Run the following command to start the example
```
tgo run dynamic/starter/main.go
```
