package policy_test

import (
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/bugfixes/authorizer/service/policy"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAllow(t *testing.T) {
	if os.Getenv("GITHUB_ACTOR") == "" {
		err := godotenv.Load()
		if err != nil {
			t.Errorf("godotenv err: %w", err)
		}
	}

	tests := []struct {
		name    string
		request events.APIGatewayCustomAuthorizerRequest
		expect  events.APIGatewayCustomAuthorizerResponse
	}{
		{
			name: "allowed",
			request: events.APIGatewayCustomAuthorizerRequest{
				Type:               "TOKEN",
				AuthorizationToken: "tester-37259d99-5747-4feb-9261-2764c8cfc326",
				MethodArn:          "arn:aws:execute-api:eu-west-2:123456789:wmcwzleu0i/ESTestInvoke-stage/GET/",
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
			resp := policy.GenerateAllow(test.request)
			passed := assert.IsType(t, test.expect, resp)
			if !passed {
				t.Errorf("allow policy type failed: %+v, %+v", test.expect, resp)
			}
			passed = assert.Equal(t, test.expect, resp)
			if !passed {
				t.Errorf("allow policy equal failed: %+v, %+v", test.expect, resp)
			}
		})
	}
}

func TestGenerateDeny(t *testing.T) {
	if os.Getenv("GITHUB_ACTOR") == "" {
		err := godotenv.Load()
		if err != nil {
			t.Errorf("godotenv err: %w", err)
		}
	}

	tests := []struct {
		name    string
		request events.APIGatewayCustomAuthorizerRequest
		expect  events.APIGatewayCustomAuthorizerResponse
	}{
		{
			name: "denied",
			request: events.APIGatewayCustomAuthorizerRequest{
				Type:               "TOKEN",
				AuthorizationToken: "tester-37259d99-5747-4feb-9261-2764c8cfc326",
				MethodArn:          "arn:aws:execute-api:eu-west-2:123456789:wmcwzleu0i/ESTestInvoke-stage/GET/",
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
			resp := policy.GenerateDeny(test.request)
			passed := assert.IsType(t, test.expect, resp)
			if !passed {
				t.Errorf("denied policy type failed: %+v, %+v", test.expect, resp)
			}
			passed = assert.Equal(t, test.expect, resp)
			if !passed {
				t.Errorf("denied policy equal failed: %+v, %+v", test.expect, resp)
			}
		})
	}
}

func BenchmarkGenerateAllow(b *testing.B) {
	b.ReportAllocs()

	if os.Getenv("GITHUB_ACTOR") == "" {
		err := godotenv.Load()
		if err != nil {
			b.Errorf("godotenv err: %w", err)
		}
	}

	tests := []struct {
		request events.APIGatewayCustomAuthorizerRequest
		expect  events.APIGatewayCustomAuthorizerResponse
	}{
		{
			request: events.APIGatewayCustomAuthorizerRequest{
				Type:               "TOKEN",
				AuthorizationToken: "tester-37259d99-5747-4feb-9261-2764c8cfc326",
				MethodArn:          "arn:aws:execute-api:eu-west-2:123456789:wmcwzleu0i/ESTestInvoke-stage/GET/",
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

		resp := policy.GenerateAllow(test.request)
		assert.IsType(b, test.expect, resp)
		assert.Equal(b, test.expect, resp)

		b.StopTimer()
	}
}

func BenchmarkGenerateDeny(b *testing.B) {
	b.ReportAllocs()

	if os.Getenv("GITHUB_ACTOR") == "" {
		err := godotenv.Load()
		if err != nil {
			b.Errorf("godotenv err: %w", err)
		}
	}

	tests := []struct {
		request events.APIGatewayCustomAuthorizerRequest
		expect  events.APIGatewayCustomAuthorizerResponse
	}{
		{
			request: events.APIGatewayCustomAuthorizerRequest{
				Type:               "TOKEN",
				AuthorizationToken: "tester-37259d99-5747-4feb-9261-2764c8cfc326",
				MethodArn:          "arn:aws:execute-api:eu-west-2:123456789:wmcwzleu0i/ESTestInvoke-stage/GET/",
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
		b.StartTimer()

		resp := policy.GenerateDeny(test.request)
		assert.IsType(b, test.expect, resp)
		assert.Equal(b, test.expect, resp)

		b.StopTimer()
	}
}
