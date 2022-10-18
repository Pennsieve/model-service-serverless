package labels

const (
	ORGANIZATION_LABEL      = "Organization"
	DATASET_LABEL           = "Dataset"
	USER_LABEL              = "User"
	MODEL_LABEL             = "Model"
	MODEL_PROPERTY_LABEL    = "ModelProperty"
	RECORD_LABEL            = "Record"
	PACKAGE_LABEL           = "Package"
	PROXY_RELATIONSHIP_TYPE = "belongs_to"
	MODEL_RELATIONSHIP_STUB = "ModelRelationshipStub"
)

const (
	InstanceOf     = "@INSTANCE_OF"
	HasProperty    = "@HAS_PROPERTY"
	InOrganization = "@IN_ORGANIZATION"
	InDataset      = "@IN_DATASET"
	CreatedBy      = "@CREATED_BY"
	UpdatedBy      = "@UPDATED_BY"
	InPackage      = "@IN_PACKAGE"
	RelatedTo      = "@RELATED_TO"
)

var RESERVED_SCHEMA_RELATIONSHIPS = []string{
	InstanceOf,
	HasProperty,
	InOrganization,
	InDataset,
	CreatedBy,
	UpdatedBy,
	InPackage,
	RelatedTo,
}
