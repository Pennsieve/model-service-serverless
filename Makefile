.PHONY: build-lambda

build-lambda:
	cd lambda/service && \
	env GOOS=linux GOARCH=amd64 go build -o ../bin/modelService/model_service