package main

// TODO return error when unexpected response is returned - Sometimes the SQL query will expecute succesfully however the stored procedure will return an
// error which is interpreted as a sucesful query
import (
	// GO DEFAULTS
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	//
	"time"
	// GITHUB
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	_ "github.com/denisenkom/go-mssqldb"
)

var (
	//  Setup time based on timezone offset
	aws_timezone                 string = os.Getenv("AWS_TIMEZONE")
	source_db_name               string = os.Getenv("AWS_RDS_DATABASENAMES")
	server                       string = os.Getenv("AWS_RDS_HOSTNAME")
	user                         string = os.Getenv("AWS_RDS_USERNAME")
	port                         string = os.Getenv("AWS_RDS_PORT") // TO DO - Port needs to be an int and not string
	aws_s3_bucket_arn            string = os.Getenv("AWS_S3_BUCKET_ARN")
	encrypted_kms_master_key_arn string = os.Getenv("AWS_KMS_KEY_ARN")
	encrypted_password           string = os.Getenv("AWS_RDS_PASSWORD")
	decrypted                    string
)

type DB_Response struct { //TODO mask KMS key in response
	Col_task_id                  string         `json:"task_id"`
	Col_task_type                string         `json:"task_type"`
	Col_lifecycle                string         `json:"lifecycle"`
	Col_created_at               string         `json:"created_at"`
	Col_last_updated             string         `json:"last_updated"`
	Col_database_name            string         `json:"database_name"`
	Col_s3_object_arn            string         `json:"s3_object_arn"`
	Col_overwrite_S3_backup_file string         `json:"overwrite_S3_backup_file"`
	Col_kms_master_key_arn       string         `json:"kms_master_key_arn"`
	Col_task_progress            string         `json:"task_progress"`
	Col_task_info                sql.NullString `json:"task_info"`
}

// Decrypt AWS Lambda at rest
func AWS_Decrypt(aws_env_var string) (string, error) {
	kmsClient := kms.New(session.New())
	decodedBytes, err := base64.StdEncoding.DecodeString(aws_env_var)
	if err != nil {
		panic(err)
	}
	input := &kms.DecryptInput{
		CiphertextBlob: decodedBytes,
	}
	response, err := kmsClient.Decrypt(input)
	if err != nil {
		panic(err)
	}
	// Plaintext is a byte array, so convert to string
	decrypted = string(response.Plaintext[:])
	return decrypted, nil
}

// func containsEmpty(ss ...string) bool {
// 	for _, s := range ss {
// 		if s == "" {
// 			return true
// 		}
// 	}
// 	return false
// }

type Request struct {
	Databasenames string `json:"databasename"`
	Comment       string `json:"comment"`
}

func Handler(request Request) (DB_Response, error) {
	source_db_name := request.Databasenames
	comment := request.Comment
	// Define Location
	loc, _ := time.LoadLocation(aws_timezone)
	datetime := time.Now().In(loc).Format("2006-01-02_15-04-05")
	// Decrypt the encrypted lambda env vars using a KMS Key defined in the lambda settings
	password, err := AWS_Decrypt(encrypted_password)
	kms_master_key_arn, err := AWS_Decrypt(encrypted_kms_master_key_arn)

	// Check Env vars
	if containsEmpty(server, user, password, port, datetime, kms_master_key_arn, aws_s3_bucket_arn) {
		log.Fatal("ERROR!..Ensure ENV VARS exist AWS_RDS_HOSTNAME AWS_RDS_USERNAME AWS_RDS_PASSWORD AWS_RDS_PORT  AWS_KMS_KEY_ARN AWS_S3_BUCKET_ARN")
	}

	// Define connection string
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s", server, user, password, port)
	log.Println(fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s", server, user, encrypted_password, port))

	conn, err := sql.Open("mssql", connString)
	if err != nil {
		log.Fatal("Open connection failed:", err.Error())
	}
	defer conn.Close()

	// Set context
	ctx := context.Background()

	// Define columns in the respose from the SQL query
	var (
		col_task_id                  string
		col_task_type                string
		col_lifecycle                string
		col_created_at               string
		col_last_updated             string
		col_database_name            string
		col_s3_object_arn            string
		col_overwrite_S3_backup_file string
		col_kms_master_key_arn       string
		col_task_progress            string
		col_task_info                sql.NullString
	)

	// Define slice for db response
	// db_response := []*DB_Response{}

	log.Println(fmt.Sprintf("Backing up DB: %s", source_db_name))
	s3_arn_to_backup_to := fmt.Sprintf("%s/%s-%s-%s.bak", aws_s3_bucket_arn, source_db_name, comment, datetime)
	log.Println(fmt.Sprintf("Backup Comment: %s", comment))

	overwrite_S3_backup_file := 1
	backup_type := "FULL"
	query := fmt.Sprintf("exec msdb.dbo.rds_backup_database \n@source_db_name='%s',\n@s3_arn_to_backup_to='%s',\n@kms_master_key_arn='%s',\n@overwrite_S3_backup_file=%d,\n@type='%s';", source_db_name, s3_arn_to_backup_to, kms_master_key_arn, overwrite_S3_backup_file, backup_type)
	log.Println(query)

	// Execute stored procedure
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&col_task_id,
			&col_task_type,
			&col_lifecycle,
			&col_created_at,
			&col_last_updated,
			&col_database_name,
			&col_s3_object_arn,
			&col_overwrite_S3_backup_file,
			&col_kms_master_key_arn,
			&col_task_progress,
			&col_task_info)
		if err != nil {
			log.Fatal(err)
		}
	}
	// Close connection
	defer conn.Close()

	return DB_Response{
		Col_task_id:                  col_task_id,
		Col_task_type:                col_task_type,
		Col_lifecycle:                col_lifecycle,
		Col_created_at:               col_created_at,
		Col_last_updated:             col_last_updated,
		Col_database_name:            col_database_name,
		Col_s3_object_arn:            col_s3_object_arn,
		Col_overwrite_S3_backup_file: col_overwrite_S3_backup_file,
		Col_kms_master_key_arn:       col_kms_master_key_arn,
		Col_task_progress:            col_task_progress,
		Col_task_info:                col_task_info,
	}, nil

	// return db_response, nil
}

func main() {
	lambda.Start(Handler)
}
