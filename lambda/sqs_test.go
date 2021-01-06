package main

import (
	"flag"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type MyEvent struct {
	Address string `json:"address"`
	MessageType string `json:"messageType"`
	Content string `json:"content"`
	EmailSubject string `json:"email_subject"`
	Source string `json:"source"`
}

var queueName = "testqueue"

func GetQueueURL(sess *session.Session, queue *string) (*sqs.GetQueueUrlOutput, error) {
	// Create an SQS service client
	svc := sqs.New(sess)

	result, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: queue,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func SendMsg(sess *session.Session, queueURL *string, address string, EmailSubject string, Source string, messageType string, content string) error {
	// Create an SQS service client
	svc := sqs.New(sess)

	_, err := svc.SendMessage(&sqs.SendMessageInput{
		DelaySeconds: aws.Int64(10),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"address": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(address),
			},
			"messageType": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(messageType),
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
		MessageBody: aws.String(content),
		QueueUrl:    queueURL,
	})

	if err != nil {
		return err
	}

	return nil
}

func handler(data MyEvent) (string, error) {
	// Convert queue name type
	queue := flag.String("q", queueName, "The name of the queue")
	flag.Parse()

	if *queue == "" {
		return fmt.Sprintf("You must supply the name of a queue (-q QUEUE)"),nil
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Get URL of queue
	result, err := GetQueueURL(sess, queue)
	if err != nil {
		return fmt.Sprintf("Got an error getting the queue URL: %s", err),nil
	}

	queueURL := result.QueueUrl

	err = SendMsg(sess, queueURL, data.Address, data.EmailSubject, data.Source, data.MessageType, data.Content)
	if err != nil {
		return fmt.Sprintf("Got an error sending the message: %s", err), nil
	}

	return fmt.Sprintf("Sent message to queue"), nil
}

func main(){
	lambda.Start(handler)
}
