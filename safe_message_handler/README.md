# Synchronized message handlers

ClusterManagerWorkflow uses native notification channels to delay message
handling until startup and sync.Mutex to protect state across yielding activity
calls. Updates can interleave while waiting, so lock ownership matters even with
single-token dispatch. Assignment updates check idempotency before scheduling.
Before completion/Continue-As-New, workflow.WaitHandlersFinished waits for accepted
update completions. Signal goroutines require their own coordination. State is
carried explicitly into the next run.

Follow [the root setup](../README.md#setup), start a local Temporal server, then
run these commands from the repository root in separate worker/starter terminals:

```sh
mkdir -p bin
tgo build -o bin/safe_message_handler-worker ./safe_message_handler/worker
./bin/safe_message_handler-worker
```

```sh
tgo run ./safe_message_handler/starter
```

Package tests: `tgo test ./safe_message_handler`.
