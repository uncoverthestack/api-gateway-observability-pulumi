package main

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigateway"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	awslambda "github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		environment := cfg.Get("environment")
		if environment == "" {
			environment = "dev"
		}
		projectName := cfg.Get("projectName")
		if projectName == "" {
			projectName = "example"
		}
		namePrefix := fmt.Sprintf("%s-%s", projectName, environment)

		tags := pulumi.StringMap{
			"Environment": pulumi.String(environment),
			"Project":     pulumi.String(projectName),
		}

		// IAM Role for Lambda
		assumeRolePolicy, _ := json.Marshal(map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []map[string]interface{}{
				{
					"Action":    "sts:AssumeRole",
					"Principal": map[string]string{"Service": "lambda.amazonaws.com"},
					"Effect":    "Allow",
				},
			},
		})

		lambdaRole, err := iam.NewRole(ctx, namePrefix+"-lambda-role", &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(string(assumeRolePolicy)),
		})
		if err != nil {
			return err
		}

		_, err = iam.NewRolePolicyAttachment(ctx, namePrefix+"-lambda-basic", &iam.RolePolicyAttachmentArgs{
			Role:      lambdaRole.Name,
			PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
		})
		if err != nil {
			return err
		}

		// Lambda Functions
		helloFn, err := awslambda.NewFunction(ctx, namePrefix+"-hello", &awslambda.FunctionArgs{
			Runtime:    pulumi.String("nodejs20.x"),
			Handler:    pulumi.String("hello.handler"),
			Role:       lambdaRole.Arn,
			Timeout:    pulumi.Int(10),
			MemorySize: pulumi.Int(128),
			Code: pulumi.NewAssetArchive(map[string]interface{}{
				"hello.js": pulumi.NewFileAsset("handlers/hello.js"),
			}),
			Tags: tags,
		})
		if err != nil {
			return err
		}

		workFn, err := awslambda.NewFunction(ctx, namePrefix+"-work", &awslambda.FunctionArgs{
			Runtime:    pulumi.String("nodejs20.x"),
			Handler:    pulumi.String("work.handler"),
			Role:       lambdaRole.Arn,
			Timeout:    pulumi.Int(30),
			MemorySize: pulumi.Int(128),
			Code: pulumi.NewAssetArchive(map[string]interface{}{
				"work.js": pulumi.NewFileAsset("handlers/work.js"),
			}),
			Tags: tags,
		})
		if err != nil {
			return err
		}

		// REST API
		api, err := apigateway.NewRestApi(ctx, namePrefix+"-api", &apigateway.RestApiArgs{
			Description: pulumi.String(fmt.Sprintf("Example API for monitoring template testing (%s)", environment)),
			Tags:        tags,
		})
		if err != nil {
			return err
		}

		// /hello endpoint
		helloResource, err := apigateway.NewResource(ctx, namePrefix+"-hello-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: api.RootResourceId,
			PathPart: pulumi.String("hello"),
		})
		if err != nil {
			return err
		}

		helloMethod, err := apigateway.NewMethod(ctx, namePrefix+"-hello-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    helloResource.ID(),
			HttpMethod:    pulumi.String("GET"),
			Authorization: pulumi.String("NONE"),
		})
		if err != nil {
			return err
		}

		helloIntegration, err := apigateway.NewIntegration(ctx, namePrefix+"-hello-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            helloResource.ID(),
			HttpMethod:            helloMethod.HttpMethod,
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   helloFn.InvokeArn,
		})
		if err != nil {
			return err
		}

		// /work endpoint
		workResource, err := apigateway.NewResource(ctx, namePrefix+"-work-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: api.RootResourceId,
			PathPart: pulumi.String("work"),
		})
		if err != nil {
			return err
		}

		workMethod, err := apigateway.NewMethod(ctx, namePrefix+"-work-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    workResource.ID(),
			HttpMethod:    pulumi.String("GET"),
			Authorization: pulumi.String("NONE"),
		})
		if err != nil {
			return err
		}

		workIntegration, err := apigateway.NewIntegration(ctx, namePrefix+"-work-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            workResource.ID(),
			HttpMethod:            workMethod.HttpMethod,
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   workFn.InvokeArn,
		})
		if err != nil {
			return err
		}

		// Deployment & Stage
		deployment, err := apigateway.NewDeployment(ctx, namePrefix+"-deployment", &apigateway.DeploymentArgs{
			RestApi: api.ID(),
			Triggers: pulumi.StringMap{
				"redeployment": pulumi.All(helloIntegration.ID(), workIntegration.ID()).ApplyT(
					func(ids []interface{}) string {
						return fmt.Sprintf("%v-%v", ids[0], ids[1])
					},
				).(pulumi.StringOutput),
			},
		}, pulumi.DependsOn([]pulumi.Resource{helloIntegration, workIntegration}))
		if err != nil {
			return err
		}

		stage, err := apigateway.NewStage(ctx, namePrefix+"-stage", &apigateway.StageArgs{
			RestApi:    api.ID(),
			Deployment: deployment.ID(),
			StageName:  pulumi.String(environment),
			Tags:       tags,
		})
		if err != nil {
			return err
		}

		// Lambda Permissions for API Gateway
		fns := map[string]*awslambda.Function{"hello": helloFn, "work": workFn}
		for name, fn := range fns {
			_, err = awslambda.NewPermission(ctx, fmt.Sprintf("%s-%s-permission", namePrefix, name), &awslambda.PermissionArgs{
				Action:    pulumi.String("lambda:InvokeFunction"),
				Function:  fn.Name,
				Principal: pulumi.String("apigateway.amazonaws.com"),
				SourceArn: api.ExecutionArn.ApplyT(func(arn string) string {
					return arn + "/*/*"
				}).(pulumi.StringOutput),
			})
			if err != nil {
				return err
			}
		}

		// Stack Outputs
		ctx.Export("apiUrl", stage.InvokeUrl)
		ctx.Export("helloEndpoint", stage.InvokeUrl.ApplyT(func(u string) string { return u + "/hello" }))
		ctx.Export("workEndpoint", stage.InvokeUrl.ApplyT(func(u string) string { return u + "/work" }))
		ctx.Export("apiGatewayName", api.Name)
		ctx.Export("lambdaFunctionNames",
			pulumi.All(helloFn.Name, workFn.Name).ApplyT(func(args []interface{}) string {
				return fmt.Sprintf("%v,%v", args[0], args[1])
			}),
		)
		ctx.Export("stageName", stage.StageName)

		return nil
	})
}
