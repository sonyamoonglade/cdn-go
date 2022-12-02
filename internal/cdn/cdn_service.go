package cdn

import (
	"context"
	"fmt"
	"io"
	"time"

	"animakuro/cdn/internal/cdn/dto"
	cdnutil "animakuro/cdn/internal/cdn/util"
	"animakuro/cdn/internal/entities"
	"animakuro/cdn/internal/formdata"
	"animakuro/cdn/internal/fs"
	cache "animakuro/cdn/pkg/cache/bucket"
	filecache "animakuro/cdn/pkg/cache/file"

	"github.com/gabriel-vasile/mimetype"
	"github.com/sonyamoonglade/dealer-go/v2"
	"go.uber.org/zap"
)

const (
	saveRetries   = 5
	deleteRetries = 5
)

var (
	// todo: make configurable
	saveTimeout = time.Second * 10
)

var ()

type Service interface {
	//DB logic
	GetBucketDB(ctx context.Context, bucketName string) (*entities.Bucket, error)
	GetFileDB(ctx context.Context, bucket string, uuid string) (*entities.File, error)
	GetAllBucketsDB(ctx context.Context) ([]*entities.Bucket, error)
	SaveFileDB(ctx context.Context, dto dto.SaveFileDto) error
	SaveBucketDB(ctx context.Context, dto dto.CreateBucketDto) (*entities.Bucket, error)
	DeleteFileDB(ctx context.Context, bucket string, uuid string) error
	InitBuckets(ctx context.Context) error

	//Internal CDN logic
	UploadMany(ctx context.Context, bucket string, files []*formdata.UploadFile) ([]string, []string, error)
	ReadFile(isOrig bool, path string, hosts []string) ([]byte, error)
	DeleteAll(path string) error
	MustSave(buff []byte, path string)
	TryReadExisting(path string) ([]byte, bool, error)
	TryDeleteLocally(dirPath string)

	ParseMime(buff []byte) string
}

type CdnService struct {
	logger     *zap.SugaredLogger
	repository Repository
	bc         *cache.BucketCache
	fc         *filecache.FileCache
	domain     string
	dealer     *dealer.Dealer
}

func NewService(logger *zap.SugaredLogger,
	repo Repository,
	bucketCache *cache.BucketCache,
	fileCache *filecache.FileCache,
	domain string,
	dealer *dealer.Dealer) *CdnService {
	return &CdnService{
		logger:     logger,
		repository: repo,

		bc:     bucketCache,
		fc:     fileCache,
		domain: domain,
		dealer: dealer,
	}
}

func (s *CdnService) SetSaveTimeout(dur time.Duration) {
	saveTimeout = dur
}

func (s *CdnService) GetBucketDB(ctx context.Context, name string) (*entities.Bucket, error) {
	bucket, err := s.repository.GetBucket(ctx, name)
	if err != nil {
		return nil, cdnutil.WrapInternal(err, "CdnService.GetBucket")
	}

	//No bucket was found
	if bucket == nil {
		return nil, entities.ErrBucketNotFound
	}

	return bucket, nil
}

func (s *CdnService) GetFileDB(ctx context.Context, bucket string, name string) (*entities.File, error) {

	file, err := s.repository.GetFile(ctx, bucket, name)
	if err != nil {
		return nil, cdnutil.WrapInternal(err, "CdnService.GetFile")
	}

	//No file was found
	if file == nil {
		return nil, entities.ErrFileNotFound
	}

	return file, nil
}

func (s *CdnService) SaveFileDB(ctx context.Context, dto dto.SaveFileDto) error {
	ok, err := s.repository.SaveFile(ctx, dto)
	if err != nil {
		return cdnutil.WrapInternal(err, "CdnService.SaveFileDB.s.repository.SaveFile")
	}
	// Duplicate
	if ok == false {
		return entities.ErrFileAlreadyExists
	}

	return nil
}

func (s *CdnService) SaveBucketDB(ctx context.Context, dto dto.CreateBucketDto) (*entities.Bucket, error) {
	b, err := s.repository.SaveBucket(ctx, dto)
	if err != nil {
		return nil, cdnutil.WrapInternal(err, "CdnService.SaveBucketDB")
	}

	// Duplicate
	if b == nil {
		return nil, entities.ErrBucketAlreadyExists
	}

	return b, nil
}

func (s *CdnService) GetAllBucketsDB(ctx context.Context) ([]*entities.Bucket, error) {
	buckets, err := s.repository.GetAllBuckets(ctx)
	if err != nil {
		return nil, cdnutil.WrapInternal(err, "CdnService.GetAllBucketsDB.s.repository.GetAllBuckets")
	}
	//No buckets are present
	if buckets == nil {
		return nil, entities.ErrBucketsNotDefined
	}

	return buckets, nil
}

func (s *CdnService) DeleteFileDB(ctx context.Context, bucket string, uuid string) error {
	ok, err := s.repository.DeleteFile(ctx, bucket, uuid)
	if err != nil {
		return cdnutil.WrapInternal(err, "CdnService.DeleteFileDB.s.repository.DeleteFile")
	}

	if !ok {
		return entities.ErrFileNotFound
	}

	return nil
}

