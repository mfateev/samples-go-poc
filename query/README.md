This sample workflow demos how to use query API to get the current state of running workflow.

`query_workflow.go` shows how to setup a custom workflow query handler

`query_workflow_test.go` shows how to unit-test query functionality

Steps to run this sample:
1) Run a [Temporal service](../README.md#setup).
2) Run the following command to start worker
```
mkdir -p bin
tgo build -o bin/query-worker query/worker/main.go
./bin/query-worker
```
3) Run the following command to trigger a workflow execution. You should see workflowID and runID print out on screen.
```
tgo run query/starter/main.go
```
4) Run the following command to see current workflow state on the screen.
```
tgo run query/query/main.go
```
5) You could also specify the query type "__stack_trace" to dump the call stack for the workflow.
```
tgo run query/query/main.go -t __stack_trace
```
