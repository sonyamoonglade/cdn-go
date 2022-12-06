package formdata

import (
	"mime/multipart"
	"strings"

	"animakuro/cdn/internal/fs"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrInvalidExtension = errors.New("file has invalid extension")
	ErrNoFiles          = errors.New("no files")
)

type UploadFile struct {
	// Generated from fs.DefaultName and Extension
	UploadName string
	Extension  string
	UUID       string
	MimeType   string
	Size       int64
	Open       func() (multipart.File, error)
}

func ParseFiles(form *multipart.Form) ([]*UploadFile, error) {

	fields := form.File
	parsed := make([]*UploadFile, 0, len(form.File))

	var spl []string
	for _, field := range fields {

		for _, v := range field {
			var upl UploadFile

			spl = strings.Split(v.Filename, ".")
			if len(spl) == 1 {
				return nil, ErrInvalidExtension
			}

			upl.Open = v.Open
			upl.UUID = uuid.New().String()
			upl.Extension = spl[len(spl)-1]
			upl.Size = v.Size
			upl.UploadName = fs.DefaultName + "." + upl.Extension
			upl.MimeType = v.Header.Get("Content-Type")
			parsed = append(parsed, &upl)
		}

	}

	// TODO: test
	if len(parsed) == 0 {
		return nil, ErrNoFiles
	}

	return parsed, nil
}
