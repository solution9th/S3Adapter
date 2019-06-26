package cos

// import (
// 	"bufio"
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"os"
// 	"testing"
// 	"time"

// 	"github.com/aws/aws-sdk-go/aws"
// 	"github.com/aws/aws-sdk-go/aws/credentials"
// 	"github.com/aws/aws-sdk-go/aws/session"
// 	"github.com/aws/aws-sdk-go/service/s3"
// 	. "github.com/smartystreets/goconvey/convey"
// )

// type Provider struct {
// 	// The error to be returned from Retrieve
// 	Err error

// 	// The provider name to set on the Retrieved returned Value
// 	ProviderName string
// }

// // Retrieve will always return the error that the ErrorProvider was created with.
// func (p Provider) Retrieve() (credentials.Value, error) {
// 	return credentials.Value{
// 		//s3
// 		//AccessKeyID:"AKIAIPET2JNKIR4L3DQA",
// 		//SecretAccessKey:"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",
// 		//cos
// 		AccessKeyID:     "AKIDXt0wd5JkiQyHMBc48OSHqCJH0Tnt3jrk",
// 		SecretAccessKey: "pQlOnxsiKhogQ2RelV0GGBsKXeW9dw73",
// 		ProviderName:    p.ProviderName}, p.Err
// }

// // IsExpired will always return not expired.
// func (p Provider) IsExpired() bool {
// 	return false

// }

// func TestS3(t *testing.T) {
// 	Convey("连接到s3", t, func() {
// 		fmt.Println("连接到s3")
// 		sess := session.Must(session.NewSession(&aws.Config{
// 			//Region: aws.String("ap-northeast-1"),
// 			Region:      aws.String("ap-northeast-1"),
// 			Credentials: credentials.NewCredentials(Provider{}),
// 			//Endpoint: aws.String("s3.amazonaws.com"),
// 			Endpoint:   aws.String("http://mybee.com:9091"),
// 			DisableSSL: aws.Bool(false),
// 		}))
// 		service := s3.New(sess)
// 		Convey("上传文件", func() {
// 			fp, err := os.Open("s3_test.go")
// 			So(err, ShouldBeNil)
// 			defer fp.Close()

// 			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
// 			defer cancel()

// 			_, err = service.PutObjectWithContext(ctx, &s3.PutObjectInput{
// 				//Bucket: aws.String("mafeng"),
// 				Bucket: aws.String("mafeng-1258705469"),
// 				Key:    aws.String("test/s3_test.go"),
// 				Body:   fp,
// 			})
// 			So(err, ShouldBeNil)
// 		})

// 		Convey("HEAD文件", func() {
// 			fmt.Println("HEAD文件")
// 			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
// 			defer cancel()
// 			out, err := service.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
// 				//Bucket: aws.String("mafeng"),
// 				Bucket: aws.String("mafeng-1258705469"),
// 				Key:    aws.String("test/s3_test.go"),
// 			})
// 			So(err, ShouldBeNil)
// 			fmt.Println(out)
// 		})

// 		Convey("下载文件", func() {
// 			fmt.Println("下载文件")
// 			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
// 			defer cancel()
// 			out, err := service.GetObjectWithContext(ctx, &s3.GetObjectInput{
// 				//Bucket: aws.String("mafeng"),
// 				Bucket: aws.String("mafeng-1258705469"),
// 				Key:    aws.String("test/s3_test.go"),
// 			})
// 			So(err, ShouldBeNil)
// 			defer out.Body.Close()
// 			scanner := bufio.NewScanner(out.Body)
// 			for scanner.Scan() {
// 				Println(scanner.Text())
// 			}
// 		})

// 		//
// 		//Convey("遍历目录 ListObjectsPages", func() {
// 		//	fmt.Println("遍历目录 ListObjectsPages")
// 		//	var objkeys []string
// 		//
// 		//	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
// 		//	defer cancel()
// 		//
// 		//	err := service.ListObjectsPagesWithContext(ctx, &s3.ListObjectsInput{
// 		//		Bucket: aws.String("mafeng"),
// 		//		Prefix: aws.String("test/"),
// 		//	}, func(output *s3.ListObjectsOutput, b bool) bool {
// 		//		for _, content := range output.Contents {
// 		//			objkeys = append(objkeys, aws.StringValue(content.Key))
// 		//		}
// 		//		return true
// 		//	})
// 		//	So(err, ShouldBeNil)
// 		//	Println(objkeys)
// 		//})
// 		//
// 		//
// 		//Convey("遍历目录 ListObjects", func() {
// 		//	fmt.Println("遍历目录 ListObjects")
// 		//	var objkeys []string
// 		//
// 		//	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
// 		//	defer cancel()
// 		//
// 		//	out, err := service.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
// 		//		Bucket: aws.String("mafeng"),
// 		//		Prefix: aws.String("test/"),
// 		//	})
// 		//	So(err, ShouldBeNil)
// 		//	for _, content := range out.Contents {
// 		//		objkeys = append(objkeys, aws.StringValue(content.Key))
// 		//	}
// 		//	Println(objkeys)
// 		//})
// 	})
// }

// func Test_cosHeaderToS3Header(t *testing.T) {
// 	var header = make(http.Header)
// 	header.Set("x-cos-meta-mafeng", "hhhh")
// 	m := cosHeaderToS3Header(header)
// 	fmt.Println(m)
// }

// func Test_s3HeaderToCosMeta(t *testing.T) {
// 	var header = make(map[string]*string)
// 	header["x-amz-meta-mafeng"] = aws.String("hhhh")
// 	m := s3HeaderToCosMeta(header)
// 	fmt.Println(m)
// }
