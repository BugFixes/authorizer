package service

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/authorizer/service/policy"
	"github.com/bugfixes/authorizer/service/validator"
)

// Handler process request
func Handler(ctx context.Context, event events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	token := ""

	service := ""
	authHeader := strings.ToLower(os.Getenv("AUTH_HEAD"))
	for Key, Value := range event.Headers {
		key := strings.ToLower(Key)
		if key == authHeader {
			token = Value
		}

		if Key == "Host" {
			service = Value
		}
	}

	// Token sent
	fmt.Println(fmt.Sprintf("AUTH Key: %s", token))
	fmt.Println(fmt.Sprintf("Event: %v", event))

	newEvent := events.APIGatewayCustomAuthorizerRequest{
		Type:               event.Type,
		AuthorizationToken: token,
		MethodArn:          event.MethodArn,
	}

	// Test token
	if strings.Contains(token, os.Getenv("AUTH_PREF")) {
		if validator.Key(token, service) {
			fmt.Println(fmt.Sprintf("allowed: %s", token))
			return policy.GenerateAllow(newEvent), nil
		}
		fmt.Println(fmt.Sprintf("denied: %s", token))
		return policy.GenerateDeny(newEvent), nil
	}

	fmt.Println(fmt.Sprintf("Pref: %s, key: %s", os.Getenv("AUTH_PREF"), token))
	return events.APIGatewayCustomAuthorizerResponse{}, fmt.Errorf("%s", "Unauthorized")
}
