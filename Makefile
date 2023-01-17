.PHONY: help clean test test-ci package publish start-neo4j-empty

LAMBDA_BUCKET ?= "pennsieve-cc-lambda-functions-use1"
WORKING_DIR   ?= "$(shell pwd)"
API_DIR ?= "api"
SERVICE_NAME  ?= "model-service-serverless"
PACKAGE_NAME  ?= "${SERVICE_NAME}-${IMAGE_TAG}.zip"
NEO4J_APOC_VERSION ?= "3.5.0.13"

.DEFAULT: help

help:
	@echo "Make Help for $(SERVICE_NAME)"
	@echo ""
	@echo "make start-neo4j-empty   	- Start a clean NEO4J container for local testing"
	@echo "make clean			- spin down containers and remove db files"
	@echo "make test			- run dockerized tests locally"
	@echo "make test-ci			- run dockerized tests for Jenkins"
	@echo "make package			- create venv and package lambda function"
	@echo "make publish			- package and publish lambda function"

# Start a clean NEO4J container for local testing
start-neo4j-empty: docker-clean install-apoc
	docker-compose -f docker-compose.test.yml --compatibility up neo4j

# Run dockerized tests (can be used locally)
test:
	docker-compose -f docker-compose.test.yml down --remove-orphans
	mkdir -p data conf
	chmod -R 777 data conf
	docker-compose -f docker-compose.test.yml up --exit-code-from local_tests local_tests
	make clean

# Run dockerized tests (used on Jenkins)
test-ci: install-apoc
	docker-compose -f docker-compose.test.yml down --remove-orphans
	mkdir -p data plugins conf logs
	chmod -R 777 conf
	@IMAGE_TAG=$(IMAGE_TAG) docker-compose -f docker-compose.test.yml up --exit-code-from=ci-tests ci-tests

# Remove folders created by NEO4J docker container
clean: docker-clean
	rm -rf conf
	rm -rf data
	rm -rf plugins

# Spin down active docker containers.
docker-clean:
	docker-compose -f docker-compose.test.yml down

# Installing apoc plugin for NEO4J database
install-apoc:
	mkdir -p $(PWD)/plugins
	bin/install-neo4j-apoc.sh $(PWD)/plugins $(NEO4J_APOC_VERSION)

# Run local NEO4J container with exposed ports
neo4j:
	docker-compose -f docker-compose.test.yml up -d neo4j

# Build lambda and create ZIP file
package:
	@echo ""
	@echo "***********************"
	@echo "*   Building lambda   *"
	@echo "***********************"
	@echo ""
	cd lambda/service; \
  		env GOOS=linux GOARCH=amd64 go build -o $(WORKING_DIR)/lambda/bin/modelService/model_service; \
		cd $(WORKING_DIR)/lambda/bin/modelService/ ; \
			zip -r $(WORKING_DIR)/lambda/bin/modelService/$(PACKAGE_NAME) .

# Copy Service lambda to S3 location
publish:
	@make package
	@echo ""
	@echo "*************************"
	@echo "*   Publishing lambda   *"
	@echo "*************************"
	@echo ""
	aws s3 cp $(WORKING_DIR)/lambda/bin/modelService/$(PACKAGE_NAME) s3://$(LAMBDA_BUCKET)/$(SERVICE_NAME)/
	rm -rf $(WORKING_DIR)/lambda/bin/modelService/$(PACKAGE_NAME)