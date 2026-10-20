# Ordered signals with native select

AwaitSignalsWorkflow receives Signal1, Signal2 and Signal3 on typed native
channels. One Go select loop records early arrivals and processes them in order.
Consumed signal channels become nil. A resettable time.Timer bounds the next
required signal by SignalToSignalTimeout and the remaining FromFirstSignalTimeout.
Context cancellation and timer expiry return errors. There is no SDK Selector or
Await coroutine in this implementation.

Follow [the root setup](../README.md#setup), start a local Temporal server, then
run these commands from the repository root in separate worker/starter terminals:

```sh
mkdir -p bin
tgo build -o bin/await-signals-worker ./await-signals/worker
./bin/await-signals-worker
```

```sh
tgo run ./await-signals/starter
```

Package tests: `tgo test ./await-signals`.
