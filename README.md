# Temporal Go samples on the isolate POC SDK

This module ports two examples from [Temporal's samples-go](https://github.com/temporalio/samples-go)
at commit `aaf79b6`: [helloworld](https://github.com/temporalio/samples-go/tree/aaf79b6/helloworld)
and [choice-exclusive](https://github.com/temporalio/samples-go/tree/aaf79b6/choice-exclusive).
The adapted sample code uses the upstream Apache 2.0 license in `LICENSE`.

**Repositories:** the modified compiler and runtime live in
[`mfateev/golang-go`](https://github.com/mfateev/golang-go/tree/task/modify-go-runtime-for-isolates),
the workflow API and Temporal adapter in
[`mfateev/sdk-go-poc`](https://github.com/mfateev/sdk-go-poc/tree/task/modify-go-runtime-for-isolates),
and these examples in
[`mfateev/samples-go-poc`](https://github.com/mfateev/samples-go-poc/tree/task/modify-go-runtime-for-isolates).
All three use the `task/modify-go-runtime-for-isolates` branch. The samples
pin the SDK module to a commit from that branch.

## 1. Prepare a Linux machine

These commands use Bash on Debian or Ubuntu Linux, on either ARM64 or x86-64.
The build and run sequence below was verified on Linux ARM64. Have a few
gigabytes of free space for the Go source build and module cache. Install Git,
curl, the archive tools, and a C compiler:

```sh
sudo apt-get update
sudo apt-get install -y ca-certificates curl git tar coreutils build-essential
```

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
    POC_ARCH=arm64
    POC_GO_SHA=5a4ec883379d51ee9ce1040d5e87f8d35e20387574dd8c947feb01eabc3c1b37
    ;;
  Linux/x86_64|Linux/amd64)
    POC_ARCH=amd64
    POC_GO_SHA=ffb5f8de10c62550dfddab66b36b57030721e0a44a3218e9e1181d7b59f121ca
    ;;
  *) echo "This guide supports Linux ARM64 or x86-64" >&2; exit 1 ;;
esac

POC_GO_ARCHIVE="$POC_ROOT/go1.26.7.linux-$POC_ARCH.tar.gz"
curl -fL --retry 3 -o "$POC_GO_ARCHIVE" "https://go.dev/dl/go1.26.7.linux-$POC_ARCH.tar.gz"
printf '%s  %s\n' "$POC_GO_SHA" "$POC_GO_ARCHIVE" | sha256sum -c -
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
cd "$POC_ROOT/samples-go-poc"
```

Use `../golang-go/bin/go` for all commands in this module. A system `go` command would
not recognize `-isolate-dir`.

## 3. Build both sample workers and the replayer

Run from `samples-go-poc` after the previous step:

```bash
../golang-go/bin/go mod download
../golang-go/bin/go test ./...
mkdir -p bin
../golang-go/bin/go build -isolate-dir=./helloworld/workflow -o bin/helloworld-worker ./helloworld/worker
../golang-go/bin/go build -isolate-dir=./choice-exclusive/workflow -o bin/choice-worker ./choice-exclusive/worker
../golang-go/bin/go build -isolate-dir=./helloworld/workflow -isolate-dir=./choice-exclusive/workflow -o bin/replay ./replay
```

Each workflow directory contains a normal Go `main` and an `isolate.json`.
The activities and workers are host-side; the workflow programs import
`sdk-go-poc/workflow`. The SDK and Temporal Go SDK versions are pinned in the
module files.

## 4. Install the Temporal CLI and run the samples

Download Temporal CLI 1.9.1 from [Temporal's Linux archive](https://github.com/temporalio/cli#install-via-download).
These commands still run in the setup terminal:

```bash
curl -fL --retry 3 -o "$POC_ROOT/temporal-cli.tar.gz" \
  "https://temporal.download/cli/archive/v1.9.1?platform=linux&arch=$POC_ARCH"
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
../golang-go/bin/go run ./helloworld/starter Temporal
../golang-go/bin/go run ./choice-exclusive/starter
```

The first starter prints `Hello Temporal!`; the second prints
`order completed`. Each also prints its Workflow ID and Run ID. The choice
workflow schedules `GetOrder`, then one of `OrderApple`, `OrderBanana`,
`OrderCherry`, or `OrderOrange`. `GetOrder` chooses randomly in a host activity,
and Temporal records that result for replay. `TEMPORAL_ADDRESS` can point the
workers and starters at another server; the default is `localhost:7233`.

## 5. Replay the completed histories

Keep the server running so the CLI can export histories. In Terminal 4, paste
the Workflow IDs printed by the starters when prompted:

```bash
cd "$HOME/temporal-isolates-poc/samples-go-poc"
read -r -p 'Hello Workflow ID: ' HELLO_WORKFLOW_ID
read -r -p 'Choice Workflow ID: ' CHOICE_WORKFLOW_ID
"$HOME/temporal-isolates-poc/temporal-cli/temporal" workflow show \
  --workflow-id "$HELLO_WORKFLOW_ID" --output json > bin/hello-history.json
"$HOME/temporal-isolates-poc/temporal-cli/temporal" workflow show \
  --workflow-id "$CHOICE_WORKFLOW_ID" --output json > bin/choice-history.json
./bin/replay helloworld-poc HelloWorld bin/hello-history.json
./bin/replay choice-exclusive-poc ExclusiveChoice bin/choice-history.json
```

Both replays should print `replay passed`. The workers and server can then be
stopped with Ctrl-C in their terminals.

These ports use the POC SDK's serial, byte-oriented API: upstream
`workflow.Context` and futures become blocking calls from ordinary Go `main`
functions. Native quiescence and deterministic concurrent goroutines remain
future work.
