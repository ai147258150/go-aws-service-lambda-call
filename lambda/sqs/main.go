package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"os"
)

const (
	Success = "Success"
	Error   = "Error"
	SCode   = 200
	ECode   = 400
)

type PostBody struct {
	MessageType  string 	 `json:"messageType"`
	EmailSubject string 	 `json:"email_subject"`
	Content      string 	 `json:"content"`
	Source       string 	 `json:"source"`
	Address      interface{} `json:"address"`
}

var svc *sqs.SQS

func init() {
	rand.Seed(time.Now().UnixNano())

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create an SQS service client
	svc = sqs.New(sess)
}

func GetRandomString(n int) string {
	randBytes := make([]byte, n/2)
	rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}

func Convert(array interface{}) string {
	return strings.Replace(strings.Trim(fmt.Sprint(array), "[]"), " ", ",", -1)
}

func SendMsg(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse,error) {

	randStr  := GetRandomString(10)

	queueURL := flag.String(randStr, os.Getenv("queueUrl"), "The url of the queue")
	flag.Parse()

	if *queueURL == "" {
		log.Printf("You must supply the url of a queue (-q QUEUE)")
		return events.APIGatewayProxyResponse{Body: Error, StatusCode: ECode}, nil
	}

	postData := request.Body

	var Address,EmailSubject,Source,MessageType,Content = "","no-subject","","",""

	// json str to map
	var dat PostBody

	if err := json.Unmarshal([]byte(postData), &dat); err == nil {
		Address 	= Convert(dat.Address)
		Source 		= dat.Source
		MessageType = dat.MessageType
		Content 	= dat.Content

		if dat.EmailSubject != "" {
			EmailSubject = dat.EmailSubject
		}
	}

	_, err := svc.SendMessage(&sqs.SendMessageInput{
		DelaySeconds: aws.Int64(10),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"address": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(Address),
			},
			"messageType": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(MessageType),
			},
			"email_subject": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(EmailSubject),
			},
			"source": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(Source),
			},
		},
		MessageBody: aws.String(Content),
		QueueUrl:    queueURL,
	})

	if err != nil {
		log.Printf("Got an error sending the message: %s", err)
		return events.APIGatewayProxyResponse{Body: Error, StatusCode: ECode}, nil
	}

	log.Printf("Sent message to queue")
	return events.APIGatewayProxyResponse{Body: Success, StatusCode: SCode}, nil
}

func main(){
	lambda.Start(SendMsg)
}