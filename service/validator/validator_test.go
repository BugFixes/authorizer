package validator_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/bugfixes/authorizer/service/validator"
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
				N: aws.String(fmt.Sprintf("%d", expires.Unix())),
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

func TestKey(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		t.Errorf("godotenv err: %w", err)
	}

	type request struct {
		key         string
		expires     time.Time
		service     string
		serviceTest string
	}

	tests := []struct {
		name string
		request
		expect bool
	}{
		{
			name: "tester +10 min",
			request: request{
				key:         "tester-69e668a5-b11f-405b-ae8a-e0eb3e6f371a",
				expires:     time.Now().Add(10 * time.Minute),
				service:     "tester",
				serviceTest: "tester",
			},
			expect: true,
		},
		{
			name: "tester -10 min",
			request: request{
				key:         "tester-69e668a5-b11f-405b-ae8a-e0eb3e6f371a",
				expires:     time.Now().Add(-10 * time.Minute),
				service:     "tester",
				serviceTest: "tester",
			},
		},
		{
			name: "tester something else",
			request: request{
				key:         "tester-69e668a5-b11f-405b-ae8a-e0eb3e6f371a",
				expires:     time.Now().Add(10 * time.Minute),
				service:     "tester",
				serviceTest: "somethingelse",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = injectKey(test.key, test.expires, test.service)

			resp := validator.Key(test.key, test.serviceTest)
			passed := assert.IsType(t, test.expect, resp)
			if !passed {
				t.Errorf("validator type test failed: %+v", test.expect)
			}
			passed = assert.Equal(t, test.expect, resp)
			if !passed {
				t.Errorf("validator equal test failed: %+v, resp: %+v", test.expect, resp)
			}

			_ = deleteKey(test.key)
		})
	}
}

func BenchmarkKey(b *testing.B) {
	b.ReportAllocs()

	err := godotenv.Load()
	if err != nil {
		b.Errorf("godotenv err: %w", err)
	}

	type request struct {
		key         string
		expires     time.Time
		service     string
		serviceTest string
	}

	tests := []struct {
		request
		expect bool
	}{
		{
			request: request{
				key:         "tester-69e668a5-b11f-405b-ae8a-e0eb3e6f371a",
				expires:     time.Now().Add(10 * time.Minute),
				service:     "tester",
				serviceTest: "tester",
			},
			expect: true,
		},
		{
			request: request{
				key:         "tester-69e668a5-b11f-405b-ae8a-e0eb3e6f371a",
				expires:     time.Now().Add(-10 * time.Minute),
				service:     "tester",
				serviceTest: "tester",
			},
		},
		{
			request: request{
				key:         "tester-69e668a5-b11f-405b-ae8a-e0eb3e6f371a",
				expires:     time.Now().Add(10 * time.Minute),
				service:     "tester",
				serviceTest: "somethingelse",
			},
		},
	}

	b.ResetTimer()
	for _, test := range tests {
		b.StopTimer()

		_ = injectKey(test.key, test.expires, test.service)

		resp := validator.Key(test.key, test.serviceTest)
		assert.IsType(b, test.expect, resp)
		assert.Equal(b, test.expect, resp)

		_ = deleteKey(test.key)

		b.StartTimer()

	}
}
