# MongoDB Dump Scheduler

## Overview

This Go application automates the process of creating MongoDB dumps, compressing them, and uploading them to an AWS S3 bucket. It runs on a schedule (default: every 30 minutes) and is designed to be deployed locally, on an AWS EC2 instance, or in a containerized environment (e.g., Docker).

---

## Features

- **MongoDB Dump**: Creates MongoDB dumps using the `mongodump` utility.
- **Compression**: Compresses dump files into `.tar.gz` format, with the option to use `pigz` for parallel compression.
- **AWS S3 Upload**: Uploads compressed dump files to an AWS S3 bucket.
- **Scheduling**: Configures the dump process to run at regular intervals (default: every 30 minutes).
- **Automatic Cleanup**: Deletes local dump files after successful upload to S3.

---

## Prerequisites

To run this application, you need to have the following libraries and tools installed:

### 1. **Go (Golang)**

Make sure Go is installed on your system (version 1.16+ recommended). You can check this with:

```bash
go version
```

If you need to install Go, follow the [official Go installation guide](https://golang.org/doc/install).

### 2. **MongoDB Database Tools (mongodump)**

Install `mongodump` from the MongoDB database tools package:

- **Ubuntu/Debian**:
  ```bash
  wget -qO - https://www.mongodb.org/static/pgp/server-6.0.asc | sudo apt-key add -
  echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu focal/mongodb-org/6.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-6.0.list
  sudo apt-get update
  sudo apt-get install -y mongodb-database-tools
  ```

Verify the installation:

```bash
mongodump --version
```

### 3. **AWS CLI (for local development)**

If you're running the application locally, install and configure the AWS CLI to manage AWS credentials:

```bash
sudo apt-get install awscli -y
aws configure
```

### 4. **pigz (Optional)**

For faster, parallel compression of dump files, install `pigz`:

```bash
sudo apt-get install pigz -y
```

### 5. **AWS IAM Role (for EC2 deployments)**

If you're deploying the application on an EC2 instance, you'll need to create an IAM role with S3 permissions and attach it to your instance. The role should have access to upload files to the S3 bucket where the dumps will be stored.

---

## Environment Variables

You must configure the application using environment variables. These variables can be set in a `.env` file (for local development) or as system environment variables (for production).

### Example `.env` File

```env
MONGO_URL="mongodb+srv://<username>:<password>@<cluster-url>/<db-name>?tls=true&authSource=admin&replicaSet=<replica-set>"
AWS_REGION="us-east-1"
S3_BUCKET="your-s3-bucket-name"
DUMP_DIR="/tmp/cyclix-dumps"
```

### Required Environment Variables

- **MONGO_URL**: MongoDB connection string for accessing your database.
- **AWS_REGION**: The AWS region where your S3 bucket is located.
- **S3_BUCKET**: The name of your S3 bucket where dump files will be uploaded.
- **DUMP_DIR**: Local directory for temporarily storing MongoDB dumps before uploading them to S3.

---

## Setup and Installation

### 1. **Clone the Repository**

```bash
git clone https://github.com/yourusername/mongo-dump-scheduler.git
cd mongo-dump-scheduler
```

### 2. **Install Dependencies**

Ensure all Go dependencies are installed:

```bash
go mod tidy
```

### 3. **Set Up AWS Credentials**

- **For Local Development**: Set up AWS credentials using `aws configure` or environment variables:

  ```bash
  export AWS_ACCESS_KEY_ID="YOUR_ACCESS_KEY_ID"
  export AWS_SECRET_ACCESS_KEY="YOUR_SECRET_ACCESS_KEY"
  export AWS_REGION="your-aws-region"
  export DUMP_DIR="your-folder-dir"
  ```

- **For EC2 Deployment**: Ensure your EC2 instance is assigned an IAM role with appropriate S3 permissions.

### 4. **Build the Application**

```bash
make build
```

### 5. **Run the Application**

```bash
make run
```

---

## Usage

### Running Locally

1. Ensure AWS credentials are set up (via `aws configure` or environment variables).
2. Start the application:

   ```bash
   make run
   ```

This will start the scheduler, which will create MongoDB dumps and upload them to S3 every 30 minutes by default.

### Running on AWS EC2

1. Attach the IAM role to the EC2 instance.
2. SSH into your EC2 instance, navigate to the project directory, and run:

   ```bash
   make build
   make run
   ```

```bash
docker build -t mongo-dump-scheduler .
```

## Logs and Monitoring

The application logs key events (e.g., dump creation, S3 upload success/failure) to the console.

- **Local Logs**: View directly in the terminal.
- **EC2 Logs**: Use CloudWatch Logs or SSH into the instance to view logs.

---

## Testing

### Local Testing

1. Ensure MongoDB is running with some sample data.
2. Run the application and verify that dump files are created in the `DUMP_DIR` and uploaded to S3.
3. Check S3 to confirm the dumps are uploaded successfully.

### Testing with Large Dumps

1. Populate your MongoDB with a large dataset.
2. Monitor the compression and upload processes for performance.

---

## Error Handling

- **MongoDB Dump Errors**: Ensure `mongodump` is installed and accessible in your `PATH`.
- **S3 Upload Errors**: Ensure AWS credentials are set and have necessary permissions for the S3 bucket.

---

## Libraries and Tools

This project uses several Go libraries and external utilities:

### Go Libraries

- **AWS SDK for Go v2**: Used to interact with AWS services like S3. [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2)
- **robfig/cron**: Used for scheduling tasks like creating MongoDB dumps. [robfig/cron](https://github.com/robfig/cron)
- **Logrus**: Provides structured logging. [Logrus](https://github.com/sirupsen/logrus)
- **godotenv**: Loads environment variables from a `.env` file. [godotenv](https://github.com/joho/godotenv)

### External Utilities

- **mongodump**: From MongoDB Database Tools, used to create MongoDB backups. [MongoDB Database Tools](https://www.mongodb.com/docs/database-tools/)
- **tar and pigz**: Used to compress dump files into `.tar.gz` format, with `pigz` providing parallel compression.
