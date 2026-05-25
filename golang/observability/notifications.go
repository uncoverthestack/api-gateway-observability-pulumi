package observability

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/sns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// NotificationArgs holds inputs for creating the notification infrastructure.
type NotificationArgs struct {
	Environment string
	ProjectName string
	AlertEmail  string
}

// NotificationResult holds the created resources for downstream use.
type NotificationResult struct {
	AlertTopic    *sns.Topic
	AlertTopicArn pulumi.StringOutput
}

// CreateNotifications provisions an SNS topic and email subscription.
// After deployment, check your inbox and confirm the subscription.
func CreateNotifications(ctx *pulumi.Context, args NotificationArgs) (*NotificationResult, error) {
	namePrefix := fmt.Sprintf("%s-%s", args.ProjectName, args.Environment)
	tags := pulumi.StringMap{
		"Environment": pulumi.String(args.Environment),
		"Project":     pulumi.String(args.ProjectName),
	}

	// SNS Topic
	alertTopic, err := sns.NewTopic(ctx, namePrefix+"-alerts", &sns.TopicArgs{
		DisplayName: pulumi.String(fmt.Sprintf("%s API Alerts (%s)", args.ProjectName, args.Environment)),
		Tags:        tags,
	})
	if err != nil {
		return nil, fmt.Errorf("creating SNS topic: %w", err)
	}

	// Email subscription
	_, err = sns.NewTopicSubscription(ctx, namePrefix+"-email-sub", &sns.TopicSubscriptionArgs{
		Topic:    alertTopic.Arn,
		Protocol: pulumi.String("email"),
		Endpoint: pulumi.String(args.AlertEmail),
	})
	if err != nil {
		return nil, fmt.Errorf("creating email subscription: %w", err)
	}

	// Topic policy — allow CloudWatch to publish
	policyDoc := alertTopic.Arn.ApplyT(func(arn string) (string, error) {
		policy := map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []map[string]interface{}{
				{
					"Effect":    "Allow",
					"Principal": map[string]string{"Service": "cloudwatch.amazonaws.com"},
					"Action":    "sns:Publish",
					"Resource":  arn,
				},
			},
		}
		bytes, err := json.Marshal(policy)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}).(pulumi.StringOutput)

	_, err = sns.NewTopicPolicy(ctx, namePrefix+"-alert-policy", &sns.TopicPolicyArgs{
		Arn:    alertTopic.Arn,
		Policy: policyDoc,
	})
	if err != nil {
		return nil, fmt.Errorf("creating topic policy: %w", err)
	}

	return &NotificationResult{
		AlertTopic:    alertTopic,
		AlertTopicArn: alertTopic.Arn,
	}, nil
}
