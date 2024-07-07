package obj_stg

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type ObjectStorage struct {
	s3Client *s3.Client
	bucket   string
}

func NewObjectStorage(s3Client *s3.Client, bucket string) *ObjectStorage {
	return &ObjectStorage{
		s3Client: s3Client,
		bucket:   bucket,
	}
}

func readMimeType(f multipart.File) (string, error) {
	buf := make([]byte, 512)
	_, err := f.Read(buf)
	if err != nil {
		return "", err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return "", err
	}
	return http.DetectContentType(buf), nil
}

func makeKey(userId int32, fileId uuid.UUID, fileExt string) string {
	return fmt.Sprintf("%d/%s.%s", userId, fileId.String(), fileExt)
}

func (o *ObjectStorage) uploadFile(ctx context.Context, userId int32, ext string, r io.Reader) (uuid.UUID, error) {
	fileId, err := uuid.NewRandom()
	if err != nil {
		return uuid.UUID{}, err
	}

	input := s3.PutObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(makeKey(userId, fileId, ext)),
		Body:   r,
	}
	if _, err := o.s3Client.PutObject(ctx, &input); err != nil {
		return uuid.UUID{}, err
	}

	return fileId, nil
}

func (o *ObjectStorage) DeleteFile(ctx context.Context, userId int32, fileId uuid.UUID, fileExt string) error {
	input := s3.DeleteObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(makeKey(userId, fileId, fileExt)),
	}
	_, err := o.s3Client.DeleteObject(ctx, &input)
	return err
}
