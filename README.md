# API Observability & Monitoring Platform

A reusable **Pulumi** template that provisions CloudWatch dashboards, automated alarms, and SNS alerting for any AWS API Gateway in under 30 seconds.

## What It Does

Point this at any existing API Gateway, run `pulumi up`, and get:

- **CloudWatch Dashboard** — 10 widgets covering request count, latency percentiles (p50/p90/p99), error rates (4XX/5XX), Lambda duration, invocations, and throttles
- **5 CloudWatch Alarms** — high latency, high error rate, API downtime, Lambda errors, and request throttling
- **SNS Email Alerts** — alarm notifications delivered to your inbox with direct links to the affected alarm

No console clicking. No manual setup. No copy-pasting JSON. Just config and deploy.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    CloudWatch Dashboard                      │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐   │
│  │ Traffic   │ │ Latency  │ │ Errors   │ │ Lambda Stats │   │
│  │ Volume    │ │ p50/p90  │ │ 4xx/5xx  │ │ Duration     │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────────┘   │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│              Your Existing API Gateway                       │
│         (this template monitors it — doesn't create it)     │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    CloudWatch Alarms → SNS → Email           │
│  High Latency │ High Error Rate │ API Downtime              │
│  Lambda Errors │ Throttling                                  │
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites
- Python 3.9+
- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- AWS CLI configured (`aws configure`)

### Deploy

```bash
# Clone and setup
git clone https://github.com/YOUR_USERNAME/api-observability-platform.git
cd api-observability-platform
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# Configure — point at your existing API Gateway
pulumi stack init prod
pulumi config set aws:region eu-west-2
pulumi config set alert_email your@email.com
pulumi config set api_gateway_name your-api-gateway-name
pulumi config set stage_name prod
pulumi config set lambda_function_names my-create-fn,my-read-fn,my-update-fn

# Deploy
pulumi up

# Confirm the SNS email subscription (check your inbox)
```

### Finding Your API Gateway Name

In the AWS console, go to API Gateway → your API. The name shown at the top is what you pass to `api_gateway_name`. Or via CLI:

```bash
aws apigateway get-rest-apis --query "items[].name" --output text
```

### Finding Your Lambda Function Names

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

## What Gets Created

| Resource              | Count | Purpose                              |
|-----------------------|-------|--------------------------------------|
| CloudWatch Dashboard  | 1     | 10 monitoring widgets                |
| CloudWatch Alarms     | 5+    | Latency, errors, downtime, throttle  |
| SNS Topic             | 1     | Alarm notification delivery          |
| SNS Subscription      | 1     | Email endpoint                       |

**Estimated cost:** CloudWatch dashboards are $3/month each (first 3 free). Alarms are $0.10/month each. SNS email is free.

## Alarms Explained

| Alarm               | Triggers When                                    | Why It Matters                                    |
|----------------------|--------------------------------------------------|---------------------------------------------------|
| **High Latency**     | Avg response time > threshold for 2 minutes      | Users are waiting too long                        |
| **High Error Rate**  | 5XX errors exceed threshold for 2 minutes        | Your API is returning server errors               |
| **API Downtime**     | Zero requests for 5 minutes                      | Your API may be unreachable                       |
| **Lambda Errors**    | Any unhandled exception in a Lambda function     | Code is crashing                                  |
| **Throttling**       | API Gateway starts rejecting requests            | You've hit rate limits                            |

## Project Structure

```
api-observability-platform/
├── __main__.py                       # Entry point — reads config, wires modules
├── Pulumi.yaml                       # Project definition
├── Pulumi.dev.yaml                   # Example config
├── requirements.txt
└── src/
    └── observability/
        ├── dashboard.py              # CloudWatch dashboard (10 widgets)
        ├── alarms.py                 # 5 alarm types with SNS actions
        └── notifications.py          # SNS topic + email subscription
```

## Multi-Environment Usage

Deploy the same monitoring to dev, staging, and prod:

```bash
# Dev
pulumi stack init dev
pulumi config set api_gateway_name my-api-dev
pulumi config set stage_name dev
pulumi up

# Prod
pulumi stack init prod
pulumi config set api_gateway_name my-api-prod
pulumi config set stage_name prod
pulumi config set latency_threshold_ms 1000    # Stricter threshold for prod
pulumi config set error_rate_threshold 1       # Lower tolerance in prod
pulumi up
```

## Cleanup

```bash
pulumi destroy
pulumi stack rm dev
```

## Built With

- [Pulumi](https://www.pulumi.com/) — Infrastructure as Code (Python)
- [AWS CloudWatch](https://aws.amazon.com/cloudwatch/) — Monitoring & Observability
- [AWS SNS](https://aws.amazon.com/sns/) — Alert Notifications

## License

MIT
