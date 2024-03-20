package utilfunc

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// DownloadFromS3
// take an AWS session that has access to S3 in an AWS account
// and download to RAM the passed key object within the passed bucket.
func DownloadFromS3(sess *session.Session, bucket, key string) (data *[]byte, err error) {

	// instantiate an S3 downloader and a writer
	buf := aws.NewWriteAtBuffer([]byte{})
	downloader := s3manager.NewDownloader(sess)

	// download the file
	_, err = downloader.Download(buf, &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	if err != nil {
		return
	}

	// parse the data contents
	bufb := buf.Bytes()
	data = &bufb

	return
}

// UploadToS3
// take an AWS session that has access to S3 in an AWS account
// and upload the passed bytes content to the passed bucket as an
// object within the passed key, with the correct Content-Type metadata.
func UploadToS3(sess *session.Session, bucket, key, contentType string, data *[]byte) (err error) {

	// instantiate an S3 uploader
	uploader := s3manager.NewUploader(sess)

	// prepare the upload
	input := &s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(*data),
		ContentType: aws.String(contentType),
	}

	// do the upload
	_, err = uploader.UploadWithContext(context.Background(), input)

	return
}
