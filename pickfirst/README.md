### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
mkdir -p bin
tgo build -o bin/pickfirst-worker pickfirst/worker/main.go
./bin/pickfirst-worker
```
3) Run the following command to start the example
```
tgo run pickfirst/starter/main.go
```
