package clients

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Config holds everything needed to build an S3-compatible client.
// MinIO and LocalStack are S3-compatible: set Endpoint to their S3 API base
// URL (e.g. http://localhost:9000) and provide the access/secret keys.
// Leave Endpoint empty to talk to real AWS S3.
type S3Config struct {
	Region         string
	Bucket         string
	Endpoint       string
	AccessKey      string
	SecretKey      string
	ForcePathStyle bool
}

type S3Client struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}


func NewS3Client(ctx context.Context, region, bucket string) (*S3Client, error) {
	if bucket == "" {
		return nil, fmt.Errorf("s3 bucket name is required but was empty")
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	loadOpts := []func(*config.LoadOptions) error{config.WithRegion(cfg.Region)}
	if cfg.Endpoint != "" {
		// S3-compatible endpoint (MinIO/LocalStack): use static credentials.
		loadOpts = append(loadOpts,
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		if cfg.ForcePathStyle {
			o.UsePathStyle = true
		}
	})

	return &S3Client{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bucket:        cfg.Bucket,
	}, nil
}

func (c *S3Client) Key(userID, fileID string) string {
	return fmt.Sprintf("resumes/%s/%s", userID, fileID)
}

// Put writes raw bytes to the bucket under key. Used for server-side uploads
// where the client sends the file to the API rather than directly to storage.
func (c *S3Client) Put(ctx context.Context, key string, body []byte, contentType string) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("uploading key %s: %w", key, err)
	}
	return nil
}

func (c *S3Client) PresignUpload(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	request, err := c.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("presigning upload: %w", err)
	}
	return request.URL, nil
}

func (c *S3Client) PresignDownload(ctx context.Context, key string, expiry time.Duration) (string, error) {
	request, err := c.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("presigning download: %w", err)
	}
	return request.URL, nil
}

func (c *S3Client) Delete(ctx context.Context, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("deleting object: %w", err)
	}
	return nil
}
<<<<<<< HEAD
=======

func (c *S3Client) Download(ctx context.Context, key string) ([]byte, error) {
	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("downloading object: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("reading downloaded content: %w", err)
	}
	return data, nil
}
func (c *S3Client) Ping(ctx context.Context) error {
	_, err := c.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(c.bucket),
	})
	if err != nil {
		return fmt.Errorf("checking bucket availability: %w", err)
	}
	return nil
}
>>>>>>> dev
