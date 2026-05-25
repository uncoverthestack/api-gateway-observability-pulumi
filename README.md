# API Gateway Observability Pulumi Templates

A reusable Pulumi template that provisions CloudWatch dashboards, automated alarms, and SNS email alerts for any AWS API Gateway in under 30 seconds.

Available in **Python**, **Node.js / TypeScript** and **Go** available.

## What It Does

Point this at any existing API Gateway, run `pulumi up`, and get a complete monitoring setup:

- **CloudWatch Dashboard** — request count, latency percentiles (p50/p90/p99), 4XX/5XX errors, Lambda performance
- **5 CloudWatch Alarms** — high latency, high error rate, API downtime, Lambda errors, throttling
- **SNS Email Alerts** — instant notifications when something breaks

No console clicking. No manual JSON. Just config and deploy.

## Choose Your Language

| Language | Status | Folder |
|----------|--------|--------|
| Python | ✅ Available | [`python/`](./python/) |
| Node.js / TypeScript | ✅ Available| [`nodejs/`](./nodejs/) |
| Go | ✅ Available | [`golang/`](./golang/) |

Each language folder is self-contained with its own README and deployment instructions.

## Why Multiple Languages?

Pulumi lets you write infrastructure in real programming languages not config files. Teams already have a preferred language for their codebase, and forcing them to learn a new one just for infrastructure is friction. This template provides identical functionality across Python, Node.js, and Go so any team can adopt it.

## Quick Start (Python)

```bash
cd python/
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
pulumi stack init dev
pulumi config set alert_email your@email.com
pulumi config set api_gateway_name your-api-name
pulumi up
```

See the [Python README](./python/README.md) for full instructions.

## Don't Have an API to Monitor?

Each language version includes an example API in its `examples/` folder. Deploy that first, then point the monitoring template at it.

## What You Get

A CloudWatch dashboard like this:

> Screenshots coming soon  see `docs/images/`

An alarm email like this when something breaks:

> Screenshots coming soon — see `docs/images/`

## Project Structure

```
api-gateway-observability-pulumi/
├── python/              # Python implementation (available)
│   ├── src/
│   ├── examples/
│   ├── __main__.py
│   └── README.md
├── nodejs/              # Node.js implementation (coming soon)
├── golang/              # Go implementation (coming soon)
├── docs/                # Shared documentation and screenshots
├── README.md            # This file
└── LICENSE
```

## Contributing

Found a bug or want to add a feature? Open an issue or pull request. If you're porting the template to another language not listed here, that contribution is especially welcome.

## License

MIT
