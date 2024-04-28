# Final Project Markas

This final project is the culmination of the Beta intensive program.

## How to Run the Code

1. Run `go mod tidy` to install all dependencies.
2. Create a `.env` file and fill the key as per the `.env.example` file.
3. Create `db_backup.json` file and fill the key as per the `db_backup.json.example` file.
4. Run `go run main.go` to execute the program


### db_backup.json file example

```json
     {
         "data": [
             {
                 "database_name": "db_2",
                 "latest_backup": {
                     "file_name": "hahaahahs.zip",
                     "id": 2,
                     "timestamp": "2024-04-06T14:10:53.94+08:00"
                 }
             },
             {
                 "database_name": "db_5",
                 "latest_backup": {
                     "file_name": "mysql-2023-10-29-00-00-00-cv_kucing_oren-8634bf3f-23b5-45a7-8b78-fe9b1a3bcf66.sql.zip",
                     "id": 12,
                     "timestamp": "2024-04-08T06:05:20.031+08:00"
                 }
             }
         ],
         "message": "Success"
     }
     ```