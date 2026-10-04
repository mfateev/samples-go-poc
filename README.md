# Temporal Go samples on the isolate POC SDK

This module ports three examples from [Temporal's samples-go](https://github.com/temporalio/samples-go)
at commit `aaf79b6`: [helloworld](https://github.com/temporalio/samples-go/tree/aaf79b6/helloworld)
and [choice-exclusive](https://github.com/temporalio/samples-go/tree/aaf79b6/choice-exclusive).
The [sleep-for-days](https://github.com/temporalio/samples-go/tree/aaf79b6/sleep-for-days)
port is a concurrency fixture; its live execution and replay require the
quiescence work described below.
The adapted sample code uses the upstream Apache 2.0 license in `LICENSE`.

**Repositories:** the modified compiler and runtime live in
[`mfateev/golang-go`](https://github.com/mfateev/golang-go/tree/task/modify-go-runtime-for-isolates),
the workflow API and Temporal adapter in
[`mfateev/sdk-go-poc`](https://github.com/mfateev/sdk-go-poc/tree/task/modify-go-runtime-for-isolates),
and these examples in
[`mfateev/samples-go-poc`](https://github.com/mfateev/samples-go-poc/tree/task/modify-go-runtime-for-isolates).
All three use the `task/modify-go-runtime-for-isolates` branch. The checked-in
`go.work` makes the samples use the sibling SDK checkout directly, without
pinning an isolate SDK version or commit in `go.mod`. Keep the repositories
beside each other as shown below. The upstream Temporal Go SDK remains pinned
because the adapter uses its internal workflow extension API.

## 1. Prepare a Linux or macOS machine

The commands below support ARM64 and x86-64 on Debian or Ubuntu Linux and
macOS. The build and run sequence was verified on Linux ARM64. Have a few
gigabytes of free space for the Go source build and module cache.

On Debian or Ubuntu, install Git, curl, the archive tools, and a C compiler:

```sh
sudo apt-get update
sudo apt-get install -y ca-certificates curl git tar coreutils build-essential
```

On macOS, install the Xcode Command Line Tools if they are missing:

```sh
xcode-select -p || xcode-select --install
```

Wait for the installer to finish, then verify `xcode-select -p` succeeds before
continuing. The tools provide Git and Clang; macOS also includes curl, tar, and
`shasum`. The following shell blocks work in Bash or the default macOS zsh.

The fork is Go 1.28 development source and requires Go 1.26.0 or newer for
bootstrapping. These commands install Go 1.26.7 into a new directory under
your home directory. The SHA-256 values are from [the Go downloads page](https://go.dev/dl/).

```bash
set -e
POC_ROOT="$HOME/temporal-isolates-poc"
mkdir -p "$POC_ROOT"
test ! -e "$POC_ROOT/go"
test ! -e "$POC_ROOT/golang-go"
test ! -e "$POC_ROOT/sdk-go-poc"
test ! -e "$POC_ROOT/samples-go-poc"

case "$(uname -s)/$(uname -m)" in
  Linux/aarch64|Linux/arm64)
    POC_PLATFORM=linux
    POC_ARCH=arm64
    POC_GO_SHA=5a4ec883379d51ee9ce1040d5e87f8d35e20387574dd8c947feb01eabc3c1b37
    ;;
  Linux/x86_64|Linux/amd64)
    POC_PLATFORM=linux
    POC_ARCH=amd64
    POC_GO_SHA=ffb5f8de10c62550dfddab66b36b57030721e0a44a3218e9e1181d7b59f121ca
    ;;
  Darwin/arm64)
    POC_PLATFORM=darwin
    POC_ARCH=arm64
    POC_GO_SHA=020a1e8224811be75163e920bc77e0926a1390a6aeea19bdcf23f74b9d749f6d
    ;;
  Darwin/x86_64)
    POC_PLATFORM=darwin
    POC_ARCH=amd64
    POC_GO_SHA=92e8b34bff3c89ab16404c595669ac8cb004cc2f676dcbd1f5b87a6b8def3b47
    ;;
  *) echo "This guide supports Linux or macOS on ARM64 or x86-64" >&2; exit 1 ;;
esac

POC_GO_ARCHIVE="$POC_ROOT/go1.26.7.$POC_PLATFORM-$POC_ARCH.tar.gz"
curl -fL --retry 3 -o "$POC_GO_ARCHIVE" "https://go.dev/dl/go1.26.7.$POC_PLATFORM-$POC_ARCH.tar.gz"
if [ "$POC_PLATFORM" = darwin ]; then
  printf '%s  %s\n' "$POC_GO_SHA" "$POC_GO_ARCHIVE" | shasum -a 256 -c -
else
  printf '%s  %s\n' "$POC_GO_SHA" "$POC_GO_ARCHIVE" | sha256sum -c -
fi
tar -C "$POC_ROOT" -xzf "$POC_GO_ARCHIVE"
"$POC_ROOT/go/bin/go" version
```

## 2. Clone the three repositories and build the toolchain

Continue in the same terminal. Use the POC branch for each repository.

```bash
git clone --depth 1 --single-branch --branch task/modify-go-runtime-for-isolates \
  https://github.com/mfateev/golang-go.git "$POC_ROOT/golang-go"
git clone --depth 1 --single-branch --branch task/modify-go-runtime-for-isolates \
  https://github.com/mfateev/sdk-go-poc.git "$POC_ROOT/sdk-go-poc"
git clone --depth 1 --single-branch --branch task/modify-go-runtime-for-isolates \
  https://github.com/mfateev/samples-go-poc.git "$POC_ROOT/samples-go-poc"
cd "$POC_ROOT/golang-go/src"
GOROOT_BOOTSTRAP="$POC_ROOT/go" ./make.bash
"$POC_ROOT/golang-go/bin/go" version
export GOCACHE="$POC_ROOT/go-build-cache"
mkdir -p "$GOCACHE"
"$POC_ROOT/golang-go/bin/go" env GOCACHE
cd "$POC_ROOT/samples-go-poc"
"$POC_ROOT/golang-go/bin/go" env GOWORK
```

The `go version` line from the fork ends with `(isolates POC)`.

Go's build cache already accounts for compiler changes, so a separate cache
is not required for correctness. `make.bash` manages its own bootstrap cache;
this `GOCACHE` applies to the later sample builds. It makes cleanup targeted,
but can increase total disk use if your usual Go cache is also populated.

The printed workspace path should end in `samples-go-poc/go.work`. That file
loads this module and `../sdk-go-poc`, so builds use the SDK branch you cloned.
Go module requirements accept versions rather than branch names; this
workspace is how these samples consume the branch's checked-out source.
If you already created a parent workspace using the older instructions, the
samples' own workspace takes precedence. Use `../golang-go/bin/go` for all
commands in this module. A system `go` command would not recognize
`-isolate-dir`.

To update an existing checkout to the latest POC branch, first commit or stash
any local edits, then run:

```bash
git -C "$POC_ROOT/golang-go" pull --ff-only origin task/modify-go-runtime-for-isolates
git -C "$POC_ROOT/sdk-go-poc" pull --ff-only origin task/modify-go-runtime-for-isolates
git -C "$POC_ROOT/samples-go-poc" pull --ff-only origin task/modify-go-runtime-for-isolates
cd "$POC_ROOT/golang-go/src"
GOROOT_BOOTSTRAP="$POC_ROOT/go" ./make.bash
cd "$POC_ROOT/samples-go-poc"
```

Then repeat the builds in step 3. Pulling the SDK branch updates the source
used by the workspace; no `go get` or SDK version edit is needed.

## 3. Build the sample workers and the replayer

Run from `samples-go-poc` after the previous step:

```bash
../golang-go/bin/go mod download
../golang-go/bin/go test ./...
mkdir -p bin
../golang-go/bin/go build -isolate-dir=./helloworld/workflow -o bin/helloworld-worker ./helloworld/worker
../golang-go/bin/go build -isolate-dir=./choice-exclusive/workflow -o bin/choice-worker ./choice-exclusive/worker
../golang-go/bin/go build -isolate-dir=./sleep-for-days/workflow -o bin/sleep-for-days-worker ./sleep-for-days/worker
../golang-go/bin/go build -isolate-dir=./helloworld/workflow -isolate-dir=./choice-exclusive/workflow -isolate-dir=./sleep-for-days/workflow -o bin/replay ./replay
```

Each workflow directory contains an `isolate.json`, named workflow functions,
and a small Go `main` dispatcher. Functions register from `init` using
`workflow.RegisterTyped` or `workflow.RegisterTyped0`, and `main` calls
`workflow.Run`. Arguments and results use Temporal's default data converter
inside the isolate: `HelloWorld` takes and returns a string, `ExclusiveChoice`
takes no arguments and returns the selected fruit, and `SleepForDays` takes a
duration string and returns `"done"`. The activities and workers are host-side;
activity and signal operations still use byte slices. The isolate SDK comes
from the sibling branch checkout; upstream dependencies use the module files.

The isolate adapter requires the default converter. Custom converters and
protobuf message values remain outside this POC; nil, bytes, and ordinary JSON
values are supported. Ordinary Temporal workflow registration remains available
through the standard worker API.

## 4. Install the Temporal CLI and run the samples

Download Temporal CLI 1.9.1 from [Temporal's archive](https://github.com/temporalio/cli#install-via-download).
These commands still run in the setup terminal:

```bash
curl -fL --retry 3 -o "$POC_ROOT/temporal-cli.tar.gz" \
  "https://temporal.download/cli/archive/v1.9.1?platform=$POC_PLATFORM&arch=$POC_ARCH"
mkdir -p "$POC_ROOT/temporal-cli"
tar -C "$POC_ROOT/temporal-cli" -xzf "$POC_ROOT/temporal-cli.tar.gz"
"$POC_ROOT/temporal-cli/temporal" --version
```

Open four terminals. Use **Terminal 1** for an in-memory local server:

```bash
"$HOME/temporal-isolates-poc/temporal-cli/temporal" server start-dev --headless
```

Use **Terminal 2** for the hello world worker:

```bash
cd "$HOME/temporal-isolates-poc/samples-go-poc"
./bin/helloworld-worker
```

Use **Terminal 3** for the exclusive choice worker:

```bash
cd "$HOME/temporal-isolates-poc/samples-go-poc"
./bin/choice-worker
```

Once both workers report `Started Worker`, use **Terminal 4** for the starters:

```bash
cd "$HOME/temporal-isolates-poc/samples-go-poc"
export GOCACHE="$HOME/temporal-isolates-poc/go-build-cache"
../golang-go/bin/go run ./helloworld/starter Temporal
../golang-go/bin/go run ./choice-exclusive/starter
```

The first starter prints `Hello Temporal!`; the second prints
`order completed: <fruit>`. Each also prints its Workflow ID and Run ID. The choice
workflow schedules `GetOrder`, then one of `OrderApple`, `OrderBanana`,
`OrderCherry`, or `OrderOrange`. `GetOrder` chooses randomly in a host activity,
and Temporal records that result for replay. `TEMPORAL_ADDRESS` can point the
workers and starters at another server; the default is `localhost:7233`.

The typed versions of both serial samples completed on Temporal CLI 1.9.1's
development server on Linux ARM64, and their exported histories replayed in
fresh processes. All three sample workers and the combined replayer built with
the SDK branch workspace.

## 5. Replay the completed histories

Keep the server running so the CLI can export histories. In Terminal 4, paste
the Workflow IDs printed by the starters when prompted:

```bash
cd "$HOME/temporal-isolates-poc/samples-go-poc"
printf 'Hello Workflow ID: '
read -r HELLO_WORKFLOW_ID
printf 'Choice Workflow ID: '
read -r CHOICE_WORKFLOW_ID
"$HOME/temporal-isolates-poc/temporal-cli/temporal" workflow show \
  --workflow-id "$HELLO_WORKFLOW_ID" --output json > bin/hello-history.json
"$HOME/temporal-isolates-poc/temporal-cli/temporal" workflow show \
  --workflow-id "$CHOICE_WORKFLOW_ID" --output json > bin/choice-history.json
./bin/replay helloworld-poc HelloWorld bin/hello-history.json
./bin/replay choice-exclusive-poc ExclusiveChoice bin/choice-history.json
```

Both replays should print `replay passed`. Use histories from the current typed
samples; histories recorded by the earlier byte-only workflow versions use a
different argument/result contract. The workers and server can then be stopped
with Ctrl-C in their terminals.

The first two ports use named typed workflow functions with blocking activity
calls. Upstream `workflow.Context` and futures become ordinary Go calls inside
the isolate. Native quiescence and deterministic concurrent goroutines remain
future work.

## 6. Concurrent sleep-for-days fixture

The upstream `sleep-for-days` workflow starts an activity future, then selects
between a timer future and a `complete` signal channel. The port keeps those
operations concurrent: `workflow.ExecuteActivityAsync` returns a result channel,
native `time.After` returns a host-driven durable timer channel, and
`workflow.GetSignalChannel("complete")` wraps the signal call. Ordinary Go
`select` chooses the event. Its interval defaults to 30 days; the starter accepts a
shorter positive Go duration such as `1m` for experiments.

The worker builds, but **this workflow is not yet supported for live execution
or replay**. The current Temporal bridge returns from a Workflow Task when it
sees the first blocked host call. It cannot yet observe that all isolate
goroutines have become quiescent, so the other concurrent calls can be missed
or processed in a later task. Deterministic scheduling and an exact quiescence
barrier are needed before this sample is an acceptance test.

After that bridge work, run the fixture from `samples-go-poc` with the server
started as in step 4:

```bash
./bin/sleep-for-days-worker
```

In another terminal, start the workflow and send its completion signal using
the Workflow ID printed by the starter:

```bash
cd "$HOME/temporal-isolates-poc/samples-go-poc"
../golang-go/bin/go run ./sleep-for-days/starter 1m
"$HOME/temporal-isolates-poc/temporal-cli/temporal" workflow signal \
  --workflow-id YOUR_WORKFLOW_ID --name complete
"$HOME/temporal-isolates-poc/temporal-cli/temporal" workflow show \
  --workflow-id YOUR_WORKFLOW_ID --output json > bin/sleep-for-days-history.json
./bin/replay sleep-for-days-poc SleepForDays bin/sleep-for-days-history.json
```

The expected final result is `done` and `replay passed: SleepForDays`.

If you need to reclaim disk space after all Go commands finish, clear only
the POC build cache:

```sh
GOCACHE="$HOME/temporal-isolates-poc/go-build-cache" \
  "$HOME/temporal-isolates-poc/golang-go/bin/go" clean -cache
```
