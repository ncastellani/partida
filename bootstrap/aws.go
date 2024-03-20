package bootstrap

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
)

// instantiate an AWS session
var AWSSession, _ = session.NewSession(&aws.Config{
	Region: aws.String(regionToAWSRegion[os.Getenv("APP_REGION")]),
	Credentials: credentials.NewStaticCredentials(
		os.Getenv("AWS_IAM_ID"),
		os.Getenv("AWS_IAM_KEY"),
		"",
	),
})

// create a session for DynamoDB
var DB = dynamo.New(AWSSession, &aws.Config{Region: aws.String(regionToAWSRegion[os.Getenv("APP_REGION")])})

// create a session for SQS
var QUEUE = sqs.New(AWSSession, &aws.Config{Region: aws.String(os.Getenv("SQS_AWS_REGION"))})

// update the opened AWS connections and the AWS session
func UpdateAWSConnections() {
	AWSSession, _ = session.NewSession(&aws.Config{
		Region: aws.String(regionToAWSRegion[os.Getenv("APP_REGION")]),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_IAM_ID"),
			os.Getenv("AWS_IAM_KEY"),
			"",
		),
	})

	DB = dynamo.New(AWSSession, &aws.Config{Region: aws.String(regionToAWSRegion[os.Getenv("APP_REGION")])})
	QUEUE = sqs.New(AWSSession, &aws.Config{Region: aws.String(os.Getenv("SQS_AWS_REGION"))})
}

// add an event into the SQS queue
func AppendToQueue(l *log.Logger, key string, priority int, category string, data map[string]interface{}) {

	// marshal the JSON to send
	body, _ := sonic.MarshalString(data)

	// assemble the deduplication ID
	dedup := fmt.Sprintf(
		"%v.%v%v",
		key,
		uuid.NewString(),
		uuid.NewString(),
	)

	// send to the queue
	out, err := QUEUE.SendMessage(&sqs.SendMessageInput{
		QueueUrl:               aws.String(os.Getenv("SQS_SERVICE_QUEUE")),
		MessageAttributes:      map[string]*sqs.MessageAttributeValue{"Category": {DataType: aws.String("String"), StringValue: &category}},
		MessageBody:            &body,
		MessageDeduplicationId: &dedup,
		MessageGroupId:         aws.String(fmt.Sprintf("%v-%v", key, priority)),
	})

	// panic in case of put queue error
	if err != nil {
		panic(err)
	}

	l.Printf("sucessfully added an event into the SQS queue [id: %v]", &out.MessageId)

}

// return an AWS session created upon the passed parameters
func GetAWSSession(endpoint, region, token, key string) (sess *session.Session, err error) {
	return session.NewSession(&aws.Config{
		Endpoint:    aws.String(endpoint),
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(token, key, ""),
	})
}
