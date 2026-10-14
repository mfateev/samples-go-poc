### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
mkdir -p bin
tgo build -o bin/greetings-worker greetings/worker/main.go
./bin/greetings-worker
```
3) Run the following command to start the example
```
tgo run greetings/starter/main.go
```
