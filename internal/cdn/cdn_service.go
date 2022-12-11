package cdn

import (
	"context"
	"fmt"
	"io"
	"path"

	"animakuro/cdn/internal/cdn/cdnutil"
	"animakuro/cdn/internal/cdn/dto"
	"animakuro/cdn/internal/entities"
	"animakuro/cdn/internal/formdata"
	"animakuro/cdn/internal/fs"

	bucketcache "animakuro/cdn/pkg/cache/bucket"
	filecache "animakuro/cdn/pkg/cache/file"
	"animakuro/cdn/pkg/dealer"

	"github.com/gabriel-vasile/mimetype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

const (
	saveRetries   = 5
	deleteRetries = 5
)

type Service interface {
	//DB logic
	InitBuckets(ctx context.Context) error
	GetBucketDB(ctx context.Context, bucketName string) (*entities.Bucket, error)
	GetAllBucketsDB(ctx context.Context) ([]*entities.Bucket, error)
	SaveBucketDB(ctx context.Context, dto dto.CreateBucketDto) (*entities.Bucket, error)

	GetFileDB(ctx context.Context, bucket string, uuid string) (*entities.File, error)
	SaveFileDB(ctx context.Context, dto dto.SaveFileDto) error
	MarkAsDeletableDB(ctx context.Context, bucket string, mongoID primitive.ObjectID) error
	DeleteFileDB(ctx context.Context, bucket string, uuid string) error

	//Internal CDN logic
	UploadMany(ctx context.Context, bucket string, files []*formdata.UploadFile) ([]string, []string, error)
	MustSave(buff []byte, path string)

	ReadFile(path string, hosts []string) ([]byte, error)
	ReadExisting(path string) ([]byte, bool, error)

	DeleteAll(path string) error
	TryDeleteLocally(dirPath string)

	ParseMime(buff []byte) string
}

type cdnService struct {
	domain     string
	repository Repository
	logger     *zap.SugaredLogger
	dealer     *dealer.Dealer
	bc         *bucketcache.BucketCache
	fc         filecache.FileCache
}

func NewService(logger *zap.SugaredLogger,
	repo Repository,
	bucketCache *bucketcache.BucketCache,
	fileCache filecache.FileCache,
	domain string,
	dealer *dealer.Dealer) Service {
	return &cdnService{
		logger:     logger,
		repository: repo,
		bc:         bucketCache,
		fc:         fileCache,
		domain:     domain,
		dealer:     dealer,
	}
}

func (s *cdnService) GetBucketDB(ctx context.Context, name string) (*entities.Bucket, error) {
	bucket, err := s.repository.GetBucket(ctx, name)
	if err != nil {
		return nil, cdnutil.WrapInternal(err, "cdnService.GetBucket")
	}

	//No bucket was found
	if bucket == nil {
		return nil, entities.ErrBucketNotFound
	}

	return bucket, nil
}

func (s *cdnService) GetFileDB(ctx context.Context, bucket string, name string) (*entities.File, error) {
	file, err := s.repository.GetFile(ctx, bucket, name)
	if err != nil {
		return nil, cdnutil.WrapInternal(err, "cdnService.GetFile")
	}

	// No file was found
	if file == nil {
		return nil, entities.ErrFileNotFound
	}

	// Marked for deletion
	if file.IsDeletable == true {
		// Return NotFound because clients shouldn't know that file is deleted.
		// If it's marked, for them it's equal to NotFound.
		// This is a special case.
		return nil, entities.ErrFileNotFound
	}

	return file, nil
}

func (s *cdnService) SaveFileDB(ctx context.Context, dto dto.SaveFileDto) error {
	ok, err := s.repository.SaveFile(ctx, dto)
	if err != nil {
		return cdnutil.WrapInternal(err, "cdnService.SaveFileDB.s.repository.SaveFile")
	}

	// Duplicate
	if ok == false {
		return entities.ErrFileAlreadyExists
	}

	return nil
}

func (s *cdnService) SaveBucketDB(ctx context.Context, dto dto.CreateBucketDto) (*entities.Bucket, error) {
	b, err := s.repository.SaveBucket(ctx, dto)
	if err != nil {
		return nil, cdnutil.WrapInternal(err, "cdnService.SaveBucketDB")
	}

	// Duplicate
	if b == nil {
		return nil, entities.ErrBucketAlreadyExists
	}

	return b, nil
}

func (s *cdnService) GetAllBucketsDB(ctx context.Context) ([]*entities.Bucket, error) {
	buckets, err := s.repository.GetAllBuckets(ctx)
	if err != nil {
		return nil, cdnutil.WrapInternal(err, "cdnService.GetAllBucketsDB.s.repository.GetAllBuckets")
	}

	//No buckets are present
	if buckets == nil {
		return nil, entities.ErrBucketsNotDefined
	}

	return buckets, nil
}

func (s *cdnService) DeleteFileDB(ctx context.Context, bucket string, uuid string) error {
	ok, err := s.repository.DeleteFile(ctx, bucket, uuid)
	if err != nil {
		return cdnutil.WrapInternal(err, "cdnService.DeleteFileDB.s.repository.DeleteFile")
	}

	if !ok {
		return entities.ErrFileNotFound
	}

	return nil
}

func (s *cdnService) MarkAsDeletableDB(ctx context.Context, bucket string, mongoID primitive.ObjectID) error {
	if err := s.repository.MarkAsDeletable(ctx, bucket, mongoID); err != nil {
		return cdnutil.WrapInternal(err, "cdnService.MarkAsDeletableDB.s.repository.MarkAsDeletable")
	}
	return nil
}

func (s *cdnService) UploadMany(ctx context.Context, bucket string, files []*formdata.UploadFile) ([]string, []string, error) {
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

		j := s.dealer.Run(func() *dealer.JobResult {
			// todo: move path to var
			return dealer.NewJobResult(nil, fs.WriteFileToBucket(buff, bucket, file.UUID, file.UploadName))
		})

		res := j.Wait()
		if err := res.Err; err != nil {
			return nil, nil, cdnutil.WrapInternal(err, "cdnService.UploadFiles.fs.WriteFileToBucket")
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

		err = s.SaveFileDB(ctx, fdto)
		if err != nil {
			// If saving to DB has failed then delete file locally.
			defer func() {
				// TODO: move to cdn/path
				// use path from cdn_service.go:176
				pathToDelete := path.Join(fs.BucketsPath(), bucket, file.UUID)
				if err := s.DeleteAll(pathToDelete); err != nil {
					err = cdnutil.ChainInternal(err, "cdnService.UploadMany->cdnService.DeleteAll")
					s.logger.Errorf(err.Error())
				}
			}()

			// Return to client that something went wrong
			return nil, nil, cdnutil.ChainInternal(err, "cdnService.UploadMany->cdnService.SaveFileDB")
		}

		fileURL := fmt.Sprintf("%s/%s/%s", s.domain, bucket, file.UUID)

		urls = append(urls, fileURL)
		ids = append(ids, file.UUID)

		s.logger.Debugf("saved: %s to %s at: %s", fdto.Name, bucket, s.domain)
	}

	return urls, ids, nil
}

func (s *cdnService) ReadFile(path string, hosts []string) ([]byte, error) {

	availableHost, isSelfHosting := cdnutil.IsAvailable(hosts, s.domain)
	if !isSelfHosting {
		// download file's bits here from availableHost
		_ = availableHost
	}

	// Lookup for locally cached original file
	bits, isCached := s.fc.Lookup(path)
	if isCached {
		s.logger.Debugf("original file is found in cache: %s", path)
		return bits, nil
	}

	j := s.dealer.Run(func() *dealer.JobResult {
		// Read file by path from disk
		return dealer.NewJobResult(fs.ReadFile(path))
	})

	res := j.Wait()
	bits, err := res.Out.([]byte), res.Err
	if err != nil {
		return nil, cdnutil.WrapInternal(err, "cdnService.ReadFile.fs.ReadFile")
	}

	return bits, nil
}

func (s *cdnService) MustSave(buff []byte, path string) {

	var ok bool
	for i := 0; i < saveRetries; i++ {
		if ok {
			break
		}

		j := s.dealer.Run(func() *dealer.JobResult {
			return dealer.NewJobResult(nil, fs.WriteFile(path, buff))
		})

		res := j.Wait()
		if err := res.Err; err != nil {
			s.logger.Errorf("could not save file: %s. err: %s. Retries left: %d", path, err.Error(), saveRetries-i)
			continue
		}

		ok = true
	}

	if !ok {
		s.logger.Errorf("could not save file: %s. Fatal", path)
		return
	}

	s.logger.Debug("saved file: %s", path)
}

func (s *cdnService) InitBuckets(ctx context.Context) error {

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

func (s *cdnService) ReadExisting(path string) ([]byte, bool, error) {
	// Lookup in cache firstly
	bits, isCached := s.fc.Lookup(path)
	if isCached {
		return bits, true, nil
	}

	// Checkout for locally resolved file in disk
	ok := fs.IsExists(path)
	if ok == true {

		// Read file from disk
		j := s.dealer.Run(func() *dealer.JobResult {
			return dealer.NewJobResult(fs.ReadFile(path))
		})

		res := j.Wait()
		bits, err := res.Out.([]byte), res.Err
		if err != nil {
			return nil, false, cdnutil.WrapInternal(err, "cdnService.TryReadExisting.fs.ReadFile")
		}

		return bits, true, nil
	}

	// File does not exist locally or in cache
	return nil, false, nil
}

func (s *cdnService) TryDeleteLocally(dirPath string) {

	s.logger.Debugf("trying to delete locally: %s", dirPath)
	j := s.dealer.Run(func() *dealer.JobResult {
		return dealer.NewJobResult(nil, fs.TryDelete(dirPath))
	})

	res := j.Wait()
	if err := res.Err; err != nil {
		s.logger.Errorf("could not delete locally at: %s", err.Error())
	}
	return
}

func (s *cdnService) DeleteAll(path string) error {

	s.logger.Debugf("deleting all at: %s", path)

	var ok bool
	for i := 0; i < deleteRetries; i++ {

		j := s.dealer.Run(func() *dealer.JobResult {
			return dealer.NewJobResult(nil, fs.TryDelete(path))
		})

		res := j.Wait()
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

	return entities.ErrFileCantDelete
}

func (s *cdnService) ParseMime(buff []byte) string {
	return mimetype.Detect(buff).String()
}
