package clients_test

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/awstesting/mock"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/defgenx/funicular/pkg/clients"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"

	//. "github.com/jvshahid/mock4go"
)

var _ = Describe("Aws", func() {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	var awsConfig = &aws.Config{
		DisableSSL: aws.Bool(true),
		Endpoint:   aws.String(server.URL),
	}

	Describe("Using AWS Manager", func() {

		var manager *AWSManager

		BeforeEach(func() {
			manager = NewAWSManager(awsConfig)
		})

		Context("From constructor function", func() {

			It("should create a valid instance", func() {
				Expect(manager).To(BeAssignableToTypeOf(&AWSManager{}))
			})

			It("should contain same S3 client", func() {
				Expect(manager.S3Manager).To(BeAssignableToTypeOf(&S3Manager{}))
			})
		})
	})

	Describe("Using AWS S3 Manager", func() {

		var s3Manager = NewS3Manager(mock.Session)

		Context("From constructor function", func() {

			It("should create a valid instance", func() {
				Expect(s3Manager).To(BeAssignableToTypeOf(&S3Manager{}))
			})

			It("should have no wrapper", func() {
				Expect(s3Manager.S3).To(HaveLen(0))
			})

			It("should have a S3 client", func() {
				Expect(s3Manager.Client).To(BeAssignableToTypeOf(&s3.S3{}))
			})

			//It("should upload a file", func() {
			//	s3Wrapper := s3Manager.AddS3BucketManager("test-bucket")
			//	Expect(s3Wrapper).To(BeAssignableToTypeOf(&S3Wrapper{}))
			//
			//	_, upError := s3Wrapper.UploadFile(
			//		"",
			//		"foo.bar",
			//		strings.NewReader("foo:bar"),
			//	)
			//	fmt.Print(upError)
			//})
		})
	})

	Describe("Using AWS S3 Wrapper", func() {

		var s3Wrapper = NewS3Wrapper("test-bucket", mock.Session)

		Context("From constructor function", func() {

			It("should create a valid instance", func() {
				Expect(s3Wrapper).To(BeAssignableToTypeOf(&S3Wrapper{}))
			})

			It("should have an uploader", func() {
				Expect(s3Wrapper.Uploader).To(BeAssignableToTypeOf(&s3manager.Uploader{}))
			})

			It("should have an downloader", func() {
				Expect(s3Wrapper.Downloader).To(BeAssignableToTypeOf(&s3manager.Downloader{}))
			})
		})
	})
})
