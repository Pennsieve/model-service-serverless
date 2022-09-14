package handler

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"os"
)

// SSMGetParameterAPI defines the interface for the GetParameter function.
// We use this interface to test the function using a mocked service.
type SSMGetParameterAPI interface {
	GetParameter(ctx context.Context,
		params *ssm.GetParameterInput,
		optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

// FindParameter retrieves an AWS Systems Manager string parameter
// Inputs:
//     c is the context of the method call, which includes the AWS Region
//     api is the interface that defines the method call
//     input defines the input arguments to the service call.
// Output:
//     If success, a GetParameterOutput object containing the result of the service call and nil
//     Otherwise, nil and an error from the call to GetParameter
func FindParameter(c context.Context, api SSMGetParameterAPI, input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	return api.GetParameter(c, input)
}

// SSMVars contains variables from AWS SSM used by serverless model-service.
type SSMVars struct {
	dbUri         *string
	neo4jUserName *string
	neo4jPassword *string
}

// fetchSSMVariables returns a SSMVars struct with values from AWS SSM
func fetchSSMVariables() (*SSMVars, error) {
	env := os.Getenv("ENV")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := ssm.NewFromConfig(cfg)

	// GET NEO4J HOST
	input := &ssm.GetParameterInput{
		Name: aws.String(fmt.Sprintf("/%s/model-service-serverless/db-host", env)),
	}

	results, err := FindParameter(context.TODO(), client, input)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	dbUri := results.Parameter.Value

	// GET USERNAME
	input = &ssm.GetParameterInput{
		Name: aws.String(fmt.Sprintf("/%s/model-service-serverless/neo4j-bolt-user", env)),
	}

	results, err = FindParameter(context.TODO(), client, input)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	neo4jUserName := results.Parameter.Value

	// GET PASSWORD
	input = &ssm.GetParameterInput{
		WithDecryption: true,
		Name:           aws.String(fmt.Sprintf("/%s/model-service-serverless/neo4j-bolt-password", env)),
	}

	results, err = FindParameter(context.TODO(), client, input)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	neo4jPassword := results.Parameter.Value

	return &SSMVars{
		dbUri:         dbUri,
		neo4jUserName: neo4jUserName,
		neo4jPassword: neo4jPassword,
	}, nil
}
