# Example API

A minimal AWS API Gateway deployment for testing the [api-gateway-observability-pulumi](../README.md) monitoring template.

If you don't already have an API to monitor, deploy this first — it gives you a working API Gateway with two endpoints that generate realistic traffic patterns (successful responses, slow responses, and errors).

## What This Deploys

- 1 API Gateway REST API
- 2 Lambda functions (Node.js)
- IAM role with basic execution permissions

Total deployment: **about 20 seconds**. Estimated cost: **under £1/month** if left running.

## Endpoints

| Endpoint | Behaviour                                              | Useful For Testing       |
|----------|--------------------------------------------------------|--------------------------|
| `/hello` | Always returns 200 quickly                             | Healthy baseline traffic |
| `/work`  | 70% normal, 15% slow (3-5s), 15% errors (500)          | Latency & error alarms   |

## Quick Start

```bash
# From the project root, navigate into this folder
cd examples/example-api

# Setup
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# Deploy
pulumi stack init dev
pulumi config set aws:region eu-west-2
pulumi up

# Test it
curl $(pulumi stack output hello_endpoint)
curl $(pulumi stack output work_endpoint)
```

## Using With the Monitoring Template

After deploying this example API, note the outputs:

```bash
pulumi stack output api_gateway_name
pulumi stack output lambda_function_names
pulumi stack output stage_name
```

Then go to the parent directory and configure the monitoring template with those values:

```bash
cd ../..

pulumi stack init dev
pulumi config set alert_email your@email.com
pulumi config set api_gateway_name <value from output>
pulumi config set stage_name <value from output>
pulumi config set lambda_function_names <value from output>
pulumi up
```

Now the monitoring template is watching the example API. Hit the endpoints with curl a few times and watch your CloudWatch dashboard populate.

## Generating Test Traffic

Quick traffic burst:

```bash
API_URL=$(pulumi stack output api_url)
for i in {1..50}; do curl -s $API_URL/work > /dev/null; done
```

This will generate enough traffic to trigger both the latency and error rate alarms within a couple of minutes.

## Cleanup

```bash
pulumi destroy
pulumi stack rm dev
```

---

This is an example deployment for testing only — do not use these endpoints in production.
