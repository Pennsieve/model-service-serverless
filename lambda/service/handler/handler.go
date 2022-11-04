package handler

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/pennsieve/model-service-serverless/api/service"
	"github.com/pennsieve/model-service-serverless/api/store"
	"github.com/pennsieve/pennsieve-go-api/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/permissions"
	"log"
	"regexp"
	"time"
)

var neo4jDriver neo4j.DriverWithContext

// init runs on cold start of lambda and fetches variables and created neo4j driver.
func init() {

	// Get the Model-Service variables from SSM.
	ssmVars, err := fetchSSMVariables()
	if err != nil {
		log.Fatalln(err)
	}

	neo4jDriver, err = neo4j.NewDriverWithContext(*ssmVars.dbUri,
		neo4j.BasicAuth(*ssmVars.neo4jUserName, *ssmVars.neo4jPassword, ""),
		func(config *neo4j.Config) {
			config.MaxConnectionPoolSize = 10
			config.MaxConnectionLifetime = 5 * time.Minute
			config.ConnectionAcquisitionTimeout = 10 * time.Second
		})
	if err != nil {
		panic(err)
	}

	// We are not closing the driver to allow the driver to be used across lambda calls while lambda is hot. We will
	// need to depend on neo4j to close stale connections after a time-out.
}

func ModelServiceHandler(request events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	var err error
	var apiResponse *events.APIGatewayV2HTTPResponse

	r := regexp.MustCompile(`(?P<method>) (?P<pathKey>.*)`)
	routeKeyParts := r.FindStringSubmatch(request.RouteKey)
	routeKey := routeKeyParts[r.SubexpIndex("pathKey")]

	claims := authorizer.ParseClaims(request.RequestContext.Authorizer.Lambda)
	authorized := false

	// Initiate NEO4j session
	db := neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer db.Close(context.Background())

	// Create GraphStore object with initiated db.
	graphStore := store.NewGraphStore(db)
	graphService := service.NewGraphService(graphStore)

	switch routeKey {
	case "/metadata/models":
		switch request.RequestContext.HTTP.Method {
		case "GET":
			//	Return all models for a specific dataset
			if authorized = authorizer.HasRole(*claims, permissions.ViewGraphSchema); authorized {
				apiResponse, err = getDatasetModelsRoute(graphService, request, claims)
			}
		}
	case "/metadata/query":
		switch request.RequestContext.HTTP.Method {
		case "POST":
			fmt.Println("Handling POST /graph/query request")
			authorized = true
			//if authorized = hasRole(*claims, permissions.CreateDeleteFiles); authorized {
			apiResponse, err = postGraphQueryRoute(graphService, request, claims)
			//}
		}
	case "/metadata/records/relationships":
		switch request.RequestContext.HTTP.Method {
		case "POST":
			fmt.Println("Handling POST /metadata/records/relationships")
			if authorized = authorizer.HasRole(*claims, permissions.CreateDeleteRecord); authorized {
				apiResponse, err = postGraphRecordRelationshipRoute(graphService, request, claims)
			}
		}
	}

	// Return unauthorized if
	if !authorized {
		apiResponse := events.APIGatewayV2HTTPResponse{
			StatusCode: 403,
			Body:       `{"message": "User is not authorized to perform this action on the dataset."}`,
		}
		return &apiResponse, nil
	}

	// Response
	if err != nil {
		log.Fatalln("Something is wrong with creating the response", err)
	}
	return apiResponse, nil

}
