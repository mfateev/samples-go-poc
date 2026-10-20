### Steps to run this sample:
1) Run a [Temporal service](../README.md#setup).
2) Run the following command to start the worker
```
mkdir -p bin
tgo build -o bin/greetingslocal-worker greetingslocal/worker/main.go
./bin/greetingslocal-worker
```
3) Run the following command to start the example
```
tgo run greetingslocal/starter/main.go
```
