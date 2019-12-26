#!/usr/bin/env bash

BUILD_BUCKET=bugfixes-builds-eu
STACK_NAME=authorizer

function build()
{
    echo "Build"
    GOOS=linux GOARCH=amd64 go build .
    zip ${STACK_NAME}.zip ${STACK_NAME}
}

function moveFile()
{
    echo "Move File"
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
    echo "UpdateStack"
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
    echo "DeleteStack"
    awslocal cloudformation delete-stack --stack-name ${STACK_NAME}
}

build
moveFile

STACK_EXISTS=$(aws cloudformation list-stacks --region eu-west-2 --stack-status-filter ROLLBACK_COMPLETE UPDATE_ROLLBACK_COMPLETE | jq '.StackSummaries[].StackName//empty' | grep "${STACK_NAME}")
if [[ -z ${STACK_EXISTS} ]] || [[ "${STACK_EXISTS}" == "" ]]; then
    echo "No Stack"
    createStack
else
    STACK_ROLLBACK=$(aws cloudformation list-stacks --stack-status-filter ROLLBACK_COMPLETE | jq '.StackSummaries[].StackName//empty' | grep "${STACK_NAME}")
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
