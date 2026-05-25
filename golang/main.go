// API Observability & Monitoring Platform
//
// Provisions CloudWatch dashboards, automated alarms, and SNS alerting
// for any existing AWS API Gateway.
//
// Usage:
//   pulumi stack init dev
//   pulumi config set alertEmail your@email.com
//   pulumi config set apiGatewayName your-api-name
//   pulumi config set stageName prod
//   pulumi config set lambdaFunctionNames function-1,function-2
//   pulumi up
package main

import (
	"fmt"
	"strings"

	"api-observability-platform/observability"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		awsCfg := config.New(ctx, "aws")

		// Configuration with defaults
		environment := cfg.Get("environment")
		if environment == "" {
			environment = "dev"
		}

		projectName := cfg.Get("projectName")
		if projectName == "" {
			projectName = "api-observability"
		}

		alertEmail := cfg.Require("alertEmail")

		latencyThresholdMs := cfg.GetInt("latencyThresholdMs")
		if latencyThresholdMs == 0 {
			latencyThresholdMs = 3000
		}

		errorRateThreshold := cfg.GetInt("errorRateThreshold")
		if errorRateThreshold == 0 {
			errorRateThreshold = 5
		}

		region := awsCfg.Get("region")
		if region == "" {
			region = "eu-west-2"
		}

		// Existing API Gateway details
		apiGatewayName := cfg.Require("apiGatewayName")
		stageName := cfg.Get("stageName")
		if stageName == "" {
			stageName = "dev"
		}

		lambdaNamesRaw := cfg.Get("lambdaFunctionNames")
		var lambdaFunctionNames []pulumi.StringInput
		if lambdaNamesRaw != "" {
			for _, n := range strings.Split(lambdaNamesRaw, ",") {
				trimmed := strings.TrimSpace(n)
				if trimmed != "" {
					lambdaFunctionNames = append(lambdaFunctionNames, pulumi.String(trimmed))
				}
			}
		}

		// 1. Setup Notifications
		notifications, err := observability.CreateNotifications(ctx, observability.NotificationArgs{
			Environment: environment,
			ProjectName: projectName,
			AlertEmail:  alertEmail,
		})
		if err != nil {
			return fmt.Errorf("creating notifications: %w", err)
		}

		// 2. Create Alarms
		_, err = observability.CreateAlarms(ctx, observability.AlarmArgs{
			Environment:         environment,
			ProjectName:         projectName,
			ApiName:             pulumi.String(apiGatewayName),
			StageName:           stageName,
			AlertTopicArn:       notifications.AlertTopicArn,
			LambdaFunctionNames: lambdaFunctionNames,
			LatencyThresholdMs:  latencyThresholdMs,
			ErrorRateThreshold:  errorRateThreshold,
		})
		if err != nil {
			return fmt.Errorf("creating alarms: %w", err)
		}

		// 3. Create Dashboard
		dashboard, err := observability.CreateDashboard(ctx, observability.DashboardArgs{
			Environment:         environment,
			ProjectName:         projectName,
			ApiName:             pulumi.String(apiGatewayName),
			StageName:           stageName,
			LambdaFunctionNames: lambdaFunctionNames,
			Region:              region,
		})
		if err != nil {
			return fmt.Errorf("creating dashboard: %w", err)
		}

		// Stack Outputs
		dashboardURL := fmt.Sprintf(
			"https://%s.console.aws.amazon.com/cloudwatch/home?region=%s#dashboards:name=%s",
			region, region, dashboard.DashboardName,
		)
		ctx.Export("dashboardUrl", pulumi.String(dashboardURL))
		ctx.Export("alertTopicArn", notifications.AlertTopicArn)
		ctx.Export("monitoredApi", pulumi.String(apiGatewayName))
		ctx.Export("monitoredStage", pulumi.String(stageName))
		ctx.Export("environment", pulumi.String(environment))

		return nil
	})
}
