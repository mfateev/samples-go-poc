# Native Go Temporal samples

These applications use the native isolate SDK: context.Context, native goroutines,
channels, select, sync and history-driven timers inside workflows. Host clients,
activities and services use the Temporal Go SDK. [CATALOG.md](CATALOG.md) describes
included applications and remaining application ports; SDK features and sample
presence are separate concerns.

## Setup

Use sibling `golang-go`, `sdk-go-poc`, `temporal-go-sdk` and `samples-go-poc`
checkouts. Follow the SDK's
[clean source setup](https://github.com/mfateev/sdk-go-poc/blob/task/modify-go-runtime-for-isolates/README.md#install-tgo)
to check out matching revisions from `sdk-go-poc/toolchain/go-source.json` and
build/install tgo. Prerequisites are Git, Python 3.11+, a C compiler and bootstrap
Go 1.26.7 or newer; published precompiled bundles need no bootstrap Go.

Run from this repository root after placing the matching tgo on PATH:

```bash
tgo version
tgo env GOROOT GOWORK
tgo mod download
tgo test -p=2 ./... ./codec-server/... ./grpc-proxy/...
sample_build_dir=$(mktemp -d)
while IFS= read -r sample_package; do
    if [ -n "$sample_package" ]; then
        tgo build -p=2 -o "$sample_build_dir/program" "$sample_package" || exit
    fi
done < <(tgo list -f '{{if eq .Name "main"}}{{.ImportPath}}{{end}}' ./... ./codec-server/... ./grpc-proxy/...)
```

GOWORK should identify this checkout's `go.work`. It includes the root samples,
sibling SDK and nested codec-server/grpc-proxy modules; its replacement selects
the sibling host SDK fork. A root wildcard alone omits the nested modules. The loop builds one executable
at a time so native registration is actually generated; a multi-package build
only compiles packages and does not exercise that linking path.
Cloud collectors/deployments need additional infrastructure documented per sample.
Unit/build success does not establish those deployments work.

`tgo build`, `tgo test` and `tgo run` discover workflow markers, including test-only
entries. Run builds a temporary marked executable. Stock Go and `tgo install`
do not generate native entry registration. Marked test coverage mode is unsupported.

## First application

Install Temporal CLI separately and start a development server in one terminal:

```sh
temporal server start-dev --headless
```

From this repository root in another terminal:

```sh
mkdir -p bin
tgo build -o bin/helloworld-worker ./helloworld/worker
./bin/helloworld-worker
```

From a third terminal in the same checkout:

```sh
tgo run ./helloworld/starter Temporal
```

The starter prints its workflow ID and greeting. `TEMPORAL_ADDRESS` changes this
sample's default `localhost:7233` target. Other samples have their own documented
flags/task queues. A native entry is an ordinary concrete function:

```go
//temporal:workflow
func HelloWorld(ctx context.Context, name string) (string, error) {
    ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
        StartToCloseTimeout: 10*time.Second,
    })
    var greeting string
    err := workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, &greeting)
    return greeting, err
}
```

Use `github.com/mfateev/sdk-go-poc/workflow` and its worker wrapper. The context
is injected, never a serialized input. Host clients start native workflows by
registered string name. Activities may accept context first or omit it.

## Other small examples

| Application | Behavior |
| --- | --- |
| `choice-exclusive` | Host activity chooses an order activity; workflow propagates failures |
| `sleep-for-days` | Native select over email completion, 30-day timer, completion signal and cancellation |
| `goroutines` | Native fan-out activities with results returned in input order |
| `goroutine` | Native concurrent activity sequences |
| `random-streams` | Independent named streams with replay-stable draws |
| `dynamic-workflows` | Native fallback registration and runtime names |
| `logger-interceptor`, `opentelemetry`, `opentelemetry-v2`, `datadog` | Private workflow interception plus host observation/export |
| `codec-server`, `snappycompress`, `pso` | Host codecs and matching local value conversion |
| `safe_message_handler`, `shoppingcart` | Synchronized update state and handler draining |

For example, build a worker and run its starter from the repository root:

```sh
tgo build -o bin/goroutines-worker ./goroutines/worker
./bin/goroutines-worker
```

In another terminal: `tgo run ./goroutines/starter Ada Grace Linus`.
See each sample's README for additional services and inputs.

## Replay and tests

Build the combined small-sample replayer and export a completed history:

```sh
tgo build -o bin/replay ./replay
temporal workflow show --workflow-id YOUR_WORKFLOW_ID --output json > bin/history.json
./bin/replay HelloWorld bin/history.json
```

It also registers ExclusiveChoice, SleepForDays and GreetAll. Use the matching
registered name and input/result contract. Replay is a required application
change check, not a promise of general compatibility with every SDK history.
Do not remove executable saved-history fixtures as documentation cleanup.

Tests use the native SDK TestWorkflowEnvironment with host virtual time and
activities/mocks. Supply `t` or `s.T()` when creating the environment for cleanup.
Register workflows before SDK mock controls that need registration metadata.
Host mock callbacks cannot call native workflow APIs; ReturnIsolate selects a
marked native replacement. Decode copied update results through testsuite helpers.

Selected live tests are opt-in:

```sh
SAMPLES_DEV_SERVER=1 tgo test -count=1 -v ./helloworld
SAMPLES_DEV_SERVER=1 tgo test -count=1 -v ./dynamic-workflows
SAMPLES_DEV_SERVER=1 tgo test -count=1 -v ./nexus-standalone-activity ./nexus-standalone-operations
```

SDK [TESTING.md](https://github.com/mfateev/sdk-go-poc/blob/task/modify-go-runtime-for-isolates/TESTING.md)
explains deeper compiled/race/replay and cache-enabled/disabled checks.

## Workflow rules

Selected package globals and allocations are private to each execution. Pass
configuration as inputs or copied factory config; keep filesystem, environment,
network and secrets on the host. Google protobuf, native context propagation,
custom data/failure converters, dynamic handlers and workflow streams are supported
by the SDK. External payload references and variadic handlers remain outside its
native contract. Complete applications using every supported feature need not
be present here; see the catalog.

Supported maps/select/random/time are deterministic. Use native Go synchronization
correctly. Random bytes, including crypto-named APIs, are predictable replay data.
Queries/validators read workflow state and mutate only scratch values. Runtime
ownership/effect/resource faults fail the Workflow Task and cannot be recovered
into application progress. Non-yielding code has no hard-kill guarantee.

Wait for accepted update handlers with workflow.WaitHandlersFinished before
returning or continuing as new; separately drain admitted signals and coordinate
signal goroutines. Native timer Stop/Reset suppress local delivery but do not
cancel an already issued host timer. Matching serializers/codecs/propagators are
needed on both host and native sides.

For the compiler/runtime algorithms and a guide to reimplementing the POC, start
at the Go fork's
[documentation index](https://github.com/mfateev/golang-go/blob/task/modify-go-runtime-for-isolates/doc/isolates/README.md)
and the SDK's
[implementation guide](https://github.com/mfateev/sdk-go-poc/blob/task/modify-go-runtime-for-isolates/IMPLEMENTATION.md).
