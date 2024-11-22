package limitless

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

// ScanRequest represents the parameters for a DynamoDB scan operation.
type ScanRequest struct {
	// TableName is the name of the DynamoDB table to scan.
	TableName string

	// IndexName is an optional name of a secondary index to scan instead of the base table.
	IndexName *string

	// ConsistentRead specifies whether to use strongly consistent reads.
	// If set to true, the operation uses strongly consistent reads; otherwise, eventually consistent reads are used.
	ConsistentRead *bool

	// Limit sets the maximum number of items to evaluate (not necessarily the number of matching items).
	// If DynamoDB processes the number of items up to the limit while processing the results, it stops the operation and returns the matching values up to that point.
	Limit *int32

	// ExclusiveStartKey is the primary key of the item where the scan operation will start.
	// Used for pagination to resume a scan operation.
	ExclusiveStartKey map[string]types.AttributeValue
}

// QueryRequest represents the parameters for a DynamoDB query operation.
type QueryRequest struct {
	// TableName is the name of the DynamoDB table to query.
	TableName string

	// Condition is the query condition expression. The condition expression has the format "#KEY operator :VALUE".
	// The "key" is inferred by removing the leading "#" in "#KEY". You provide the value of ":VALUE" in the [Values] field.
	// For example, "#name = :nameValue" would match items with the attribute "name" equal to the value provided in the "Values" map.
	Condition string

	// Values is a map of attribute names to attribute values, representing the query parameters.
	// For example, { ":nameValue": "Soumya" } would set the value of ":nameValue" equal to "Soumya".
	Values map[string]any

	// IndexName is an optional name of a secondary index to query instead of the base table.
	IndexName *string

	// ConsistentRead specifies whether to use strongly consistent reads. Must be false if using an index.
	// If set to true, the operation uses strongly consistent reads; otherwise, eventually consistent reads are used.
	ConsistentRead *bool

	// Ascending determines the order of the query results.
	// If set to true, the query processes items in ascending order; if false, in descending order.
	Ascending *bool

	// Limit sets the maximum number of items to evaluate (not necessarily the number of matching items).
	// If DynamoDB processes the number of items up to the limit while processing the results, it stops the operation and returns the matching values up to that point.
	Limit *int32

	// ProjectionExpression specifies the attributes to be returned in the query result.
	// This is a string identifying one or more attributes to retrieve from the table.
	ProjectionExpression *string

	// ExpressionAttributeNames provides name substitution for reserved words in the ProjectionExpression.
	// Use this if you need to specify reserved words as attribute names.
	ExpressionAttributeNames map[string]string

	// ExclusiveStartKey is the primary key of the item where the query operation will start.
	// Used for pagination to resume a query operation.
	ExclusiveStartKey map[string]types.AttributeValue
}

// PutItemRequest represents the parameters for a DynamoDB PutItem operation.
type PutItemRequest struct {
	// TableName is the name of the DynamoDB table where the item will be put.
	TableName string

	// Item is the item to be put into the table. It should be a map or a struct
	// that can be marshaled into a DynamoDB-compatible format. Use the "dynamodbav" tag.
	Item interface{}
}
