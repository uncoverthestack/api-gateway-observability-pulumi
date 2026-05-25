from typing import List
import pulumi
import pulumi_aws as aws


def create_alarms(
    environment: str,
    project_name: str,
    api_name: pulumi.Input[str],
    stage_name: str,
    alert_topic_arn: pulumi.Input[str],
    lambda_function_names: List[pulumi.Input[str]],
    latency_threshold_ms: int = 3000,
    error_rate_threshold: int = 5,
) -> dict:
    name_prefix = f"{project_name}-{environment}"
    base_tags = {"Environment": environment, "Project": project_name}

    latency_alarm = aws.cloudwatch.MetricAlarm(
        f"{name_prefix}-high-latency",
        alarm_description=f"API latency exceeded {latency_threshold_ms}ms ({environment})",
        namespace="AWS/ApiGateway",
        metric_name="Latency",
        dimensions={"ApiName": api_name, "Stage": stage_name},
        statistic="Average",
        period=60,
        evaluation_periods=2,
        threshold=latency_threshold_ms,
        comparison_operator="GreaterThanThreshold",
        alarm_actions=[alert_topic_arn],
        ok_actions=[alert_topic_arn],
        treat_missing_data="notBreaching",
        tags={**base_tags, "AlarmType": "latency"},
    )

    error_alarm = aws.cloudwatch.MetricAlarm(
        f"{name_prefix}-high-error-rate",
        alarm_description=f"API 5xx error rate exceeded {error_rate_threshold}% ({environment})",
        namespace="AWS/ApiGateway",
        metric_name="5XXError",
        dimensions={"ApiName": api_name, "Stage": stage_name},
        statistic="Average",
        period=60,
        evaluation_periods=2,
        threshold=error_rate_threshold / 100,
        comparison_operator="GreaterThanThreshold",
        alarm_actions=[alert_topic_arn],
        ok_actions=[alert_topic_arn],
        treat_missing_data="notBreaching",
        tags={**base_tags, "AlarmType": "errors"},
    )

    downtime_alarm = aws.cloudwatch.MetricAlarm(
        f"{name_prefix}-api-downtime",
        alarm_description=f"No API requests received for 5 minutes ({environment})",
        namespace="AWS/ApiGateway",
        metric_name="Count",
        dimensions={"ApiName": api_name, "Stage": stage_name},
        statistic="Sum",
        period=300,
        evaluation_periods=1,
        threshold=0,
        comparison_operator="LessThanOrEqualToThreshold",
        alarm_actions=[alert_topic_arn],
        treat_missing_data="breaching",
        tags={**base_tags, "AlarmType": "downtime"},
    )

    lambda_alarms = []
    for i, fn_name in enumerate(lambda_function_names):
        alarm = aws.cloudwatch.MetricAlarm(
            f"{name_prefix}-lambda-errors-{i}",
            alarm_description=pulumi.Output.from_input(fn_name).apply(
                lambda n: f"Lambda errors detected on {n} ({environment})"
            ),
            namespace="AWS/Lambda",
            metric_name="Errors",
            dimensions={"FunctionName": fn_name},
            statistic="Sum",
            period=60,
            evaluation_periods=1,
            threshold=1,
            comparison_operator="GreaterThanOrEqualToThreshold",
            alarm_actions=[alert_topic_arn],
            ok_actions=[alert_topic_arn],
            treat_missing_data="notBreaching",
            tags={**base_tags, "AlarmType": "lambda"},
        )
        lambda_alarms.append(alarm)

    throttle_alarm = aws.cloudwatch.MetricAlarm(
        f"{name_prefix}-throttling",
        alarm_description=f"API requests being throttled ({environment})",
        namespace="AWS/ApiGateway",
        metric_name="Count",
        dimensions={"ApiName": api_name, "Stage": stage_name},
        statistic="Sum",
        period=60,
        evaluation_periods=1,
        threshold=1,
        comparison_operator="GreaterThanOrEqualToThreshold",
        alarm_actions=[alert_topic_arn],
        treat_missing_data="notBreaching",
        tags={**base_tags, "AlarmType": "throttle"},
    )

    return {
        "latency_alarm": latency_alarm,
        "error_alarm": error_alarm,
        "downtime_alarm": downtime_alarm,
        "lambda_alarms": lambda_alarms,
        "throttle_alarm": throttle_alarm,
    }
