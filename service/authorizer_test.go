package service_test

import (
  "database/sql"
	"fmt"
  "os"
	"testing"

  "github.com/aws/aws-lambda-go/events"
  "github.com/bugfixes/authorizer/service"
  "github.com/joho/godotenv"
  "github.com/stretchr/testify/assert"
	_ "github.com/lib/pq"
)

type AgentData struct {
	ID        string
	Key       string
	Secret    string
	CompanyID string
	Name      string
}

var connectDetails = ""

func injectAgent(data AgentData) error {
  db, err := sql.Open("postgres", connectDetails)
  if err != nil {
    return fmt.Errorf("injectAgent db.open: %w", err)
  }
  defer func() {
    err := db.Close()
    if err != nil {
      fmt.Printf("injectAgent db.close: %v", err)
    }
  }()
  _, err = db.Exec(
    "INSERT INTO agent (id, key, secret, company_id, name) VALUES ($1, $2, $3, $4, $5)",
    data.ID,
    data.Key,
    data.Secret,
    data.CompanyID,
    data.Name)
  if err != nil {
    return fmt.Errorf("injectAgent db.exec: %w", err)
  }

  return nil
}

func deleteAgent(id string) error {
  db, err := sql.Open("postgres", connectDetails)
  if err != nil {
    return fmt.Errorf("deleteAgent db.open: %w", err)
  }
  defer func() {
    err := db.Close()
    if err != nil {
      fmt.Printf("deleteAgent db.close: %v", err)
    }
  }()
  _, err = db.Exec("DELETE FROM agent WHERE id = $1", id)
  if err != nil {
    return fmt.Errorf("deleteAgent db.exec: %w", err)
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
	connectDetails = fmt.Sprintf(
    "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
    os.Getenv("DB_HOSTNAME"),
    os.Getenv("DB_PORT"),
    os.Getenv("DB_USERNAME"),
    os.Getenv("DB_PASSWORD"),
    os.Getenv("DB_DATABASE"))

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

func BenchmarkHandler(b *testing.B) {
	b.ReportAllocs()

  if os.Getenv("GITHUB_ACTOR") == "" {
    err := godotenv.Load()
    if err != nil {
      b.Errorf("godotenv err: %w", err)
    }
  }

  connectDetails = fmt.Sprintf(
    "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
    os.Getenv("DB_HOSTNAME"),
    os.Getenv("DB_PORT"),
    os.Getenv("DB_USERNAME"),
    os.Getenv("DB_PASSWORD"),
    os.Getenv("DB_DATABASE"))

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

  b.ResetTimer()

  for _, test := range tests {
    b.Run(test.name, func(t *testing.B) {
      b.StopTimer()

      // inject key
      injErr := injectAgent(test.agent)
      if injErr != nil {
        t.Errorf("inject err: %w", injErr)
      }

      b.StartTimer()
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
      b.StopTimer()

      // delete the tester key
      delErr := deleteAgent(test.agent.ID)
      if delErr != nil {
        t.Errorf("delete err: %w", delErr)
      }
    })
  }
}
