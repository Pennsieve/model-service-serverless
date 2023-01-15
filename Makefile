.PHONY: help clean test testdocker package publish

LAMBDA_BUCKET ?= "pennsieve-cc-lambda-functions-use1"
WORKING_DIR   ?= "$(shell pwd)"
API_DIR ?= "api"
SERVICE_NAME  ?= "model-service-serverless"
PACKAGE_NAME  ?= "${SERVICE_NAME}-${VERSION}.zip"

.DEFAULT: help

help:
	@echo "Make Help for $(SERVICE_NAME)"
	@echo ""
	@echo "make clean   - removes node_modules directory"
	@echo "make test    - run tests"
	@echo "make package - create venv and package lambda function"
	@echo "make publish - package and publish lambda function"

#test:
#	@echo ""
#	@echo "*******************"
#	@echo "*   Testing API   *"
#	@echo "*******************"
#	@echo ""
#	@cd $(API_DIR); \
#		go test ./... ;
#	@echo ""
#	@echo "***********************"
#	@echo "*   Testing Lambda    *"
#	@echo "***********************"
#	@echo ""
#	@cd $(WORKING_DIR)/lambda/service; \
#		go test ./... ;

test:
	cd $(API_DIR); \
		docker-compose -f docker-compose.test.yml up --build --exit-code-from api_tests; e1=$$? ;\
		docker-compose -f docker-compose.test.yml down --volumes; \
		sudo rm -rf ${HOME}/neo4j/testdata
		exit "$$(( e1 ))"

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

publish:
	@make package
	@echo ""
	@echo "*************************"
	@echo "*   Publishing lambda   *"
	@echo "*************************"
	@echo ""
	aws s3 cp $(WORKING_DIR)/lambda/bin/modelService/$(PACKAGE_NAME) s3://$(LAMBDA_BUCKET)/$(SERVICE_NAME)/
	rm -rf $(WORKING_DIR)/lambda/bin/modelService/$(PACKAGE_NAME)