### Steps to run this sample:
1) Configure a [Temporal Server](https://github.com/temporalio/samples-go/tree/main/#how-to-use) (such as Temporal Cloud) with apiKey.

2) Run the following command to start the worker
```
mkdir -p bin
tgo build -o bin/helloworld-apiKey-worker ./helloworld-apiKey/worker
./bin/helloworld-apiKey-worker \
    -target-host my.namespace.tmprl.cloud:7233 \
    -namespace my.namespace \
    -api-key CLIENT_API_KEY
```
3) Run the following command to start the example
```
tgo run ./helloworld-apiKey/starter \
    -target-host my.namespace.tmprl.cloud:7233 \
    -namespace my.namespace \
    -api-key CLIENT_API_KEY
```