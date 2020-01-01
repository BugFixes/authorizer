#!/usr/bin/env bash

BUILD_BUCKET=bugfixes-builds-eu
STACK_NAME=authorizer

function build()
{
    echo "build"
    GOOS=linux GOARCH=amd64 go build .
    zip ${STACK_NAME}-${GITHUB_SHA}.zip ${STACK_NAME}
}

function moveFiles()
{
    echo "moveFiles"
    aws s3 cp ./${STACK_NAME}-${GITHUB_SHA}.zip s3://${BUILD_BUCKET}/${STACK_NAME}-${GITHUB_SHA}.zip
}

function createStack()
{
    echo "CreateStack"
    aws cloudformation create-stack \
	    --template-body file://.ci/cloud/cloudformation.yaml \
	    --stack-name ${STACK_NAME} \
	    --capabilities CAPABILITY_NAMED_IAM \
	    --parameters \
		    ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
		    ParameterKey=Environment,ParameterValue=live \
		    ParameterKey=BuildBucket,ParameterValue=${BUILD_BUCKET} \
		    ParameterKey=BuildKey,ParameterValue=${STACK_NAME}-${GITHUB_SHA}.zip \
		    ParameterKey=DBHostname,ParameterValue=${DB_HOSTNAME} \
		    ParameterKey=DBPort,ParameterValue=${DB_PORT} \
		    ParameterKey=DBUsername,ParameterValue=${DB_USERNAME} \
		    ParameterKey=DBPassword,ParameterValue=${DB_PASSWORD} \
		    ParameterKey=DBTable,ParameterValue=agent \
		    ParameterKey=DBDatabase,ParameterValue=bugfixes
}

function updateStack()
{
    echo "updateStack"
    aws cloudformation update-stack \
	    --template-body file://.ci/cloud/cloudformation.yaml \
	    --stack-name ${STACK_NAME} \
	    --capabilities CAPABILITY_NAMED_IAM \
	    --parameters \
	      ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
		    ParameterKey=Environment,ParameterValue=live \
		    ParameterKey=BuildBucket,ParameterValue=${BUILD_BUCKET} \
		    ParameterKey=BuildKey,ParameterValue=${STACK_NAME}-${GITHUB_SHA}.zip \
		    ParameterKey=DBHostname,ParameterValue=${DB_HOSTNAME} \
		    ParameterKey=DBPort,ParameterValue=${DB_PORT} \
		    ParameterKey=DBUsername,ParameterValue=${DB_USERNAME} \
		    ParameterKey=DBPassword,ParameterValue=${DB_PASSWORD} \
		    ParameterKey=DBTable,ParameterValue=agent \
		    ParameterKey=DBDatabase,ParameterValue=bugfixes
}

function deleteStack()
{
    echo "deleteStack"
    awslocal cloudformation delete-stack --stack-name ${STACK_NAME}
}

function testCode()
{
    TEST_CODE=true DB_DATABASE=tester DB_TABLE=agent DB_HOSTNAME=0.0.0.0 DB_PORT=5432 DB_USERNAME=postgres DB_PASSWORD=tester go test ./...
    echo "----"
    echo "---- Benchmarks ----"
    echo "----"
    TEST_CODE=true DB_DATABASE=tester DB_TABLE=agent DB_HOSTNAME=0.0.0.0 DB_PORT=5432 DB_USERNAME=postgres DB_PASSWORD=tester go test ./... -bench=. -run=$$$
}

function testDatabase()
{
    echo "testDatabase"
    docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=tester -e POSGRES_USERNAME=tester -e POSTGRES_DB=tester --name tester_postgres postgres:11.5
    sleep 10
    docker exec -e PGPASSWORD=tester tester_postgres psql -U postgres -d tester -c "DROP TABLE "public"."agent";"
    docker exec -e PGPASSWORD=tester tester_postgres psql -U postgres -d tester -c "CREATE TABLE "public"."agent" ("id" uuid, "name" varchar(200), "key" uuid, "secret" uuid, "company_id" uuid, PRIMARY KEY ("id"));"
}

function cloudFormation()
{
  echo "cloudFormation"
  STACK_EXISTS=$(aws cloudformation list-stacks --region eu-west-2 --stack-status-filter ROLLBACK_COMPLETE UPDATE_ROLLBACK_COMPLETE CREATE_COMPLETE UPDATE_COMPLETE | jq '.StackSummaries[].StackName//empty' | grep "${STACK_NAME}")
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



