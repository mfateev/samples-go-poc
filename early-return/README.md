### Early-Return Sample

This sample demonstrates an early-return from a workflow.

By utilizing Update-with-Start, a client can start a new workflow and synchronously receive
a response mid-workflow, while the workflow continues to run to completion.

See [shopping cart](https://github.com/temporalio/samples-go/tree/main/shoppingcart)
for Update-with-Start being used for lazy initialization.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).

   NOTE: frontend.enableExecuteMultiOperation=true must be configured for the server
in order to use Update-with-Start. For example:
```
temporal server start-dev --dynamic-config-value frontend.enableExecuteMultiOperation=true
```

2) Run the following command to start the worker
```
mkdir -p bin
tgo build -o bin/early-return-worker early-return/worker/main.go
./bin/early-return-worker
```
3) Run the following command to start the example
```
tgo run early-return/starter/main.go
```
