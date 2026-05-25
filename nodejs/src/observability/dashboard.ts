import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";

export interface DashboardArgs {
    environment: string;
    projectName: string;
    apiName: pulumi.Input<string>;
    stageName: string;
    lambdaFunctionNames: pulumi.Input<string>[];
    region: string;
}

export interface DashboardResult {
    dashboard: aws.cloudwatch.Dashboard;
    dashboardName: string;
}

export function createDashboard(args: DashboardArgs): DashboardResult {
    const { environment, projectName, apiName, stageName, lambdaFunctionNames, region } = args;
    const namePrefix = `${projectName}-${environment}`;
    const dashboardName = `${namePrefix}-api-observability`;

    // Resolve all outputs before building JSON
    const dashboardBody = pulumi
        .all([apiName, ...lambdaFunctionNames])
        .apply(([resolvedApiName, ...resolvedFnNames]) =>
            buildDashboardJson(resolvedApiName, resolvedFnNames, stageName, region)
        );

    const dashboard = new aws.cloudwatch.Dashboard(`${namePrefix}-dashboard`, {
        dashboardName,
        dashboardBody,
    });

    return { dashboard, dashboardName };
}

function buildDashboardJson(
    resolvedApiName: string,
    resolvedFnNames: string[],
    stageName: string,
    region: string
): string {
    const lambdaMetrics = (metricName: string): any[] => {
        const stat = metricName === "Duration" ? "Average" : "Sum";
        return resolvedFnNames.map((fn) => [
            "AWS/Lambda",
            metricName,
            "FunctionName",
            fn,
            { label: fn, stat },
        ]);
    };

    const widgets = [
        // Row 1: Traffic & Errors
        {
            type: "metric",
            x: 0,
            y: 0,
            width: 8,
            height: 6,
            properties: {
                title: "API Request Count",
                metrics: [
                    [
                        "AWS/ApiGateway",
                        "Count",
                        "ApiName",
                        resolvedApiName,
                        "Stage",
                        stageName,
                        { stat: "Sum", period: 60 },
                    ],
                ],
                view: "timeSeries",
                stacked: false,
                region,
                period: 60,
            },
        },
        {
            type: "metric",
            x: 8,
            y: 0,
            width: 8,
            height: 6,
            properties: {
                title: "5XX Errors",
                metrics: [
                    [
                        "AWS/ApiGateway",
                        "5XXError",
                        "ApiName",
                        resolvedApiName,
                        "Stage",
                        stageName,
                        { stat: "Sum", period: 60, color: "#d62728" },
                    ],
                ],
                view: "timeSeries",
                region,
                period: 60,
            },
        },
        {
            type: "metric",
            x: 16,
            y: 0,
            width: 8,
            height: 6,
            properties: {
                title: "4XX Errors",
                metrics: [
                    [
                        "AWS/ApiGateway",
                        "4XXError",
                        "ApiName",
                        resolvedApiName,
                        "Stage",
                        stageName,
                        { stat: "Sum", period: 60, color: "#ff7f0e" },
                    ],
                ],
                view: "timeSeries",
                region,
                period: 60,
            },
        },
        // Row 2: Latency
        {
            type: "metric",
            x: 0,
            y: 6,
            width: 12,
            height: 6,
            properties: {
                title: "API Latency (ms)",
                metrics: [
                    [
                        "AWS/ApiGateway",
                        "Latency",
                        "ApiName",
                        resolvedApiName,
                        "Stage",
                        stageName,
                        { stat: "Average", label: "Avg", color: "#2ca02c" },
                    ],
                    ["...", { stat: "p50", label: "p50", color: "#1f77b4" }],
                    ["...", { stat: "p90", label: "p90", color: "#ff7f0e" }],
                    ["...", { stat: "p99", label: "p99", color: "#d62728" }],
                ],
                view: "timeSeries",
                region,
                period: 60,
                yAxis: { left: { min: 0, label: "ms" } },
            },
        },
        {
            type: "metric",
            x: 12,
            y: 6,
            width: 12,
            height: 6,
            properties: {
                title: "Integration Latency (ms)",
                metrics: [
                    [
                        "AWS/ApiGateway",
                        "IntegrationLatency",
                        "ApiName",
                        resolvedApiName,
                        "Stage",
                        stageName,
                        { stat: "Average", label: "Avg" },
                    ],
                    ["...", { stat: "p90", label: "p90", color: "#ff7f0e" }],
                ],
                view: "timeSeries",
                region,
                period: 60,
                yAxis: { left: { min: 0, label: "ms" } },
            },
        },
        // Row 3: Lambda Performance
        {
            type: "metric",
            x: 0,
            y: 12,
            width: 12,
            height: 6,
            properties: {
                title: "Lambda Duration (ms)",
                metrics: lambdaMetrics("Duration"),
                view: "timeSeries",
                region,
                period: 60,
            },
        },
        {
            type: "metric",
            x: 12,
            y: 12,
            width: 12,
            height: 6,
            properties: {
                title: "Lambda Errors",
                metrics: lambdaMetrics("Errors"),
                view: "timeSeries",
                region,
                period: 60,
            },
        },
        // Row 4: Lambda Invocations & Throttles
        {
            type: "metric",
            x: 0,
            y: 18,
            width: 12,
            height: 6,
            properties: {
                title: "Lambda Invocations",
                metrics: lambdaMetrics("Invocations"),
                view: "timeSeries",
                stacked: true,
                region,
                period: 60,
            },
        },
        {
            type: "metric",
            x: 12,
            y: 18,
            width: 12,
            height: 6,
            properties: {
                title: "Lambda Throttles",
                metrics: lambdaMetrics("Throttles"),
                view: "timeSeries",
                region,
                period: 60,
            },
        },
    ];

    return JSON.stringify({ widgets });
}
