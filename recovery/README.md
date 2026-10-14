
### Recovery Sample
This sample implements a `RecoveryWorkflow` which is designed to restart all `TripWorkflow` executions which are currently
outstanding and replay all signals from previous run.  This is useful where a bad code change is rolled out which
causes workflows to get stuck or state is corrupted.

### Steps to run this sample
1) Run the following command to start worker
```
mkdir -p bin
tgo build -o bin/recovery-worker recovery/worker/main.go
./bin/recovery-worker
```
2) Run the following command to start trip workflow
```
tgo run recovery/starter/main.go
```
3) Run the following command to query trip workflow
```
tgo run recovery/query/main.go
```
4) Run the following command to send signal to trip workflow
```
tgo run recovery/signal/main.go -s '{"ID": "Trip1", "Total": 10}'
```
4) Run the following command to start recovery workflow
```
tgo run recovery/starter/main.go -w recovery_workflow -wt recoveryworkflow -i '{"Type": "TripWorkflow", "Concurrency": 2}'
```