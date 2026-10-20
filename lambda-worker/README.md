# Lambda Worker

This sample demonstrates how to run a Temporal Worker inside an AWS Lambda function using
the [`lambdaworker`](https://pkg.go.dev/go.temporal.io/sdk/contrib/aws/lambdaworker)
contrib package. It includes optional OpenTelemetry instrumentation that exports traces
and metrics through AWS Distro for OpenTelemetry (ADOT).

The sample registers a simple greeting Workflow and Activity, but the pattern applies to
any Workflow/Activity definitions.

## Prerequisites

- A [Temporal Cloud](https://temporal.io/cloud) namespace (or a self-hosted Temporal
  cluster accessible from your Lambda)
- AWS CLI configured with permissions to create Lambda functions, IAM roles, and
  CloudFormation stacks
- mTLS client certificate and key for your Temporal namespace (place as `client.pem` and
  `client.key` in this directory). See more [here](https://docs.temporal.io/cloud/certificates).
- Matching tgo and the sibling SDK/host-SDK workspace from [root setup](../README.md#setup)

## Files

| File | Description |
|------|-------------|
| `worker/main.go` | Lambda worker entry point -- configures the worker, registers Workflows/Activities, and starts the Lambda handler |
| `greeting/workflow.go` | Sample Workflow that executes a greeting Activity |
| `greeting/activity.go` | Sample Activity that returns a greeting string |
| `starter/main.go` | Helper program to start a Workflow execution against the Lambda worker |
| `temporal.toml` | Temporal client connection configuration (update with your namespace) |
| `otel-collector-config.yaml` | OpenTelemetry Collector sidecar configuration for ADOT |
| `deploy-lambda.sh` | Builds and deploys the Lambda function |
| `mk-iam-role.sh` | Creates the IAM role that allows Temporal Cloud to invoke the Lambda |
| `iam-role-for-temporal-lambda-invoke.yaml` | CloudFormation template for the IAM role |
| `extra-setup-steps` | Additional IAM and Lambda configuration for OpenTelemetry support |

## Setup

The instructions here are a slimmed down version of the more complete getting started guide which
you can find [here](https://docs.temporal.io/production-deployment/worker-deployments/serverless-workers/aws-lambda).

### 1. Create a custom-runtime Lambda function

Create a provided.al2023 function for the Go bootstrap executable:

```bash
aws lambda create-function \
  --function-name my-temporal-worker \
  --runtime provided.al2023 \
  --handler lambda_function.lambda_handler \
  --role arn:aws:iam::<YOUR_ACCOUNT_ID>:role/my-temporal-worker-execution \
  --timeout 600 \
  --memory-size 256
```

### 2. Configure Temporal connection

Edit `temporal.toml` with your Temporal Cloud namespace address and credentials. In production,
we'd recommend reading your credentials from a secret store, but to keep this example simple
the toml file defaults to reading them from keys bundled along with the Lambda code. You may
also set environment variables if you prefer not to use a file. For more information, see
our [docs on environment config](https://docs.temporal.io/develop/environment-configuration).

### 3. Create the IAM role

This creates the IAM role that Temporal Cloud assumes to invoke your Lambda function:

```bash
./mk-iam-role.sh <stack-name> <external-id> <lambda-arn>
```

The External ID is provided by Temporal Cloud in your namespace's serverless worker
configuration.

### 4. (Optional) Enable OpenTelemetry

If you want traces, metrics, and logs, run the extra setup to grant the Lambda role the
necessary permissions:

```bash
./extra-setup-steps <role-name> <function-name> <region> <account-id>
```

Update `otel-collector-config.yaml` with your function name and region as needed.

You'll also want to ensure the Lambda function has the additional OTel collector layer required for
exporting OTel metrics and traces. See [here](https://aws-otel.github.io/docs/getting-started/lambda)
for more.

### 5. Deploy the Lambda function

The checked-in deploy-lambda.sh invokes stock Go. For the native POC use this
build/package sequence from lambda-worker with matching tgo on PATH:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 tgo build -tags lambda.norpc -o bootstrap ./worker
zip function.zip bootstrap client.pem client.key temporal.toml otel-collector-config.yaml
aws lambda update-function-code --function-name <function-name> --zip-file fileb://function.zip
```

This builds the native workflow entry metadata as well as the host Lambda
handler. Match the function architecture to the artifact. Cloud deployment and
collector behavior require their own environment; unit tests do not deploy them.

### 6. Configure Temporal to be able to invoke your lambda function

Refer to the docs [here](https://docs.temporal.io/production-deployment/worker-deployments/serverless-workers/aws-lambda#create-worker-deployment-version).

### 7. Start a Workflow

Use the starter program to execute a Workflow on the Lambda worker, using
the same config file the Lambda uses for connecting to the server:

```bash
TEMPORAL_CONFIG_FILE=temporal.toml tgo run starter/main.go
```
