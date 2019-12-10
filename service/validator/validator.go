package validator

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// KeyData object
type KeyData struct {
	Key        string `json:"authKey"`
	ExpireTime int    `json:"expires"`
	Service    string `json:"service"`
}

func matchKey(key string) KeyData {
	s, err := session.NewSession(&aws.Config{
		Region:   aws.String(os.Getenv("DB_REGION")),
		Endpoint: aws.String(os.Getenv("DB_ENDPOINT")),
	})
	if err != nil {
		fmt.Printf("Key Session Error: %+v\n", err)
		return KeyData{}
	}
	svc := dynamodb.New(s)
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"authKey": {
				S: aws.String(key),
			},
		},
		TableName: aws.String(os.Getenv("DB_TABLE")),
	})
	if err != nil {
		fmt.Printf("Key Get Error: %+v\n", err)
		return KeyData{}
	}
	returnData := KeyData{}
	unErr := dynamodbattribute.UnmarshalMap(result.Item, &returnData)
	if unErr != nil {
		fmt.Printf("Key Unmarshall Error: %+v\n", unErr)
		return KeyData{}
	}

	return returnData
}

func (k KeyData) validKey() bool {
	t := time.Now().Unix()
	return int(t) <= k.ExpireTime
}

// Key validate the key
func Key(key string, service string) bool {
	keyFound := matchKey(key)

	if keyFound.validKey() {
		if keyFound.Service == service {
			return true
		}
	}

	return false
}
