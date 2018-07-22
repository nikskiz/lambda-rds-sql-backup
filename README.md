# README #

### What is this repository for? ###

* Lambda function written in GO to take a backup of a single or multiple databases within MS SQL
* Backups are encrypted and stored in S3
* Environment variables are required via lambda.
* Version 1

### How do I get set up? ###

#### Summary of set up ####
* Lambda function in golang which utilizes the following imports
      * go-mssqldb - Used for connecting to the MSQL database engine
      * github.com/aws/aws-lambda-go/lambda - Used to handle lambda requests
#### Configuration ####
* Create lambda using GO 1.X
* In AWS setup a Lambda function, ensure the following permissions:
    * Cloudwatch - create groups and streams. Put logs
    * AWS Key Management Service - Access to KMS for the key to encrypt the snapshots
* create a test with the following json. You can replace the Databasename with your own.
```json
{
  "Databasename": [
    "foo",
    "bar"
  ]
}
```
* Configure lambda timeout to nte 30 Seconds
#### Dependencies ####
* N/A
* Deployment instructions
  * Clone the repo and ensure your $GOPATH is set to the 'go' folder
  * Compile binary `go install rds-sql-backup`
  * Copy the binary into a zip and upload to lambda function

### Who do I talk to? ###

* Repo owner or admin
* Nikola Sepentulevski
