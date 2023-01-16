# model-service-serverless
New serverless service handling Neo4J .

## Testing
You can test the service locally by:

1. Spin up the NEO4J container using ```make start-neo4j-empty```. This will start a Docker container with NEO4J ports exposed.
2. Run ```go test ./...``` in the ```api``` folder. This will run all tests in all subdirectories. Navigate to the ```service`` or ```store``` folder to run only specific tests.

Alternatively, you can run ```make test``` to run all tests in a dockerized container. This will mimic how the tests are run on Jenkins. 

## CI

__Testing__

The tests are automatically run by Jenkins once you push to a feature branch. Successful tests are required to merge a feature branch into the main branch.

__Build and Development Deployment__

Artifacts are built in Jenkins and published to S3. The dev build triggers a deployment of the Lambda function and creates a "Lambda version" that is used by the model-service.

__Deployment of an Artifact__

1. Deployements to *development* are automatically done by Jenkins once you merge a feature branch into main.

2. Deployments to *production* are done via Jenkins.

   1. Determine the artifact version you want to deploy (you can find the latest version number in the development deployment job).
   2. Deploy the Lambda function via Jenkins.