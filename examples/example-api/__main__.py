"""
Example API — Minimal API Gateway for testing the monitoring template.

Deploys:
    - 1 REST API (API Gateway)
    - 2 Lambda functions (hello, work)
    - IAM role with basic execution permissions

After deploying, the stack outputs the API name and Lambda function names —
plug those into the monitoring template config to test the dashboards and alarms.
"""

import json
import pulumi
import pulumi_aws as aws

# ── Configuration ────────────────────────────────────────────────────────
config = pulumi.Config()
environment = config.get("environment") or "dev"
project_name = config.get("project_name") or "example"
name_prefix = f"{project_name}-{environment}"

# ── IAM Role for Lambda ──────────────────────────────────────────────────
lambda_role = aws.iam.Role(
    f"{name_prefix}-lambda-role",
    assume_role_policy=json.dumps(
        {
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Action": "sts:AssumeRole",
                    "Principal": {"Service": "lambda.amazonaws.com"},
                    "Effect": "Allow",
                }
            ],
        }
    ),
)

aws.iam.RolePolicyAttachment(
    f"{name_prefix}-lambda-basic",
    role=lambda_role.name,
    policy_arn=aws.iam.ManagedPolicy.AWS_LAMBDA_BASIC_EXECUTION_ROLE,
)

# ── Lambda Functions ─────────────────────────────────────────────────────
hello_fn = aws.lambda_.Function(
    f"{name_prefix}-hello",
    runtime=aws.lambda_.Runtime.NODE_JS20D_X,
    handler="hello.handler",
    role=lambda_role.arn,
    timeout=10,
    memory_size=128,
    code=pulumi.AssetArchive({"hello.js": pulumi.FileAsset("handlers/hello.js")}),
    tags={"Environment": environment, "Project": project_name},
)

work_fn = aws.lambda_.Function(
    f"{name_prefix}-work",
    runtime=aws.lambda_.Runtime.NODE_JS20D_X,
    handler="work.handler",
    role=lambda_role.arn,
    timeout=30,
    memory_size=128,
    code=pulumi.AssetArchive({"work.js": pulumi.FileAsset("handlers/work.js")}),
    tags={"Environment": environment, "Project": project_name},
)

# ── REST API ─────────────────────────────────────────────────────────────
api = aws.apigateway.RestApi(
    f"{name_prefix}-api",
    description=f"Example API for monitoring template testing ({environment})",
    tags={"Environment": environment, "Project": project_name},
)

# ── /hello endpoint ──────────────────────────────────────────────────────
hello_resource = aws.apigateway.Resource(
    f"{name_prefix}-hello-resource",
    rest_api=api.id,
    parent_id=api.root_resource_id,
    path_part="hello",
)

hello_method = aws.apigateway.Method(
    f"{name_prefix}-hello-method",
    rest_api=api.id,
    resource_id=hello_resource.id,
    http_method="GET",
    authorization="NONE",
)

hello_integration = aws.apigateway.Integration(
    f"{name_prefix}-hello-integration",
    rest_api=api.id,
    resource_id=hello_resource.id,
    http_method=hello_method.http_method,
    integration_http_method="POST",
    type="AWS_PROXY",
    uri=hello_fn.invoke_arn,
)

# ── /work endpoint ───────────────────────────────────────────────────────
work_resource = aws.apigateway.Resource(
    f"{name_prefix}-work-resource",
    rest_api=api.id,
    parent_id=api.root_resource_id,
    path_part="work",
)

work_method = aws.apigateway.Method(
    f"{name_prefix}-work-method",
    rest_api=api.id,
    resource_id=work_resource.id,
    http_method="GET",
    authorization="NONE",
)

work_integration = aws.apigateway.Integration(
    f"{name_prefix}-work-integration",
    rest_api=api.id,
    resource_id=work_resource.id,
    http_method=work_method.http_method,
    integration_http_method="POST",
    type="AWS_PROXY",
    uri=work_fn.invoke_arn,
)

# ── Deployment & Stage ───────────────────────────────────────────────────
deployment = aws.apigateway.Deployment(
    f"{name_prefix}-deployment",
    rest_api=api.id,
    triggers={
        "redeployment": pulumi.Output.all(
            hello_integration.id, work_integration.id
        ).apply(lambda ids: "-".join(ids)),
    },
    opts=pulumi.ResourceOptions(depends_on=[hello_integration, work_integration]),
)

stage = aws.apigateway.Stage(
    f"{name_prefix}-stage",
    rest_api=api.id,
    deployment=deployment.id,
    stage_name=environment,
    tags={"Environment": environment, "Project": project_name},
)

# ── Lambda Permissions for API Gateway ───────────────────────────────────
for name, fn in [("hello", hello_fn), ("work", work_fn)]:
    aws.lambda_.Permission(
        f"{name_prefix}-{name}-permission",
        action="lambda:InvokeFunction",
        function=fn.name,
        principal="apigateway.amazonaws.com",
        source_arn=api.execution_arn.apply(lambda arn: f"{arn}/*/*"),
    )

# ── Stack Outputs ────────────────────────────────────────────────────────
pulumi.export("api_url", stage.invoke_url)
pulumi.export("hello_endpoint", stage.invoke_url.apply(lambda u: f"{u}/hello"))
pulumi.export("work_endpoint", stage.invoke_url.apply(lambda u: f"{u}/work"))
pulumi.export("api_gateway_name", api.name)
pulumi.export("lambda_function_names", pulumi.Output.all(hello_fn.name, work_fn.name).apply(lambda names: ",".join(names)))
pulumi.export("stage_name", stage.stage_name)
