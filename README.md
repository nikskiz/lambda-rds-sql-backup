# README #

### What is this repository for? ###

* Lambda function written in GO to take a backup of a single database
* Environment variables are required via lambda
* Version 1

### How do I get set up? ###
1. Install from source
```bash
git config --global url."git@bitbucket.org:".insteadOf "https://bitbucket.org/"
go get bitbucket.org/energyonelimited/sage-db-backup
```
2. Create lambda function with the following
    * Permissions
        * S3
        * Cloudwatch
        * VPC
        * KMS
    * Add ENV VARS's
        *  AWS_KMS_KEY_ARN
        *  AWS_RDS_HOSTNAME
        *  AWS_RDS_PASSWORD
        *  AWS_RDS_PORT
        *  AWS_RDS_USERNAME
        *  AWS_TIMEZONE i.e `Australia/Sydney`
    * Encrypt the following ENV VARS's at rest and transit via lambda using a KMS key
        *  AWS_RDS_PASSWORD
        *  AWS_KMS_KEY_ARN
    * Create lambda in Private Subnet. Needs to be behind a NAT that can reach t he AWS endpoints and RDS

3. Create test event with "example_event.json"

### Who do I talk to? ###

* Repo owner or admin
