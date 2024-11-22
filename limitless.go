package limitless

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Limitless struct {
	ddb *dynamodb.Client
}

func NewClient(dynamodb *dynamodb.Client) *Limitless {
	client := new(Limitless)
	client.ddb = dynamodb
	return client
}

func GetItem[T any](l *Limitless, tableName string, key *T) (*T, error) {
	nout := new(T)

	Key, err := attributevalue.MarshalMap(key)
	if err != nil {
		return nil, err
	}
	res, err := l.ddb.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &tableName,
		Key:       Key,
	})
	if err != nil {
		return nil, err
	}

	if len(res.Item) == 0 {
		return nil, nil
	}

	if err := attributevalue.UnmarshalMap(res.Item, &nout); err != nil {
		return nil, err
	}
	return nout, nil
}

func PutItem[T any](l *Limitless, tableName string, item T) error {
	Item, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}
	_, err = l.ddb.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      Item,
	})
	if err != nil {
		return err
	}
	return nil
}

func PutItemConditional[T any](l *Limitless, tableName string, conditionExpression string, item T) error {
	Item, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}
	_, err = l.ddb.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName:           &tableName,
		Item:                Item,
		ConditionExpression: &conditionExpression,
	})
	if err != nil {
		return err
	}
	return nil
}

func BatchPutItem[T any](l *Limitless, tableName string, items []T) error {
	writeRequests := make([]types.WriteRequest, len(items))
	for i, v := range items {
		Item, err := attributevalue.MarshalMap(v)
		if err != nil {
			return err
		}

		writeRequests[i] = types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: Item,
			},
		}
	}

	_, err := l.ddb.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			tableName: writeRequests,
		},
	})
	return err
}

func BatchGetItem[T any](l *Limitless, tableName string, items []T) ([]T, error) {
	keys := make([]map[string]types.AttributeValue, len(items))
	for i, v := range items {
		Item, err := attributevalue.MarshalMap(v)
		if err != nil {
			return nil, err
		}

		keys[i] = Item
	}

	response, err := l.ddb.BatchGetItem(context.TODO(), &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			tableName: {
				Keys: keys,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	result := make([]T, 0)
	for _, r := range response.Responses {
		for _, v := range r {
			item := new(T)
			err = attributevalue.UnmarshalMap(v, item)
			if err != nil {
				return nil, err
			}
			result = append(result, *item)
		}
	}
	return result, err
}

// Query performs a query operation on a DynamoDB table using the provided query request.
// It processes the condition and values, builds the necessary input for the DynamoDB Query API,
// executes the query, and unmarshals the results into a slice of type T.
//
// Parameters:
//   - l: A pointer to the Limitless client.
//   - queryRequest: A QueryRequest struct containing the query parameters.
//
// Returns:
//   - []T: A slice of type T containing the query results.
//   - map[string]types.AttributeValue: The LastEvaluatedKey from the query, which can be used for pagination.
//   - error: An error if the query operation fails, or nil if successful.
func Query[T any](l *Limitless, queryRequest QueryRequest) ([]T, map[string]types.AttributeValue, error) {
	condition := queryRequest.Condition
	values := queryRequest.Values

	keyNameList := make([]string, 0)
	valueNameList := make([]string, 0)
	parsedCondition := ""
	for i := 0; i < len(condition); i = i + 1 {
		if condition[i] != ':' && condition[i] != '#' {
			parsedCondition += string(condition[i])
			continue
		} else {
			keyName := ""
			isKeyName := false
			if condition[i] == '#' {
				isKeyName = true
			}

			for j := i; j < len(condition) && condition[j] != ' '; j += 1 {
				keyName += string(condition[j])
				i += 1
			}
			if isKeyName {
				keyNameList = append(keyNameList, keyName)
			} else {
				valueNameList = append(valueNameList, keyName)
			}
			i = i - 1
		}
	}

	expressionAttributeNames := make(map[string]string)
	for i, k := range keyNameList {
		replacedKey := "#KEY" + fmt.Sprintf("%d", i)
		expressionAttributeNames[replacedKey] = strings.TrimPrefix(k, "#")
		condition = strings.ReplaceAll(condition, k, replacedKey)
	}
	ExpressionAttributeValues := make(map[string]any)
	for i, k := range valueNameList {
		replacedKey := ":VALUE" + fmt.Sprintf("%d", i)
		ExpressionAttributeValues[replacedKey] = values[k]
		condition = strings.ReplaceAll(condition, k, replacedKey)
	}

	valueMap, err := attributevalue.MarshalMap(ExpressionAttributeValues)
	if err != nil {
		return nil, nil, err
	}

	// Copy user send ExpressionAttributeNames
	for k, v := range queryRequest.ExpressionAttributeNames {
		expressionAttributeNames[k] = v
	}
	queryInput := new(dynamodb.QueryInput)
	queryInput.TableName = &queryRequest.TableName
	queryInput.KeyConditionExpression = &condition
	queryInput.ExpressionAttributeNames = expressionAttributeNames
	queryInput.ExpressionAttributeValues = valueMap
	queryInput.ConsistentRead = queryRequest.ConsistentRead
	queryInput.IndexName = queryRequest.IndexName
	queryInput.Limit = queryRequest.Limit
	queryInput.ScanIndexForward = queryRequest.Ascending
	queryInput.ExclusiveStartKey = queryRequest.ExclusiveStartKey
	queryInput.ProjectionExpression = queryRequest.ProjectionExpression

	result, err := l.ddb.Query(context.TODO(), queryInput)
	if err != nil {
		return nil, nil, err
	}

	res := make([]T, 0)
	for _, k := range result.Items {
		temp := new(T)
		err := attributevalue.UnmarshalMap(k, temp)
		if err != nil {
			return nil, nil, err
		}
		res = append(res, *temp)
	}
	return res, result.LastEvaluatedKey, nil
}

func TransactWriteItems(l *Limitless, items *[]PutItemRequest, idempotencyToken *string) error {
	transactItems := make([]types.TransactWriteItem, 0)
	for _, k := range *items {
		temp, err := attributevalue.MarshalMap(k.Item)
		if err != nil {
			return err
		}

		transactItems = append(transactItems, types.TransactWriteItem{
			Put: &types.Put{
				TableName: &k.TableName,
				Item:      temp,
			},
		})
	}

	_, err := l.ddb.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems:      transactItems,
		ClientRequestToken: idempotencyToken,
	})
	if err != nil {
		fmt.Println("Error = ", err)
	}
	return err
}

