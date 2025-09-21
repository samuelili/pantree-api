package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

var _sesClient *sesv2.Client
var _ctx *context.Context

func loadAws(ctx *context.Context) error {
	// aws
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Error loading aws config", err)
		return err
	}

	_ctx = ctx
	_sesClient = sesv2.NewFromConfig(cfg)

	return nil
}

func sendEmail(to string, subject string, body string) error {
	from := cfg.Email.From
	arn := cfg.Email.ARN

	var params *sesv2.SendEmailInput = &sesv2.SendEmailInput{
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data: &subject,
				},
				Body: &types.Body{
					Text: &types.Content{
						Data: &body,
					},
				},
			},
		},
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		FromEmailAddressIdentityArn: &arn,
		FromEmailAddress:            &from,
	}

	_, err := _sesClient.SendEmail(ctx, params)
	if err != nil {
		log.Println("could not send email", err)
		return err
	}

	return nil
}
