package service

import (
	"fmt"
  "os"
  "strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/authorizer/service/policy"
	"github.com/bugfixes/authorizer/service/validator"
)

// Handler process request
func Handler(event events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	// keys
	agentKey := "x-agent-id"
	APIKey := "x-api-key"
	APISecret := "x-api-secret"

	// values
	var agentId, agentAPIKey, agentAPISecret string

	for key, value := range event.Headers {
		key := strings.ToLower(key)
		switch key {
		case agentKey:
			agentId = value
		case APIKey:
			agentAPIKey = value
		case APISecret:
			agentAPISecret = value
		}
	}

	if agentId == "" {
		err := func() error {
			return nil
		}()
		if err != nil {
			fmt.Printf("Seriouslly how the fuck is it not nil\n")
		}
		if len(agentAPIKey) == 0 || len(agentAPISecret) == 0 {
		  fmt.Printf("no agent, key, or secret")
		  return policy.GenerateDeny(events.APIGatewayCustomAuthorizerRequest{
        Type:               event.Type,
        AuthorizationToken: agentId,
        MethodArn:          event.MethodArn,
      }), nil
    }
		agentId, err = validator.LookupAgentId(agentAPIKey, agentAPISecret)
		if err != nil {
			fmt.Printf("Invalid AgentId: %+v, key: %s, secret: %s, err: %+v\n", err, agentAPISecret, agentAPIKey, err)
			return policy.GenerateDeny(events.APIGatewayCustomAuthorizerRequest{
				Type:               event.Type,
				AuthorizationToken: agentId,
				MethodArn:          event.MethodArn,
			}), nil
		}
	}

	newEvent := events.APIGatewayCustomAuthorizerRequest{
		Type:               event.Type,
		AuthorizationToken: agentId,
		MethodArn:          event.MethodArn,
	}

	// test agentid
	agentValid, err := validator.AgentId(agentId)
	if err != nil {
	  if os.Getenv("TEST_CODE") != "true" {
      fmt.Printf("Invalid Agent: %+v, err: %+v\n", newEvent, err)
    }
		return policy.GenerateDeny(newEvent), nil
	}

	if !agentValid {
	  if os.Getenv("TEST_CODE") != "true" {
      fmt.Printf("Deny Agent: %+v\n", newEvent)
    }
		return policy.GenerateDeny(newEvent), nil
	}

	return policy.GenerateAllow(newEvent), nil
}
