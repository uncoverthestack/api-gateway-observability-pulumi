import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";

export interface AlarmArgs {
    environment: string;
    projectName: string;
    apiName: pulumi.Input<string>;
    stageName: string;
    alertTopicArn: pulumi.Input<string>;
    lambdaFunctionNames: pulumi.Input<string>[];
    latencyThresholdMs?: number;
    errorRateThreshold?: number;
}

export interface AlarmResult {
    latencyAlarm: aws.cloudwatch.MetricAlarm;
    errorAlarm: aws.cloudwatch.MetricAlarm;
    downtimeAlarm: aws.cloudwatch.MetricAlarm;
    lambdaAlarms: aws.cloudwatch.MetricAlarm[];
    throttleAlarm: aws.cloudwatch.MetricAlarm;
}

export function createAlarms(args: AlarmArgs): AlarmResult {
    const {
        environment,
        projectName,
        apiName,
        stageName,
        alertTopicArn,
        lambdaFunctionNames,
        latencyThresholdMs = 3000,
        errorRateThreshold = 5,
    } = args;

    const namePrefix = `${projectName}-${environment}`;
    const baseTags = { Environment: environment, Project: projectName };

    // 1. High Latency Alarm
    const latencyAlarm = new aws.cloudwatch.MetricAlarm(`${namePrefix}-high-latency`, {
        alarmDescription: `API latency exceeded ${latencyThresholdMs}ms (${environment})`,
        namespace: "AWS/ApiGateway",
        metricName: "Latency",
        dimensions: { ApiName: apiName, Stage: stageName },
        statistic: "Average",
        period: 60,
        evaluationPeriods: 2,
        threshold: latencyThresholdMs,
        comparisonOperator: "GreaterThanThreshold",
        alarmActions: [alertTopicArn],
        okActions: [alertTopicArn],
        treatMissingData: "notBreaching",
        tags: { ...baseTags, AlarmType: "latency" },
    });

    // 2. High Error Rate Alarm
    const errorAlarm = new aws.cloudwatch.MetricAlarm(`${namePrefix}-high-error-rate`, {
        alarmDescription: `API 5xx error rate exceeded ${errorRateThreshold}% (${environment})`,
        namespace: "AWS/ApiGateway",
        metricName: "5XXError",
        dimensions: { ApiName: apiName, Stage: stageName },
        statistic: "Average",
        period: 60,
        evaluationPeriods: 2,
        threshold: errorRateThreshold / 100,
        comparisonOperator: "GreaterThanThreshold",
        alarmActions: [alertTopicArn],
        okActions: [alertTopicArn],
        treatMissingData: "notBreaching",
        tags: { ...baseTags, AlarmType: "errors" },
    });

    // 3. API Downtime Alarm
    const downtimeAlarm = new aws.cloudwatch.MetricAlarm(`${namePrefix}-api-downtime`, {
        alarmDescription: `No API requests received for 5 minutes (${environment})`,
        namespace: "AWS/ApiGateway",
        metricName: "Count",
        dimensions: { ApiName: apiName, Stage: stageName },
        statistic: "Sum",
        period: 300,
        evaluationPeriods: 1,
        threshold: 0,
        comparisonOperator: "LessThanOrEqualToThreshold",
        alarmActions: [alertTopicArn],
        treatMissingData: "breaching",
        tags: { ...baseTags, AlarmType: "downtime" },
    });

    // 4. Lambda Error Alarms — one per function
    const lambdaAlarms = lambdaFunctionNames.map((fnName, i) => {
        return new aws.cloudwatch.MetricAlarm(`${namePrefix}-lambda-errors-${i}`, {
            alarmDescription: pulumi
                .output(fnName)
                .apply((n) => `Lambda errors detected on ${n} (${environment})`),
            namespace: "AWS/Lambda",
            metricName: "Errors",
            dimensions: { FunctionName: fnName },
            statistic: "Sum",
            period: 60,
            evaluationPeriods: 1,
            threshold: 1,
            comparisonOperator: "GreaterThanOrEqualToThreshold",
            alarmActions: [alertTopicArn],
            okActions: [alertTopicArn],
            treatMissingData: "notBreaching",
            tags: { ...baseTags, AlarmType: "lambda" },
        });
    });

    // 5. Throttling Alarm
    const throttleAlarm = new aws.cloudwatch.MetricAlarm(`${namePrefix}-throttling`, {
        alarmDescription: `API requests being throttled (${environment})`,
        namespace: "AWS/ApiGateway",
        metricName: "Count",
        dimensions: { ApiName: apiName, Stage: stageName },
        statistic: "Sum",
        period: 60,
        evaluationPeriods: 1,
        threshold: 1,
        comparisonOperator: "GreaterThanOrEqualToThreshold",
        alarmActions: [alertTopicArn],
        treatMissingData: "notBreaching",
        tags: { ...baseTags, AlarmType: "throttle" },
    });

    return {
        latencyAlarm,
        errorAlarm,
        downtimeAlarm,
        lambdaAlarms,
        throttleAlarm,
    };
}
