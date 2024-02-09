# AWS S3 CloudFront Sync

The `aws-s3-cloudfront-sync` tool is designed to synchronize files from an AWS S3 bucket to a local directory, leveraging Amazon CloudFront for enhanced speed and reduced costs. It supports downloading all files within a specified prefix (acting as a directory) from an S3 bucket and ensures that your local machine's directory remains in sync with the S3 bucket.

## Features

- **S3 and CloudFront Integration**: Utilizes S3 for storage and CloudFront for accelerated file delivery.
- **Selective Sync**: Synchronizes files based on the specified prefix within a bucket, allowing for partial bucket sync.
- **Efficient Syncing**: Checks for file existence and size before downloading to avoid unnecessary transfers.
- **Progress Tracking**: Displays download progress for each file using a progress bar.

## Requirements

- Go 1.15 or higher
- AWS SDK for Go v2
- An AWS account and credentials configured (e.g., via AWS CLI)
- An S3 bucket with public access or appropriate bucket policies for CloudFront distribution
- A CloudFront distribution configured to serve files from the S3 bucket

## Installation

Clone this repository or download the source code to your local machine. Then, navigate to the project directory and build the binary:

```sh
go build -o bin/aws-s3-cloudfront-sync
```

## Usage

To use `aws-s3-cloudfront-sync`, you'll need to specify the S3 bucket (and optional prefix), the CloudFront domain mapped to your S3 bucket, and the local directory where the files should be synced to.

```sh
./aws-s3-cloudfront-sync -bucketName "bucket-name/prefix-path" -cloudFrontDomain "https://sub.domain.com" -localSyncDir "./syncedFiles"
```

## Parameters

- bucketName: The S3 bucket and optional prefix in the format 'bucket-name/prefix-path'. The prefix allows for syncing a specific directory within the bucket.
- cloudFrontDomain: The CloudFront domain mapping to the S3 bucket. It should be in the format 'https://sub.domain.com'.
- localSyncDir: The local directory where files will be synchronized to. Defaults to './syncedFiles' if not specified.

## How It Works

The tool performs the following steps:

1. Parses the provided bucket name and prefix.
2. Ensures the specified local directory exists or creates it.
3. Lists all objects in the specified S3 bucket and prefix.
4. For each object, checks if it exists locally and matches the remote size.
5. If the local file is missing or the size differs, downloads the file from CloudFront and updates the local copy.