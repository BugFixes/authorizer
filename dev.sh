#!/usr/bin/env bash

STACK_NAME=authorizer
BUILD_BUCKET=builds

export AWS_DEFAULT_REGION=us-east-1

function removeFiles()
{
    echo "RemoveFiles"
    if [[ -f "${STACK_NAME}.zip" ]]; then
        rm ${STACK_NAME}
        rm ${STACK_NAME}.zip
    fi
}

function createStack()
{
    echo "CreateStack"
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
}

function updateStack()
{
    echo "UpdateStack"
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
}

function deleteStack()
{
    echo "DeleteStack"
    awslocal cloudformation delete-stack --stack-name ${STACK_NAME}
}

function justDB()
{
    echo "justDB"
    awslocal dynamodb delete-table \
        --table-name authorizer-dynamo-dev \
        --endpoint-url http://0.0.0.0:4569

    awslocal dynamodb create-table \
        --table-name authorizer-dynamo-dev \
        --attribute-definitions AttributeName=id,AttributeType=S \
        --key-schema AttributeName=id,KeyType=HASH \
        --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
        --endpoint-url http://0.0.0.0:4569
}

function build()
{
    echo "Build"
    GOOS=linux GOARCH=amd64 go build .
    zip ${STACK_NAME}.zip ${STACK_NAME}
}

function bucket()
{
    echo "Bucket"
    BUCKET_EXISTS=$(awslocal s3api list-buckets | jq '.Buckets[].Name//empty' | grep "${BUILD_BUCKET}")
    if [[ -z "${BUCKET_EXISTS}" ]] || [[ "${BUCKET_EXISTS}" == "" ]]; then
        awslocal s3api create-bucket --bucket ${BUILD_BUCKET}
    fi
}

function cloudFormation()
{
    STACK_EXISTS=$(awslocal cloudformation list-stacks --stack-status-filter ROLLBACK_COMPLETE UPDATE_ROLLBACK_COMPLETE | jq '.StackSummaries[].StackName//empty' | grep "${STACK_NAME}")
    if [[ -z "${STACK_EXISTS}" ]] || [[ "${STACK_EXISTS}" == "" ]]; then
        createStack
    else
        STACK_ROLLBACK=$(awslocal cloudformation list-stacks --stack-status-filter ROLLBACK_COMPLETE | jq '.StackSummaries[].StackName//empty' | grep "${STACK_NAME}")
        if [[ -z "${STACK_ROLLBACK}" ]] || [[ "${STACK_ROLLBACK}" == "" ]]; then
            updateStack
        else
            deleteStack
            createStack
        fi
    fi
}

function testIt()
{
    go test ./...
    go test ./... -bench=. -run=$$$
}


if [[ ! -z ${1} ]] || [[ "${1}" != "" ]]; then
    ${1}
else
    removeFiles
    build
    bucket
    cloudFormation
    removeFiles
fi

