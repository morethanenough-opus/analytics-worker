package analytics_worker

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// DownloadS3Object downloads an object from S3
func DownloadS3Object(sess *session.Session, bucket, key string) (string, error) {
	s3Client := s3.New(sess)
	req, err := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err!= nil {
		return "", err
	}

	fileName := filepath.Base(key)
	tmpDir, err := os.MkdirTemp("", "tmp")
	if err!= nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, fileName)
	f, err := os.Create(filePath)
	if err!= nil {
		return "", err
	}
	defer f.Close()

	log.Printf("Downloading %s from S3 to %s\n", key, filePath)

	_, err = req.WriteTo(f)
	if err!= nil {
		return "", err
	}

	return filePath, nil
}

// UploadS3Object uploads an object to S3
func UploadS3Object(sess *session.Session, bucket, key string, filePath string) error {
	s3Client := s3.New(sess)
	file, err := os.Open(filePath)
	if err!= nil {
		return err
	}
	defer file.Close()

	log.Printf("Uploading %s to S3\n", filePath)

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err!= nil {
		return err
	}

	return nil
}

// ExtractZip extracts a zip file
func ExtractZip(filePath string, destDir string) error {
	f, err := os.Open(filePath)
	if err!= nil {
		return err
	}
	defer f.Close()

	r, err := zip.NewReader(f, f.Size())
	if err!= nil {
		return err
	}

	for _, f := range r.File {
		rc, err := f.Open()
		if err!= nil {
			return err
		}
		defer rc.Close()

		destPath := filepath.Join(destDir, f.Name)
		if!strings.HasSuffix(destPath, "/") {
			os.MkdirAll(filepath.Dir(destPath), 0755)
		}

		if err := copyFile(rc, destPath); err!= nil {
			return err
		}
	}

	return nil
}

func copyFile(src *zip.Reader, dst string) error {
	f, err := os.Create(dst)
	if err!= nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, src)
	if err!= nil {
		return err
	}

	return nil
}

// GetS3ObjectMetadata retrieves metadata of an S3 object
func GetS3ObjectMetadata(sess *session.Session, bucket, key string) (*s3.GetObjectOutput, error) {
	s3Client := s3.New(sess)

	req, err := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err!= nil {
		return nil, err
	}

	return req.Send()
}

// GetS3BucketList retrieves a list of S3 buckets
func GetS3BucketList(sess *session.Session) ([]string, error) {
	s3Client := s3.New(sess)

	input := &s3.ListBucketsInput{}
	output, err := s3Client.ListBuckets(input)
	if err!= nil {
		return nil, err
	}

	buckets := make([]string, len(output.Buckets))
	for i, bucket := range output.Buckets {
		buckets[i] = *bucket.Name
	}

	return buckets, nil
}

func GetS3ObjectURL(sess *session.Session, bucket, key string) (string, error) {
	s3Client := s3.New(sess)

	req, err := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err!= nil {
		return "", err
	}

	return req.Presign(10 * time.Minute)
}

func GetS3ObjectMimeType(sess *session.Session, bucket, key string) (string, error) {
	s3Client := s3.New(sess)

	req, err := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err!= nil {
		return "", err
	}

	return req.Presign(10 * time.Minute)
}