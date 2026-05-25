# API Observability Template Node.js / TypeScript

The TypeScript implementation of the API Gateway observability template, built with Pulumi.

## What This Builds

When you run `pulumi up`, this template provisions monitoring infrastructure for an existing AWS API Gateway:

- **CloudWatch Dashboard** — 10 widgets covering request count, latency percentiles (p50/p90/p99), error rates (4XX/5XX), and Lambda performance
- **5 CloudWatch Alarms** — high latency, high error rate, API downtime, Lambda errors, and request throttling
- **SNS Email Alerts** — alarm notifications delivered to your inbox

Total deployment time: under 30 seconds. No console clicking. No JSON copying.

## Prerequisites

- Node.js 18+
- [Pulumi CLI](https://www.pulumi.com/docs/install/) installed
- AWS CLI configured (`aws configure`)
- An existing AWS API Gateway to monitor (or use the [example API](./examples/example-api/) included)

## Quick Start

```bash
# From the nodejs/ folder
npm install

# Initialise the stack
pulumi stack init dev
pulumi config set aws:region eu-west-2
pulumi config set alertEmail your@email.com

# Point at your existing API Gateway
pulumi config set apiGatewayName your-api-gateway-name
pulumi config set stageName prod
pulumi config set lambdaFunctionNames function-1,function-2,function-3

# Deploy
pulumi up

# Confirm the SNS email subscription (check your inbox)
```

## Don't Have an API to Monitor?

Use the included example API in [`examples/example-api/`](./examples/example-api/). It deploys in 20 seconds.

```bash
cd examples/example-api
npm install
pulumi stack init dev
pulumi up

# Then use the outputs to configure the monitoring template
```

## Finding Your API Gateway Name

```bash
aws apigateway get-rest-apis --query "items[].name" --output text
```

## Finding Your Lambda Function Names

```bash
aws lambda list-functions --query "Functions[].FunctionName" --output text
```

## Configuration

| Config Key              | Required | Default | Description                            |
|-------------------------|----------|---------|----------------------------------------|
| `alertEmail`            | Yes      | —       | Email for alarm notifications          |
| `apiGatewayName`        | Yes      | —       | Name of your existing API Gateway      |
| `stageName`             | No       | `dev`   | API Gateway stage to monitor           |
| `lambdaFunctionNames`   | No       | —       | Comma-separated Lambda function names  |
| `environment`           | No       | `dev`   | Stack environment label                |
| `projectName`           | No       | `api-observability` | Resource name prefix        |
| `latencyThresholdMs`    | No       | `3000`  | Latency alarm threshold (ms)           |
| `errorRateThreshold`    | No       | `5`     | Error rate alarm threshold (%)         |

## Project Structure

```
nodejs/
├── index.ts                          # Entry point — reads config, wires modules
├── Pulumi.yaml                       # Project definition
├── Pulumi.dev.yaml                   # Example config
├── package.json                      # Node.js dependencies
├── tsconfig.json                     # TypeScript config
├── src/
│   └── observability/
│       ├── dashboard.ts              # CloudWatch dashboard (10 widgets)
│       ├── alarms.ts                 # 5 alarm types with SNS actions
│       └── notifications.ts          # SNS topic + email subscription
└── examples/
    └── example-api/                  # Optional test API
```

## Alarms Explained

| Alarm               | Triggers When                                    | Why It Matters                                    |
|---------------------|--------------------------------------------------|---------------------------------------------------|
| **High Latency**     | Avg response time > threshold for 2 minutes      | Users are waiting too long                        |
| **High Error Rate**  | 5XX errors exceed threshold for 2 minutes        | Your API is returning server errors               |
| **API Downtime**     | Zero requests for 5 minutes                      | Your API may be unreachable                       |
| **Lambda Errors**    | Any unhandled exception in a Lambda function     | Code is crashing                                  |
| **Throttling**       | API Gateway starts rejecting requests            | You've hit rate limits                            |

## Multi-Environment Usage

```bash
# Dev
pulumi stack init dev
pulumi config set apiGatewayName my-api-dev
pulumi up

# Prod (stricter thresholds)
pulumi stack init prod
pulumi config set apiGatewayName my-api-prod
pulumi config set latencyThresholdMs 1000
pulumi config set errorRateThreshold 1
pulumi up
```

## Cleanup

```bash
pulumi destroy
pulumi stack rm dev
```

## Looking for Another Language?

This template is also available in:
- [Python](../python/)
- [Go](../golang/) (Available)

