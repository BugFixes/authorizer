package service_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/bugfixes/authorizer/service"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func injectKey(key string, expires time.Time, service string) error {
	s, err := session.NewSession(&aws.Config{
		Region:   aws.String(os.Getenv("DB_REGION")),
		Endpoint: aws.String(os.Getenv("DB_ENDPOINT")),
	})
	if err != nil {
		return err
	}
	svc := dynamodb.New(s)
	input := &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE")),
		Item: map[string]*dynamodb.AttributeValue{
			"authKey": {
				S: aws.String(key),
			},
			"expires": {
				N: aws.String(fmt.Sprintf("%v", expires.Unix())),
			},
			"service": {
				S: aws.String(service),
			},
		},
		ConditionExpression: aws.String("attribute_not_exists(#AUTHKEY)"),
		ExpressionAttributeNames: map[string]*string{
			"#AUTHKEY": aws.String("authKey"),
		},
	}
	_, err = svc.PutItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				return fmt.Errorf("ErrCodeConditionalCheckFailedException: %w", aerr)
			case "ValidationException":
				return fmt.Errorf("validation error: %w", aerr)
			default:
				fmt.Println(fmt.Sprintf("unknown code err reason: %v", input))
				return fmt.Errorf("unknown code err: %w", aerr)
			}
		}
	}

	return nil
}

func deleteKey(key string) error {
	s, err := session.NewSession(&aws.Config{
		Region:   aws.String(os.Getenv("DB_REGION")),
		Endpoint: aws.String(os.Getenv("DB_ENDPOINT")),
	})
	if err != nil {
		return err
	}
	svc := dynamodb.New(s)
	_, err = svc.DeleteItem(&dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"authKey": {
				S: aws.String(key),
			},
		},
		TableName: aws.String(os.Getenv("DB_TABLE")),
	})
	if err != nil {
		return err
	}

	return nil
}

func TestHandler(t *testing.T) {
	if len(os.Args) >= 1 {
		for _, env := range os.Args {
			if env == "localDev" {
				err := godotenv.Load()
				if err != nil {
					fmt.Println(fmt.Sprintf("godotenv err: %v", err))
				}
			}
		}
	}

	type inject struct {
		key     string
		expires time.Time
		service string
	}

	tests := []struct {
		name string
		inject
		request events.APIGatewayCustomAuthorizerRequestTypeRequest
		expect  events.APIGatewayCustomAuthorizerResponse
		err     error
	}{
		{
			name: "+10 min",
			inject: inject{
				key:     "tester-69e668a5-b11f-405b-ae8a-e0eb3e6f371a",
				expires: time.Now().Add(10 * time.Minute),
				service: "tester.test.com",
			},
			request: events.APIGatewayCustomAuthorizerRequestTypeRequest{
				Type: "TOKEN",
				Headers: map[string]string{
					"Host":                 "tester.test.com",
					os.Getenv("AUTH_HEAD"): "tester-69e668a5-b11f-405b-ae8a-e0eb3e6f371a",
				},
				MethodArn: "arn:aws:execute-api:eu-west-2:123456789:wmcwzleu0i/ESTestInvoke-stage/GET/",
			},
			expect: events.APIGatewayCustomAuthorizerResponse{
				PrincipalID: "system",
				PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
					Version: "2012-10-17",
					Statement: []events.IAMPolicyStatement{
						{
							Action:   []string{"execute-api:Invoke"},
							Effect:   "Allow",
							Resource: []string{"arn:aws:execute-api:eu-west-2:123456789:wmcwzleu0i/ESTestInvoke-stage/GET/"},
						},
					},
				},
				Context: map[string]interface{}{
					"booleanKey": true,
					"numberKey":  123,
					"stringKey":  "stringval",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// inject key
			injectKey(test.inject.key, test.inject.expires, test.inject.service)

			// do the test
			resp, err := service.Handler(context.Background(), test.request)
			passed := assert.IsType(t, test.err, err)
			if !passed {
				t.Errorf("%s type failed: %w", test.name, err)
			}
			passed = assert.Equal(t, test.expect, resp)
			if !passed {
				t.Errorf("%s equal failed: %v, resp: %v", test.name, test.expect, resp)
			}

			// delete the tester key
			deleteKey(test.inject.key)
		})
	}
}

func BenchmarkHandler(b *testing.B) {
	b.ReportAllocs()
	if len(os.Args) >= 1 {
		for _, env := range os.Args {
			if env == "localDev" {
				err := godotenv.Load()
				if err != nil {
					fmt.Println(fmt.Sprintf("godotenv err: %v", err))
				}
			}
		}
	}

	type inject struct {
		key     string
		expires time.Time
		service string
	}

	tests := []struct {
		inject
		request events.APIGatewayCustomAuthorizerRequestTypeRequest
		expect  events.APIGatewayCustomAuthorizerResponse
		err     error
	}{
		{
			inject: inject{
				key:     "tester-69e668a5-b11f-405b-ae8a-e0eb3e6f371a",
				expires: time.Now().Add(10 * time.Minute),
				service: "tester.test.com",
			},
			request: events.APIGatewayCustomAuthorizerRequestTypeRequest{
				Type: "TOKEN",
				Headers: map[string]string{
					"Host":                 "tester.test.com",
					os.Getenv("AUTH_HEAD"): "tester-69e668a5-b11f-405b-ae8a-e0eb3e6f371a",
				},
				MethodArn: "arn:aws:execute-api:eu-west-2:123456789:wmcwzleu0i/ESTestInvoke-stage/GET/",
			},
			expect: events.APIGatewayCustomAuthorizerResponse{
				PrincipalID: "system",
				PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
					Version: "2012-10-17",
					Statement: []events.IAMPolicyStatement{
						{
							Action:   []string{"execute-api:Invoke"},
							Effect:   "Allow",
							Resource: []string{"arn:aws:execute-api:eu-west-2:123456789:wmcwzleu0i/ESTestInvoke-stage/GET/"},
						},
					},
				},
				Context: map[string]interface{}{
					"booleanKey": true,
					"numberKey":  123,
					"stringKey":  "stringval",
				},
			},
		},
	}

	b.ResetTimer()

	for _, test := range tests {
		b.StartTimer()
		// inject key
		injectKey(test.inject.key, test.inject.expires, test.inject.service)

		// do the test
		resp, err := service.Handler(context.Background(), test.request)
		passed := assert.IsType(b, test.err, err)
		if !passed {
			fmt.Println(fmt.Sprintf("test: %v, expect: %v, resp: %v, err: %v", test.request, test.expect, resp, err))
		}
		assert.Equal(b, test.expect, resp)

		// delete the tester key
		deleteKey(test.inject.key)

		b.StartTimer()
	}
}
