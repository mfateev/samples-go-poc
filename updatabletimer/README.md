# Updatable Timer Sample

A helper structure that supports blocking sleep that can be rescheduled at any moment.

Demonstrates:

* A native time.Timer with Stop/Reset
* A typed native time.Time signal channel
* Go select over timer, update and context cancellation channels

### Steps to run this sample:

1) Run a [Temporal service](../README.md#setup).
2) Run the following command to start the worker

```
mkdir -p bin
tgo build -o bin/updatabletimer-worker updatabletimer/worker/main.go
./bin/updatabletimer-worker
```

3) Run the following command to start the example

```
tgo run updatabletimer/starter/main.go
```

4)  Run the following command to update the timer wake-up time

```
tgo run updatabletimer/updater/main.go
```
Timer Stop/Reset suppress local delivery; an already issued host timer is not
canceled by this POC. Run commands from the repository root after setup.
