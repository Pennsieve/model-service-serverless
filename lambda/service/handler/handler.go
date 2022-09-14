package handler

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/dbTable"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/organization"
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

// Claims is an object containing claims and user info
type Claims struct {
	orgClaim       organization.Claim
	datasetClaim   dataset.Claim
	userId         int64
	isSuperAdmin   bool
	organizationId int64
}

func ModelServiceHandler(request events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	var err error

	var apiResponse *events.APIGatewayV2HTTPResponse

	r := regexp.MustCompile(`(?P<method>) (?P<pathKey>.*)`)
	routeKeyParts := r.FindStringSubmatch(request.RouteKey)
	routeKey := routeKeyParts[r.SubexpIndex("pathKey")]

	claims := parseClaims(request)
	authorized := false
	switch routeKey {
	case "/metadata/models":
		switch request.RequestContext.HTTP.Method {
		case "GET":
			//	Return all models for a specific dataset
			if authorized = hasRole(*claims, permissions.ViewGraphSchema); authorized {
				apiResponse, err = getDatasetModelsRoute(request, claims)
			}
		}
	case "/metadata/query":
		switch request.RequestContext.HTTP.Method {
		case "POST":
			fmt.Println("Handling POST /graph/query request")
			authorized = true
			//if authorized = hasRole(*claims, permissions.CreateDeleteFiles); authorized {
			apiResponse, err = postGraphQueryRoute(request, claims)
			//}
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

// parseClaims parses the claims in the provided request.
func parseClaims(request events.APIGatewayV2HTTPRequest) *Claims {

	claims := request.RequestContext.Authorizer.Lambda

	var orgClaim organization.Claim
	if val, ok := claims["org_claim"]; ok {
		orgClaims := val.(map[string]interface{})
		orgRole := int64(orgClaims["Role"].(float64))
		orgClaim = organization.Claim{
			Role:            dbTable.DbPermission(orgRole),
			IntId:           int64(orgClaims["IntId"].(float64)),
			EnabledFeatures: nil,
		}
	}

	var datasetClaim dataset.Claim
	if val, ok := claims["dataset_claim"]; ok {
		if val != nil {
			datasetClaims := val.(map[string]interface{})
			datasetRole := int64(datasetClaims["Role"].(float64))
			datasetClaim = dataset.Claim{
				Role:   dataset.Role(datasetRole),
				NodeId: datasetClaims["NodeId"].(string),
				IntId:  int64(datasetClaims["IntId"].(float64)),
			}
		}
	}

	userId := int64(claims["user_id"].(float64))
	isSuperAdmin := claims["is_super_admin"].(bool)
	organizationId := int64(claims["organization_id"].(float64))

	returnedClaims := Claims{
		orgClaim:       orgClaim,
		datasetClaim:   datasetClaim,
		userId:         userId,
		isSuperAdmin:   isSuperAdmin,
		organizationId: organizationId,
	}

	return &returnedClaims

}

// hasRole returns a boolean indicating whether the user has the correct permissions.
func hasRole(claims Claims, permission permissions.DatasetPermission) bool {

	//hasOrgRole := claims.orgClaim.Role >= dbTable.Delete

	hasValidPermissions := permissions.HasDatasetPermission(claims.datasetClaim.Role, permission)

	return hasValidPermissions

}
