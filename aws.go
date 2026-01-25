package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"mime"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

var _sesClient *sesv2.Client
var _s3Client *s3.Client
var _ctx *context.Context

func loadAws(ctx *context.Context) error {
	// aws
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.S3.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWS.AccessKeyID,
			cfg.AWS.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		log.Fatal("Error loading aws config", err)
		return err
	}

	_ctx = ctx
	_sesClient = sesv2.NewFromConfig(awsCfg)
	_s3Client = s3.NewFromConfig(awsCfg)

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

func uploadS3(key string, contentType string, data io.Reader) error {
	bucket := cfg.S3.Bucket

	// Read all data into buffer to get size
	buf := &bytes.Buffer{}
	size, err := io.Copy(buf, data)
	if err != nil {
		log.Println("Error reading data for S3 upload:", err)
		return err
	}

	_, err = _s3Client.PutObject(*_ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(buf.Bytes()),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})

	if err != nil {
		log.Println("Error uploading to S3:", err)
		return err
	}

	return nil
}

func deleteS3(key string) error {
	bucket := cfg.S3.Bucket

	_, err := _s3Client.DeleteObject(*_ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		log.Println("Error deleting from S3:", err)
		return err
	}

	return nil
}

func getS3PresignedURL(key string, expiration time.Duration) (string, error) {
	bucket := cfg.S3.Bucket

	presigner := s3.NewPresignClient(_s3Client)

	req, err := presigner.PresignGetObject(*_ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))

	if err != nil {
		log.Println("Error generating presigned URL:", err)
		return "", err
	}

	return req.URL, nil
}

// Helper function to upload image file with automatic content type detection
func uploadImageS3(key string, data io.Reader) error {
	// Detect content type from file extension
	ext := filepath.Ext(key)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream" // fallback
	}

	return uploadS3(key, contentType, data)
}
