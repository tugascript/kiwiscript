package objstg

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"mime/multipart"

	"github.com/google/uuid"
)

const maxSize = 256 * 1024

const imageExt = "jpeg"

var qualities = [10]int{100, 90, 80, 70, 60, 50, 40, 30, 20, 10}

const (
	pngMime  string = "image/png"
	jpegMime string = "image/jpeg"
)

func valImgMime(mimeType string) bool {
	return mimeType == pngMime || mimeType == jpegMime
}

func compressImage(data image.Image) (bytes.Buffer, error) {
	var jpegImage bytes.Buffer

	for _, quality := range qualities {
		jpegImage.Reset()
		err := jpeg.Encode(&jpegImage, data, &jpeg.Options{Quality: quality})
		if err != nil {
			return bytes.Buffer{}, err
		}
		if jpegImage.Len() < maxSize {
			break
		}
	}

	return jpegImage, nil
}

func decodeImage(f multipart.File) (image.Image, error) {
	img, format, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	switch format {
	case "jpeg", "png":
		return img, nil
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}
}

func (o *ObjectStorage) UploadImage(ctx context.Context, userId int32, f multipart.File) (uuid.UUID, string, error) {
	fileId := uuid.UUID{}
	if mimeType, err := readMimeType(f); err != nil || !valImgMime(mimeType) {
		return fileId, "", fmt.Errorf("mime type not supported")
	}

	img, err := decodeImage(f)
	if err != nil {
		return fileId, "", err
	}

	compressedImg, err := compressImage(img)
	if err != nil {
		return fileId, "", err
	}

	fileId, err = o.uploadFile(ctx, userId, imageExt, bytes.NewReader(compressedImg.Bytes()))
	if err != nil {
		return uuid.UUID{}, "", err
	}

	return fileId, imageExt, nil
}
