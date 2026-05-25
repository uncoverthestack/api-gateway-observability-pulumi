import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";

export interface NotificationArgs {
    environment: string;
    projectName: string;
    alertEmail: string;
}

export interface NotificationResult {
    alertTopic: aws.sns.Topic;
    alertTopicArn: pulumi.Output<string>;
}

export function createNotifications(args: NotificationArgs): NotificationResult {
    const { environment, projectName, alertEmail } = args;
    const namePrefix = `${projectName}-${environment}`;

    // SNS Topic
    const alertTopic = new aws.sns.Topic(`${namePrefix}-alerts`, {
        displayName: `${projectName} API Alerts (${environment})`,
        tags: { Environment: environment, Project: projectName },
    });

    // Email subscription
    new aws.sns.TopicSubscription(`${namePrefix}-email-sub`, {
        topic: alertTopic.arn,
        protocol: "email",
        endpoint: alertEmail,
    });

    // Topic policy — allow CloudWatch to publish
    new aws.sns.TopicPolicy(`${namePrefix}-alert-policy`, {
        arn: alertTopic.arn,
        policy: alertTopic.arn.apply((arn) =>
            JSON.stringify({
                Version: "2012-10-17",
                Statement: [
                    {
                        Effect: "Allow",
                        Principal: { Service: "cloudwatch.amazonaws.com" },
                        Action: "sns:Publish",
                        Resource: arn,
                    },
                ],
            })
        ),
    });

    return {
        alertTopic,
        alertTopicArn: alertTopic.arn,
    };
}