func (s *CdnService) UploadMany(ctx context.Context, bucket string, files []*formdata.UploadFile) ([]string, []string, error) {

	var urls []string
	var ids []string

	for _, file := range files {
		osfile, err := file.Open()
		if err != nil {
			return nil, nil, cdnutil.WrapInternal(err, "UploadFiles.file.Open")
		}

		buff, err := io.ReadAll(osfile)
		if err != nil {
			return nil, nil, cdnutil.WrapInternal(err, "UploadFiles.io.ReadAll")
		}

		j := dealer.NewJob(func() *dealer.JobResult {
			return dealer.NewJobResult(nil, fs.WriteFileToBucket(buff, bucket, file.UUID, file.UploadName))
		})
		s.dealer.AddJob(j)

		res := j.WaitResult()
		s.logger.Debugf("job result: %+v\n", res)
		if err := res.Err; err != nil {
			return nil, nil, cdnutil.WrapInternal(err, "CdnService.UploadFiles.fs.WriteFileToBucket")
		}

		//todo: get host from env
		fdto := dto.SaveFileDto{
			Name:        file.UploadName,
			Bucket:      bucket,
			AvailableIn: []string{s.domain},
			MimeType:    file.MimeType,
			UUID:        file.UUID,
			Extension:   "." + file.Extension,
		}

		s.logger.Debugf("dto: %+v\n", fdto)

		err = s.SaveFileDB(ctx, fdto)
		if err != nil {
			return nil, nil, err
		}

		//todo: path
		path := fmt.Sprintf("%s/%s/%s", s.domain, bucket, file.UUID)

		s.logger.Debugf("path: %s\n", path)
		urls = append(urls, path)
		ids = append(ids, file.UUID)
	}

	return urls, ids, nil
}

func (s *CdnService) ReadFile(isOriginal bool, path string, hosts []string) ([]byte, error) {

	availableHost, isSelfHosting := cdnutil.IsAvailable(hosts, s.domain)
	if !isSelfHosting {
		//download file's bits here from availableHost
		_ = availableHost
	}

	//Lookup for cached locally original file
	bits, isCached := s.fc.Lookup(path)
	if isCached && isOriginal {
		s.logger.Debugf("found in cache: %s", path)
		return bits, nil
	}

	j := dealer.NewJob(func() *dealer.JobResult {
		//Read original file from os
		return dealer.NewJobResult(fs.ReadFile(path))
	})
	s.dealer.AddJob(j)

	res := j.WaitResult()
	bits, err := res.Out.([]byte), res.Err
	if err != nil {
		return nil, cdnutil.WrapInternal(err, "CdnService.Read.fs.ReadFile")
	}

	//User has requested for original file (no need for resolver processing)
	if isOriginal {
		s.logger.Debugf("found locally: %s", path)
		return bits, nil
	}

	return bits, nil
}

func (s *CdnService) MustSave(buff []byte, path string) {
	ctx, cancel := context.WithTimeout(context.Background(), saveTimeout)
	defer cancel()

	select {
	case <-ctx.Done():
		s.logger.Errorf("could not save file: %s. Reached timeout: %s", path, ctx.Err().Error())
		return
	default:
		var ok bool
		for i := 0; i < saveRetries; i++ {
			if ok {
				return
			}
			j := dealer.NewJob(func() *dealer.JobResult {
				return dealer.NewJobResult(nil, fs.WriteFile(path, buff))
			})
			s.dealer.AddJob(j)

			res := j.WaitResult()
			if err := res.Err; err != nil {
				s.logger.Errorf("could not save file: %s. err: %s. Retries left: %d", path, err.Error(), saveRetries-i)
				continue
			}

			ok = true
			s.logger.Debugf("saved: %s", path)
		}

		if !ok {
			s.logger.Errorf("could not save file: %s. Fatal", path)
		}
	}

}

func (s *CdnService) InitBuckets(ctx context.Context) error {

	buckets, err := s.GetAllBucketsDB(ctx)
	if err != nil {
		//Do not wrap
		return err
	}

	for _, bucket := range buckets {
		s.bc.Add(bucket)
		s.logger.Debugf("bucket: '%s' is added to cache", bucket.Name)
	}

	return nil
}

func (s *CdnService) TryReadExisting(path string) ([]byte, bool, error) {
	// Lookup in cache firstly
	bits, isCached := s.fc.Lookup(path)
	if isCached {
		return bits, true, nil
	}

	// Checkout for locally resolved file in disk
	ok := fs.IsExists(path)
	if ok == true {

		// Read file from disk
		j := dealer.NewJob(func() *dealer.JobResult {
			return dealer.NewJobResult(fs.ReadFile(path))
		})

		s.dealer.AddJob(j)

		res := j.WaitResult()
		bits, err := res.Out.([]byte), res.Err
		if err != nil {
			return nil, false, cdnutil.WrapInternal(err, "CdnService.TryReadExisting.fs.ReadFile")
		}

		return bits, true, nil
	}

	// File does not exist locally or in cache
	return nil, false, nil
}

func (s *CdnService) TryDeleteLocally(dirPath string) {

	s.logger.Debugf("trying to delete locally: %s", dirPath)
	j := dealer.NewJob(func() *dealer.JobResult {
		return dealer.NewJobResult(nil, fs.TryDelete(dirPath))
	})
	s.dealer.AddJob(j)

	res := j.WaitResult()
	if err := res.Err; err != nil {
		s.logger.Errorf("could not delete locally at: %s", err.Error())
	}
	return
}

func (s *CdnService) DeleteAll(path string) error {

	s.logger.Debugf("deleting all at: %s", path)

	var ok bool
	for i := 0; i < deleteRetries; i++ {

		j := dealer.NewJob(func() *dealer.JobResult {
			return dealer.NewJobResult(nil, fs.TryDelete(path))
		})
		s.dealer.AddJob(j)

		res := j.WaitResult()
		if err := res.Err; err != nil {
			s.logger.Errorf("could not delete at: %s err: %s. Retries left: %d", path, err.Error(), deleteRetries-i)
			continue
		}

		//Deleted successfully
		ok = true
	}

	if ok {
		return nil
	}

	return entities.ErrFileCantRemove
}

func (s *CdnService) ParseMime(buff []byte) string {
	return mimetype.Detect(buff).String()
}
