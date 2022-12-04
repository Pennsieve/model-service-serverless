.PHONY: build-lambda test

LAMBDA_BUCKET ?= "pennsieve-cc-lambda-functions-use1"
WORKING_DIR   ?= "$(shell pwd)"
API_DIR ?= "api"

package:
	cd lambda/service && \
	env GOOS=linux GOARCH=amd64 go build -o ../bin/modelService/model_service

test:
	@echo ""
	@echo "*******************"
	@echo "*   Testing API   *"
	@echo "*******************"
	@echo ""
	@\
		cd $(API_DIR); \
		go test ./... ;
	@echo ""
	@echo "***********************"
	@echo "*   Testing Lambda    *"
	@echo "***********************"
	@echo ""
	@\
		cd ${WORKING_DIR}/lambda/service; \
		go test ./... ;

publish:
	@make package
	@echo ""
	@echo "*************************"
	@echo "*   Publishing lambda   *"
	@echo "*************************"
	@echo ""
	@aws s3 cp $(WORKING_DIR)/$(PACKAGE_NAME) s3://${LAMBDA_BUCKET}/$(SERVICE_NAME)/
	@rm -rf $(WORKING_DIR)/$(PACKAGE_NAME)