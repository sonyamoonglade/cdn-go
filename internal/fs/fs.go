package fs

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"animakuro/cdn/internal/cdn/cdnutil"
	"animakuro/cdn/internal/entities"

	"github.com/pkg/errors"
)

const (
	DefaultName = "data"
)

var bucketsPath string

func SetBucketsPath(path string) {
	bucketsPath = path
}

func BucketsPath() string {
	return bucketsPath
}

func TryDelete(dirPath string) error {
	err := os.RemoveAll(dirPath)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return nil
		}

		err = fmt.Errorf("could not remove dir at: %s: %w", dirPath, err)
		return cdnutil.WrapInternal(err, "fs.TryDelete.os.RemoveAll")
	}

	return nil
}

func IsExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		}

		return false
	}

	return true
}

func ReadFile(path string) ([]byte, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, cdnutil.WrapInternal(err, "fs.ReadFile.os.Open")
	}

	bits, err := io.ReadAll(f)
	if err != nil {
		return nil, cdnutil.WrapInternal(err, "fs.ReadFile.io.ReadAll")
	}

	defer f.Close()
	return bits, nil
}

func WriteFileToBucket(buff []byte, bucket string, uuid string, fileName string) error {
	dirPath := path.Join(bucketsPath, bucket, uuid)
	fullPath := path.Join(dirPath, fileName)

	entr, err := os.ReadDir(dirPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return cdnutil.WrapInternal(err, "fs.WriteFileToBucket.os.ReadDir")
	}

	// No items (folder does not exist)
	if len(entr) == 0 {
		err = createDir(dirPath)
		if err != nil {
			return cdnutil.ChainInternal(err, "fs.WriteFileToBucket->fs.createDir")
		}
	}

	err = WriteFile(fullPath, buff)
	if err != nil {
		return cdnutil.ChainInternal(err, "fs.WriteFileToBucket->fs.WriteFile")
	}

	return nil
}

func WriteFile(path string, buff []byte) error {
	err := os.WriteFile(path, buff, 0777)
	if err != nil {
		return cdnutil.WrapInternal(err, "fs.WriteFile.os.WriteFile")
	}

	return nil
}

func CreateBucket(bucket string) error {
	p := path.Join(bucketsPath, bucket)

	entr, err := os.ReadDir(p)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return cdnutil.WrapInternal(err, "fs.CreateBucket.os.ReadDir")
	}

	//Bucket exists
	if len(entr) != 0 {
		return entities.ErrBucketAlreadyExists
	}

	err = createDir(p)
	if err != nil {
		return cdnutil.ChainInternal(err, "fs.CreateBucket->fs.createDir")
	}

	metafilePath := path.Join(p, "meta")

	//todo: delete bucket folder if fails
	_, err = os.Create(metafilePath)
	if err != nil {
		return cdnutil.WrapInternal(err, "fs.CreateBucket.os.Create")
	}

	return nil

}

func createDir(path string) error {
	err := os.MkdirAll(path, 0777)
	if err != nil {
		return cdnutil.WrapInternal(err, "fs.createDir.os.MkdirAll")
	}

	return nil
}
