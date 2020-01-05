package service

import (
	"fmt"
  "os"

  "github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/authorizer/service/policy"
	"github.com/bugfixes/agent"
)

// Handler process request
func Handler(event events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
  c := agent.ConnectDetails{
    Host:     os.Getenv("DB_HOSTNAME"),
    Port:     os.Getenv("DB_PORT"),
    Username: os.Getenv("DB_USERNAME"),
    Password: os.Getenv("DB_PASSWORD"),
    Database: os.Getenv("DB_DATABASE"),
  }

	agentId, err := c.FindAgentFromHeaders(event.Headers)
	if err != nil {
    fmt.Printf("couldnt find agentId from headers: %+v, err: %+v\n", event.Headers, err)
    return policy.GenerateDeny(events.APIGatewayCustomAuthorizerRequest{
      Type:               event.Type,
      AuthorizationToken: agentId,
      MethodArn:          event.MethodArn,
    }), nil
  }

	newEvent := events.APIGatewayCustomAuthorizerRequest{
		Type:               event.Type,
		AuthorizationToken: agentId,
		MethodArn:          event.MethodArn,
	}

	return policy.GenerateAllow(newEvent), nil
}
