### Steps to run this sample:
1) Run a [Temporal service](../README.md#setup).
2) Run the following command to start the worker
```
mkdir -p bin
tgo build -o bin/timer-worker timer/worker/main.go
./bin/timer-worker
```
3) Run the following command to start the example
```
tgo run timer/starter/main.go
```
