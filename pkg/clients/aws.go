package clients

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"log"
	"os"
)

type AWSManager struct {
	config       *aws.Config
	disconnected chan bool
	closed       bool
	S3Manager    *S3Manager
	log          *log.Logger
}

func NewAWSManager(config *aws.Config) *AWSManager {
	sess := session.Must(session.NewSession(config))
	return &AWSManager{
		config:    config,
		S3Manager: NewS3Manager(sess),
		log:       log.New(os.Stdout, "AWSManager", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

//------------------------------------------------------------------------------

type S3Manager struct {
	session *session.Session
	Client  *s3.S3
	S3      []*S3Wrapper
}

func NewS3Manager(session *session.Session) *S3Manager {
	s3Client := s3.New(session)
	return &S3Manager{
		session: session,
		Client:  s3Client,
		S3:      make([]*S3Wrapper, 0),
	}
}

func (s3m *S3Manager) AddS3BucketManager(bucketName string) *S3Wrapper {
	s3Wrapper := NewS3Wrapper(bucketName, s3m.session)
	s3m.S3 = append(s3m.S3, s3Wrapper)
	return s3Wrapper
}

//------------------------------------------------------------------------------

// S3 Adapter
type S3Wrapper struct {
	bucketName string
	Uploader   *s3manager.Uploader
	Downloader *s3manager.Downloader
}

func NewS3Wrapper(bucketName string, s3Session *session.Session) *S3Wrapper {
	uploader := s3manager.NewUploader(s3Session)
	downloader := s3manager.NewDownloader(s3Session)
	return &S3Wrapper{
		bucketName: bucketName,
		Uploader:   uploader,
		Downloader: downloader,
	}
}

func (s3w *S3Wrapper) UploadFile(path string, filename string, data io.Reader) (*s3manager.UploadOutput, error) {
	upParams := &s3manager.UploadInput{
		Bucket: aws.String(s3w.bucketName),
		Key:    aws.String(path + filename),
		Body:   data,
	}
	result, err := s3w.Uploader.Upload(upParams)
	return result, err
}
