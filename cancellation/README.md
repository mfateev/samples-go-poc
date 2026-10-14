## Cancellation
<!-- @@@SNIPSTART samples-go-cancellation-readme -->
Make sure the [Temporal Server is running locally](https://learn.temporal.io/getting_started/go/dev_environment/#set-up-a-local-temporal-service-for-development-with-temporal-cli).

From the root of the project, start a Worker:

```bash
mkdir -p bin
tgo build -o bin/cancellation-worker cancellation/worker/main.go
./bin/cancellation-worker
```

Start the Workflow Execution:

```bash
tgo run cancellation/starter/main.go
```

Cancel the Workflow Execution:

```bash
tgo run cancellation/cancel/main.go
```
<!-- @@@SNIPEND -->
