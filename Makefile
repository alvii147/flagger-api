GO=go
BIN=bin
EXE=cmd
SRC=./...
PSQL=PGPASSWORD=${FLAGGERAPI_POSTGRES_PASSWORD} psql --username=${FLAGGERAPI_POSTGRES_USERNAME}

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
	@$(PSQL) --file=db/drop_db.sql

.PHONY: create-db
create-db:
	@$(PSQL) --file=db/create_db.sql
	@$(PSQL) --dbname=${FLAGGERAPI_POSTGRES_DATABASE_NAME} --file=db/create_tables.sql

.PHONY: drop-test-db
drop-test-db:
	@$(PSQL) --file=db/drop_test_db.sql

.PHONY: create-test-db
create-test-db:
	@$(PSQL) --file=db/create_test_db.sql
	@$(PSQL) --dbname=test_${FLAGGERAPI_POSTGRES_DATABASE_NAME} --file=db/create_tables.sql

.PHONY: test
test: create-test-db
	-FLAGGERAPI_POSTGRES_DATABASE_NAME=test_${FLAGGERAPI_POSTGRES_DATABASE_NAME} FLAGGERAPI_MAIL_CLIENT_TYPE=inmem $(GO) test ${TEST_OPTS} $(SRC)
	@$(PSQL) --file=db/drop_test_db.sql

.PHONY: cover
cover: test
	$(GO) tool cover -func coverage.out
