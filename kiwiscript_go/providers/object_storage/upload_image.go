package objstg

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"mime/multipart"

	"github.com/google/uuid"
)

const imageExt = "jpeg"

const (
	pngMime  string = "image/png"
	jpegMime string = "image/jpeg"
)

func valImgMime(mimeType string) bool {
	return mimeType == pngMime || mimeType == jpegMime
}

func compressImage(data image.Image) (bytes.Buffer, error) {
	var jpegImage bytes.Buffer

	if err := jpeg.Encode(&jpegImage, data, &jpeg.Options{Quality: 75}); err != nil {
		return bytes.Buffer{}, err
	}

	return jpegImage, nil
}

func decodeImage(f multipart.File) (image.Image, error) {
	img, format, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	switch format {
	case "jpeg", "png":
		return img, nil
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}
}

type UploadImageOptions struct {
	RequestID string
	UserID    int32
	FH        *multipart.FileHeader
}

func (o *ObjectStorage) UploadImage(ctx context.Context, opts UploadImageOptions) (uuid.UUID, string, error) {
	log := o.buildLogger(opts.RequestID, "UploadImage").With("userId", opts.UserID)
	log.DebugContext(ctx, "Uploading image...")

	f, err := opts.FH.Open()
	if err != nil {
		log.ErrorContext(ctx, "Error opening file", "error", err)
		return uuid.UUID{}, "", fmt.Errorf("error opening file")
	}
	defer o.closeFile(f)

	if mimeType, err := readMagicBytes(f); err != nil || !valImgMime(mimeType) {
		return uuid.UUID{}, "", fmt.Errorf("mime type not supported")
	}

	img, err := decodeImage(f)
	if err != nil {
		log.ErrorContext(ctx, "Error decoding image", "error", err)
		return uuid.UUID{}, "", err
	}

	compressedImg, err := compressImage(img)
	if err != nil {
		log.ErrorContext(ctx, "Error compressing image", "error", err)
		return uuid.UUID{}, "", err
	}

	fileId, err := o.uploadFile(ctx, opts.UserID, imageExt, bytes.NewReader(compressedImg.Bytes()))
	if err != nil {
		log.ErrorContext(ctx, "Error uploading image", "error", err)
		return uuid.UUID{}, "", err
	}

	return fileId, imageExt, nil
}
