This sample Workflow Definition demonstrates how to use multiple workflow safe goroutines (instead of native ones) to
process multiple sequences of activities in parallel.
In Temporal Workflow Definition, you should not use `go` keyword to start goroutines. Instead, you use the `workflow.Go`
function, which spawns a coroutine that is never run in parallel, but instead deterministically.

To see more information on goroutines and multithreading, see our
[docs on Go SDK multithreading](https://docs.temporal.io/develop/go/go-sdk-multithreading).

### Steps to run this sample:

1) Run a [Temporal Service](https://github.com/temporalio/samples-go/tree/main/#how-to-use)
2) Run the following command to start the worker

```
mkdir -p bin
tgo build -o bin/goroutine-worker goroutine/worker/main.go
./bin/goroutine-worker
```

3) Run the following command to start the example

```
tgo run goroutine/starter/main.go
```
