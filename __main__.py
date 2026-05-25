import pulumi

from src.observability.notifications import create_notifications
from src.observability.alarms import create_alarms
from src.observability.dashboard import create_dashboard

# Configuration 
config = pulumi.Config()
aws_config = pulumi.Config("aws")

environment = config.get("environment") or "dev"
project_name = config.get("project_name") or "api-observability"
alert_email = config.require("alert_email")
latency_threshold_ms = config.get_int("latency_threshold_ms") or 3000
error_rate_threshold = config.get_int("error_rate_threshold") or 5
region = aws_config.get("region") or "eu-west-2"

# Your existing API Gateway details
api_gateway_name = config.require("api_gateway_name")
stage_name = config.get("stage_name") or "dev"
lambda_names_raw = config.get("lambda_function_names") or ""
lambda_function_names = [n.strip() for n in lambda_names_raw.split(",") if n.strip()]

# 1. Setup Notifications 
notifications = create_notifications(
    environment=environment,
    project_name=project_name,
    alert_email=alert_email,
)

# 2. Create Alarms
alarms = create_alarms(
    environment=environment,
    project_name=project_name,
    api_name=api_gateway_name,
    stage_name=stage_name,
    alert_topic_arn=notifications["alert_topic_arn"],
    lambda_function_names=lambda_function_names,
    latency_threshold_ms=latency_threshold_ms,
    error_rate_threshold=error_rate_threshold,
)

# 3. Create Dashboard
dashboard = create_dashboard(
    environment=environment,
    project_name=project_name,
    api_name=api_gateway_name,
    stage_name=stage_name,
    lambda_function_names=lambda_function_names,
    region=region,
)

# Stack Outputs
pulumi.export(
    "dashboard_url",
    pulumi.Output.from_input(
        f"https://{region}.console.aws.amazon.com/cloudwatch/home"
        f"?region={region}#dashboards:name={dashboard['dashboard_name']}"
    ),
)
pulumi.export("alert_topic_arn", notifications["alert_topic_arn"])
pulumi.export("monitored_api", api_gateway_name)
pulumi.export("monitored_stage", stage_name)
pulumi.export("monitored_functions", lambda_function_names)
pulumi.export("environment", environment)
