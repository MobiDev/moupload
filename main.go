package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"mime"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"upload.mobius.ovh/config"

	"github.com/spf13/viper"
)

func guid() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
		return "xxxx-xxxx-xxxx-xxxx"
	}
	guid := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return guid
}

func main() {

	viper.SetConfigName("config")

	// Set the path to look for the configurations file
	viper.AddConfigPath("$HOME/moupload")

	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()

	viper.SetConfigType("yml")
	var configuration config.Configurations

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&configuration)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	// Initialize minio client object.
	minioClient, err := minio.New(viper.GetString("endpoint"), &minio.Options{
		Creds:  credentials.NewStaticV4(viper.GetString("accessKeyID"), viper.GetString("secretAccessKey"), ""),
		Secure: viper.GetBool("useSSL"),
	})

	fmt.Println(viper.GetString("accessKeyID"), viper.GetString("secretAccessKey"))
	if err != nil {
		log.Fatalln(err)
	}

	guid := guid()
	log.Printf(os.Args[1])

	ln := strings.Split(os.Args[1], ".")
	pa := strings.Split(os.Args[1], "/")
	contentType := mime.TypeByExtension(ln[len(ln)-1])
	log.Printf("Content-Type:", contentType)

	fileName := regexp.MustCompile("[^a-zA-Z0-9-_.]").ReplaceAllString(pa[len(pa)-1], "")
	// log.Printf(fileName)

	ctx := context.Background()
	bucketName := "uploads"
	objectName := guid + "/" + fileName

	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	filePath := path.Join(currentWorkingDirectory, os.Args[1])

	// Upload the test file with FPutObject
	info, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: "contentType"})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)

	fmt.Println("Find the upload at ", viper.GetString("uploadPrefix")+objectName)
}
