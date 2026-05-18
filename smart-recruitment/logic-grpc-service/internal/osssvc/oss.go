package osssvc

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"recruitment/logic-grpc-service/internal/config"
)

type Client struct {
	api    *s3.Client
	bucket string
}

func New(ctx context.Context, cfg *config.Config) (*Client, error) {
	if cfg.OSSBucket == "" || cfg.OSSAccessKey == "" || cfg.OSSSecretKey == "" {
		return nil, fmt.Errorf("OSS_BUCKET, OSS_ACCESS_KEY_ID, OSS_SECRET_ACCESS_KEY required")
	}
	creds := credentials.NewStaticCredentialsProvider(cfg.OSSAccessKey, cfg.OSSSecretKey, "")
	var api *s3.Client
	if cfg.OSSEndpoint != "" {
		pathStyle := usePathStyleForEndpoint(cfg.OSSEndpoint)
		resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               cfg.OSSEndpoint,
				HostnameImmutable: pathStyle, // 阿里云 virtual-hosted 必须为 false，否则 host 锁在 endpoint 上，仍会生成 path-style URL
				SigningRegion:     cfg.OSSRegion,
			}, nil
		})
		awsCfg := aws.Config{
			Region:                      cfg.OSSRegion,
			Credentials:                 creds,
			EndpointResolverWithOptions: resolver,
		}
		api = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			// 阿里云 OSS 多地域禁止 path-style（endpoint/bucket/key），会 403 SecondLevelDomainForbidden；
			// 须 virtual-hosted（bucket.endpoint/key）。MinIO 等本地网关仍用 path-style。
			o.UsePathStyle = pathStyle
		})
	} else {
		lc, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(cfg.OSSRegion),
			awsconfig.WithCredentialsProvider(creds),
		)
		if err != nil {
			return nil, err
		}
		api = s3.NewFromConfig(lc)
	}
	return &Client{api: api, bucket: cfg.OSSBucket}, nil
}

func usePathStyleForEndpoint(endpoint string) bool {
	e := strings.ToLower(strings.TrimSpace(endpoint))
	if strings.Contains(e, "aliyuncs.com") {
		return false
	}
	return true
}

func (c *Client) PresignPut(ctx context.Context, objectKey, contentType string, expires time.Duration) (url string, headers map[string]string, err error) {
	ps := s3.NewPresignClient(c.api)
	out, err := ps.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(objectKey),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", nil, err
	}
	h := make(map[string]string)
	for k, v := range out.SignedHeader {
		if len(v) > 0 {
			h[k] = v[0]
		}
	}
	return out.URL, h, nil
}

func (c *Client) PresignGet(ctx context.Context, objectKey string, expires time.Duration) (string, error) {
	ps := s3.NewPresignClient(c.api)
	out, err := ps.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func (c *Client) GetObjectHead(ctx context.Context, objectKey string, max int64) ([]byte, error) {
	rng := fmt.Sprintf("bytes=0-%d", max-1)
	out, err := c.api.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(objectKey),
		Range:  aws.String(rng),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	return io.ReadAll(io.LimitReader(out.Body, max))
}

func (c *Client) GetObjectBytes(ctx context.Context, objectKey string, max int64) ([]byte, error) {
	out, err := c.api.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	return io.ReadAll(io.LimitReader(out.Body, max+1))
}

func SafeResumeObjectKey(userID int64, fileName string) string {
	base := filepath.Base(fileName)
	base = strings.ReplaceAll(base, "..", "")
	if base == "" || base == "." {
		base = "resume.bin"
	}
	return fmt.Sprintf("resumes/%d/%d-%s", userID, time.Now().UnixNano(), base)
}
