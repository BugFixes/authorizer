#!/usr/bin/env bash

STACK_NAME=authorizer
BUILD_BUCKET=builds

removeFiles()
{
  if [[ -f "${STACK_NAME}.zip" ]]; then
    rm ${STACK_NAME}
    rm ${STACK_NAME}.zip
  fi
}

removeFiles

GOOS=linux GOARCH=amd64 go build .
zip ${STACK_NAME}.zip ${STACK_NAME}

BUCKET_EXISTS=$(aws --region us-east-1 --endpoint-url http://localhost:4572 s3api list-buckets | jq '.Buckets[].Name//empty' | grep "${BUILD_BUCKET}")
if [[ -z "${BUCKET_EXISTS}" ]] || [[ "${BUCKET_EXISTS}" == "" ]]; then
  aws --region us-east-1 --endpoint-url http://localhost:4572 s3api create-bucket --bucket ${BUILD_BUCKET}
fi

STACK_EXISTS=$(aws --region us-east-1 --endpoint-url http://localhost:4581 cloudformation list-stacks --stack-status-filter ROLLBACK_COMPLETE UPDATE_ROLLBACK_COMPLETE | jq '.StackSummaries[].StackName//empty' | grep "${STACK_NAME}")
if [[ -z "${STACK_EXISTS}" ]] || [[ "${STACK_EXISTS}" == "" ]]; then
  aws --region us-east-1 --endpoint-url http://localhost:4572 s3 cp ./${STACK_NAME}.zip s3://${BUILD_BUCKET}/${STACK_NAME}.zip
  aws --region us-east-1 --endpoint-url http://localhost:4580 route53 create-hosted-zone --name docker.devel --caller-reference devStuff
  aws --region us-east-1 --endpoint-url http://localhost:4581 cloudformation create-stack \
	  --template-body file://cf.yaml \
	  --stack-name ${STACK_NAME} \
	  --capabilities CAPABILITY_NAMED_IAM \
	  --parameters \
		  ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
		  ParameterKey=Environment,ParameterValue=dev \
		  ParameterKey=BuildBucket,ParameterValue=${BUILD_BUCKET} \
		  ParameterKey=BuildKey,ParameterValue=${STACK_NAME}.zip \
		  ParameterKey=DBEndpoint,ParameterValue=http://localhost:4569
else
  aws --region us-east-1 --endpoint-url http://localhost:4572 s3 cp ./${STACK_NAME}.zip s3://${BUILD_BUCKET}/${STACK_NAME}.zip
  aws --region us-east-1 --endpoint-url http://localhost:4581 cloudformation update-stack \
	  --template-body file://cf.yaml \
	  --stack-name ${STACK_NAME} \
	  --capabilities CAPABILITY_NAMED_IAM \
	  --parameters \
		  ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
		  ParamterKey=Environment,ParamterValue=dev \
		  ParameterKey=BuildBucket,ParamterValue=${BUILD_BUCKET} \
		  ParameterKey=BuildKey,ParameterValue=${STACK_NAME}.zip \
		  ParameterKey=DBEndpoint,ParameterValue=http://localhost:4569
fi

go test ./...
go test ./... -bench=. -run=$$$

removeFiles


