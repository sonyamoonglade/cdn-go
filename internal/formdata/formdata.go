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
)

type UploadFile struct {
	Extension  string
	UploadName string
	MimeType   string
	Size       int64
	UUID       string
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

	return parsed, nil
}
