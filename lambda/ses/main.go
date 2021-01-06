package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
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

var fromEmail = "test@email.com"
var sesClient *ses.SES

func init() {
	sesClient = ses.New(session.New(), aws.NewConfig().WithRegion("eu-central-1"))
}

func Convert(array interface{}) string {
	return strings.Replace(strings.Trim(fmt.Sprint(array), "[]"), " ", ",", -1)
}

func SendEmail(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse,error) {

	postData := request.Body

	var EmailSubject = "no-subject"

	// json str to map
	var dat PostBody

	json.Unmarshal([]byte(postData), &dat)

	if dat.EmailSubject != "" {
		EmailSubject = dat.EmailSubject
	}

	emailParams := &ses.SendEmailInput{
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Data:aws.String(dat.Content),
				},
			},
			Subject: &ses.Content{
				Data:aws.String(EmailSubject),
			},
		},
		Destination: &ses.Destination{
			ToAddresses:[]*string{aws.String(Convert(dat.Address))},
		},
		Source:aws.String(fromEmail),
	}

	_, err := sesClient.SendEmail(emailParams)

	if err != nil {
		log.Printf("Failed to send mail")
		return events.APIGatewayProxyResponse{Body: Error, StatusCode: ECode}, nil
	}

	log.Printf("Mail sent successfully")
	return events.APIGatewayProxyResponse{Body: Success, StatusCode: SCode}, nil
}

func main(){
	lambda.Start(SendEmail)
}