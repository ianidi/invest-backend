package s3

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/minio/minio-go"
	"github.com/spf13/viper"
)

func Upload_minio(filename string, file []byte) error {
	endpoint := viper.GetString("s3_endpoint")
	key := viper.GetString("s3_key")
	secret := viper.GetString("s3_secret")
	useSSL := true
	bucketName := "invest"

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, key, secret, useSSL)
	if err != nil {
		return err
	}

	// Upload the file
	filePath := "/var/server/static/" + filename

	err = ioutil.WriteFile(filePath, file, 0644)
	if err != nil {
		return err
	}

	_, err = minioClient.FPutObject(bucketName, filename, filePath, minio.PutObjectOptions{})

	os.Remove("/var/server/static/temp_" + filename)

	if err != nil {
		return err
	}

	return nil
}

func Upload(filename string, file []byte) (string, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("ru-msk"),
		Credentials: credentials.NewStaticCredentials(viper.GetString("s3_key"), viper.GetString("s3_secret"), ""),
		Endpoint:    aws.String(viper.GetString("s3_endpoint")),
	}))
	svc := s3.New(sess)

	params := &s3.PutObjectInput{
		Bucket: aws.String(viper.GetString("s3_bucket")),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(file),
	}

	//Make file available publically
	params.SetACL("public-read")

	_, err := svc.PutObject(params)

	if err != nil {
		return "", err
	}

	return viper.GetString("s3_cdn_url") + filename, nil
}
