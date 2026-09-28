package storage

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	s3 *s3.Client
}

func NewClient(region string, accessKey string, secretKey string, endpoint string, usePathStyle bool) *Client {
	cfg := aws.Config{
		Region:       region,
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		BaseEndpoint: aws.String(endpoint),
	}

	return &Client{
		s3: s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.UsePathStyle = usePathStyle
		}),
	}
}

func (c *Client) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	getObjInput := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	getObjOutput, err := c.s3.GetObject(ctx, getObjInput)
	if err != nil {
		return nil, err
	}
	defer getObjOutput.Body.Close()

	body, err := io.ReadAll(getObjOutput.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (c *Client) PutObject(ctx context.Context, bucket, key string, body []byte, contentType string) error {
	_, err := c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	return err
}
