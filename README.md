# Limitless

A thin wrapper around AWS DynamoDB. Uses AWS Go SDK V2.

Makes life easier and less verbose. Somewhat battle-tested. Removes need to manually Marshal/Unmarshall to/from AttributeValues. Uses the "dynamodb" tag for serialization and deserialization.

How to use:

```
import(
    "github.com/0Xero7/limitless"
    "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type User struct {
    UserId string `dynamodbav:"user_id"`
    Name   string `dynamodbav:"name,omitempty"`
}

func main() {
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        return err
    }
    limitlessClient := limitless.NewClient(dynamodb.NewFromConfig(cfg))

    // Scan
	results, lastEvaluatedKey, err := limitless.Scan[User](limitlessClient, ScanRequest{
		TableName: "users",
	})
}

```