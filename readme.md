# MongoDB Dump Scheduler

## Overview

**MongoDB Dump Scheduler** is a Go-based application designed to automate the process of creating backups of your MongoDB databases. It performs scheduled dumps, compresses them, and securely uploads the archives to an AWS S3 bucket. This setup ensures that your data is backed up regularly and stored reliably in the cloud, facilitating easy recovery and scalability.

---

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [1. Clone the Repository](#1-clone-the-repository)
  - [2. Install Go](#2-install-go)
  - [3. Install MongoDB Database Tools (`mongodump`)](#3-install-mongodb-database-tools-mongodump)
  - [4. Install `pigz` (Optional for Parallel Compression)](#4-install-pigz-optional-for-parallel-compression)
- [Configuration](#configuration)
  - [1. Environment Variables](#1-environment-variables)
  - [2. AWS Credentials](#2-aws-credentials)
- [Running the Application](#running-the-application)
  - [Using Makefile](#using-makefile)
  - [As a Systemd Service (Optional)](#as-a-systemd-service-optional)
  - [Using Docker (Optional)](#using-docker-optional)
- [Testing](#testing)
  - [1. Functional Testing](#1-functional-testing)
  - [2. Performance Testing](#2-performance-testing)
  - [3. Reliability Testing](#3-reliability-testing)
- [Troubleshooting](#troubleshooting)
- [Security Best Practices](#security-best-practices)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgements](#acknowledgements)
- [Contact](#contact)

---

## Features

- **Scheduled Dumps:** Automatically creates MongoDB dumps at specified intervals using cron expressions.
- **Compression:** Compresses dumps into `.tar.gz` archives for efficient storage and transfer.
- **Parallel Compression:** Utilizes `pigz` for multi-threaded compression, enhancing performance for large datasets.
- **AWS S3 Integration:** Uploads compressed archives to AWS S3 buckets for secure and scalable storage.
- **Logging:** Implements structured logging using `logrus` for easy monitoring and debugging.
- **Retry Mechanism:** Incorporates retry logic with exponential backoff to handle transient failures during uploads.

---

## Prerequisites

Before setting up the MongoDB Dump Scheduler, ensure you have the following prerequisites:

1. **Go (Golang) Installed:**

   - Version: 1.16 or higher.
   - [Download Go](https://golang.org/dl/)

2. **MongoDB Database Tools (`mongodump`):**

   - Essential for creating MongoDB backups.

3. **`pigz` (Optional):**

   - A parallel implementation of gzip for faster compression.

4. **AWS Account:**

   - Access to AWS S3 for storing backups.

5. **AWS CLI (Optional but Recommended):**

   - For configuring AWS credentials.
   - [Install AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html)

6. **Git Installed:**
   - For cloning the repository.
   - [Download Git](https://git-scm.com/downloads)

---

## Installation

### 1. Clone the Repository

Start by cloning the repository to your local machine:

```bash
git clone https://github.com/GiorgiMakharadze/mongo-dump-scheduler.git
cd mongo-dump-scheduler
```

### 2. Install Go

If you haven't installed Go yet, follow these steps:

1. **Download Go:**

   Visit the [official Go download page](https://golang.org/dl/) and download the appropriate installer for your operating system.

2. **Install Go:**

   Follow the installation instructions provided for your OS.

3. **Verify Installation:**

   ```bash
   go version
   ```

   **Expected Output:**

   ```
   go version go1.20.5 linux/amd64
   ```

### 3. Install MongoDB Database Tools (`mongodump`)

`mongodump` is part of the MongoDB Database Tools suite. Follow these steps to install it on Ubuntu/Debian systems:

1. **Import the MongoDB Public GPG Key:**

   ```bash
   wget -qO - https://www.mongodb.org/static/pgp/server-6.0.asc | sudo apt-key add -
   ```

   **Note:** If you receive a warning about `apt-key` being deprecated, proceed, but be aware that future versions may require different steps.

2. **Create a List File for MongoDB:**

   ```bash
   echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu focal/mongodb-org/6.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-6.0.list
   ```

   **Replace `focal` with your Ubuntu codename if you're not using Ubuntu 20.04 (Focal Fossa).** For example, use `jammy` for Ubuntu 22.04.

3. **Update the Package Database:**

   ```bash
   sudo apt-get update
   ```

4. **Install MongoDB Database Tools:**

   ```bash
   sudo apt-get install -y mongodb-database-tools
   ```

5. **Verify Installation:**

   ```bash
   mongodump --version
   ```

   **Expected Output:**

   ```
   mongodump version: x.x.x
   git version: xxxxxxxx
   Go version: go1.xx.x
   ```

### 4. Install `pigz` (Optional for Parallel Compression)

For enhanced compression performance, especially with large datasets, install `pigz`:

```bash
sudo apt-get install pigz -y
```

**Verify Installation:**

```bash
pigz --version
```

**Expected Output:**

```
pigz 2.6
```

---

## Configuration

### 1. Environment Variables

Create a `.env` file in the project's root directory to store configuration variables. This file should contain all necessary environment variables required by the application.

**Create `.env` File:**

```bash
touch .env
```

**Edit `.env` File:**

```env
# MongoDB Connection URI
MONGO_URL="mongodb+srv://doadmin:yourpassword@cluster0.mongodb.net/yourdbname?retryWrites=true&w=majority"

# AWS Configuration
AWS_REGION="us-east-1"                # Replace with your AWS region
S3_BUCKET="your-s3-bucket-name"       # Replace with your S3 bucket name

# Dump Directory
DUMP_DIR="/tmp/cyclix-dumps"          # Temporary directory for dumps

# Schedule (Optional)
# Default schedule is every 30 minutes. You can modify the cron expression as needed.
# For example, "0 0,30 * * * *" runs at minute 0 and 30 of every hour.
SCHEDULE="0 */30 * * * *"
```

**Notes:**

- **Security:** Ensure that the `.env` file is **excluded from version control** to protect sensitive information. Add `.env` to your `.gitignore` file.

  ```gitignore
  # .gitignore
  .env
  ```

### 2. AWS Credentials

The application requires AWS credentials to upload files to S3. There are multiple ways to provide these credentials:

#### A. Using IAM Roles (Recommended for AWS EC2 Instances)

1. **Create an IAM Role with S3 Permissions:**

   - **Navigate to IAM in AWS Console:**

     - Go to [IAM Roles](https://console.aws.amazon.com/iam/home#/roles).
     - Click on **"Create role"**.

   - **Select Trusted Entity:**

     - Choose **"AWS service"**.
     - Select **"EC2"** as the use case.
     - Click **"Next: Permissions"**.

   - **Attach Permissions Policy:**

     - **Option 1:** Attach the managed policy `AmazonS3FullAccess` for testing purposes.
     - **Option 2:** Create a custom policy with least privilege.

     **Example Custom Policy (`S3DumperPolicy`):**

     ```json
     {
       "Version": "2012-10-17",
       "Statement": [
         {
           "Effect": "Allow",
           "Action": ["s3:PutObject", "s3:PutObjectAcl", "s3:ListBucket"],
           "Resource": [
             "arn:aws:s3:::your-s3-bucket-name",
             "arn:aws:s3:::your-s3-bucket-name/*"
           ]
         }
       ]
     }
     ```

     **Replace `your-s3-bucket-name` with the actual name of your S3 bucket.**

   - **Review and Create:**
     - Provide a name for the role, e.g., `EC2MongoDumpS3Access`.
     - Review the role and create it.

2. **Attach the IAM Role to Your EC2 Instance:**

   - **Navigate to EC2 Instances:**

     - Go to the [EC2 Dashboard](https://console.aws.amazon.com/ec2/).
     - Select your instance.

   - **Modify IAM Role:**
     - Click on **"Actions"** > **"Security"** > **"Modify IAM role"**.
     - Select the IAM role you created (`EC2MongoDumpS3Access`).
     - Click **"Update IAM role"**.

3. **Verify IAM Role Assignment:**

   SSH into your EC2 instance and run:

   ```bash
   curl http://169.254.169.254/latest/meta-data/iam/security-credentials/
   ```

   **Expected Output:**

   ```
   EC2MongoDumpS3Access
   ```

   To retrieve the credentials:

   ```bash
   curl http://169.254.169.254/latest/meta-data/iam/security-credentials/EC2MongoDumpS3Access
   ```

   **Note:** Avoid manually handling these credentials; the AWS SDK for Go automatically retrieves them.

#### B. Using AWS Credentials File (For Local Development)

1. **Install AWS CLI (If Not Already Installed):**

   ```bash
   sudo apt-get update
   sudo apt-get install awscli -y
   ```

2. **Configure AWS Credentials:**

   Run the following command and enter your AWS Access Key ID and Secret Access Key when prompted:

   ```bash
   aws configure
   ```

   **This creates a credentials file at `~/.aws/credentials` with the following structure:**

   ```ini
   [default]
   aws_access_key_id = YOUR_ACCESS_KEY_ID
   aws_secret_access_key = YOUR_SECRET_ACCESS_KEY
   ```

3. **Set AWS Region (Optional):**

   You can set the AWS region globally or specify it in the `.env` file.

   ```bash
   aws configure set region us-east-1
   ```

#### C. Using Environment Variables

Alternatively, set AWS credentials using environment variables. This method is useful for containerized environments or temporary credentials.

1. **Set Environment Variables:**

   ```bash
   export AWS_ACCESS_KEY_ID="YOUR_ACCESS_KEY_ID"
   export AWS_SECRET_ACCESS_KEY="YOUR_SECRET_ACCESS_KEY"
   export AWS_REGION="us-east-1" # Replace with your AWS region
   ```

2. **Persist Environment Variables (Optional):**

   Add the above exports to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

   ```bash
   echo 'export AWS_ACCESS_KEY_ID="YOUR_ACCESS_KEY_ID"' >> ~/.bashrc
   echo 'export AWS_SECRET_ACCESS_KEY="YOUR_SECRET_ACCESS_KEY"' >> ~/.bashrc
   echo 'export AWS_REGION="us-east-1"' >> ~/.bashrc
   source ~/.bashrc
   ```

---

## Running the Application

Once you've completed the installation and configuration steps, you're ready to build and run the MongoDB Dump Scheduler.

### Using Makefile

The project includes a `Makefile` to streamline the build and run processes.

1. **Build the Application:**

   ```bash
   make build
   ```

   **This command:**

   - Creates a `bin` directory if it doesn't exist.
   - Compiles the Go application and places the binary in `./bin/mongo-dump-scheduler`.

2. **Run the Application:**

   ```bash
   make run
   ```

   **This command:**

   - Builds the application (if not already built).
   - Executes the binary to start the dump scheduler.

   **Expected Output:**

   ```
   2024/10/15 22:10:00 MongoDB dump scheduler started.
   INFO[2024-10-15T22:10:05+04:00] Successfully created dump at /tmp/cyclix-dumps/2024-10-15/22-10-05
   INFO[2024-10-15T22:10:05+04:00] Successfully uploaded dump to s3://your-s3-bucket-name/2024-10-15/22-10-05.tar.gz
   ```

### As a Systemd Service (Optional)

Running the application as a `systemd` service ensures it starts on boot and restarts on failure.

1. **Create a `systemd` Service File:**

   ```bash
   sudo nano /etc/systemd/system/mongo-dump-scheduler.service
   ```

   **Add the Following Content:**

   ```ini
   [Unit]
   Description=MongoDB Dump Scheduler
   After=network.target

   [Service]
   Type=simple
   User=your-username
   WorkingDirectory=/home/your-username/Desktop/mongodumper
   ExecStart=/home/your-username/Desktop/mongodumper/bin/mongo-dump-scheduler
   Restart=on-failure
   EnvironmentFile=/home/your-username/Desktop/mongodumper/.env

   [Install]
   WantedBy=multi-user.target
   ```

   **Replace `your-username` with your actual username.**

2. **Reload `systemd` and Start the Service:**

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable mongo-dump-scheduler
   sudo systemctl start mongo-dump-scheduler
   ```

3. **Check Service Status:**

   ```bash
   sudo systemctl status mongo-dump-scheduler
   ```

   **Expected Output:**

   ```
   ● mongo-dump-scheduler.service - MongoDB Dump Scheduler
        Loaded: loaded (/etc/systemd/system/mongo-dump-scheduler.service; enabled; vendor preset: enabled)
        Active: active (running) since Thu 2024-10-15 22:10:00 UTC; 5s ago
      Main PID: 12345 (mongo-dump-sched)
         Tasks: 10 (limit: 1153)
        Memory: 20.0M
        CGroup: /system.slice/mongo-dump-scheduler.service
                └─12345 /home/your-username/Desktop/mongodumper/bin/mongo-dump-scheduler
   ```

4. **View Logs:**

   ```bash
   sudo journalctl -u mongo-dump-scheduler -f
   ```

   **This command streams the service logs in real-time.**

### Using Docker (Optional)

Containerizing the application ensures consistency across different environments and simplifies deployment.

1. **Create a `Dockerfile`:**

   ```dockerfile
   # Use official Go image as builder
   FROM golang:1.20 AS builder

   # Set environment variables
   ENV GO111MODULE=on

   # Set working directory
   WORKDIR /app

   # Copy go.mod and go.sum
   COPY go.mod go.sum ./

   # Download dependencies
   RUN go mod download

   # Copy source code
   COPY . .

   # Build the application
   RUN go build -o bin/mongo-dump-scheduler

   # Use minimal base image
   FROM alpine:latest

   # Install required packages
   RUN apk --no-cache add ca-certificates tar gzip pigz

   # Set working directory
   WORKDIR /root/

   # Copy binary from builder
   COPY --from=builder /app/bin/mongo-dump-scheduler /bin/mongo-dump-scheduler

   # Copy .env file
   COPY --from=builder /app/.env .

   # Set entrypoint
   ENTRYPOINT ["./bin/mongo-dump-scheduler"]
   ```

2. **Build the Docker Image:**

   ```bash
   docker build -t mongo-dump-scheduler .
   ```

3. **Run the Docker Container:**

   ```bash
   docker run -d --name mongo-dump-scheduler \
       -v /path/to/.env:/root/.env \
       mongo-dump-scheduler
   ```

   **Replace `/path/to/.env` with the actual path to your `.env` file.**

4. **Verify the Container is Running:**

   ```bash
   docker ps
   ```

   **Expected Output:**

   ```
   CONTAINER ID   IMAGE                   COMMAND                  CREATED          STATUS          PORTS     NAMES
   abcdef123456   mongo-dump-scheduler    "./bin/mongo-dump-…"    10 seconds ago   Up 9 seconds              mongo-dump-scheduler
   ```

5. **View Container Logs:**

   ```bash
   docker logs -f mongo-dump-scheduler
   ```

---

## Testing

Thorough testing ensures that the application functions correctly and efficiently, especially when handling large dumps.

### 1. Functional Testing

- **Basic Dump and Upload:**

  1. **Run the Application:**

     ```bash
     make run
     ```

  2. **Check Temporary Directory:**

     ```bash
     ls -lh /tmp/cyclix-dumps/2024-10-15/
     ```

     **Should List Files Like:**

     ```
     22-10-05.tar.gz
     ```

  3. **Verify S3 Upload:**

     - **Via AWS Console:** Navigate to your S3 bucket and verify the presence of the uploaded `.tar.gz` file.
     - **Via AWS CLI:**

       ```bash
       aws s3 ls s3://your-s3-bucket-name/2024-10-15/
       ```

       **Should List Files Like:**

       ```
       2024-10-15/22-10-05.tar.gz
       ```

### 2. Performance Testing

- **Simulate Large Dumps:**

  1. **Populate MongoDB with Large Data:**

     Insert substantial data into your MongoDB database to simulate a large dump scenario.

  2. **Run the Dump Scheduler:**

     ```bash
     make run
     ```

  3. **Monitor Performance:**

     - **CPU and Memory Usage:**

       ```bash
       htop
       ```

       Ensure that `pigz` is utilizing multiple CPU cores for compression.

  4. **Verify Successful Upload:**

     - Check the S3 bucket for the large `.tar.gz` file.
     - Download and extract it to ensure data integrity.

### 3. Reliability Testing

- **Simulate Network Failures:**

  1. **Disconnect Network During Upload:**

     Temporarily disable network connectivity while the upload is in progress.

  2. **Observe Retry Mechanism:**

     The application should retry the upload based on the implemented retry logic with exponential backoff.

- **Disk Space Management:**

  1. **Monitor Disk Usage:**

     Ensure that temporary dump files are being cleaned up after successful uploads.

     ```bash
     df -h /tmp/cyclix-dumps
     ```

  2. **Attempt Dumps with Low Disk Space:**

     Fill up the disk and observe how the application handles insufficient disk space during dumps.

---

## Troubleshooting

### 1. `mongodump` Not Found

- **Error Message:**

  ```
  mongodump failed: exec: "mongodump": executable file not found in $PATH
  ```

- **Solution:**

  - Ensure that MongoDB Database Tools are installed.
  - Verify installation by running:

    ```bash
    mongodump --version
    ```

### 2. AWS Credential Errors

- **Error Message:**

  ```
  Failed to upload dump to S3: failed to upload file to S3: operation error S3: PutObject, get identity: get credentials: failed to refresh cached credentials, no EC2 IMDS role found, operation error ec2imds: GetMetadata, request canceled, context deadline exceeded
  ```

- **Solution:**

  - **If Running on EC2:** Ensure an IAM role with necessary S3 permissions is attached to the instance.
  - **If Running Locally:** Configure AWS credentials using the AWS CLI (`aws configure`) or set environment variables.
  - **Verify Permissions:** Ensure the IAM role or user has `s3:PutObject` and `s3:ListBucket` permissions for the target bucket.

### 3. Compression Failures

- **Error Message:**

  ```
  failed to compress dump using tar and pigz: exit status 1
  ```

- **Solution:**

  - Ensure that `tar` and `pigz` are installed and accessible in the system's `PATH`.
  - Verify their versions:

    ```bash
    tar --version
    pigz --version
    ```

### 4. S3 Upload Failures

- **Error Message:**

  ```
  failed to upload file to S3 after retries: [error details]
  ```

- **Solution:**

  - Check network connectivity.
  - Verify AWS credentials and permissions.
  - Ensure the S3 bucket exists and is correctly specified in the `.env` file.

### 5. Application Not Starting

- **Error Message:**

  ```
  Failed to create dump directory: [error details]
  ```

- **Solution:**

  - Ensure the specified `DUMP_DIR` has appropriate permissions.
  - Verify that the user running the application has write access to the directory.

---

## Security Best Practices

1. **Secure AWS Credentials:**

   - **Use IAM Roles:** Prefer IAM roles over hardcoding credentials.
   - **Least Privilege:** Assign only necessary permissions to IAM roles or users.
   - **Rotate Credentials:** Regularly rotate AWS access keys to minimize security risks.

2. **Protect Sensitive Data:**

   - **Encrypt Data at Rest:** Enable server-side encryption (SSE) for your S3 bucket.
   - **Encrypt Data in Transit:** Ensure all data transfers use HTTPS.

3. **Restrict S3 Bucket Access:**

   - **Bucket Policies:** Implement strict bucket policies to allow access only from authorized sources.
   - **VPC Endpoints:** Use VPC endpoints to restrict S3 access within your VPC.

4. **Secure Application Logs:**

   - Avoid logging sensitive information such as AWS credentials or database URIs.
   - Implement log rotation and secure storage for logs.

5. **Regular Audits:**
   - Periodically review IAM policies and S3 bucket permissions.
   - Monitor access logs to detect unauthorized access attempts.

---

## Contributing

Contributions are welcome! If you'd like to contribute to the MongoDB Dump Scheduler, please follow these steps:

1. **Fork the Repository:**

   Click on the "Fork" button at the top-right corner of the repository page.

2. **Create a Feature Branch:**

   ```bash
   git checkout -b feature/YourFeatureName
   ```

3. **Commit Your Changes:**

   ```bash
   git commit -m "Add your detailed description of changes"
   ```

4. **Push to Your Fork:**

   ```bash
   git push origin feature/YourFeatureName
   ```

5. **Open a Pull Request:**

   Navigate to the original repository and click on "New pull request". Provide a clear description of your changes and submit the pull request.

---

## License

This project is licensed under the [MIT License](LICENSE).

---

## Acknowledgements

- **[AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/docs/)**
- **[Logrus Logging Library](https://github.com/sirupsen/logrus)**
- **[Robfig Cron](https://github.com/robfig/cron)**
- **[Backoff Retry Library](https://github.com/cenkalti/backoff)**
- **[Pigz - Parallel Gzip](https://github.com/madler/pigz)**

---

## Contact

For any questions, issues, or feature requests, please open an issue on the [GitHub repository](https://github.com/GiorgiMakharadze/mongo-dump-scheduler/issues) or contact [Giorgi Makharadze](mailto:your-email@example.com).

---

**Happy Dumping!**

```

---

### Notes:

- **Replace Placeholder Values:**
  - **`your-s3-bucket-name`**: Replace with your actual S3 bucket name.
  - **`yourpassword`**, **`yourdbname`**: Replace with your MongoDB credentials and database name.
  - **`your-username`**: Replace with your actual system username when configuring `systemd`.
  - **`your-email@example.com`**: Replace with your actual contact email.

- **Security Considerations:**
  - Ensure that the `.env` file is excluded from version control by adding it to `.gitignore`.
  - Use IAM roles with the least privilege necessary when deploying to AWS environments.
  - Regularly rotate your AWS credentials and monitor access logs.

- **Optional Sections:**
  - **Systemd Service** and **Docker** sections are optional and intended for users who prefer running the application as a service or within a containerized environment. You can omit these sections if they are not applicable to your setup.

- **Testing Instructions:**
  - The testing section provides steps to verify the application's functionality, performance, and reliability. It's crucial to perform these tests to ensure the scheduler operates as expected, especially when handling large datasets.

Feel free to modify and extend this `README.md` to better fit your project's specific needs and configurations.
```
