// Package cdnpath contains utility functions for making paths to different cdn files
package cdnpath

import "path"

type Existing struct {
	BucketsPath string
	Bucket      string
	UUID        string
	SHA1        string
}

// ToExistingFile makes path to existing file
// e.g. /local/buckets/site-content/abcd-arft/ash1371ahsdahd17236ah
func ToExistingFile(ex *Existing) string {
	return path.Join(ex.BucketsPath, ex.Bucket, ex.UUID, ex.SHA1)
}

type Original struct {
	BucketsPath string
	Bucket      string
	UUID        string
	DefaultName string
}

// ToOriginalFile makes path to original file
// e.g. /local/buckets/site-content/abcd-eafs/{DefaultName}
func ToOriginalFile(og *Original) string {
	return path.Join(og.BucketsPath, og.Bucket, og.UUID, og.DefaultName)
}

func ToDir(bucket, UUID string) string {
	return path.Join(bucket, UUID)
}
