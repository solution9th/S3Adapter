package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {

	var (
		accessKey  = "<new accessKey>"
		secretKey  = "<new secretKey>"
		region     = "<region>"
		endPoint   = "<endpoint>"
		bucketName = "<bucket name>"
	)

	sess, err := session.NewSession(&aws.Config{
		Endpoint:    aws.String(endPoint),
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		panic(err)
	}

	client := s3.New(sess)

	output, err := client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(output.String())
}
