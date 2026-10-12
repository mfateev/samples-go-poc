# Named random streams

`workflow.GetRandomStream(ctx, name)` returns a persistent `*math/rand/v2.Rand`
inside a native workflow. `Plan` uses separate `choices`, `jitter`, and `audit`
streams. Drawing additional audit values leaves the other streams unchanged.
Repeated lookup of a name continues its sequence; derived or canceled contexts
share that execution's stream.

Seeds derive from the original run identity recorded in history and the exact
stream name. Replay reconstructs the generators through the same workflow calls,
without recording individual draws. Different executions usually produce different
results, so compare audit independence within the same execution identity, as the
unit tests do. Changes to draw order in an existing stream require ordinary
workflow versioning when they change replayed commands.

The `next` query draws through a captured workflow generator using a scratch copy.
Repeated queries return the same next choice without changing the workflow stream.

These generators are predictable. Use host activities for cryptographic secrets.

With a Temporal server at `localhost:7233`, run from the repository root:

```sh
mkdir -p bin
../golang-go/bin/go build -o bin/random-streams-worker ./random-streams/worker
./bin/random-streams-worker
```

In another terminal:

```sh
../golang-go/bin/go run ./random-streams/starter -count 10 -audit-draws 123
```

Set `TEMPORAL_ADDRESS` in both terminals to use another server. Native worker
executables must be built with the isolate-aware toolchain.

```sh
../golang-go/bin/go test ./random-streams/...
SAMPLES_DEV_SERVER=1 ../golang-go/bin/go test -count=1 ./random-streams
```

The development-server test checks completed queries and compares recorded outputs
with replay using both the default workflow cache and cache size zero. This is an
additional POC sample, outside the upstream migration inventory.
