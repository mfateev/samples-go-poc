This sample workflow demonstrates how to execute multiple activities in parallel and merge their results using futures.
The futures are awaited using Get method in the same order the activities are invoked. See `splitmerge-selector` sample
to see how to process them in the order of activity completion instead.

### Steps to run this sample:

1) Run a [Temporal service](../README.md#setup).
2) Run the following command to start the worker

```
mkdir -p bin
tgo build -o bin/splitmerge-future-worker splitmerge-future/worker/main.go
./bin/splitmerge-future-worker
```

3) Run the following command to start the example

```
tgo run splitmerge-future/starter/main.go
```
