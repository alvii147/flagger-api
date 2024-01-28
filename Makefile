GO=go
BIN=bin
EXE=cmd
SRC=./...

FLAGGERAPI_POSTGRES_HOSTNAME ?= localhost
FLAGGERAPI_POSTGRES_PORT ?= 5432
FLAGGERAPI_POSTGRES_USERNAME ?= postgres
FLAGGERAPI_POSTGRES_PASSWORD ?= postgres
FLAGGERAPI_POSTGRES_DATABASE_NAME ?= flaggerdb

PSQL=PGPASSWORD=${FLAGGERAPI_POSTGRES_PASSWORD} psql --username=${FLAGGERAPI_POSTGRES_USERNAME} --host=${FLAGGERAPI_POSTGRES_HOSTNAME} --port=${FLAGGERAPI_POSTGRES_PORT}

TEST_OPTS=-coverprofile coverage.out
ifdef REGEX
	TEST_OPTS=-run ${REGEX}
endif

ifeq ($(VERBOSE), 1)
	TEST_OPTS:=${TEST_OPTS} -v
endif

.PHONY: build
build:
	$(GO) build -o $(BIN)/ $(SRC)

.PHONY: clean
clean:
	rm -rf $(BIN)/*

.PHONY: server
server: build
	./$(BIN)/$(EXE)

.PHONY: drop-db
drop-db:
	@$(PSQL) --command="DROP DATABASE ${FLAGGERAPI_POSTGRES_DATABASE_NAME};"

.PHONY: create-db
create-db:
	@$(PSQL) --command="CREATE DATABASE ${FLAGGERAPI_POSTGRES_DATABASE_NAME};"
	@$(PSQL) --dbname=${FLAGGERAPI_POSTGRES_DATABASE_NAME} --file=db/create_tables.sql

.PHONY: drop-test-db
drop-test-db:
	@$(PSQL) --command="DROP DATABASE test_${FLAGGERAPI_POSTGRES_DATABASE_NAME};"

.PHONY: create-test-db
create-test-db:
	@$(PSQL) --command="CREATE DATABASE test_${FLAGGERAPI_POSTGRES_DATABASE_NAME};"
	@$(PSQL) --dbname=test_${FLAGGERAPI_POSTGRES_DATABASE_NAME} --file=db/create_tables.sql

.PHONY: test
test: create-test-db
	-FLAGGERAPI_POSTGRES_DATABASE_NAME=test_${FLAGGERAPI_POSTGRES_DATABASE_NAME} FLAGGERAPI_MAIL_CLIENT_TYPE=inmem $(GO) test ${TEST_OPTS} $(SRC)
	@$(PSQL) --command="DROP DATABASE test_${FLAGGERAPI_POSTGRES_DATABASE_NAME};"

.PHONY: cover
cover: test
	$(GO) tool cover -func coverage.out
