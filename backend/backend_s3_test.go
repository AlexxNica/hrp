package backend

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zlangbert/hrp/config"
	"testing"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/aws"
	"bytes"
	"io/ioutil"
	"errors"
	"github.com/zlangbert/hrp/util"
	"path/filepath"
	"io"
)

func TestS3_New(t *testing.T) {

	cfg := config.New()
	cfg.S3.Bucket = "test"
	cfg.S3.LocalSyncPath = "/tmp/hrp"

	b, err := newS3(cfg)

	assert.NotNil(t, b, "backend not nil")
	assert.Nil(t, err, "err nil")
}

func TestS3_New_ConfigVerify_MissingBucket(t *testing.T) {

	cfg := config.New()

	_, err := newS3(cfg)

	assert.Error(t, err, "missing config returns error")
	assert.Contains(t,
		err.Error(),
		"bucket missing",
		"expected bucket missing error")
}

func TestS3_New_ConfigVerify_MissingLocalSyncPath(t *testing.T) {

	cfg := config.New()
	cfg.S3.Bucket = "test"

	_, err := newS3(cfg)

	assert.Error(t, err, "missing config returns error")
	assert.Contains(t,
		err.Error(),
		"local sync path missing",
		"expected local sync path missing error")
}

func TestS3Backend_Initialize(t *testing.T) {

	cfg := testConfig()
	b, _ := newS3(cfg)

	indexData := bytes.NewReader([]byte{0, 1, 2, 3})

	// mock
	s3Api := new(s3Mock)
	s3Api.On("PutObject", &s3.PutObjectInput{
		Bucket: aws.String("bucket-test"),
		Key: aws.String("prefix/index.yaml"),
		Body: indexData,
	}).Return(
		&s3.PutObjectOutput{},
		nil,
	)
	b.svc = s3Api

	awsUtil := new(awsUtilMock)
	awsUtil.On(
		"Sync",
		"s3://" + filepath.Join(cfg.S3.Bucket, cfg.S3.Prefix),
		cfg.S3.LocalSyncPath,
	).Return(nil)
	b.awsUtil = awsUtil

	helmUtil := new(helmUtilMock)
	helmUtil.On(
		"GenerateIndex",
		cfg.BaseURL,
		cfg.S3.LocalSyncPath,
	).Return(nil)
	helmUtil.On(
		"ReadIndex",
		cfg.S3.LocalSyncPath,
	).Return(indexData, nil)

	// run
	err := b.Initialize()

	// check
	assert.Nil(t, err, "nil err")
}

func TestS3Backend_GetIndex(t *testing.T) {

	cfg := testConfig()
	b, _ := newS3(cfg)
	objectData := []byte{0, 1, 2, 3, 4}

	// mock
	s3Api := new(s3Mock)
	s3Api.On("GetObject", &s3.GetObjectInput{
		Bucket: aws.String("bucket-test"),
		Key: aws.String("prefix/test"),
	}).Return(
		&s3.GetObjectOutput{
			Body: ioutil.NopCloser(bytes.NewReader(objectData)),
		},
		nil,
	)
	b.svc = s3Api

	// run
	result, err := b.GetChart("test")

	// check
	assert.Nil(t, err, "nil err")
	assert.Equal(t, result, objectData)
}

func TestS3Backend_GetIndex_S3Error(t *testing.T) {

	cfg := testConfig()
	b, _ := newS3(cfg)

	// mock
	s3Api := new(s3Mock)
	s3Api.On("GetObject", &s3.GetObjectInput{
		Bucket: aws.String("bucket-test"),
		Key: aws.String("prefix/test"),
	}).Return(
		nil,
		errors.New("fail"),
	)
	b.svc = s3Api

	// run
	result, err := b.GetChart("test")

	// check
	assert.Nil(t, result, "nil result")
	assert.Contains(t, err.Error(), "fail")
}

//
// helpers
//

func testConfig() *config.AppConfig {
	cfg := config.New()
	cfg.S3.Bucket = "bucket-test"
	cfg.S3.Prefix = "prefix"
	cfg.S3.LocalSyncPath = "/tmp/hrp"

	return cfg
}

// s3Mock
type s3Mock struct{
	mock.Mock
	s3iface.S3API
}

func (m *s3Mock) GetObject(i *s3.GetObjectInput) (*s3.GetObjectOutput, error)  {
	args := m.Called(i)

	var out *s3.GetObjectOutput
	var err error

	if o, ok := args.Get(0).(*s3.GetObjectOutput); ok {
		out = o
	} else {
		out = nil
	}

	if e, ok := args.Get(1).(error); ok {
		err = e
	} else {
		err = nil
	}

	return out, err
}

func (m *s3Mock) PutObject(i *s3.PutObjectInput) (*s3.PutObjectOutput, error)  {
	args := m.Called(i)

	var out *s3.PutObjectOutput
	var err error

	if o, ok := args.Get(0).(*s3.PutObjectOutput); ok {
		out = o
	} else {
		out = nil
	}

	if e, ok := args.Get(1).(error); ok {
		err = e
	} else {
		err = nil
	}

	return out, err
}

// awsUtilMock
type awsUtilMock struct {
	mock.Mock
	util.AwsUtil
}

func (m *awsUtilMock) Sync(source string, target string) error  {
	args := m.Called(source, target)
	return args.Error(0)
}

// helmUtilMock
type helmUtilMock struct {
	mock.Mock
	util.HelmUtil
}

func (m *helmUtilMock) GenerateIndex(baseUrl string, path string) error {
	args := m.Called(baseUrl, path)
	return args.Error(0)
}

func (m *helmUtilMock) ReadIndex(path string) (io.ReadSeeker, error) {
	args := m.Called(path)
	return args.Get(0).(io.ReadSeeker), args.Error(1)
}