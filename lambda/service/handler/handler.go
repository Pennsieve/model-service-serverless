package handler

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/pennsieve/model-service-serverless/api/shared"
	"github.com/pennsieve/model-service-serverless/api/store"
	"github.com/pennsieve/pennsieve-go-core/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/permissions"
	pgQueries "github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
	"os"
	"regexp"
	"time"
)

var neo4jDriver neo4j.DriverWithContext

// init runs on cold start of lambda and fetches variables and created neo4j driver.
func init() {

	log.SetFormatter(&log.JSONFormatter{})
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(ll)
	}

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
	neoDb := shared.NewNeo4jSession(neo4jDriver.NewSession(context.Background(), neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	}))
	defer neoDb.Close(context.Background())

	db, err := pgQueries.ConnectRDSWithOrg(int(claims.OrgClaim.IntId))
	defer db.Close()
	if err != nil {
		return apiResponse, err
	}

	// Create GraphStore object with initiated db.
	graphStore := store.NewModelServiceStore(db, neoDb)

	switch routeKey {
	case "/metadata_legacy/models":
		switch request.RequestContext.HTTP.Method {
		case "GET":
			//	Return all models for a specific dataset
			if authorized = authorizer.HasRole(*claims, permissions.ViewGraphSchema); authorized {
				apiResponse, err = getDatasetModelsRoute(graphStore, request, claims)
			}
		}
	case "/metadata_legacy/query":
		switch request.RequestContext.HTTP.Method {
		case "POST":
			if authorized = authorizer.HasRole(*claims, permissions.ViewRecords); authorized {
				apiResponse, err = postGraphQueryRoute(graphStore, request, claims)
			}
		}
	case "/metadata_legacy/query/autocomplete":
		switch request.RequestContext.HTTP.Method {
		case "POST":
			if authorized = authorizer.HasRole(*claims, permissions.ViewRecords); authorized {
				apiResponse, err = postAutocompleteRoute(graphStore, request, claims)
			}
		}

	case "/metadata_legacy/records/relationships":
		switch request.RequestContext.HTTP.Method {
		case "POST":
			if authorized = authorizer.HasRole(*claims, permissions.CreateDeleteRecord); authorized {
				apiResponse, err = postGraphRecordRelationshipRoute(graphStore, request, claims)
			}
		}

	case "/metadata_legacy/package":
		switch request.RequestContext.HTTP.Method {
		case "GET":
			//	Return all models for a specific dataset
			if authorized = authorizer.HasRole(*claims, permissions.ViewRecords); authorized {
				apiResponse, err = getMetaDataForPackage(graphStore, request, claims)
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
