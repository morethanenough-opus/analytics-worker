Here's the refactored and improved version of the code:

```go
package analytics_worker

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	tempDirPrefix = "analytics-worker"
	defaultPresignDuration = 10 * time.Minute
)

// DownloadS3Object downloads an object from S3 using downloader for better performance
func DownloadS3Object(sess *session.Session, bucket, key string) (string, error) {
	downloader := s3manager.NewDownloader(sess)
	
	tmpDir, err := os.MkdirTemp("", tempDirPrefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	fileName := filepath.Base(key)
	filePath := filepath.Join(tmpDir, fileName)
	
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	log.Printf("Downloading %s from S3 to %s", key, filePath)

	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}

	return filePath, nil
}

// UploadS3Object uploads an object to S3 using uploader for better performance
func UploadS3Object(sess *session.Session, bucket, key, filePath string) error {
	uploader := s3manager.NewUploader(sess)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	log.Printf("Uploading %s to S3 bucket %s as %s", filePath, bucket, key)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// ExtractZip extracts a zip file to destination directory
func ExtractZip(filePath, destDir string) error {
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		destPath := filepath.Join(destDir, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open zip entry: %w", err)
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			srcFile.Close()
			return fmt.Errorf("failed to create destination file: %w", err)
		}

		if _, err := io.Copy(destFile, srcFile); err != nil {
			srcFile.Close()
			destFile.Close()
			return fmt.Errorf("failed to copy file content: %w", err)
		}

		srcFile.Close()
		destFile.Close()
	}

	return nil
}

// GetS3ObjectMetadata retrieves metadata of an S3 object
func GetS3ObjectMetadata(sess *session.Session, bucket, key string) (*s3.HeadObjectOutput, error) {
	svc := s3.New(sess)
	return svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
}

// GetS3BucketList retrieves a list of S3 buckets
func GetS3BucketList(sess *session.Session) ([]string, error) {
	svc := s3.New(sess)
	result, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	buckets := make([]string, len(result.Buckets))
	for i, b := range result.Buckets {
		buckets[i] = aws.StringValue(b.Name)
	}

	return buckets, nil
}

// GetS3ObjectURL generates a presigned URL for an S3 object
func GetS3ObjectURL(sess *session.Session, bucket, key string) (string, error) {
	svc := s3.New(sess)
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return req.Presign(defaultPresignDuration)
}

// GetS3ObjectMimeType retrieves the MIME type of an S3 object
func GetS3ObjectMimeType(sess *session.Session, bucket, key string) (string, error) {
	metadata, err := GetS3ObjectMetadata(sess, bucket, key)
	if err != nil {
		return "", fmt.Errorf("failed to get object metadata: %w", err)
	}
	return aws.StringValue(metadata.ContentType), nil
}
```