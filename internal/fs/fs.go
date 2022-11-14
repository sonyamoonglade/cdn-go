package fs

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"animakuro/cdn/pkg/cdn_errors"
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
		text := err.Error()
		if strings.Contains(text, "no such file or directory") {
			return nil
		}
		return fmt.Errorf("could not remove dir at: %s. %s", dirPath, err.Error())
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
		return nil, cdn_errors.WrapInternal(err, "Service.ReadOriginalFile.os.Open")
	}

	bits, err := io.ReadAll(f)
	if err != nil {
		return nil, cdn_errors.WrapInternal(err, "Service.ReadOriginalFile.io.ReadAll")
	}

	defer f.Close()
	return bits, nil
}

func WriteFileToBucket(buff []byte, bucket string, uuid string, fileName string) error {

	dirPath := path.Join(bucketsPath, bucket, uuid)
	fullPath := path.Join(dirPath, fileName)

	entr, err := os.ReadDir(dirPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return cdn_errors.WrapInternal(err, "fs.WriteFileToBucket.os.ReadDir")
	}

	//No items (folder does not exist)
	if len(entr) == 0 {

		err = createDir(dirPath)
		if err != nil {
			return cdn_errors.WrapInternal(err, "fs.WriteFIleToBucket.createDir")
		}

	}

	return WriteFile(fullPath, buff)
}

func WriteFile(path string, buff []byte) error {
	err := os.WriteFile(path, buff, 0777)
	if err != nil {
		return cdn_errors.WrapInternal(err, "fs.WriteNew.os.WriteFile")
	}
	return nil
}

func CreateBucket(bucket string) error {

	p := path.Join(bucketsPath, bucket)
	entr, err := os.ReadDir(p)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return cdn_errors.WrapInternal(err, "fs.CreateBucket.os.ReadDir")
	}
	//Bucket exists
	if len(entr) != 0 {
		return cdn_errors.ErrBucketAlreadyExists
	}

	err = createDir(p)
	if err != nil {
		return cdn_errors.WrapInternal(err, "fs.CreateBucket.fs.createDir")
	}

	metafilePath := path.Join(p, "meta")

	//todo: delete bucket folder if fails
	_, err = os.Create(metafilePath)
	if err != nil {
		return cdn_errors.WrapInternal(err, "fs.CreateBucket.os.Create")
	}

	return nil

}

func createDir(path string) error {
	return os.MkdirAll(path, 0777)
}
