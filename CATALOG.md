# Sample catalog

The table lists included applications and unported sample groups. Directory
presence means source/tests exist; it does not mean the cloud deployments,
collectors, authentication services or production integrations were exercised.
Run the workspace checks in [README.md](README.md) and use each application's
setup instructions for its external services.

## Application-port scope

The absent groups are not all missing SDK features. Native context propagation,
per-context data converters, encryption through host codecs and workflow streams
are implemented in sdk-go-poc. The corresponding complete sample applications
have not all been ported into this repository.

| Unported group | Current scope |
| --- | --- |
| `ctxpropagation` | Native mechanism exists; SDK example/contextpropagation demonstrates it |
| `encryption` | Native factories/context conversion and host encryption codecs exist; SDK example/converter demonstrates them |
| `external-storage` | External payload-reference transport at the native boundary is unsupported |
| `googleadk` | No native adapter for the contrib workflow.Context integration |
| `nexus-context-propagation` | Complete Nexus application not ported; native context propagators exist |
| `workflow-security-interceptor` | Application-specific native security interceptor/header integration not ported |
| `workflowstreams` | SDK native contrib/workflowstreams and compiled examples exist; this full sample group is not included |

Native ports target their documented current input/output contracts. Do not
assume all ordinary SDK histories replay after replacing coroutine execution with
native primitives; validate the intended application's recorded histories.

## Included patterns

Use context.Context, native go/channels/select/sync and history-driven time.
Signal channels are typed and filled before ordinary continuations resume.
Future.ToChannel is the native select adapter. WaitHandlersFinished drains
accepted update completions; coordinate signal goroutines separately before
completion/Continue-As-New. Host starters use registered workflow string names.
Serializer/interceptor/propagator factories create private state explicitly.

## Inventory


