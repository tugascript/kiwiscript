package objstg

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/kiwiscript/kiwiscript_go/utils"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"time"
)

type ObjectStorage struct {
	putClient *s3.Client
	getClient *s3.PresignClient
	bucket    string
	log       *slog.Logger
}

func NewObjectStorage(
	log *slog.Logger,
	s3Client *s3.Client,
	bucket string,
) *ObjectStorage {
	return &ObjectStorage{
		putClient: s3Client,
		getClient: s3.NewPresignClient(s3Client),
		bucket:    bucket,
		log:       log,
	}
}

func readMimeType(f multipart.File) (string, error) {
	mtype, err := mimetype.DetectReader(f)
	if err != nil {
		return "", err
	}
	return mtype.String(), nil
}

func readMagicBytes(f multipart.File) (string, error) {
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
	if _, err := o.putClient.PutObject(ctx, &input); err != nil {
		return uuid.UUID{}, err
	}

	return fileId, nil
}

func (o *ObjectStorage) DeleteFile(ctx context.Context, userId int32, fileId uuid.UUID, fileExt string) error {
	input := s3.DeleteObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(makeKey(userId, fileId, fileExt)),
	}
	_, err := o.putClient.DeleteObject(ctx, &input)
	return err
}

func (o *ObjectStorage) closeFile(f io.Closer) {
	if err := f.Close(); err != nil {
		o.log.Error("Failed to close file", "error", err)
	}
}

func (o *ObjectStorage) buildLogger(requestID, function string) *slog.Logger {
	return utils.BuildLogger(o.log, utils.LoggerOptions{
		Layer:     utils.ProvidersLogLayer,
		Location:  "object_storage",
		Function:  function,
		RequestID: requestID,
	})
}

type GetFileURLOptions struct {
	RequestID string
	UserID    int32
	FileID    uuid.UUID
	FileExt   string
}

func (o *ObjectStorage) GetFileUrl(ctx context.Context, opts GetFileURLOptions) (string, error) {
	log := o.buildLogger(opts.RequestID, "GetFileUrl").With(
		"userID", opts.UserID,
		"fileID", opts.FileID.String(),
		"fileExt", opts.FileExt,
	)
	log.DebugContext(ctx, "Getting file URL...")

	req, err := o.getClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(o.bucket),
			Key:    aws.String(makeKey(opts.UserID, opts.FileID, opts.FileExt)),
		},
		s3.WithPresignExpires(time.Hour*24),
	)
	if err != nil {
		log.ErrorContext(ctx, "Error getting file URL", "error", err)
		return "", err
	}

	return req.URL, nil
}
