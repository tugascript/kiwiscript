package objstg

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

type UploadDocumentOptions struct {
	RequestID string
	UserId    int32
	FH        *multipart.FileHeader
}

func (o *ObjectStorage) UploadDocument(ctx context.Context, opts UploadDocumentOptions) (uuid.UUID, string, error) {
	log := o.buildLogger(opts.RequestID, "UploadDocument").With("userId", opts.UserId)
	log.InfoContext(ctx, "Uploading document...")

	f, err := opts.FH.Open()
	if err != nil {
		log.ErrorContext(ctx, "Error opening file", "error", err)
		return uuid.UUID{}, "", fmt.Errorf("error opening file")
	}
	defer o.closeFile(f)

	mimeType, err := readMimeType(f)
	if err != nil {
		return uuid.UUID{}, "", err
	}
	if !valDocMime(mimeType) {
		return uuid.UUID{}, "", fmt.Errorf("mime type not supported")
	}

	docExt := selectDocExt(mimeType)
	fileId, err := o.uploadFile(ctx, opts.UserId, docExt, f)
	if err != nil {
		return uuid.UUID{}, "", err
	}

	return fileId, docExt, err
}
