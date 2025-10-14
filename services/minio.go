package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	appcontext "github.com/cloakd/common/context"
	serviceContext "github.com/cloakd/common/services"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

type MinIOService struct {
	serviceContext.DefaultService
	client     *minio.Client
	bucketName string
	endpoint   string
	accessKey  string
	secretKey  string
	useSSL     bool
}

const MINIO_SVC = "minio_svc"

func (svc MinIOService) Id() string {
	return MINIO_SVC
}

func (svc *MinIOService) Configure(ctx *appcontext.Context) error {
	svc.endpoint = os.Getenv("MINIO_ENDPOINT")
	if svc.endpoint == "" {
		svc.endpoint = "localhost:9000"
	}

	svc.accessKey = os.Getenv("MINIO_ACCESS_KEY")
	if svc.accessKey == "" {
		svc.accessKey = "admin"
	}

	svc.secretKey = os.Getenv("MINIO_SECRET_KEY")
	if svc.secretKey == "" {
		svc.secretKey = "password123"
	}

	svc.useSSL = os.Getenv("MINIO_USE_SSL") == "true"

	svc.bucketName = os.Getenv("MINIO_BUCKET_NAME")
	if svc.bucketName == "" {
		svc.bucketName = "ven-learning"
	}

	return svc.DefaultService.Configure(ctx)
}

func (svc *MinIOService) Start() error {
	client, err := minio.New(svc.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(svc.accessKey, svc.secretKey, ""),
		Secure: svc.useSSL,
	})
	if err != nil {
		return fmt.Errorf("failed to create MinIO client: %v", err)
	}

	svc.client = client

	if err := svc.ensureBucket(); err != nil {
		return fmt.Errorf("failed to ensure bucket exists: %v", err)
	}

	log.Printf("MinIO service started successfully with endpoint: %s", svc.endpoint)
	return nil
}

func (svc *MinIOService) ensureBucket() error {
	ctx := context.Background()

	exists, err := svc.client.BucketExists(ctx, svc.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %v", err)
	}

	if !exists {
		err = svc.client.MakeBucket(ctx, svc.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %v", err)
		}
		log.Printf("Created MinIO bucket: %s", svc.bucketName)
	}

	return nil
}

func (svc *MinIOService) UploadFile(objectName string, reader io.Reader, objectSize int64, contentType string) (*minio.UploadInfo, error) {
	ctx := context.Background()

	uploadInfo, err := svc.client.PutObject(ctx, svc.bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to MinIO: %v", err)
	}

	return &uploadInfo, nil
}

func (svc *MinIOService) GetFileURL(objectName string, expiry time.Duration) (string, error) {
	ctx := context.Background()

	presignedURL, err := svc.client.PresignedGetObject(ctx, svc.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return presignedURL.String(), nil
}

func (svc *MinIOService) DeleteFile(objectName string) error {
	ctx := context.Background()

	err := svc.client.RemoveObject(ctx, svc.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file from MinIO: %v", err)
	}

	return nil
}

func (svc *MinIOService) GetFileInfo(objectName string) (*minio.ObjectInfo, error) {
	ctx := context.Background()

	objInfo, err := svc.client.StatObject(ctx, svc.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	return &objInfo, nil
}

func (svc *MinIOService) ListFiles(prefix string) ([]minio.ObjectInfo, error) {
	ctx := context.Background()

	var objects []minio.ObjectInfo
	objectCh := svc.client.ListObjects(ctx, svc.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %v", object.Err)
		}
		objects = append(objects, object)
	}

	return objects, nil
}

func (svc *MinIOService) GetBucketName() string {
	return svc.bucketName
}
