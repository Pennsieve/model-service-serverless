# model-service-serverless
New serverless service handling Neo4J 

## Testing

The tests are automatically run by Jenkins once you push to a feature branch. Successful tests are required to merge a feature branch into the main branch. 

## Deployment

__Build and Development Deployment__

Artifacts are built in Jenkins and published to S3. The dev build triggers a deployment of the Lambda function and creates a "Lambda version" that is used by the model-service.

__Deployment of an Artifact__

1. Deployements to *development* are automatically done by Jenkins once you merge a feature branch into main.

2. Deployments to *production* are done via Jenkins.

   1. Determine the artifact version you want to deploy (you can find the latest version number in the development deployment job).
   2. Deploy the Lambda function via Jenkins.