| Sample group | State | Current implementation |
| --- | --- | --- |
| [await-signals](./await-signals/) | Included | One native select loop retains signals received out of order. |
| [batch-sliding-window](./batch-sliding-window/) | Included | One native completion channel controls the sliding window; nonblocking drains preserve state before Continue-As-New. |
| [branch](./branch/) | Included | Native workflow entries and host worker/client APIs. |
| [cancellation](./cancellation/) | Included | Cleanup uses context.WithoutCancel for both scheduling and waiting. |
| [child-workflow](./child-workflow/) | Included | Native workflow entries and host worker/client APIs. |
| [child-workflow-continue-as-new](./child-workflow-continue-as-new/) | Included | Native workflow entries and host worker/client APIs. |
| [choice-exclusive](./choice-exclusive/) | Included | Native signal-based choice implementation. |
| [choice-multi](./choice-multi/) | Included | Native workflow entries and host worker/client APIs. |
| [codec-server](./codec-server/) | Included | Separate workspace module; codec/server code stays on the host and the workflow is in native/. |
| [cron](./cron/) | Included | Native workflow entries and host worker/client APIs. |
| `ctxpropagation` | Unported | See application-port scope above. |
| [datadog](./datadog/) | Included | Native workflow interceptor, host exporter, and Datadog v2 tracer. |
| [dsl](./dsl/) | Included | Native goroutines and a mutex protect shared parallel-branch bindings. |
| [dynamic](./dynamic/) | Included | String-named activities ported; this is distinct from dynamic workflow registration. |
| [dynamic-workflows](./dynamic-workflows/) | Included | Native fallback workflow registration with isolate-owned converter.EncodedValues; dynamic activities and runtime versioning options remain on the host. Unit tests cover runtime names and decoding; the opt-in live test verifies all starter scenarios and history replay. |
| [dynamicconfig](./dynamicconfig/) | Included | Documentation-only sample. |
| [dynamicmtls](./dynamicmtls/) | Included | Native workflow entries and host worker/client APIs. |
| [eager-workflow-start](./eager-workflow-start/) | Included | Native workflow entries and host worker/client APIs. |
| [early-return](./early-return/) | Included | Initialization is a registered top-level local activity; updates wait on a native channel. |
| `encryption` | Unported | See application-port scope above. |
| [expense](./expense/) | Included | Native workflow entries and host worker/client APIs. |
| [external-env-conf](./external-env-conf/) | Included | Native workflow entries and host worker/client APIs. |
| `external-storage` | Unported | See application-port scope above. |
| [fileprocessing](./fileprocessing/) | Included | Native workflow entries and host worker/client APIs. |
| [gcp](./gcp/) | Included | Cloud Run workers use the native SDK wrapper; identity, telemetry, and deployment remain host concerns. |
| `googleadk` | Unported | See application-port scope above. |
| [goroutine](./goroutine/) | Included | Native goroutines and a completion channel; results preserve input order. |
| [greetings](./greetings/) | Included | Native workflow entries and host worker/client APIs. |
| [greetingslocal](./greetingslocal/) | Included | Native workflow entries and host worker/client APIs. |
| [grpc-proxy](./grpc-proxy/) | Included | Separate workspace module; the gRPC proxy runs on the host. |
| [helloworld](./helloworld/) | Included | HelloWorld with activity/workflow/replay tests; dev-server test is opt-in. |
| [helloworld-apiKey](./helloworld-apiKey/) | Included | Native workflow entries and host worker/client APIs. |
| [helloworldmtls](./helloworldmtls/) | Included | Native workflow entries and host worker/client APIs. |
| [lambda-worker](./lambda-worker/) | Included | Lambda workers use the native SDK wrapper; AWS integration stays on the host. |
| [logger-interceptor](./logger-interceptor/) | Included | Native workflow interceptor factory; SDK activity interceptors stay on the host. |
| [memo](./memo/) | Included | DescribeWorkflow returns a decoded memo map instead of a protobuf execution-info value. |
| [metrics](./metrics/) | Included | Native workflow entries and host worker/client APIs. |
| [multi-history-replay](./multi-history-replay/) | Included | Native workflow replayer and worker; existing HelloWorld entry reused. |
| [mutex](./mutex/) | Included | Native request/release channels and time.Timer retain the external lock protocol; queued requests drain before coordinator completion. |
| [nexus](./nexus/) | Included | Native caller workflows and handler workflow; Nexus operations remain host registrations. |
| [nexus-cancelation](./nexus-cancelation/) | Included | Native goroutines and deterministic randomness; detached waits honor requested cancellation acknowledgments. |
| `nexus-context-propagation` | Unported | See application-port scope above. |
| [nexus-messaging](./nexus-messaging/) | Included | Both patterns use native queries, updates, typed signals, and mutexes. |
| [nexus-messaging-temporal-operation](./nexus-messaging-temporal-operation/) | Included | Both patterns ported, including the signal-with-start approval-context operation. |
| [nexus-multiple-arguments](./nexus-multiple-arguments/) | Included | Host Nexus adapter starts the native handler workflow. |
| [nexus-standalone-activity](./nexus-standalone-activity/) | Included | Host activity-backed Nexus operations; no native workflow needed. |
| [nexus-standalone-operations](./nexus-standalone-operations/) | Included | Host standalone Nexus operations; no native workflow needed. |
| [opentelemetry](./opentelemetry/) | Included | Native workflow interceptor and host span processor/exporter; setup moved to setup/. |
| [opentelemetry-v2](./opentelemetry-v2/) | Included | All three instrumentation examples ported; native spans use a sink provider and host exporter. |
| [pickfirst](./pickfirst/) | Included | Native workflow entries and host worker/client APIs. |
| [polling](./polling/) | Included | All three polling patterns ported with native timers and child workflows. |
| [pso](./pso/) | Included | Native concurrent activity batches, deterministic child IDs, and a native custom converter factory matching the host. |
| [query](./query/) | Included | Native workflow entries and host worker/client APIs. |
| [recovery](./recovery/) | Included | Native workflow entries and host worker/client APIs. |
| [reqrespactivity](./reqrespactivity/) | Included | Native request goroutines, callback activities, and completion notifications; drains admitted signals and waits for all replies before Continue-As-New. |
| [reqrespquery](./reqrespquery/) | Included | Native request goroutines with a single response-map owner; drains admitted signals and waits for all results before Continue-As-New. |
| [reqrespupdate](./reqrespupdate/) | Included | Native update-completion notifications; WaitHandlersFinished before new runs. |
| [retryactivity](./retryactivity/) | Included | Native workflow entries and host worker/client APIs. |
| [safe_message_handler](./safe_message_handler/) | Included | Native mutexes and startup notifications; WaitHandlersFinished before completion/new runs. |
| [saga](./saga/) | Included | Native workflow entries and host worker/client APIs. |
| [schedule](./schedule/) | Included | Native workflow entries and host worker/client APIs. |
| [searchattributes](./searchattributes/) | Included | Native workflow entries and host worker/client APIs. |
| [serverjwtauth](./serverjwtauth/) | Included | Native workflow entries and host worker/client APIs. |
| [session-failure](./session-failure/) | Included | Native workflow entries and host worker/client APIs. |
| [shoppingcart](./shoppingcart/) | Included | Typed checkout signals and update notifications; WaitHandlersFinished before completion/new runs. |
| [sleep-for-days](./sleep-for-days/) | Included | SleepForDays with a 90-day virtual-time test. |
| [slogadapter](./slogadapter/) | Included | Native workflow entries and host worker/client APIs. |
| [snappycompress](./snappycompress/) | Included | Codec stays on the host, native workflow in native/; compressed execution tested. |
| [splitmerge-future](./splitmerge-future/) | Included | Native workflow entries and host worker/client APIs. |
| [splitmerge-selector](./splitmerge-selector/) | Included | Native workflow entries and host worker/client APIs. |
| [standalone-activity](./standalone-activity/) | Included | Native workflow entries and host worker/client APIs. |
| [start-delay](./start-delay/) | Included | Native workflow entries and host worker/client APIs. |
| [synchronous-proxy](./synchronous-proxy/) | Included | Typed request/response signals and native select waits; proxy functions are helpers, not isolate entries. |
| [temporal-fixtures](./temporal-fixtures/) | Included | Native workflow entries and host worker/client APIs. |
| [timer](./timer/) | Included | Native workflow entries and host worker/client APIs. |
| [typed-searchattributes](./typed-searchattributes/) | Included | Documentation-only sample. |
| [updatabletimer](./updatabletimer/) | Included | Typed time.Time signals and a resettable native time.Timer. |
| [update](./update/) | Included | Native workflow entries and host worker/client APIs. |
| [worker-specific-task-queues](./worker-specific-task-queues/) | Included | Native workflow entries and host worker/client APIs. |
| [worker-versioning](./worker-versioning/) | Included | Deployment options retained; typed native signal receives. |
| `workflow-security-interceptor` | Unported | See application-port scope above. |
| `workflowstreams` | Unported | See application-port scope above. |
| [zapadapter](./zapadapter/) | Included | Host logger implementation moved to logger/ to keep process logging setup out of the isolate. |
