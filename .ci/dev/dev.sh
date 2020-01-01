#!/usr/bin/env bash

STACK_NAME=authorizer
BUILD_BUCKET=bugfixes-builds-eu
#TABLE_NAME=authorizer-dynamo-dev
DB_INSTANCE_NAME=bugfixes-dev

export AWS_DEFAULT_REGION=us-east-1

function removeFiles()
{
    echo "removeFiles"
    if [[ -f "${STACK_NAME}.zip" ]]; then
        rm ${STACK_NAME}
        rm ${STACK_NAME}.zip
    fi
}

function createStructure()
{
  echo "createStructure"
  PGPASSWORD=bugfixes psql \
    -U bugfixes \
    -d bugfixes \
    --host 0.0.0.0 \
    --port 4511 \
    -c "$(cat .ci/dev/structure.sql)"
}

function createDatabase()
{
  echo "createDatabase"
  awslocal rds create-db-instance \
    --db-instance-identifier ${DB_INSTANCE_NAME} \
    --engine postgres \
    --db-instance-class db.t2.micro \
    --db-name bugfixes \
    --master-username bugfixes \
    --master-user-password bugfixes

  createStructure
}

function wipeDatabase()
{
  echo "wipeDatabase"
  createStructure
}

function database()
{
  echo "database"
  DATABASE_EXISTS=""
  DATABASE_RESPONSE=$(awslocal rds describe-db-instances --filters Name=db-instance-id,Values=bugfixes)
  if [[ ! -z ${DATABASE_RESPONSE} ]] || [[ "${DATABASE_RESPONSE}" != "" ]]; then
    DATABASE_EXISTS=$(${DATABASE_RESPONSE} | jq ."DBInstances[]//empty")
    if [[ -z ${DATABASE_EXISTS} ]] || [[ "${DATABASE_EXISTS}" == "" ]]; then
      createDatabase
    else
      wipeDatabase
    fi
  fi
}

function moveFiles()
{
  echo "moveFiles"
  awslocal s3 cp ./${STACK_NAME}.zip s3://${BUILD_BUCKET}/${STACK_NAME}.zip
}

function createStack()
{
    echo "createStack"
    #awslocal route53 create-hosted-zone --name docker.devel --caller-reference devStuff
    awslocal cloudformation create-stack \
	    --template-body file://.ci/cloud/cloudformation.yaml \
	    --stack-name ${STACK_NAME} \
	    --capabilities CAPABILITY_NAMED_IAM \
	    --parameters \
		    ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
		    ParameterKey=Environment,ParameterValue=dev \
		    ParameterKey=BuildBucket,ParameterValue=${BUILD_BUCKET} \
		    ParameterKey=BuildKey,ParameterValue=${STACK_NAME}.zip \
		    ParameterKey=DBHostname,ParameterValue=0.0.0.0 \
		    ParameterKey=DBPort,ParameterValue=4511 \
		    ParameterKey=DBUsername,ParameterValue=bugfixes \
		    ParameterKey=DBPassword=ParameterValue=bugfixes \
		    ParameterKey=DBTable,ParameterValue=agent \
		    ParameterKey=DBDatabase,ParameterValue=bugfixes
}

function updateStack()
{
    echo "UpdateStack"
    awslocal cloudformation update-stack \
	    --template-body file://.ci/cloud/cloudformation.yaml \
	    --stack-name ${STACK_NAME} \
	    --capabilities CAPABILITY_NAMED_IAM \
	    --parameters \
		    ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
		    ParameterKey=Environment,ParameterValue=dev \
		    ParameterKey=BuildBucket,ParameterValue=${BUILD_BUCKET} \
		    ParameterKey=BuildKey,ParameterValue=${STACK_NAME}.zip \
		    ParameterKey=DBHostname,ParameterValue=0.0.0.0 \
		    ParameterKey=DBPort,ParameterValue=4511 \
		    ParameterKey=DBUsername,ParameterValue=bugfixes \
		    ParameterKey=DBPassword=ParameterValue=bugfixes \
		    ParameterKey=DBTable,ParameterValue=agent \
		    ParameterKey=DBDatabase,ParameterValue=bugfixes
}

function deleteStack()
{
    echo "DeleteStack"
    awslocal cloudformation delete-stack --stack-name ${STACK_NAME}
}

function build()
{
    echo "Build"
    GOOS=linux GOARCH=amd64 go build .
    zip ${STACK_NAME}.zip ${STACK_NAME}
}

function bucket()
{
    echo "bucket"
    BUCKET_EXISTS=$(awslocal s3api list-buckets | jq '.Buckets[].Name//empty' | grep "${BUILD_BUCKET}")
    if [[ -z "${BUCKET_EXISTS}" ]] || [[ "${BUCKET_EXISTS}" == "" ]]; then
        awslocal s3api create-bucket --bucket ${BUILD_BUCKET}
    fi
}

function cloudFormation()
{
    echo "cloudFormation"
    moveFiles
    STACK_EXISTS=$(awslocal cloudformation list-stacks --stack-status-filter ROLLBACK_COMPLETE UPDATE_ROLLBACK_COMPLETE CREATE_COMPLETE | jq '.StackSummaries[].StackName//empty' | grep "${STACK_NAME}")
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

function testCode()
{
    echo "testCode"
    go test ./...
    go test ./... -bench=. -run=$$$
}


if [[ ! -z ${1} ]] || [[ "${1}" != "" ]]; then
    ${1}
else
    removeFiles
    build
    database
    testCode
    bucket
    cloudFormation
    removeFiles
fi

