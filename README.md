# ClassiFaaS Deployment and Benchmarking Tool

ClassiFaaS is a deployment and benchmarking tool for serverless functions across multiple cloud providers (AWS, Azure, GCP, and Alibaba Cloud). It enables performance evaluations of serverless functions and analysis of heterogeneous hardware impacts on function execution.

## Folder Structure
- `cmd`: Main command-line applications for deployment and benchmarking.
- `configs`: Configuration files for deployment and benchmarking.
- `internal`: Internal packages for deployment and benchmarking logic.
- `credentials`: Credential files for cloud provider access (GCP only).
- `deployment`: Cloud provider-specific deployment scripts and function code.
  - `deployment/shared`: Benchmarks shared across all cloud providers.
  - `deployment/{provider}`: Provider-specific deployment scripts and function implementations.
  - `deployment/{provider}/manage-deployment.sh`: Interface between Go deployment code and provider-specific deployment commands. Start here if you need to modify deployment logic.

## Prerequisites

Install the following tools before getting started:

- [Go](https://go.dev/doc/install) (1.18+)
- [Node.js](https://nodejs.org/) (20+)
- [npm](https://www.npmjs.com/) (8+)

### Cloud CLIs and Frameworks

- [Azure CLI](https://learn.microsoft.com/cli/azure/install-azure-cli) and [Azure Functions Core Tools](https://learn.microsoft.com/azure/azure-functions/functions-run-local)
- [Serverless Framework](https://www.serverless.com/framework/docs/getting-started/) (AWS): `npm install -g serverless`
- [gcloud CLI](https://cloud.google.com/sdk/docs/install) (GCP)
- [Serverless Devs](https://github.com/Serverless-Devs/Serverless-Devs) (Alibaba): `npm install -g @serverless-devs/s`

## Install Dependencies for Benchmarks

Navigate to each cloud provider's `deployment/{provider}` directory and run:

```bash
npm install
```

For Alibaba Cloud, run `npm install` inside `deployment/alibaba/src` instead.

## Credentials and Authentication

Ensure you have set up your cloud accounts and logged in via their respective CLIs.

- **GCP**: Place a service account key at `credentials/gcp_service_account.json` with permissions for function deployment and management. The benchmark uses this to obtain access tokens for invoking deployed functions.
- **Azure**: Run `az login` and select the correct subscription.

## Deployment

### 1) Configure Deployment Parameters

Update `configs/deployment.yaml` to match your desired deployment configuration. This file will also containe the parameters for the generated benchmark config.

### 2) Deploy Functions

```bash
go run ./cmd/deploy deploy
```

### 3) Generate Benchmark Config

```bash
go run ./cmd/deploy generate
```

This creates `configs/generated.yaml` with parameters based on the deployment config.

### 4) Remove Deployment

Use the same configuration file used for deployment when removing resources:

```bash
go run ./cmd/deploy remove
```

## Benchmarking

Run the benchmark with either the generated config or a custom file:

```bash
go run ./cmd/bench --config configs/generated.yaml
```

## Continuous Benchmarking

We recommend scheduling runs with `cron`. For example, to run every 6 hours:

```cron
0 */6 * * * /path/to/go run /path/to/ClassiFaaS/cmd/bench --config /path/to/ClassiFaaS/configs/generated.yaml
```

Adjust paths as needed for your environment.
