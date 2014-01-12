package awsutil

import (
	"os"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
)

func auth() aws.Auth {
	return aws.Auth{
		AccessKey: os.Getenv("AWS_ACCESS_ID"),
		SecretKey: os.Getenv("AWS_SECRET_KEY"),
	}
}

// Bucket returns the default S3 bucket for this application (configured by
// env)
func Bucket() *s3.Bucket {
	// TODO: is this expensive? Should I be making it a singleton?
	s3 := s3.New(auth(), aws.USEast)
	return s3.Bucket(os.Getenv("S3_BUCKET"))
}
