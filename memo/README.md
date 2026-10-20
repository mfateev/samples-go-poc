### Steps to run this sample:

1) Before running this, you need to run [Temporal Server 1.18+](../README.md#setup).

2) Run the following command to start the worker:
```
mkdir -p bin
tgo build -o bin/memo-worker ./memo/worker/main.go
./bin/memo-worker
```

3) Run the following command to start the example:
```
tgo run ./memo/starter/main.go
```

4) Observe memo in the worker log:
```
...
2022/09/12 12:36:20 INFO  Current memo values:
description=Test upsert memo workflow
 Namespace default TaskQueue memo WorkerID 18670@Rodrigo-Zhous-MacBook-Pro.local@ WorkflowType MemoWorkflow WorkflowID memo_b1326cd2-5123-4a61-b417-435285dd7214 RunID 38128800-4c41-4d85-ba1a-a26730ebcb47 Attempt 1
2022/09/12 12:36:20 INFO  Workflow completed. Namespace default TaskQueue memo WorkerID 18670@Rodrigo-Zhous-MacBook-Pro.local@ WorkflowType MemoWorkflow WorkflowID memo_b1326cd2-5123-4a61-b417-435285dd7214 RunID 38128800-4c41-4d85-ba1a-a26730ebcb47 Attempt 1
```
