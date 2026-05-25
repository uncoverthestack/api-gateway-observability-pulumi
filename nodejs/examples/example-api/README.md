# Example API (Node.js / TypeScript)

A minimal AWS API Gateway deployment for testing the monitoring template.

## What This Deploys

- 1 API Gateway REST API
- 2 Lambda functions (Node.js)
- IAM role with basic execution permissions

Total deployment: about 20 seconds. Estimated cost: under £1/month if left running.

## Endpoints

| Endpoint | Behaviour                                              | Useful For Testing       |
|----------|--------------------------------------------------------|--------------------------|
| `/hello` | Always returns 200 quickly                             | Healthy baseline traffic |
| `/work`  | 70% normal, 15% slow (3-5s), 15% errors (500)          | Latency & error alarms   |

## Quick Start

```bash
# From the nodejs/examples/example-api folder
npm install

# Deploy
pulumi stack init dev
pulumi config set aws:region eu-west-2
pulumi up

# Test it
curl $(pulumi stack output helloEndpoint)
curl $(pulumi stack output workEndpoint)
```

## Using With the Monitoring Template

After deploying this example API, note the outputs:

```bash
pulumi stack output apiGatewayName
pulumi stack output lambdaFunctionNames
pulumi stack output stageName
```

Then go to the parent directory and configure the monitoring template:

```bash
cd ../..

pulumi stack init dev
pulumi config set alertEmail your@email.com
pulumi config set apiGatewayName <value from output>
pulumi config set stageName <value from output>
pulumi config set lambdaFunctionNames <value from output>
pulumi up
```

## Generating Test Traffic

```bash
API_URL=$(pulumi stack output apiUrl)
for i in {1..50}; do curl -s $API_URL/work > /dev/null; done
```

## Cleanup

```bash
pulumi destroy
pulumi stack rm dev
```


This is an example deployment for testing only please do not use these endpoints in production.
Thank you.
