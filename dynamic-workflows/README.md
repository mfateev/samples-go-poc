# Dynamic Workflows and Activities

Adapted from [Temporal's Go sample at d378442](https://github.com/temporalio/samples-go/tree/d3784428122e0cd4b28e91a0cd86b589749abf28/dynamic-workflows).

A single `DynamicWorkflow` handles workflow types that have no named registration.
It uses `context.Context`, is marked `//temporal:workflow`, and decodes its arguments from
`converter.EncodedValues` inside the isolate. The requested type is available as
`workflow.GetInfo(ctx).WorkflowType.Name`.

Types beginning with `dynamic-activity` call `random-activity-name`. The host's
registered `DynamicActivity` handles that unregistered activity name and decodes
its arguments on the host. Other workflow types format a result directly. Named
workflow and activity registrations take precedence over these fallbacks.

The worker also retains the upstream `LoadDynamicRuntimeOptions` callback,
selecting versioning behavior by workflow type. The starter uses workflow types
whose behavior is unspecified, so it runs with the default worker options.

Follow the [tgo installation and repository setup](../README.md#1-install-tgo-on-linux-or-macos)
first. From the repository root, with `tgo` on PATH and a Temporal service
running at `localhost:7233`, build the worker and start it:

```sh
mkdir -p bin
tgo build -o bin/dynamic-workflows-worker ./dynamic-workflows/worker
./bin/dynamic-workflows-worker
```

In another terminal, run the host starter:

```sh
tgo run ./dynamic-workflows/starter
```

The three results are:

```text
DurableExecution - Hello - Temporal
dynamic-activity - Peanut - Butter
dynamic-activity123 - Jelly - Time
```

The nearby `dynamic` sample demonstrates invocation by name. This sample shows
fallback registration for names that were never individually registered.

Run the native unit tests, or opt into a development server test that also replays
the recorded histories:

```sh
tgo test ./dynamic-workflows/...
SAMPLES_DEV_SERVER=1 tgo test -count=1 ./dynamic-workflows
```
