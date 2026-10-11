## Parallel Activities

This sample demonstrates how to kick off multiple activities in parallel and
then synchronously await their results once all have completed.

<!-- @@@SNIPSTART samples-go-branch-readme -->
Make sure the [Temporal Server is running locally](https://learn.temporal.io/getting_started/go/dev_environment/#set-up-a-local-temporal-service-for-development-with-temporal-cli).

From the root of the project, start a Worker:

```bash
mkdir -p bin
go build -o bin/branch-worker branch/worker/main.go
./bin/branch-worker
```

Start the Workflow Execution:

```bash
go run branch/starter/main.go
```
<!-- @@@SNIPEND -->
