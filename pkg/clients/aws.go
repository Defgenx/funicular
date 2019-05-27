package clients

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"os"
	"sync"
)

type AWSManager struct {
	sync.Mutex
	config       *aws.Config
	disconnected chan bool
	closed       bool
	S3Manager    *S3Manager
	log          *log.Logger
}

func NewAWSManager(clientKey string, clientSecret string, region string, maxRetries int) *AWSManager {
	config := &aws.Config {
		Credentials: credentials.NewStaticCredentials(clientKey, clientSecret, ""),
		Region: aws.String(region),
		MaxRetries: aws.Int(maxRetries),
	}
	awsManager := &AWSManager{
		config: config,
		S3Manager: NewS3Manager(config),
		log: log.New(os.Stdout, "AWSManager", log.Ldate|log.Ltime|log.Lshortfile),
	}

	go awsManager.reconnects()
}



func (awsm *AWSManager) reconnects() {
	awsm.config.Credentials.Expire()
}

type S3Manager struct {
	session    *session.Session
	Client     *s3.S3
	S3Conns     []*S3Wrapper
}

func NewS3Manager(config *aws.Config) *S3Manager {
	sess := session.Must(session.NewSession(config))
	s3Client := s3.New(sess)
	return &S3Manager{
		session: sess,
		Client: s3Client,
		S3Conns: make([]*S3Wrapper, 0),
	}
}

func (s3m *S3Manager) AddS3BucketManager(bucketName string) *S3Wrapper {
	s3m.S3Conns = append(s3m.S3Conns, NewS3Wrapper(s3m.session, bucketName))
}

// S3 Adapter
type S3Wrapper struct {
	bucketName string
	Uploader   *s3manager.Uploader
	Downloader *s3manager.Downloader
}

func NewS3Wrapper(s3Session *session.Session, bucketName string) *S3Wrapper {
	uploader := s3manager.NewUploader(s3Session)
	downloader := s3manager.NewDownloader(s3Session)
	return &S3Wrapper{
		Uploader: uploader,
		Downloader: downloader,
		bucketName: bucketName,
	}
}