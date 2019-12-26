package validator_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/bugfixes/authorizer/service/validator"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func injectAgent(data validator.AgentData) error {
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

func TestAgentId(t *testing.T) {
	if os.Getenv("GITHUB_ACTOR") == "" {
		err := godotenv.Load()
		if err != nil {
			t.Errorf("godotenv err: %w", err)
		}
	}

	tests := []struct {
		name    string
		request validator.AgentData
		expect  bool
		err     error
	}{
		{
			name: "agentid valid",
			request: validator.AgentData{
				ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c74",
				Key:       "94365b00-c6df-483f-804e-363312750500",
				Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
				CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
				Name:      "bugfixes test frontend",
			},
			expect: true,
		},
		{
			name: "agentid invalid",
			request: validator.AgentData{
				ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c75",
				Key:       "94365b00-c6df-483f-804e-363312750500",
				Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
				CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
				Name:      "bugfixes test frontend",
			},
		},
	}

	injErr := injectAgent(tests[0].request)
	if injErr != nil {
		t.Errorf("injection err: %w", injErr)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			resp, err := validator.AgentId(test.request.ID)
			passed := assert.IsType(t, test.err, err)
			if !passed {
				t.Errorf("validator err: %w", err)
			}
			passed = assert.IsType(t, test.expect, resp)
			if !passed {
				t.Errorf("validator type test failed: %+v", test.expect)
			}
			passed = assert.Equal(t, test.expect, resp)
			if !passed {
				t.Errorf("validator equal test failed: %+v, resp: %+v", test.expect, resp)
			}
		})
	}

	delErr := deleteAgent(tests[0].request.ID)
	if delErr != nil {
		t.Errorf("delete err: %w", delErr)
	}
}

func TestLookupAgentId(t *testing.T) {
	if os.Getenv("GITHUB_ACTOR") == "" {
		err := godotenv.Load()
		if err != nil {
			t.Errorf("godotenv err: %w", err)
		}
	}

	tests := []struct {
		name    string
		request validator.AgentData
		expect  string
		err     error
	}{
		{
			name: "agentid found",
			request: validator.AgentData{
				ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c72",
				Key:       "94365b00-c6df-483f-804e-363312750500",
				Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
				CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
				Name:      "bugfixes test frontend",
			},
			expect: "ad4b99e1-dec8-4682-862a-6b017e7c7c72",
		},
	}

	injErr := injectAgent(tests[0].request)
	if injErr != nil {
		t.Errorf("injection err: %w", injErr)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := validator.LookupAgentId(test.request.Key, test.request.Secret)
			passed := assert.IsType(t, test.err, err)
			if !passed {
				t.Errorf("validator err: %w", err)
			}
			passed = assert.Equal(t, test.expect, resp)
			if !passed {
				t.Errorf("validator equal: %v, resp: %v", test.expect, resp)
			}
		})
	}

	delErr := deleteAgent(tests[0].request.ID)
	if delErr != nil {
		t.Errorf("delete err: %w", delErr)
	}
}

func BenchmarkLookupAgentId(b *testing.B) {
	b.ReportAllocs()

  if os.Getenv("GITHUB_ACTOR") == "" {
    err := godotenv.Load()
    if err != nil {
      b.Errorf("godotenv err: %w", err)
    }
  }

  tests := []struct {
    name    string
    request validator.AgentData
    expect  string
    err     error
  }{
    {
      name: "agentid found",
      request: validator.AgentData{
        ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c72",
        Key:       "94365b00-c6df-483f-804e-363312750500",
        Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
        CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
        Name:      "bugfixes test frontend",
      },
      expect: "ad4b99e1-dec8-4682-862a-6b017e7c7c72",
    },
  }

  injErr := injectAgent(tests[0].request)
  if injErr != nil {
    b.Errorf("injection err: %w", injErr)
  }

  b.ResetTimer()

  for _, test := range tests {
    b.Run(test.name, func(t *testing.B) {
      t.StartTimer()
      resp, err := validator.LookupAgentId(test.request.Key, test.request.Secret)
      passed := assert.IsType(t, test.err, err)
      if !passed {
        t.Errorf("validator err: %w", err)
      }
      passed = assert.Equal(t, test.expect, resp)
      if !passed {
        t.Errorf("validator equal: %v, resp: %v", test.expect, resp)
      }
      t.StopTimer()
    })
  }

  delErr := deleteAgent(tests[0].request.ID)
  if delErr != nil {
    b.Errorf("delete err: %w", delErr)
  }
}

func BenchmarkAgentId(b *testing.B) {
  b.ReportAllocs()

  if os.Getenv("GITHUB_ACTOR") == "" {
    err := godotenv.Load()
    if err != nil {
      b.Errorf("godotenv err: %w", err)
    }
  }

  tests := []struct {
    name    string
    request validator.AgentData
    expect  bool
    err     error
  }{
    {
      name: "agentid valid",
      request: validator.AgentData{
        ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c74",
        Key:       "94365b00-c6df-483f-804e-363312750500",
        Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
        CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
        Name:      "bugfixes test frontend",
      },
      expect: true,
    },
    {
      name: "agentid invalid",
      request: validator.AgentData{
        ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c75",
        Key:       "94365b00-c6df-483f-804e-363312750500",
        Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
        CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
        Name:      "bugfixes test frontend",
      },
    },
  }

  injErr := injectAgent(tests[0].request)
  if injErr != nil {
    b.Errorf("injection err: %w", injErr)
  }

  b.ResetTimer()

  for _, test := range tests {
    b.Run(test.name, func(t *testing.B) {
      t.StartTimer()

      resp, err := validator.AgentId(test.request.ID)
      passed := assert.IsType(t, test.err, err)
      if !passed {
        t.Errorf("validator err: %w", err)
      }
      passed = assert.IsType(t, test.expect, resp)
      if !passed {
        t.Errorf("validator type test failed: %+v", test.expect)
      }
      passed = assert.Equal(t, test.expect, resp)
      if !passed {
        t.Errorf("validator equal test failed: %+v, resp: %+v", test.expect, resp)
      }

      t.StopTimer()
    })
  }

  delErr := deleteAgent(tests[0].request.ID)
  if delErr != nil {
    b.Errorf("delete err: %w", delErr)
  }
}
