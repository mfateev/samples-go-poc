This sample demonstrates how to setup a workflow to start after a delay.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
mkdir -p bin
tgo build -o bin/start-delay-worker start-delay/worker/main.go
./bin/start-delay-worker
```
3) Run the following command to start the example
```
tgo run start-delay/starter/main.go
```
