/**
 * API Observability & Monitoring Platform
 *
 * Provisions CloudWatch dashboards, automated alarms, and SNS alerting
 * for any existing AWS API Gateway.
 *
 * Point this at your API Gateway, set your thresholds, run `pulumi up`.
 *
 * Usage:
 *   pulumi stack init dev
 *   pulumi config set alertEmail your@email.com
 *   pulumi config set apiGatewayName your-api-name
 *   pulumi config set stageName prod
 *   pulumi config set lambdaFunctionNames function-1,function-2
 *   pulumi up
 */

import * as pulumi from "@pulumi/pulumi";

import { createNotifications } from "./src/observability/notifications";
import { createAlarms } from "./src/observability/alarms";
import { createDashboard } from "./src/observability/dashboard";

// Configuration
const config = new pulumi.Config();
const awsConfig = new pulumi.Config("aws");

const environment = config.get("environment") || "dev";
const projectName = config.get("projectName") || "api-observability";
const alertEmail = config.require("alertEmail");
const latencyThresholdMs = config.getNumber("latencyThresholdMs") || 3000;
const errorRateThreshold = config.getNumber("errorRateThreshold") || 5;
const region = awsConfig.get("region") || "eu-west-2";

// Your existing API Gateway details
const apiGatewayName = config.require("apiGatewayName");
const stageName = config.get("stageName") || "dev";
const lambdaNamesRaw = config.get("lambdaFunctionNames") || "";
const lambdaFunctionNames = lambdaNamesRaw
    .split(",")
    .map((n) => n.trim())
    .filter((n) => n.length > 0);

// 1. Setup Notifications
const notifications = createNotifications({
    environment,
    projectName,
    alertEmail,
});

// 2. Create Alarms
const alarms = createAlarms({
    environment,
    projectName,
    apiName: apiGatewayName,
    stageName,
    alertTopicArn: notifications.alertTopicArn,
    lambdaFunctionNames,
    latencyThresholdMs,
    errorRateThreshold,
});

// 3. Create Dashboard
const dashboard = createDashboard({
    environment,
    projectName,
    apiName: apiGatewayName,
    stageName,
    lambdaFunctionNames,
    region,
});

// Stack Outputs
export const dashboardUrl = `https://${region}.console.aws.amazon.com/cloudwatch/home?region=${region}#dashboards:name=${dashboard.dashboardName}`;
export const alertTopicArn = notifications.alertTopicArn;
export const monitoredApi = apiGatewayName;
export const monitoredStage = stageName;
export const monitoredFunctions = lambdaFunctionNames;
export const stackEnvironment = environment;