func UpdateItem[T any](l *Limitless, tableName string, key *T, item *T) error {
	value := reflect.ValueOf(*item)
	n := value.NumField()
	fields := make(map[string]string)
	values := make(map[string]any)
	updateExpressions := make([]string, 0)
	for i := range n {
		fieldName := strings.Split(value.Type().Field(i).Tag.Get("dynamodbav"), ",")[0]
		if value.Field(i).IsZero() {
			continue
		}
		fieldValue := value.Field(i).Interface()
		I := fmt.Sprintf("%d", i)

		fields["#KEY"+I] = fieldName
		values[":VALUE"+I] = fieldValue
		updateExpressions = append(updateExpressions, "#KEY"+I+" = "+":VALUE"+I)
	}
	updateExpression := "SET " + strings.Join(updateExpressions, ",")
	keyMap, _ := attributevalue.MarshalMap(*key)
	fieldValues, _ := attributevalue.MarshalMap(values)

	_, err := l.ddb.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName:                 &tableName,
		Key:                       keyMap,
		UpdateExpression:          &updateExpression,
		ExpressionAttributeNames:  fields,
		ExpressionAttributeValues: fieldValues,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// DeleteItem deletes a single item from a DynamoDB table based on the provided key.
//
// Parameters:
//   - l: A pointer to the Limitless client used to interact with DynamoDB.
//   - tableName: The name of the DynamoDB table from which to delete the item.
//   - key: A pointer to a struct of type T representing the primary key of the item to be deleted.
//
// Returns:
//   - error: An error if the deletion operation fails, or nil if successful.
func DeleteItem[T any](l *Limitless, tableName string, key *T) error {
	Key, err := attributevalue.MarshalMap(key)
	if err != nil {
		return err
	}

	_, err = l.ddb.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: &tableName,
		Key:       Key,
	})
	return err
}

// Scan performs a scan operation on a DynamoDB table using the provided scan request.
// It retrieves items from the table or a secondary index, and unmarshals the results into a slice of type T.
//
// Parameters:
//   - l: A pointer to the Limitless client.
//   - request: A ScanRequest struct containing the scan parameters.
//
// Returns:
//   - []T: A slice of type T containing the scan results.
//   - map[string]types.AttributeValue: The LastEvaluatedKey from the scan, which can be used for pagination. Is nil if there are no items left to scan.
//   - error: An error if the scan operation fails, or nil if successful.
func Scan[T any](l *Limitless, request ScanRequest) ([]T, map[string]types.AttributeValue, error) {
	resp, err := l.ddb.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName:         &request.TableName,
		IndexName:         request.IndexName,
		ConsistentRead:    request.ConsistentRead,
		Limit:             request.Limit,
		ExclusiveStartKey: request.ExclusiveStartKey,
	})
	if err != nil {
		return nil, nil, err
	}

	lek := resp.LastEvaluatedKey
	out := make([]T, 0)

	for _, v := range resp.Items {
		item := new(T)
		if attributevalue.UnmarshalMap(v, item) != nil {
			return nil, nil, err
		}
		out = append(out, *item)
	}

	return out, lek, nil
}
