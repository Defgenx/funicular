package clients

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"os"
	"sync"
)

const AWS_BASE_URL = "https://s3.console.aws.amazon.com/s3/object"

// Manager S3 connections structure
type S3Manager struct {
	clientRegion string
	config       *aws.Config
	uploaders    []*S3Wrapper
	log          *log.Logger
}


func NewS3Manager(region string) *S3Manager {
	config := &aws.Config{Region: aws.String(region)}
	return &S3Manager{
		clientRegion: region,
		config: config,
		uploaders: make([]*S3Wrapper, 0),
		log: log.New(os.Stdout, "S3Manager", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// S3 connection
type S3Wrapper struct {
	sync.Mutex
	bucketName string
	session    *session.Session
	uploader   *s3manager.Uploader
	shutdown   chan bool
	closed     bool
	reconnects uint64
}

func NewS3Wrapper(session *session.Session) {

}