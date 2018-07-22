package main
// TODO return error when unexpected response is returned - Sometimes the SQL query qill expecute succesfully however the stored procedure will return an
 // error which is interpreted as a sucesful query
import (
  // GO DEFAULTS
  "database/sql"
  "fmt"
  "log"
  "context"
  "time"
  "os"
  // GITHUB
  _ "github.com/denisenkom/go-mssqldb"
  // "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
)

var (
  source_db_name []string
)

type DB_Response struct { //TODO mask KMS key in response
  Col_task_id string
  Col_task_type string
  Col_lifecycle string
  Col_created_at string
  Col_last_updated string
  Col_database_name string
  Col_s3_object_arn string
  Col_overwrite_S3_backup_file string
  Col_kms_master_key_arn string
  Col_task_progress string
  Col_task_info sql.NullString
}

func containsEmpty(ss ...string) bool {
    for _, s := range ss {
        if s == "" {
            return true
        }
    }
    return false
}

type Request struct {
  Databasenames []string `json:"databasename"`
}

  func Handler(request Request) ([]*DB_Response, error) {
 source_db_name := request.Databasenames
 server := os.Getenv("AWS_RDS_HOSTNAME")
 user := os.Getenv("AWS_RDS_USERNAME") // TO DO - Use unicreds to get credentials
 password := os.Getenv("AWS_RDS_PASSWORD") // TO DO - Use unicreds to get credentials
 port := os.Getenv("AWS_RDS_PORT") // TO DO - Port needs to be an int and not string
 datetime := time.Now().Format("2006-01-02_15-04-05")
 aws_s3_bucket_arn := os.Getenv("AWS_S3_BUCKET_ARN")
 kms_master_key_arn := os.Getenv("AWS_KMS_KEY_ARN")

 if containsEmpty(server,user,password,port,datetime) {
   log.Fatal("ERROR!..Ensure ENV VARS exist AWS_RDS_HOSTNAME AWS_RDS_USERNAME AWS_RDS_PASSWORD AWS_RDS_PORT")
 }

 // Define connection string
 connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s", server, user, password, port)
 log.Println(connString)

 conn, err := sql.Open("mssql", connString)
 if err != nil {
  log.Fatal("Open connection failed:", err.Error())
}
defer conn.Close()

 // Set context
 ctx := context.Background()

 // Define columns in the respose from the SQL query
 var (
   col_task_id string
   col_task_type string
   col_lifecycle string
   col_created_at string
   col_last_updated string
   col_database_name string
   col_s3_object_arn string
   col_overwrite_S3_backup_file string
   col_kms_master_key_arn string
   col_task_progress string
   col_task_info sql.NullString
 )

 db_response := []*DB_Response{}

 for _, db_name := range source_db_name {
   log.Println("Backing up DB: %s", db_name)

   s3_arn_to_backup_to := fmt.Sprintf("%s/%s-%s.bak", aws_s3_bucket_arn,db_name,datetime)
   overwrite_S3_backup_file := 1
   backup_type := "FULL"

   query := fmt.Sprintf("exec msdb.dbo.rds_backup_database \n@source_db_name=\"%s\",\n@s3_arn_to_backup_to=\"%s\",\n@kms_master_key_arn=\"%s\",\n@overwrite_S3_backup_file=%d,\n@type=\"%s\";", db_name,s3_arn_to_backup_to,kms_master_key_arn,overwrite_S3_backup_file,backup_type)
   log.Println(query)

   // Execute stored procedure
   rows, err := conn.QueryContext(ctx, query)
   if err != nil {
     log.Fatal(err)
   }
   defer rows.Close()
   for rows.Next() {
     err := rows.Scan(&col_task_id,&col_task_type,&col_lifecycle,&col_created_at,&col_last_updated,&col_database_name,&col_s3_object_arn,&col_overwrite_S3_backup_file,&col_kms_master_key_arn,&col_task_progress,&col_task_info)
     if err != nil {
       log.Fatal(err)
     }
   }
     res := new(DB_Response)
     res.Col_task_id = col_task_id
     res.Col_task_type = col_task_type
     res.Col_lifecycle = col_lifecycle
     res.Col_created_at = col_created_at
     res.Col_last_updated = col_last_updated
     res.Col_database_name = col_database_name
     res.Col_s3_object_arn = col_s3_object_arn
     res.Col_overwrite_S3_backup_file = col_overwrite_S3_backup_file
     res.Col_kms_master_key_arn = col_kms_master_key_arn
     res.Col_task_progress = col_task_progress
     res.Col_task_info = col_task_info
     db_response = append(db_response, res)
     log.Println(col_task_id)
   defer conn.Close()
}

   return db_response, nil
}

func main() {
  lambda.Start(Handler)
}
