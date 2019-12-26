package service_test

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/bugfixes/authorizer/service"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type AgentData struct {
	ID        string
	Key       string
	Secret    string
	CompanyID string
	Name      string
}

func injectAgent(data AgentData) error {
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
			"id": {
				S: aws.String(data.ID),
			},
			"key": {
				S: aws.String(data.Key),
			},
			"secret": {
				S: aws.String(data.Secret),
			},
			"companyId": {
				S: aws.String(data.CompanyID),
			},
			"name": {
				S: aws.String(data.Name),
			},
		},
		ConditionExpression: aws.String("attribute_not_exists(#ID)"),
		ExpressionAttributeNames: map[string]*string{
			"#ID": aws.String("id"),
		},
	}

	_, err = svc.PutItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				return fmt.Errorf("validator ErrCodeConditionalCheckFailedException: %w", aerr)
			case "ValidationException":
				return fmt.Errorf("validator validation error: %w", aerr)
			default:
				fmt.Printf("validator unknown code err reason: %+v\n", input)
				return fmt.Errorf("validator unknown code err: %w", aerr)
			}
		}
	}

	return nil
}

func deleteAgent(id string) error {
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
			"id": {
				S: aws.String(id),
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
	if os.Getenv("GITHUB_ACTOR") == "" {
		err := godotenv.Load()
		if err != nil {
			t.Errorf("godotenv err: %w", err)
		}
	}

	tests := []struct {
		name    string
		agent   AgentData
		request events.APIGatewayCustomAuthorizerRequestTypeRequest
		expect  events.APIGatewayCustomAuthorizerResponse
		err     error
	}{
		{
			name: "allowed agentid",
			agent: AgentData{
				ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c70",
				Key:       "94365b00-c6df-483f-804e-363312750500",
				Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
				CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
				Name:      "bugfixes test frontend -- allowed agentid",
			},
			request: events.APIGatewayCustomAuthorizerRequestTypeRequest{
				Type: "TOKEN",
				Headers: map[string]string{
					"x-agent-id": "ad4b99e1-dec8-4682-862a-6b017e7c7c70",
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
		{
			name: "denied agentid",
			agent: AgentData{
				ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c71",
				Key:       "94365b00-c6df-483f-804e-363312750500",
				Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
				CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
				Name:      "bugfixes test frontend -- denied agentid",
			},
			request: events.APIGatewayCustomAuthorizerRequestTypeRequest{
				Type: "TOKEN",
				Headers: map[string]string{
					"x-agent-id": "ad4b99e1-dec8-4682-862a-6b017e7c7c70",
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
							Effect:   "Deny",
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
		{
			name: "allowed key and secret",
			agent: AgentData{
				ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c72",
				Key:       "94365b00-c6df-483f-804e-363312750500",
				Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
				CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
				Name:      "bugfixes test frontend -- allowed key",
			},
			request: events.APIGatewayCustomAuthorizerRequestTypeRequest{
				Type: "TOKEN",
				Headers: map[string]string{
					"x-api-key":    "94365b00-c6df-483f-804e-363312750500",
					"x-api-secret": "f7356946-5814-4b5e-ad45-0348a89576ef",
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
		{
			name: "denied key and secret",
			agent: AgentData{
				ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c73",
				Key:       "94365b00-c6df-483f-804e-363312750500",
				Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
				CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
				Name:      "bugfixes test frontend -- denied key",
			},
			request: events.APIGatewayCustomAuthorizerRequestTypeRequest{
				Type: "TOKEN",
				Headers: map[string]string{
					"x-api-key":    "94365b00-c6df-483f-804e-363312750501",
					"x-api-secret": "f7356946-5814-4b5e-ad45-0348a89576ef",
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
							Effect:   "Deny",
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
			injErr := injectAgent(test.agent)
			if injErr != nil {
				t.Errorf("inject err: %w", injErr)
			}

			// do the test
			resp, err := service.Handler(test.request)
			passed := assert.IsType(t, test.err, err)
			if !passed {
				t.Errorf("%s type failed: %w", test.name, err)
			}
			passed = assert.Equal(t, test.expect, resp)
			if !passed {
				t.Errorf("%s equal failed: %+v, resp: %+v", test.name, test.expect, resp)
			}

			// delete the tester key
			delErr := deleteAgent(test.agent.ID)
			if delErr != nil {
				t.Errorf("delete err: %w", delErr)
			}
		})
	}
}

//func BenchmarkHandler(b *testing.B) {
//	b.ReportAllocs()
//
//	err := godotenv.Load()
//	if err != nil {
//		b.Errorf("godotenv err: %w", err)
//	}
//
//	type inject struct {
//		key     string
//		expires time.Time
//		service string
//	}
//
//	tests := []struct {
//		inject
//		request events.APIGatewayCustomAuthorizerRequestTypeRequest
//		expect  events.APIGatewayCustomAuthorizerResponse
//		err     error
//	}{
//		{
//			inject: inject{
//				key:     "tester-69e668a5-b11f-405b-ae8a-e0eb3e6f371a",
//				expires: time.Now().Add(10 * time.Minute),
//				service: "tester.test.com",
//			},
//			request: events.APIGatewayCustomAuthorizerRequestTypeRequest{
//				Type: "TOKEN",
//				Headers: map[string]string{
//					"Host":                 "tester.test.com",
//					os.Getenv("AUTH_HEAD"): "tester-69e668a5-b11f-405b-ae8a-e0eb3e6f371a",
//				},
//				MethodArn: "arn:aws:execute-api:eu-west-2:123456789:wmcwzleu0i/ESTestInvoke-stage/GET/",
//			},
//			expect: events.APIGatewayCustomAuthorizerResponse{
//				PrincipalID: "system",
//				PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
//					Version: "2012-10-17",
//					Statement: []events.IAMPolicyStatement{
//						{
//							Action:   []string{"execute-api:Invoke"},
//							Effect:   "Allow",
//							Resource: []string{"arn:aws:execute-api:eu-west-2:123456789:wmcwzleu0i/ESTestInvoke-stage/GET/"},
//						},
//					},
//				},
//				Context: map[string]interface{}{
//					"booleanKey": true,
//					"numberKey":  123,
//					"stringKey":  "stringval",
//				},
//			},
//		},
//	}
//
//	b.ResetTimer()
//
//	for _, test := range tests {
//		b.StartTimer()
//		// inject key
//		_ = injectKey(test.inject.key, test.inject.expires, test.inject.service)
//
//		resp, err := service.Handler(test.request)
//		// do the test
//		passed := assert.IsType(b, test.err, err)
//		if !passed {
//			b.Errorf("test: %+v, expect: %+v, resp: %+v, err: %w", test.request, test.expect, resp, err)
//		}
//		assert.Equal(b, test.expect, resp)
//
//		// delete the tester key
//		_ = deleteKey(test.inject.key)
//
//		b.StartTimer()
//	}
//}
