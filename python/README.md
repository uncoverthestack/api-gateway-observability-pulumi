# API Observability Template — Python

The Python implementation of the API Gateway observability template, built with Pulumi.

## What This Builds

When you run `pulumi up`, this template provisions monitoring infrastructure for an existing AWS API Gateway:

- **CloudWatch Dashboard** — 10 widgets covering request count, latency percentiles (p50/p90/p99), error rates (4XX/5XX), and Lambda performance
- **5 CloudWatch Alarms** — high latency, high error rate, API downtime, Lambda errors, and request throttling
- **SNS Email Alerts** — alarm notifications delivered to your inbox

Total deployment time: under 30 seconds. No console clicking. No JSON copying.

## Prerequisites

- Python 3.9+
- [Pulumi CLI](https://www.pulumi.com/docs/install/) installed
- AWS CLI configured (`aws configure`)
- An existing AWS API Gateway to monitor (or use the [example API](./examples/example-api/) included)

## Quick Start

```bash
# From the python/ folder
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# Initialise the stack
pulumi stack init dev
pulumi config set aws:region eu-west-2
pulumi config set alert_email your@email.com

# Point at your existing API Gateway
pulumi config set api_gateway_name your-api-gateway-name
pulumi config set stage_name prod
pulumi config set lambda_function_names function-1,function-2,function-3

# Deploy
pulumi up

# Confirm the SNS email subscription (check your inbox)
```

## Don't Have an API to Monitor?

Use the included example API in [`examples/example-api/`](./examples/example-api/). It deploys in 20 seconds and gives you a working API Gateway with two endpoints that generate realistic test traffic.

```bash
cd examples/example-api
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
pulumi stack init dev
pulumi up

# Then use the outputs to configure the monitoring template
```

## Finding Your API Gateway Name

In the AWS console, go to API Gateway → your API. The name at the top is what you pass to `api_gateway_name`. Or via CLI:

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
| `alert_email`           | Yes      | —       | Email for alarm notifications          |
| `api_gateway_name`      | Yes      | —       | Name of your existing API Gateway      |
| `stage_name`            | No       | `dev`   | API Gateway stage to monitor           |
| `lambda_function_names` | No       | —       | Comma-separated Lambda function names  |
| `environment`           | No       | `dev`   | Stack environment label                |
| `project_name`          | No       | `api-observability` | Resource name prefix        |
| `latency_threshold_ms`  | No       | `3000`  | Latency alarm threshold (ms)           |
| `error_rate_threshold`  | No       | `5`     | Error rate alarm threshold (%)         |

## Project Structure

```
python/
├── __main__.py                       # Entry point — reads config, wires modules
├── Pulumi.yaml                       # Project definition
├── Pulumi.dev.yaml                   # Example config
├── requirements.txt                  # Python dependencies
├── src/
│   └── observability/
│       ├── dashboard.py              # CloudWatch dashboard (10 widgets)
│       ├── alarms.py                 # 5 alarm types with SNS actions
│       └── notifications.py          # SNS topic + email subscription
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

Deploy the same monitoring to dev, staging, and prod:

```bash
# Dev
pulumi stack init dev
pulumi config set api_gateway_name my-api-dev
pulumi up

# Prod (stricter thresholds)
pulumi stack init prod
pulumi config set api_gateway_name my-api-prod
pulumi config set latency_threshold_ms 1000
pulumi config set error_rate_threshold 1
pulumi up
```

## Cleanup

```bash
pulumi destroy
pulumi stack rm dev
```

## Looking for Another Language?

This template is also available in:
- [Node.js / TypeScript](../nodejs/) (coming soon)
- [Go](../golang/) (coming soon)
