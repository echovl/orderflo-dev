package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/url"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/layerhub-io/api/errors"
	"github.com/layerhub-io/api/upload"
)

type S3Uploader struct {
	client  *s3.Client
	bucket  string
	cdnBase string
}

func New(region string, bucket string, cdnBase string) (upload.SignedUploader, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	return &S3Uploader{
		client:  s3.NewFromConfig(cfg),
		bucket:  bucket,
		cdnBase: cdnBase,
	}, nil
}

func (s *S3Uploader) Upload(ctx context.Context, key string, data []byte) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(mime.TypeByExtension(path.Ext(key))),
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", errors.E(errors.KindUnexpected, err)
	}

	var uri string
	if s.cdnBase != "" {
		uri, err = joinURLs(s.cdnBase, key)
		if err != nil {
			return "", err
		}
	} else {
		uri = fmt.Sprintf("https://%v.s3.amazonaws.com/%v", s.bucket, key)
	}

	return uri, nil
}

func (s *S3Uploader) Download(ctx context.Context, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	out, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}
	defer out.Body.Close()

	return io.ReadAll(out.Body)
}

func (s *S3Uploader) GetPresignedURL(ctx context.Context, key string) (string, error) {
	pClient := s3.NewPresignClient(s.client)

	params := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(mime.TypeByExtension(path.Ext(key))),
	}

	res, err := pClient.PresignPutObject(ctx, params)
	if err != nil {
		return "", errors.E(errors.KindUnexpected, err)
	}

	return res.URL, nil
}

func joinURLs(base string, elem ...string) (result string, err error) {
	url, err := url.Parse(base)
	if err != nil {
		return
	}

	for _, e := range elem {
		url, err = url.Parse(e)
		if err != nil {
			return
		}
	}
	result = url.String()
	return
}
