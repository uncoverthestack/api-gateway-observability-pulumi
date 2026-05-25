// Package observability — CloudWatch alarms for API and Lambda monitoring.
//
// Alarms created:
//   1. High Latency    — triggers when avg latency exceeds threshold (default 3s)
//   2. High Error Rate — triggers when 5xx error rate exceeds threshold (default 5%)
//   3. API Downtime    — triggers when no requests received for 5 minutes
//   4. Lambda Errors   — triggers when any Lambda function throws unhandled errors
//   5. Throttling      — triggers when API Gateway throttles requests
package observability

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudwatch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// AlarmArgs holds inputs for creating CloudWatch alarms.
type AlarmArgs struct {
	Environment          string
	ProjectName          string
	ApiName              pulumi.StringInput
	StageName            string
	AlertTopicArn        pulumi.StringInput
	LambdaFunctionNames  []pulumi.StringInput
	LatencyThresholdMs   int
	ErrorRateThreshold   int
}

// AlarmResult holds the created alarms for inspection.
type AlarmResult struct {
	LatencyAlarm   *cloudwatch.MetricAlarm
	ErrorAlarm     *cloudwatch.MetricAlarm
	DowntimeAlarm  *cloudwatch.MetricAlarm
	LambdaAlarms   []*cloudwatch.MetricAlarm
	ThrottleAlarm  *cloudwatch.MetricAlarm
}

