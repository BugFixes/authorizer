#!/usr/bin/env bash

STACK_NAME=authorizer
BUILD_BUCKET=builds

export AWS_DEFAULT_REGION=us-east-1

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

BUCKET_EXISTS=$(awslocal s3api list-buckets | jq '.Buckets[].Name//empty' | grep "${BUILD_BUCKET}")
if [[ -z "${BUCKET_EXISTS}" ]] || [[ "${BUCKET_EXISTS}" == "" ]]; then
  awslocal s3api create-bucket --bucket ${BUILD_BUCKET}
fi

STACK_EXISTS=$(awslocal cloudformation list-stacks --stack-status-filter ROLLBACK_COMPLETE UPDATE_ROLLBACK_COMPLETE | jq '.StackSummaries[].StackName//empty' | grep "${STACK_NAME}")
if [[ -z "${STACK_EXISTS}" ]] || [[ "${STACK_EXISTS}" == "" ]]; then
  awslocal s3 cp ./${STACK_NAME}.zip s3://${BUILD_BUCKET}/${STACK_NAME}.zip
  awslocal route53 create-hosted-zone --name docker.devel --caller-reference devStuff
  awslocal cloudformation create-stack \
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
  awslocal s3 cp ./${STACK_NAME}.zip s3://${BUILD_BUCKET}/${STACK_NAME}.zip
  awslocal cloudformation update-stack \
	  --template-body file://cf.yaml \
	  --stack-name ${STACK_NAME} \
	  --capabilities CAPABILITY_NAMED_IAM \
	  --parameters \
		  ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
		  ParameterKey=Environment,ParameterValue=dev \
		  ParameterKey=BuildBucket,ParameterValue=${BUILD_BUCKET} \
		  ParameterKey=BuildKey,ParameterValue=${STACK_NAME}.zip \
		  ParameterKey=DBEndpoint,ParameterValue=http://localhost:4569
fi

go test ./...
go test ./... -bench=. -run=$$$

removeFiles


