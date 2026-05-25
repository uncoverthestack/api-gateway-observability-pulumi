import json
import pulumi
import pulumi_aws as aws


def create_notifications(environment: str, project_name: str, alert_email: str) -> dict:
    name_prefix = f"{project_name}-{environment}"

    alert_topic = aws.sns.Topic(
        f"{name_prefix}-alerts",
        display_name=f"{project_name} API Alerts ({environment})",
        tags={"Environment": environment, "Project": project_name},
    )

    aws.sns.TopicSubscription(
        f"{name_prefix}-email-sub",
        topic=alert_topic.arn,
        protocol="email",
        endpoint=alert_email,
    )

    aws.sns.TopicPolicy(
        f"{name_prefix}-alert-policy",
        arn=alert_topic.arn,
        policy=alert_topic.arn.apply(
            lambda arn: json.dumps(
                {
                    "Version": "2012-10-17",
                    "Statement": [
                        {
                            "Effect": "Allow",
                            "Principal": {"Service": "cloudwatch.amazonaws.com"},
                            "Action": "sns:Publish",
                            "Resource": arn,
                        }
                    ],
                }
            )
        ),
    )

    return {
        "alert_topic": alert_topic,
        "alert_topic_arn": alert_topic.arn,
    }
