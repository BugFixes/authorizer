#!/usr/bin/env bash

function createDatabase()
{
  echo "createDatabase"
  docker run \
    -d \
    -p 5432:5432 \
    -e POSTGRES_PASSWORD=tester \
    -e POSTGRES_USER=tester \
    -e POSTGRES_DB=postgres \
    --name tester_postgres \
    postgres:11.5
}

function injectStructure()
{
  echo "injectStructure"
  docker exec \
    -e PGPASSWORD=tester tester_postgres psql \
    -U tester \
    -d postgres \
    -c "CREATE TABLE "public"."agent" ("id" uuid, "name" varchar(200), "key" uuid, "secret" uuid, "company_id" uuid, PRIMARY KEY ("id"));"
}

function wipeDatabase()
{
  echo "wipeDatabase"
  PGPASSWORD=tester psql \
    -U tester \
    -d postgres \
    --host 0.0.0.0 \
    --port 5432 \
    -c "DROP TABLE "public"."agent";"
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
    createDatabase
    sleep 5
    injectStructure
    testCode
    wipeDatabase
fi

