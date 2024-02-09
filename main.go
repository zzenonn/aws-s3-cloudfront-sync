package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cheggaaa/pb/v3"
)

var (
	bucketArg        string // Original bucket argument containing both bucket name and optional prefix
	cloudFrontDomain string
	localSyncDir     string
)

func main() {
	flag.StringVar(&bucketArg, "bucketName", "", "The S3 bucket and optional prefix in the format 'bucket-name/prefix-path'")
	flag.StringVar(&cloudFrontDomain, "cloudFrontDomain", "", "The CloudFront domain mapping to the S3 bucket (e.g., 'https://sub.domain.com')")
	flag.StringVar(&localSyncDir, "localSyncDir", "./syncedFiles", "Local directory to sync files to")
	flag.Parse()

	// Split bucketArg into bucketName and prefix
	bucketName, prefix := parseBucketArg(bucketArg)

	// println(bucketName)
	// println(prefix)

	// Check for required flags and arguments
	if strings.TrimSpace(bucketName) == "" || strings.TrimSpace(cloudFrontDomain) == "" {
		fmt.Println("Error: bucketName and cloudFrontDomain are required.")
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.TODO()

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("unable to load SDK config, %w", err))
	}

	// Create an S3 client
	client := s3.NewFromConfig(cfg)

	// Ensure local directory exists
	if err := os.MkdirAll(localSyncDir, 0755); err != nil {
		fmt.Printf("Failed to create local directory: %v\n", err)
		return
	}

	// List objects in the bucket with prefix
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			fmt.Printf("Failed to get page: %v\n", err)
			return
		}

		for _, object := range page.Contents {
			fmt.Printf("Processing %s\n", *object.Key)
			syncFile(ctx, bucketName, prefix, *object.Key)
		}
	}
}

func parseBucketArg(bucketArg string) (bucketName, prefix string) {
	parts := strings.SplitN(bucketArg, "/", 2)
	bucketName = parts[0]
	if len(parts) > 1 {
		prefix = parts[1]
	}
	return bucketName, prefix
}

func syncFile(ctx context.Context, bucketName, prefix, key string) {
	localFilePath := filepath.Join(localSyncDir, strings.TrimPrefix(key, prefix))
	cloudFrontURL := fmt.Sprintf("%s/%s", cloudFrontDomain, strings.TrimPrefix(key, prefix+"/"))

	// Ensure the directory for the localFilePath exists
	dirPath := filepath.Dir(localFilePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		fmt.Printf("Failed to create directory for file: %s, error: %v\n", localFilePath, err)
		return
	}

	// Check if file exists locally
	info, err := os.Stat(localFilePath)
	if err == nil {
		// File exists, check size
		resp, err := http.Head(cloudFrontURL)
		if err != nil {
			fmt.Printf("Failed to get file header: %v\n", err)
			return
		}
		defer resp.Body.Close()

		remoteFileSize := resp.ContentLength
		if info.Size() == remoteFileSize {
			fmt.Printf("No changes detected, skipping: %s\n", localFilePath)
			return
		}
	}

	// Proceed to download if the file doesn't exist locally or size differs
	fmt.Printf("Downloading: %s\n", cloudFrontURL)
	resp, err := http.Get(cloudFrontURL)
	if err != nil {
		fmt.Printf("Failed to download file: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fileSize := resp.ContentLength
	bar := pb.Full.Start64(fileSize)
	bar.Set(pb.Bytes, true) // Display the progress in bytes

	// Create local file
	out, err := os.Create(localFilePath)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		return
	}
	defer out.Close()

	// Create a proxy reader to track download progress
	proxyReader := bar.NewProxyReader(resp.Body)

	// Write the body to file and track progress
	_, err = io.Copy(out, proxyReader)
	if err != nil {
		fmt.Printf("Failed to write file: %v\n", err)
		return
	}

	bar.Finish()
}
