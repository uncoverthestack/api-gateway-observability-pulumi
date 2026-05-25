import json
from typing import List
import pulumi
import pulumi_aws as aws


def create_dashboard(
    environment: str,
    project_name: str,
    api_name: pulumi.Input[str],
    stage_name: str,
    lambda_function_names: List[pulumi.Input[str]],
    region: str,
) -> dict:
    name_prefix = f"{project_name}-{environment}"
    dashboard_name = f"{name_prefix}-api-observability"

    dashboard_body = pulumi.Output.all(
        api_name, *lambda_function_names
    ).apply(lambda args: _build_dashboard_json(args[0], list(args[1:]), stage_name, region))

    dashboard = aws.cloudwatch.Dashboard(
        f"{name_prefix}-dashboard",
        dashboard_name=dashboard_name,
        dashboard_body=dashboard_body,
    )

    return {"dashboard": dashboard, "dashboard_name": dashboard_name}


def _build_dashboard_json(
    resolved_api_name: str,
    resolved_fn_names: List[str],
    stage_name: str,
    region: str,
) -> str:
    def lambda_metrics(metric_name: str) -> list:
        stat = "Average" if metric_name == "Duration" else "Sum"
        return [
            ["AWS/Lambda", metric_name, "FunctionName", fn, {"label": fn, "stat": stat}]
            for fn in resolved_fn_names
        ]

    widgets = [
        {
            "type": "metric", "x": 0, "y": 0, "width": 8, "height": 6,
            "properties": {
                "title": "API Request Count",
                "metrics": [["AWS/ApiGateway", "Count", "ApiName", resolved_api_name, "Stage", stage_name, {"stat": "Sum", "period": 60}]],
                "view": "timeSeries", "stacked": False, "region": region, "period": 60,
            },
        },
        {
            "type": "metric", "x": 8, "y": 0, "width": 8, "height": 6,
            "properties": {
                "title": "5XX Errors",
                "metrics": [["AWS/ApiGateway", "5XXError", "ApiName", resolved_api_name, "Stage", stage_name, {"stat": "Sum", "period": 60, "color": "#d62728"}]],
                "view": "timeSeries", "region": region, "period": 60,
            },
        },
        {
            "type": "metric", "x": 16, "y": 0, "width": 8, "height": 6,
            "properties": {
                "title": "4XX Errors",
                "metrics": [["AWS/ApiGateway", "4XXError", "ApiName", resolved_api_name, "Stage", stage_name, {"stat": "Sum", "period": 60, "color": "#ff7f0e"}]],
                "view": "timeSeries", "region": region, "period": 60,
            },
        },
        {
            "type": "metric", "x": 0, "y": 6, "width": 12, "height": 6,
            "properties": {
                "title": "API Latency (ms)",
                "metrics": [
                    ["AWS/ApiGateway", "Latency", "ApiName", resolved_api_name, "Stage", stage_name, {"stat": "Average", "label": "Avg", "color": "#2ca02c"}],
                    ["...", {"stat": "p50", "label": "p50", "color": "#1f77b4"}],
                    ["...", {"stat": "p90", "label": "p90", "color": "#ff7f0e"}],
                    ["...", {"stat": "p99", "label": "p99", "color": "#d62728"}],
                ],
                "view": "timeSeries", "region": region, "period": 60,
                "yAxis": {"left": {"min": 0, "label": "ms"}},
            },
        },
        {
            "type": "metric", "x": 12, "y": 6, "width": 12, "height": 6,
            "properties": {
                "title": "Integration Latency (ms)",
                "metrics": [
                    ["AWS/ApiGateway", "IntegrationLatency", "ApiName", resolved_api_name, "Stage", stage_name, {"stat": "Average", "label": "Avg"}],
                    ["...", {"stat": "p90", "label": "p90", "color": "#ff7f0e"}],
                ],
                "view": "timeSeries", "region": region, "period": 60,
                "yAxis": {"left": {"min": 0, "label": "ms"}},
            },
        },
        {
            "type": "metric", "x": 0, "y": 12, "width": 12, "height": 6,
            "properties": {
                "title": "Lambda Duration (ms)",
                "metrics": lambda_metrics("Duration"),
                "view": "timeSeries", "region": region, "period": 60,
            },
        },
        {
            "type": "metric", "x": 12, "y": 12, "width": 12, "height": 6,
            "properties": {
                "title": "Lambda Errors",
                "metrics": lambda_metrics("Errors"),
                "view": "timeSeries", "region": region, "period": 60,
            },
        },
        {
            "type": "metric", "x": 0, "y": 18, "width": 12, "height": 6,
            "properties": {
                "title": "Lambda Invocations",
                "metrics": lambda_metrics("Invocations"),
                "view": "timeSeries", "stacked": True, "region": region, "period": 60,
            },
        },
        {
            "type": "metric", "x": 12, "y": 18, "width": 12, "height": 6,
            "properties": {
                "title": "Lambda Throttles",
                "metrics": lambda_metrics("Throttles"),
                "view": "timeSeries", "region": region, "period": 60,
            },
        },
    ]

    return json.dumps({"widgets": widgets})
