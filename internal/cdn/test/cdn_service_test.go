package cdn

import (
	"context"
	"io"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"testing"

	"animakuro/cdn/internal/cdn"
	"animakuro/cdn/internal/cdn/dto"
	mock_cdn "animakuro/cdn/internal/cdn/mocks"
	"animakuro/cdn/internal/cdnpath"
	"animakuro/cdn/internal/entities"
	"animakuro/cdn/internal/formdata"
	"animakuro/cdn/internal/fs"
	cache "animakuro/cdn/pkg/cache/bucket"
	filecache "animakuro/cdn/pkg/cache/file"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	dealer "github.com/sonyamoonglade/dealer-go/v2"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

const (
	testBucketsPath = "./testdata/buckets"
	testBucket      = "plain"
)

func initDeps(ctrl *gomock.Controller) (*mock_cdn.MockRepository, *zap.SugaredLogger, *cache.BucketCache, *filecache.FileCache, string, *dealer.Dealer) {

	fs.SetBucketsPath(testBucketsPath)

	mockRepo := mock_cdn.NewMockRepository(ctrl)

	fc := filecache.NewFileCache(zap.NewNop().Sugar(), &filecache.Config{
		MaxCacheSize:   128,
		MaxCacheItems:  5,
		CacheTTL:       5,
		CacheThreshold: 1,
		FlushEvery:     360,
		CheckoutEvery:  360,
	})

	bc := cache.NewBucketCache()
	bc.Add(&entities.Bucket{
		ID:   primitive.NewObjectID(),
		Name: "images",
		Operations: []entities.Operation{
			{Name: "get", Type: "public"},
			{Name: "delete", Type: "public"},
			{Name: "post", Type: "public"},
		},
		Module: "image",
	})

	domain := "cdn.animakuro"

	d := dealer.New(zap.NewNop().Sugar(), 5)
	return mockRepo, zap.NewNop().Sugar(), bc, fc, domain, d
}

type MockFile struct {
	data []byte
	idx  int
}

func (fd *MockFile) Close() error {
	return nil
}
func (fd *MockFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, nil
}
func (fd *MockFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}
func (fd *MockFile) Read(p []byte) (n int, err error) {
	if fd.idx >= len(fd.data) {
		return 0, io.EOF
	}
	n = copy(p, fd.data[fd.idx:])
	fd.idx += n
	return n, nil
}

func TestUploadManyOk(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, logger, bc, fc, domain, d := initDeps(ctrl)
	defer ctrl.Finish()

	f := &formdata.UploadFile{
		UploadName: fs.DefaultName + "." + "txt",
		Extension:  "txt",
		MimeType:   "text/plain",
		Size:       11,
		UUID:       uuid.NewString(),
		Open: func() (multipart.File, error) {
			return &MockFile{
				data: []byte("hello world"),
			}, nil
		},
	}

	d.Start(false)
	ff := []*formdata.UploadFile{f}

	// Create bucket for test
	err := fs.CreateBucket(testBucket)
	require.NoError(t, err)

	fdto := dto.SaveFileDto{
		Name:        fs.DefaultName + "." + f.Extension,
		Bucket:      testBucket,
		AvailableIn: []string{domain},
		MimeType:    f.MimeType,
		UUID:        f.UUID,
		Extension:   "." + f.Extension,
	}

	ctx := context.TODO()
	repo.EXPECT().SaveFile(ctx, fdto).Return(true, nil).Times(len(ff))

	service := cdn.NewService(logger, repo, bc, fc, domain, d)

	urls, ids, err := service.UploadMany(ctx, testBucket, ff)
	require.NoError(t, err)
	require.NotNil(t, urls)
	require.True(t, strings.Contains(urls[0], domain))
	require.True(t, strings.Contains(urls[0], f.UUID))
	require.True(t, strings.Contains(urls[0], testBucket))
	require.NotNil(t, ids)
	require.True(t, ids[0] == f.UUID)

	uplFile, err := f.Open()
	require.NoError(t, err)

	uplBits, err := io.ReadAll(uplFile)
	require.NoError(t, err)

	osFile, err := os.Open(cdnpath.ToExistingFile(&cdnpath.Existing{
		BucketsPath: fs.BucketsPath(),
		Bucket:      testBucket,
		UUID:        f.UUID,
		SHA1:        f.UploadName, // can use instead of real sha. See impl.
	}))
	require.NoError(t, err)

	bits, err := io.ReadAll(osFile)
	require.NoError(t, err)

	// --- bits should be equal
	require.Equal(t, uplBits, bits)

	//cleanup
	defer func() {
		err = fs.TryDelete(path.Join(fs.BucketsPath(), testBucket))
		require.NoError(t, err)

		osFile.Close()
		d.Stop()
	}()

}
func TestMustSaveOk(t *testing.T) {

	ctrl := gomock.NewController(t)
	repo, logger, bc, fc, domain, d := initDeps(ctrl)
	defer ctrl.Finish()

	d.Start(false)

	service := cdn.NewService(logger, repo, bc, fc, domain, d)

	err := fs.CreateBucket(testBucket)
	require.NoError(t, err)

	mockPath := path.Join(fs.BucketsPath(), testBucket, "file.txt")

	buff := []byte("hello world")
	service.MustSave(buff, mockPath)

	f, err := os.Open(mockPath)
	require.NoError(t, err)

	bits, err := io.ReadAll(f)
	require.NoError(t, err)

	require.Equal(t, len(bits), len(buff))

	for i := range bits {
		require.Equal(t, bits[i], buff[i])
	}

	defer func() {
		err = fs.TryDelete(path.Join(fs.BucketsPath(), testBucket))
		require.NoError(t, err)
		f.Close()
		d.Stop()
	}()

}
