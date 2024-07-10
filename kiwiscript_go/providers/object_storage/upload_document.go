package obj_stg

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/google/uuid"
)

const (
	pdfMime     string = "application/pdf"
	docMime     string = "application/msword"
	docxMime    string = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	openDocMime string = "application/vnd.oasis.opendocument.text"
	zipMime     string = "application/zip"

	pdfExt     string = "pdf"
	docExt     string = "doc"
	docxExt    string = "docx"
	openDocExt string = "odt"
	zipExt     string = "zip"
)

func valDocMime(mimeType string) bool {
	switch mimeType {
	case pdfMime, docMime, docxMime, openDocMime, zipMime:
		return true
	default:
		return false
	}
}

func selectDocExt(mimeType string) string {
	switch mimeType {
	case pdfMime:
		return pdfExt
	case docMime:
		return docExt
	case docxMime:
		return docxExt
	case openDocMime:
		return openDocExt
	case zipMime:
		return zipExt
	default:
		return "txt"
	}
}

func (o *ObjectStorage) UploadDocument(ctx context.Context, userId int32, f multipart.File) (uuid.UUID, string, error) {
	fileId := uuid.UUID{}
	mimeType, err := readMimeType(f)
	if err != nil {
		return fileId, "", err
	}
	if !valDocMime(mimeType) {
		return fileId, "", fmt.Errorf("mime type not supported")
	}

	docExt := selectDocExt(mimeType)
	fileId, err = o.uploadFile(ctx, userId, docExt, f)
	if err != nil {
		return uuid.UUID{}, "", err
	}

	return fileId, docExt, err
}