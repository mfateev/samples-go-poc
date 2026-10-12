# Upstream sample migration

Audited and ported from [temporalio/samples-go at d378442](https://github.com/temporalio/samples-go/tree/d3784428122e0cd4b28e91a0cd86b589749abf28), using SDK v1.49.0 and the isolate SDK workspace checkout.

**76 of 83 upstream sample groups are present.** The other 7 depend on missing native features and are listed below. The `goroutines` and `random-streams` examples are additional POC examples, outside that upstream count. Directory presence describes migration scope, not a claim that cloud services or deployments were exercised.

## Native Go patterns

Workflow entries use `//go:isolate` and `context.Context`. Workflow concurrency uses ordinary `go`, channels, `select`, `sync.Mutex`, and `sync.WaitGroup`. Timers use `time.Timer`, `time.After`, or context-aware timer waits; time and randomness are replay deterministic in isolates. Map iteration is deterministic and requires no workflow-checker suppression or sorted-key helper.

`workflow.GetSignalChannel[T]` returns `<-chan T`. Native futures can be waited on with `Get` or selected through `ToChannel`. `workflow.WaitHandlersFinished(ctx)` waits until accepted update completions have reached the host; samples use it before returning or continuing as new. Application signal goroutines must be coordinated separately. Signals are proactively decoded into native channels before application code resumes, including signals queued before channel registration.

Build workflow workers with the modified toolchain's `go build` and run the resulting executables; `go run` does not emit isolate entry metadata. Host starters can use `go run`. The root setup instructions put the modified compiler on `PATH` for individual sample commands. Workers use the POC `worker.New`. Host clients start native workflows by their registered string names: the upstream SDK client recognizes only `workflow.Context` when validating a function reference. Host activities, clients, codecs, server integrations, and SDK worker options retain the upstream SDK APIs. Native interceptor and serializer factories are configured explicitly. Host-only logging/telemetry/codec initialization lives in separate packages where necessary.

The ports target new executions. Replacing SDK coroutines, side-effect markers, timers, or child-ID generation can change command histories. Existing histories need explicit replay/version migration validation; this catalog does not claim general compatibility with every upstream history.

## Verification

Run from this checkout with sibling `golang-go` and `sdk-go-poc` repositories:

```sh
../golang-go/bin/go test -p 3 ./... ./codec-server/... ./grpc-proxy/...
```

The compiler currently emits isolate entry metadata for built executables rather than ordinary `go test` binaries. `internal/isotest` builds upstream workflow suites in temporary directories and executes their real native workflow definitions against the SDK test environment. Original suites live in `testdata/upstream`; `migration_test.go` launches them. The catalog of native packages makes builds share a consistent compiler cache; regenerate it with `../golang-go/bin/go run ./internal/isotest/cataloggen` after adding an entry package.

The adapter supplies stable unit-test session identity and retains completed query state until each test ends, matching the live bridge's cache behavior. Mocked SDK child workflows remain host stubs; child-workflow also includes a native child integration test. Update assertions decode raw results across the isolate boundary. All 54 test files from supported upstream groups are retained, with source changes for native API signatures and payload transport. The full workspace run passes in 52 test packages and compiles another 245 packages without tests. A separate `go build -p 3 ./... ./codec-server/... ./grpc-proxy/...` validates the worker and starter executables with generated isolate metadata.

The hello-world development-server test is opt-in:

```sh
SAMPLES_DEV_SERVER=1 ../golang-go/bin/go test -count=1 -v ./helloworld
SAMPLES_DEV_SERVER=1 ../golang-go/bin/go test -count=1 -v ./dynamic-workflows
```

The opt-in hello-world and dynamic-workflows development-server tests have also been run; dynamic workflow histories replay successfully. Tests and compilation do not deploy Cloud Run/Lambda workers or run Datadog/OTel collectors, a production Temporal service, sample UIs, authentication services, or standalone codec/gRPC servers. Those require the infrastructure described in each sample README. Replay suites exercise the checked-in hello-world and rainbow-statuses histories.

## Blocked samples

| Sample | Missing feature |
| --- | --- |
| `ctxpropagation` | Application ContextPropagators across native workflow boundaries. |
| `encryption` | Application ContextPropagators and context-aware data conversion inside native workflows. Context-independent host payload codecs are supported. |
| `external-storage` | External payload storage/reference transport across the native bridge. |
| `googleadk` | A native context adapter for the contrib Google ADK integration, which uses SDK workflow.Context. |
| `nexus-context-propagation` | Application ContextPropagators into and out of native workflows used by Nexus. |
| `workflow-security-interceptor` | Application header/context propagation and an adapter for the security workflow interceptor. |
| `workflowstreams` | A native context adapter for the contrib workflow stream API, which uses SDK workflow.Context. |

Registered native signal channels are filled proactively before application goroutines resume. Late registration delivers the queued signals for that name before returning. Ordinary `select` loops with `default` can drain the admitted batch before Continue-As-New. The four samples previously blocked by host signal buffering are now migrated. Request/response workflows also wait for all accepted activities and callback replies before starting a new run.

## Complete upstream inventory

| Sample group | Status | Implementation notes |
| --- | --- | --- |
| [await-signals](./await-signals/) | Migrated | One native select loop retains signals received out of order. |
| [batch-sliding-window](./batch-sliding-window/) | Migrated | One native completion channel controls the sliding window; nonblocking drains preserve state before Continue-As-New. |
| [branch](./branch/) | Migrated | Native workflow entries and host worker/client APIs. |
| [cancellation](./cancellation/) | Migrated | Cleanup uses context.WithoutCancel for both scheduling and waiting. |
| [child-workflow](./child-workflow/) | Migrated | Native workflow entries and host worker/client APIs. |
| [child-workflow-continue-as-new](./child-workflow-continue-as-new/) | Migrated | Native workflow entries and host worker/client APIs. |
| [choice-exclusive](./choice-exclusive/) | Migrated | Existing native signal-based choice implementation retained. |
| [choice-multi](./choice-multi/) | Migrated | Native workflow entries and host worker/client APIs. |
| [codec-server](./codec-server/) | Migrated | Separate workspace module; codec/server code stays on the host and the workflow is in native/. |
| [cron](./cron/) | Migrated | Native workflow entries and host worker/client APIs. |
| `ctxpropagation` | Blocked | See missing feature above. |
| [datadog](./datadog/) | Migrated | Native workflow interceptor, host exporter, and Datadog v2 tracer. |
| [dsl](./dsl/) | Migrated | Native goroutines and a mutex protect shared parallel-branch bindings. |
| [dynamic](./dynamic/) | Migrated | String-named activities ported; this is distinct from dynamic workflow registration. |
| [dynamic-workflows](./dynamic-workflows/) | Migrated | Native fallback workflow registration with isolate-owned converter.EncodedValues; dynamic activities and runtime versioning options remain on the host. Unit tests cover runtime names and decoding; the opt-in live test verifies all starter scenarios and history replay. |
| [dynamicconfig](./dynamicconfig/) | Migrated | Documentation-only upstream sample; retained. |
| [dynamicmtls](./dynamicmtls/) | Migrated | Native workflow entries and host worker/client APIs. |
| [eager-workflow-start](./eager-workflow-start/) | Migrated | Native workflow entries and host worker/client APIs. |
| [early-return](./early-return/) | Migrated | Initialization is a registered top-level local activity; updates wait on a native channel. |
| `encryption` | Blocked | See missing feature above. |
| [expense](./expense/) | Migrated | Native workflow entries and host worker/client APIs. |
| [external-env-conf](./external-env-conf/) | Migrated | Native workflow entries and host worker/client APIs. |
| `external-storage` | Blocked | See missing feature above. |
| [fileprocessing](./fileprocessing/) | Migrated | Native workflow entries and host worker/client APIs. |
| [gcp](./gcp/) | Migrated | Cloud Run workers use the native SDK wrapper; identity, telemetry, and deployment remain host concerns. |
| `googleadk` | Blocked | See missing feature above. |
| [goroutine](./goroutine/) | Migrated | Native goroutines and a completion channel; results preserve input order. |
| [greetings](./greetings/) | Migrated | Native workflow entries and host worker/client APIs. |
| [greetingslocal](./greetingslocal/) | Migrated | Native workflow entries and host worker/client APIs. |
| [grpc-proxy](./grpc-proxy/) | Migrated | Separate workspace module; the gRPC proxy runs on the host. |
| [helloworld](./helloworld/) | Migrated | Existing native entry remains HelloWorld. Upstream activity/workflow/replay tests included; dev-server test is opt-in. |
| [helloworld-apiKey](./helloworld-apiKey/) | Migrated | Native workflow entries and host worker/client APIs. |
| [helloworldmtls](./helloworldmtls/) | Migrated | Native workflow entries and host worker/client APIs. |
| [lambda-worker](./lambda-worker/) | Migrated | Lambda workers use the native SDK wrapper; AWS integration stays on the host. |
| [logger-interceptor](./logger-interceptor/) | Migrated | Native workflow interceptor factory; SDK activity interceptors stay on the host. |
| [memo](./memo/) | Migrated | DescribeWorkflow returns a decoded memo map instead of a protobuf execution-info value. |
| [metrics](./metrics/) | Migrated | Native workflow entries and host worker/client APIs. |
| [multi-history-replay](./multi-history-replay/) | Migrated | Native workflow replayer and worker; existing HelloWorld entry reused. |
| [mutex](./mutex/) | Migrated | Native request/release channels and time.Timer retain the external lock protocol; queued requests drain before coordinator completion. |
| [nexus](./nexus/) | Migrated | Native caller workflows and handler workflow; Nexus operations remain host registrations. |
| [nexus-cancelation](./nexus-cancelation/) | Migrated | Native goroutines and deterministic randomness; detached waits honor requested cancellation acknowledgments. |
| `nexus-context-propagation` | Blocked | See missing feature above. |
| [nexus-messaging](./nexus-messaging/) | Migrated | Both patterns use native queries, updates, typed signals, and mutexes. |
| [nexus-messaging-temporal-operation](./nexus-messaging-temporal-operation/) | Migrated | Both patterns ported, including the signal-with-start approval-context operation. |
| [nexus-multiple-arguments](./nexus-multiple-arguments/) | Migrated | Host Nexus adapter starts the native handler workflow. |
| [nexus-standalone-activity](./nexus-standalone-activity/) | Migrated | Host activity-backed Nexus operations; no native workflow needed. |
| [nexus-standalone-operations](./nexus-standalone-operations/) | Migrated | Host standalone Nexus operations; no native workflow needed. |
| [opentelemetry](./opentelemetry/) | Migrated | Native workflow interceptor and host span processor/exporter; setup moved to setup/. |
| [opentelemetry-v2](./opentelemetry-v2/) | Migrated | All three instrumentation examples ported; native spans use a sink provider and host exporter. |
| [pickfirst](./pickfirst/) | Migrated | Native workflow entries and host worker/client APIs. |
| [polling](./polling/) | Migrated | All three polling patterns ported with native timers and child workflows. |
| [pso](./pso/) | Migrated | Native concurrent activity batches, deterministic child IDs, and a native custom converter factory matching the host. |
| [query](./query/) | Migrated | Native workflow entries and host worker/client APIs. |
| [recovery](./recovery/) | Migrated | Native workflow entries and host worker/client APIs. |
| [reqrespactivity](./reqrespactivity/) | Migrated | Native request goroutines, callback activities, and completion notifications; drains admitted signals and waits for all replies before Continue-As-New. |
| [reqrespquery](./reqrespquery/) | Migrated | Native request goroutines with a single response-map owner; drains admitted signals and waits for all results before Continue-As-New. |
| [reqrespupdate](./reqrespupdate/) | Migrated | Native update-completion notifications; WaitHandlersFinished before new runs. |
| [retryactivity](./retryactivity/) | Migrated | Native workflow entries and host worker/client APIs. |
| [safe_message_handler](./safe_message_handler/) | Migrated | Native mutexes and startup notifications; WaitHandlersFinished before completion/new runs. |
| [saga](./saga/) | Migrated | Native workflow entries and host worker/client APIs. |
| [schedule](./schedule/) | Migrated | Native workflow entries and host worker/client APIs. |
| [searchattributes](./searchattributes/) | Migrated | Native workflow entries and host worker/client APIs. |
| [serverjwtauth](./serverjwtauth/) | Migrated | Native workflow entries and host worker/client APIs. |
| [session-failure](./session-failure/) | Migrated | Native workflow entries and host worker/client APIs. |
| [shoppingcart](./shoppingcart/) | Migrated | Typed checkout signals and update notifications; WaitHandlersFinished before completion/new runs. |
| [sleep-for-days](./sleep-for-days/) | Migrated | Existing native entry remains SleepForDays. Upstream 90-day time-skip test included. |
| [slogadapter](./slogadapter/) | Migrated | Native workflow entries and host worker/client APIs. |
| [snappycompress](./snappycompress/) | Migrated | Codec stays on the host, native workflow in native/; compressed execution tested. |
| [splitmerge-future](./splitmerge-future/) | Migrated | Native workflow entries and host worker/client APIs. |
| [splitmerge-selector](./splitmerge-selector/) | Migrated | Native workflow entries and host worker/client APIs. |
| [standalone-activity](./standalone-activity/) | Migrated | Native workflow entries and host worker/client APIs. |
| [start-delay](./start-delay/) | Migrated | Native workflow entries and host worker/client APIs. |
| [synchronous-proxy](./synchronous-proxy/) | Migrated | Typed request/response signals and native select waits; proxy functions are helpers, not isolate entries. |
| [temporal-fixtures](./temporal-fixtures/) | Migrated | Native workflow entries and host worker/client APIs. |
| [timer](./timer/) | Migrated | Native workflow entries and host worker/client APIs. |
| [typed-searchattributes](./typed-searchattributes/) | Migrated | Documentation-only upstream sample; retained. |
| [updatabletimer](./updatabletimer/) | Migrated | Typed time.Time signals and a resettable native time.Timer. |
| [update](./update/) | Migrated | Native workflow entries and host worker/client APIs. |
| [worker-specific-task-queues](./worker-specific-task-queues/) | Migrated | Native workflow entries and host worker/client APIs. |
| [worker-versioning](./worker-versioning/) | Migrated | Deployment options retained; typed native signal receives. |
| `workflow-security-interceptor` | Blocked | See missing feature above. |
| `workflowstreams` | Blocked | See missing feature above. |
| [zapadapter](./zapadapter/) | Migrated | Host logger implementation moved to logger/ to keep process logging setup out of the isolate. |
