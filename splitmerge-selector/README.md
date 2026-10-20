# Merge activities in completion order

This workflow schedules activities and starts native goroutines to wait on
their futures. A completion channel and native select collect results as they
arrive, alongside cancellation. The directory name describes the sample family;
this port uses Go channels, not an SDK Selector. Compare splitmerge-future, which
waits on futures in scheduling order.

Follow [the root setup](../README.md#setup), start a local Temporal server, then
run these commands from the repository root in separate worker/starter terminals:

```sh
mkdir -p bin
tgo build -o bin/splitmerge-selector-worker ./splitmerge-selector/worker
./bin/splitmerge-selector-worker
```

```sh
tgo run ./splitmerge-selector/starter
```

Package tests: `tgo test ./splitmerge-selector`.
