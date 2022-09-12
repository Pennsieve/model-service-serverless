# model-service-serverless
New serverless service handling Neo4J 

## Create Lambda Build
Prior to terraforming the Lambda (which zips and uploads the lambda to AWS), the Lambda function needs
to be build for the Lambda environment. You can do this with the following command:

```env GOOS=linux GOARCH=amd64 go build -o ../bin/modelService/model_service```