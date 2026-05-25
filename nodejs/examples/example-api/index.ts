
import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";

const config = new pulumi.Config();
const environment = config.get("environment") || "dev";
const projectName = config.get("projectName") || "example";
const namePrefix = `${projectName}-${environment}`;

// IAM Role for Lambda
const lambdaRole = new aws.iam.Role(`${namePrefix}-lambda-role`, {
    assumeRolePolicy: JSON.stringify({
        Version: "2012-10-17",
        Statement: [
            {
                Action: "sts:AssumeRole",
                Principal: { Service: "lambda.amazonaws.com" },
                Effect: "Allow",
            },
        ],
    }),
});

new aws.iam.RolePolicyAttachment(`${namePrefix}-lambda-basic`, {
    role: lambdaRole.name,
    policyArn: aws.iam.ManagedPolicies.AWSLambdaBasicExecutionRole,
});

// Lambda Functions
const helloFn = new aws.lambda.Function(`${namePrefix}-hello`, {
    runtime: aws.lambda.Runtime.NodeJS20dX,
    handler: "hello.handler",
    role: lambdaRole.arn,
    timeout: 10,
    memorySize: 128,
    code: new pulumi.asset.AssetArchive({
        "hello.js": new pulumi.asset.FileAsset("handlers/hello.js"),
    }),
    tags: { Environment: environment, Project: projectName },
});

const workFn = new aws.lambda.Function(`${namePrefix}-work`, {
    runtime: aws.lambda.Runtime.NodeJS20dX,
    handler: "work.handler",
    role: lambdaRole.arn,
    timeout: 30,
    memorySize: 128,
    code: new pulumi.asset.AssetArchive({
        "work.js": new pulumi.asset.FileAsset("handlers/work.js"),
    }),
    tags: { Environment: environment, Project: projectName },
});

// REST API
const api = new aws.apigateway.RestApi(`${namePrefix}-api`, {
    description: `Example API for monitoring template testing (${environment})`,
    tags: { Environment: environment, Project: projectName },
});

// /hello endpoint
const helloResource = new aws.apigateway.Resource(`${namePrefix}-hello-resource`, {
    restApi: api.id,
    parentId: api.rootResourceId,
    pathPart: "hello",
});

const helloMethod = new aws.apigateway.Method(`${namePrefix}-hello-method`, {
    restApi: api.id,
    resourceId: helloResource.id,
    httpMethod: "GET",
    authorization: "NONE",
});

const helloIntegration = new aws.apigateway.Integration(`${namePrefix}-hello-integration`, {
    restApi: api.id,
    resourceId: helloResource.id,
    httpMethod: helloMethod.httpMethod,
    integrationHttpMethod: "POST",
    type: "AWS_PROXY",
    uri: helloFn.invokeArn,
});

// /work endpoint
const workResource = new aws.apigateway.Resource(`${namePrefix}-work-resource`, {
    restApi: api.id,
    parentId: api.rootResourceId,
    pathPart: "work",
});

const workMethod = new aws.apigateway.Method(`${namePrefix}-work-method`, {
    restApi: api.id,
    resourceId: workResource.id,
    httpMethod: "GET",
    authorization: "NONE",
});

const workIntegration = new aws.apigateway.Integration(`${namePrefix}-work-integration`, {
    restApi: api.id,
    resourceId: workResource.id,
    httpMethod: workMethod.httpMethod,
    integrationHttpMethod: "POST",
    type: "AWS_PROXY",
    uri: workFn.invokeArn,
});

// Deployment & Stage
const deployment = new aws.apigateway.Deployment(
    `${namePrefix}-deployment`,
    {
        restApi: api.id,
        triggers: {
            redeployment: pulumi
                .all([helloIntegration.id, workIntegration.id])
                .apply(([h, w]) => `${h}-${w}`),
        },
    },
    { dependsOn: [helloIntegration, workIntegration] }
);

const stage = new aws.apigateway.Stage(`${namePrefix}-stage`, {
    restApi: api.id,
    deployment: deployment.id,
    stageName: environment,
    tags: { Environment: environment, Project: projectName },
});

// Lambda Permissions for API Gateway
for (const [name, fn] of [["hello", helloFn], ["work", workFn]] as const) {
    new aws.lambda.Permission(`${namePrefix}-${name}-permission`, {
        action: "lambda:InvokeFunction",
        function: fn.name,
        principal: "apigateway.amazonaws.com",
        sourceArn: pulumi.interpolate`${api.executionArn}/*/*`,
    });
}

// Stack Outputs
export const apiUrl = stage.invokeUrl;
export const helloEndpoint = pulumi.interpolate`${stage.invokeUrl}/hello`;
export const workEndpoint = pulumi.interpolate`${stage.invokeUrl}/work`;
export const apiGatewayName = api.name;
export const lambdaFunctionNames = pulumi.all([helloFn.name, workFn.name]).apply((names) => names.join(","));
export const stageName = stage.stageName;
