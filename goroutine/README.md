# Native goroutines

SampleGoroutineWorkflow starts an ordinary Go goroutine for each activity
sequence. Each runs Step1 then Step2 with SDK futures and publishes an indexed
result through a buffered native channel. The root collects results in input
order, propagates activity errors and observes ctx.Done(). Deterministic isolate
dispatch permits native go statements; host activities run concurrently normally.

Follow [the root setup](../README.md#setup), start a local Temporal server, then
run these commands from the repository root in separate worker/starter terminals:

```sh
mkdir -p bin
tgo build -o bin/goroutine-worker ./goroutine/worker
./bin/goroutine-worker
```

```sh
tgo run ./goroutine/starter
```

Package tests: `tgo test ./goroutine`.
