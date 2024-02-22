GO=go
SRC=./...

FLAGGERAPI_POSTGRES_HOSTNAME ?= localhost
FLAGGERAPI_POSTGRES_PORT ?= 5432
FLAGGERAPI_POSTGRES_USERNAME ?= postgres
FLAGGERAPI_POSTGRES_PASSWORD ?= postgres
FLAGGERAPI_POSTGRES_DATABASE_NAME ?= flaggerdb

PSQL=psql --username=$(FLAGGERAPI_POSTGRES_USERNAME) --host=$(FLAGGERAPI_POSTGRES_HOSTNAME) --port=$(FLAGGERAPI_POSTGRES_PORT)

TEST_OPTS=-coverprofile coverage.out
ifdef REGEX
	TEST_OPTS=-run $(REGEX)
endif

ifeq ($(VERBOSE), 1)
	TEST_OPTS:=$(TEST_OPTS) -v
endif

.PHONY: create-test-db
create-test-db:
	docker exec --env PGPASSWORD=$(FLAGGERAPI_POSTGRES_PASSWORD) flagger-api-postgres $(PSQL) --command="CREATE DATABASE test_$(FLAGGERAPI_POSTGRES_DATABASE_NAME);"
	docker exec --env PGPASSWORD=$(FLAGGERAPI_POSTGRES_PASSWORD) flagger-api-postgres $(PSQL) --dbname=test_$(FLAGGERAPI_POSTGRES_DATABASE_NAME) --file=/docker-entrypoint-initdb.d/create_tables.sql

.PHONY: test
test: create-test-db
	-docker exec --env FLAGGERAPI_POSTGRES_DATABASE_NAME=test_$(FLAGGERAPI_POSTGRES_DATABASE_NAME) --env FLAGGERAPI_MAIL_CLIENT_TYPE=inmem flagger-api-server $(GO) test $(TEST_OPTS) $(SRC)
	docker exec --env PGPASSWORD=$(FLAGGERAPI_POSTGRES_PASSWORD) flagger-api-postgres $(PSQL) --command="DROP DATABASE test_$(FLAGGERAPI_POSTGRES_DATABASE_NAME);"

.PHONY: cover
cover: test
	docker exec flagger-api-server $(GO) tool cover -func coverage.out