// CreateAlarms provisions all five CloudWatch alarms with SNS actions.
func CreateAlarms(ctx *pulumi.Context, args AlarmArgs) (*AlarmResult, error) {
	if args.LatencyThresholdMs == 0 {
		args.LatencyThresholdMs = 3000
	}
	if args.ErrorRateThreshold == 0 {
		args.ErrorRateThreshold = 5
	}

	namePrefix := fmt.Sprintf("%s-%s", args.ProjectName, args.Environment)
	baseTags := pulumi.StringMap{
		"Environment": pulumi.String(args.Environment),
		"Project":     pulumi.String(args.ProjectName),
	}

	apiDimensions := pulumi.StringMap{
		"ApiName": args.ApiName.ToStringOutput(),
		"Stage":   pulumi.String(args.StageName),
	}

	// 1. High Latency Alarm
	latencyAlarm, err := cloudwatch.NewMetricAlarm(ctx, namePrefix+"-high-latency", &cloudwatch.MetricAlarmArgs{
		AlarmDescription:   pulumi.String(fmt.Sprintf("API latency exceeded %dms (%s)", args.LatencyThresholdMs, args.Environment)),
		Namespace:          pulumi.String("AWS/ApiGateway"),
		MetricName:         pulumi.String("Latency"),
		Dimensions:         apiDimensions,
		Statistic:          pulumi.String("Average"),
		Period:             pulumi.Int(60),
		EvaluationPeriods:  pulumi.Int(2),
		Threshold:          pulumi.Float64(float64(args.LatencyThresholdMs)),
		ComparisonOperator: pulumi.String("GreaterThanThreshold"),
		AlarmActions:       pulumi.StringArray{args.AlertTopicArn.ToStringOutput()},
		OkActions:          pulumi.StringArray{args.AlertTopicArn.ToStringOutput()},
		TreatMissingData:   pulumi.String("notBreaching"),
		Tags:               mergeTags(baseTags, pulumi.StringMap{"AlarmType": pulumi.String("latency")}),
	})
	if err != nil {
		return nil, fmt.Errorf("creating latency alarm: %w", err)
	}

	// 2. High Error Rate Alarm
	errorAlarm, err := cloudwatch.NewMetricAlarm(ctx, namePrefix+"-high-error-rate", &cloudwatch.MetricAlarmArgs{
		AlarmDescription:   pulumi.String(fmt.Sprintf("API 5xx error rate exceeded %d%% (%s)", args.ErrorRateThreshold, args.Environment)),
		Namespace:          pulumi.String("AWS/ApiGateway"),
		MetricName:         pulumi.String("5XXError"),
		Dimensions:         apiDimensions,
		Statistic:          pulumi.String("Average"),
		Period:             pulumi.Int(60),
		EvaluationPeriods:  pulumi.Int(2),
		Threshold:          pulumi.Float64(float64(args.ErrorRateThreshold) / 100.0),
		ComparisonOperator: pulumi.String("GreaterThanThreshold"),
		AlarmActions:       pulumi.StringArray{args.AlertTopicArn.ToStringOutput()},
		OkActions:          pulumi.StringArray{args.AlertTopicArn.ToStringOutput()},
		TreatMissingData:   pulumi.String("notBreaching"),
		Tags:               mergeTags(baseTags, pulumi.StringMap{"AlarmType": pulumi.String("errors")}),
	})
	if err != nil {
		return nil, fmt.Errorf("creating error alarm: %w", err)
	}

	// 3. API Downtime Alarm
	downtimeAlarm, err := cloudwatch.NewMetricAlarm(ctx, namePrefix+"-api-downtime", &cloudwatch.MetricAlarmArgs{
		AlarmDescription:   pulumi.String(fmt.Sprintf("No API requests received for 5 minutes (%s)", args.Environment)),
		Namespace:          pulumi.String("AWS/ApiGateway"),
		MetricName:         pulumi.String("Count"),
		Dimensions:         apiDimensions,
		Statistic:          pulumi.String("Sum"),
		Period:             pulumi.Int(300),
		EvaluationPeriods:  pulumi.Int(1),
		Threshold:          pulumi.Float64(0),
		ComparisonOperator: pulumi.String("LessThanOrEqualToThreshold"),
		AlarmActions:       pulumi.StringArray{args.AlertTopicArn.ToStringOutput()},
		TreatMissingData:   pulumi.String("breaching"),
		Tags:               mergeTags(baseTags, pulumi.StringMap{"AlarmType": pulumi.String("downtime")}),
	})
	if err != nil {
		return nil, fmt.Errorf("creating downtime alarm: %w", err)
	}

	// 4. Lambda Error Alarms — one per function
	lambdaAlarms := make([]*cloudwatch.MetricAlarm, 0, len(args.LambdaFunctionNames))
	for i, fnName := range args.LambdaFunctionNames {
		alarmName := fmt.Sprintf("%s-lambda-errors-%d", namePrefix, i)
		alarm, err := cloudwatch.NewMetricAlarm(ctx, alarmName, &cloudwatch.MetricAlarmArgs{
			AlarmDescription: fnName.ToStringOutput().ApplyT(func(n string) string {
				return fmt.Sprintf("Lambda errors detected on %s (%s)", n, args.Environment)
			}).(pulumi.StringOutput),
			Namespace:          pulumi.String("AWS/Lambda"),
			MetricName:         pulumi.String("Errors"),
			Dimensions:         pulumi.StringMap{"FunctionName": fnName.ToStringOutput()},
			Statistic:          pulumi.String("Sum"),
			Period:             pulumi.Int(60),
			EvaluationPeriods:  pulumi.Int(1),
			Threshold:          pulumi.Float64(1),
			ComparisonOperator: pulumi.String("GreaterThanOrEqualToThreshold"),
			AlarmActions:       pulumi.StringArray{args.AlertTopicArn.ToStringOutput()},
			OkActions:          pulumi.StringArray{args.AlertTopicArn.ToStringOutput()},
			TreatMissingData:   pulumi.String("notBreaching"),
			Tags:               mergeTags(baseTags, pulumi.StringMap{"AlarmType": pulumi.String("lambda")}),
		})
		if err != nil {
			return nil, fmt.Errorf("creating lambda alarm %d: %w", i, err)
		}
		lambdaAlarms = append(lambdaAlarms, alarm)
	}

	// 5. Throttling Alarm
	throttleAlarm, err := cloudwatch.NewMetricAlarm(ctx, namePrefix+"-throttling", &cloudwatch.MetricAlarmArgs{
		AlarmDescription:   pulumi.String(fmt.Sprintf("API requests being throttled (%s)", args.Environment)),
		Namespace:          pulumi.String("AWS/ApiGateway"),
		MetricName:         pulumi.String("Count"),
		Dimensions:         apiDimensions,
		Statistic:          pulumi.String("Sum"),
		Period:             pulumi.Int(60),
		EvaluationPeriods:  pulumi.Int(1),
		Threshold:          pulumi.Float64(1),
		ComparisonOperator: pulumi.String("GreaterThanOrEqualToThreshold"),
		AlarmActions:       pulumi.StringArray{args.AlertTopicArn.ToStringOutput()},
		TreatMissingData:   pulumi.String("notBreaching"),
		Tags:               mergeTags(baseTags, pulumi.StringMap{"AlarmType": pulumi.String("throttle")}),
	})
	if err != nil {
		return nil, fmt.Errorf("creating throttle alarm: %w", err)
	}

	return &AlarmResult{
		LatencyAlarm:  latencyAlarm,
		ErrorAlarm:    errorAlarm,
		DowntimeAlarm: downtimeAlarm,
		LambdaAlarms:  lambdaAlarms,
		ThrottleAlarm: throttleAlarm,
	}, nil
}

// mergeTags combines two tag maps into one.
func mergeTags(base, additional pulumi.StringMap) pulumi.StringMap {
	result := pulumi.StringMap{}
	for k, v := range base {
		result[k] = v
	}
	for k, v := range additional {
		result[k] = v
	}
	return result
}
