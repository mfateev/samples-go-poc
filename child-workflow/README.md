## Child Workflow
<!-- @@@SNIPSTART samples-go-child-workflow-example-readme -->
Make sure the [Temporal Server is running locally](https://learn.temporal.io/getting_started/go/dev_environment/#set-up-a-local-temporal-service-for-development-with-temporal-cli).

From the root of the project, start a Worker:

```bash
mkdir -p bin
tgo build -o bin/child-workflow-worker child-workflow/worker/main.go
./bin/child-workflow-worker
```

Start the Workflow Execution:

```bash
tgo run child-workflow/starter/main.go
```
<!-- @@@SNIPEND -->
