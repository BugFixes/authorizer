package validator

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"os"
)

type AgentData struct {
	ID        string
	Key       string
	Secret    string
	CompanyID string
	Name      string
}

func LookupAgentId(key, secret string) (string, error) {
	s, err := session.NewSession(&aws.Config{
		Region:   aws.String(os.Getenv("DB_REGION")),
		Endpoint: aws.String(os.Getenv("DB_ENDPOINT")),
	})
	if err != nil {
		return "", fmt.Errorf("lookupAgent sesson: %w", err)
	}

	keyFilter := expression.Name("key").Equal(expression.Value(aws.String(key)))
	secretFilter := expression.Name("secret").Equal(expression.Value(aws.String(secret)))
	proj := expression.NamesList(expression.Name("id"), expression.Name("key"), expression.Name("secret"), expression.Name("companyId"))
	expr, err := expression.NewBuilder().WithFilter(keyFilter.And(secretFilter)).WithProjection(proj).Build()
	if err != nil {
		return "", fmt.Errorf("lookupAgent expression: %w", err)
	}

	result, err := dynamodb.New(s).Scan(&dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(os.Getenv("DB_TABLE")),
	})
	if err != nil {
		return "", fmt.Errorf("lookupAgent scanItem: %w", err)
	}

	if len(result.Items) >= 1 {
		ad := AgentData{}
		unMapErr := dynamodbattribute.UnmarshalMap(result.Items[0], &ad)
		if unMapErr != nil {
			return "", fmt.Errorf("lookupAgent unmarshall: %w", err)
		}

		if ad.ID == "" {
			return "", fmt.Errorf("unknown agentId")
		}

		return ad.ID, nil
	}

	return "", fmt.Errorf("no agents found")
}

func AgentId(agentId string) (bool, error) {
	s, err := session.NewSession(&aws.Config{
		Region:   aws.String(os.Getenv("DB_REGION")),
		Endpoint: aws.String(os.Getenv("DB_ENDPOINT")),
	})
	if err != nil {
		return false, fmt.Errorf("agentId session: %w", err)
	}

	result, err := dynamodb.New(s).GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(agentId),
			},
		},
		TableName: aws.String(os.Getenv("DB_TABLE")),
	})
	if err != nil {
		return false, fmt.Errorf("agentid: %w", err)
	}

	ad := AgentData{}
	unMapErr := dynamodbattribute.UnmarshalMap(result.Item, &ad)
	if unMapErr != nil {
		return false, fmt.Errorf("agentid unmarshall: %w", err)
	}

	if ad.ID == "" {
		return false, nil
	}

	return true, nil
}
