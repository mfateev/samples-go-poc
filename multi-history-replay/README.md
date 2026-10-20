This sample demonstrates getting multiple workflow histories and replaying them.

### Steps to run this sample:
1) Run a [Temporal service](../README.md#setup).
2) Run the following command to start the worker
```shell script
mkdir -p bin
tgo build -o bin/multi-history-replay-worker multi-history-replay/worker/main.go
./bin/multi-history-replay-worker
```
3) Run the following command to start mutiple parallel workflows. This simulates having existing workflows
on the server.
```shell script
tgo run multi-history-replay/starter/main.go
```
4) Run the following command to replay the histories generated in step 3
```shell script
tgo run multi-history-replay/replayer/main.go
```
