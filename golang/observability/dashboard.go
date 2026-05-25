package observability

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudwatch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// DashboardArgs holds inputs for creating the CloudWatch dashboard.
type DashboardArgs struct {
	Environment         string
	ProjectName         string
	ApiName             pulumi.StringInput
	StageName           string
	LambdaFunctionNames []pulumi.StringInput
	Region              string
}

// DashboardResult holds the created dashboard.
type DashboardResult struct {
	Dashboard     *cloudwatch.Dashboard
	DashboardName string
}

// CreateDashboard provisions the CloudWatch dashboard with all 10 widgets.
func CreateDashboard(ctx *pulumi.Context, args DashboardArgs) (*DashboardResult, error) {
	namePrefix := fmt.Sprintf("%s-%s", args.ProjectName, args.Environment)
	dashboardName := namePrefix + "-api-observability"

	// Collect all outputs that need resolving before building the JSON.
	allInputs := make([]interface{}, 0, len(args.LambdaFunctionNames)+1)
	allInputs = append(allInputs, args.ApiName.ToStringOutput())
	for _, fn := range args.LambdaFunctionNames {
		allInputs = append(allInputs, fn.ToStringOutput())
	}

	dashboardBody := pulumi.All(allInputs...).ApplyT(func(resolved []interface{}) (string, error) {
		apiName := resolved[0].(string)
		fnNames := make([]string, 0, len(resolved)-1)
		for _, v := range resolved[1:] {
			fnNames = append(fnNames, v.(string))
		}
		return buildDashboardJSON(apiName, fnNames, args.StageName, args.Region)
	}).(pulumi.StringOutput)

	dashboard, err := cloudwatch.NewDashboard(ctx, namePrefix+"-dashboard", &cloudwatch.DashboardArgs{
		DashboardName: pulumi.String(dashboardName),
		DashboardBody: dashboardBody,
	})
	if err != nil {
		return nil, fmt.Errorf("creating dashboard: %w", err)
	}

	return &DashboardResult{
		Dashboard:     dashboard,
		DashboardName: dashboardName,
	}, nil
}

// buildDashboardJSON constructs the CloudWatch dashboard JSON body.
func buildDashboardJSON(apiName string, fnNames []string, stageName, region string) (string, error) {
	lambdaMetrics := func(metricName string) []interface{} {
		stat := "Sum"
		if metricName == "Duration" {
			stat = "Average"
		}
		metrics := make([]interface{}, 0, len(fnNames))
		for _, fn := range fnNames {
			metrics = append(metrics, []interface{}{
				"AWS/Lambda", metricName, "FunctionName", fn,
				map[string]interface{}{"label": fn, "stat": stat},
			})
		}
		return metrics
	}

	widgets := []map[string]interface{}{
		// Row 1: Traffic & Errors
		{
			"type": "metric", "x": 0, "y": 0, "width": 8, "height": 6,
			"properties": map[string]interface{}{
				"title": "API Request Count",
				"metrics": [][]interface{}{
					{"AWS/ApiGateway", "Count", "ApiName", apiName, "Stage", stageName,
						map[string]interface{}{"stat": "Sum", "period": 60}},
				},
				"view": "timeSeries", "stacked": false, "region": region, "period": 60,
			},
		},
		{
			"type": "metric", "x": 8, "y": 0, "width": 8, "height": 6,
			"properties": map[string]interface{}{
				"title": "5XX Errors",
				"metrics": [][]interface{}{
					{"AWS/ApiGateway", "5XXError", "ApiName", apiName, "Stage", stageName,
						map[string]interface{}{"stat": "Sum", "period": 60, "color": "#d62728"}},
				},
				"view": "timeSeries", "region": region, "period": 60,
			},
		},
		{
			"type": "metric", "x": 16, "y": 0, "width": 8, "height": 6,
			"properties": map[string]interface{}{
				"title": "4XX Errors",
				"metrics": [][]interface{}{
					{"AWS/ApiGateway", "4XXError", "ApiName", apiName, "Stage", stageName,
						map[string]interface{}{"stat": "Sum", "period": 60, "color": "#ff7f0e"}},
				},
				"view": "timeSeries", "region": region, "period": 60,
			},
		},
		// Row 2: Latency
		{
			"type": "metric", "x": 0, "y": 6, "width": 12, "height": 6,
			"properties": map[string]interface{}{
				"title": "API Latency (ms)",
				"metrics": []interface{}{
					[]interface{}{"AWS/ApiGateway", "Latency", "ApiName", apiName, "Stage", stageName,
						map[string]interface{}{"stat": "Average", "label": "Avg", "color": "#2ca02c"}},
					[]interface{}{"...", map[string]interface{}{"stat": "p50", "label": "p50", "color": "#1f77b4"}},
					[]interface{}{"...", map[string]interface{}{"stat": "p90", "label": "p90", "color": "#ff7f0e"}},
					[]interface{}{"...", map[string]interface{}{"stat": "p99", "label": "p99", "color": "#d62728"}},
				},
				"view": "timeSeries", "region": region, "period": 60,
				"yAxis": map[string]interface{}{
					"left": map[string]interface{}{"min": 0, "label": "ms"},
				},
			},
		},
		{
			"type": "metric", "x": 12, "y": 6, "width": 12, "height": 6,
			"properties": map[string]interface{}{
				"title": "Integration Latency (ms)",
				"metrics": []interface{}{
					[]interface{}{"AWS/ApiGateway", "IntegrationLatency", "ApiName", apiName, "Stage", stageName,
						map[string]interface{}{"stat": "Average", "label": "Avg"}},
					[]interface{}{"...", map[string]interface{}{"stat": "p90", "label": "p90", "color": "#ff7f0e"}},
				},
				"view": "timeSeries", "region": region, "period": 60,
				"yAxis": map[string]interface{}{
					"left": map[string]interface{}{"min": 0, "label": "ms"},
				},
			},
		},
		// Row 3: Lambda Performance
		{
			"type": "metric", "x": 0, "y": 12, "width": 12, "height": 6,
			"properties": map[string]interface{}{
				"title": "Lambda Duration (ms)", "metrics": lambdaMetrics("Duration"),
				"view": "timeSeries", "region": region, "period": 60,
			},
		},
		{
			"type": "metric", "x": 12, "y": 12, "width": 12, "height": 6,
			"properties": map[string]interface{}{
				"title": "Lambda Errors", "metrics": lambdaMetrics("Errors"),
				"view": "timeSeries", "region": region, "period": 60,
			},
		},
		// Row 4: Lambda Invocations & Throttles
		{
			"type": "metric", "x": 0, "y": 18, "width": 12, "height": 6,
			"properties": map[string]interface{}{
				"title": "Lambda Invocations", "metrics": lambdaMetrics("Invocations"),
				"view": "timeSeries", "stacked": true, "region": region, "period": 60,
			},
		},
		{
			"type": "metric", "x": 12, "y": 18, "width": 12, "height": 6,
			"properties": map[string]interface{}{
				"title": "Lambda Throttles", "metrics": lambdaMetrics("Throttles"),
				"view": "timeSeries", "region": region, "period": 60,
			},
		},
	}

	body := map[string]interface{}{"widgets": widgets}
	bytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshalling dashboard JSON: %w", err)
	}
	return string(bytes), nil
}
