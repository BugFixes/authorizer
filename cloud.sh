#!/usr/bin/env bash

BUILD_BUCKET=bugfixes-builds-eu
STACK_NAME=authorizer
TABLE_NAME=authorizer

function build()
{
    echo "build"
    GOOS=linux GOARCH=amd64 go build .
    zip ${STACK_NAME}.zip ${STACK_NAME}
}

function moveFiles()
{
    echo "moveFiles"
    aws s3 cp ./${STACK_NAME}.zip s3://${BUILD_BUCKET}/${STACK_NAME}.zip
}

function createStack()
{
    echo "CreateStack"
    aws cloudformation create-stack \
	    --template-body file://cf.yaml \
	    --stack-name ${STACK_NAME} \
	    --capabilities CAPABILITY_NAMED_IAM \
	    --parameters \
		    ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
		    ParameterKey=Environment,ParameterValue=dev \
		    ParameterKey=BuildBucket,ParameterValue=${BUILD_BUCKET} \
		    ParameterKey=BuildKey,ParameterValue=${STACK_NAME}.zip
}

function updateStack()
{
    echo "updateStack"
    aws cloudformation update-stack \
	    --template-body file://cf.yaml \
	    --stack-name ${STACK_NAME} \
	    --capabilities CAPABILITY_NAMED_IAM \
	    --parameters \
		    ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
		    ParameterKey=Environment,ParameterValue=dev \
		    ParameterKey=BuildBucket,ParameterValue=${BUILD_BUCKET} \
		    ParameterKey=BuildKey,ParameterValue=${STACK_NAME}.zip
}

function deleteStack()
{
    echo "deleteStack"
    awslocal cloudformation delete-stack --stack-name ${STACK_NAME}
}

function testCode()
{
    go test ./...
    go test ./... -bench=. -run=$$$
}

function testDatabase()
{
    echo "testDatabase"
    aws dynamodb create-table \
        --table-name ${TABLE_NAME} \
        --attribute-definitions AttributeName=id,AttributeType=S \
        --key-schema AttributeName=id,KeyType=HASH \
        --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
        --endpoint-url http://0.0.0.0:4569 \
        --region eu-west-2
}

function cloudFormation()
{
  echo "cloudFormation"
  STACK_EXISTS=$(aws cloudformation list-stacks --region eu-west-2 --stack-status-filter ROLLBACK_COMPLETE UPDATE_ROLLBACK_COMPLETE CREATE_COMPLETE | jq '.StackSummaries[].StackName//empty' | grep "${STACK_NAME}")
  if [[ -z ${STACK_EXISTS} ]] || [[ "${STACK_EXISTS}" == "" ]]; then
    echo "No Stack"
    createStack
  else
    STACK_ROLLBACK=$(aws cloudformation list-stacks --region eu-west-2 --stack-status-filter ROLLBACK_COMPLETE | jq '.StackSummaries[].StackName//empty' | grep "${STACK_NAME}")
    if [[ -z ${STACK_ROLLBACK} ]] || [[ "${STACK_ROLLBACK}" == "" ]]; then
        echo "Good standing"
        updateStack
    else
        echo "Failed Stack"
        deleteStack
        sleep 60
        createStack
    fi
  fi
}

if [[ ! -z ${1} ]] || [[ "${1}" != "" ]]; then
  ${1}
else
  build
  moveFiles
  cloudFormation
fi



