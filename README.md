# Final Project Markas

This final project is the culmination of the Beta intensive program.

## How to Run the Code

1. Run `go mod tidy` to install all dependencies.
2. Create a `.env` file and fill the key as per the `.env.example` file.
3. Create `db_backup.json` file and fill the key as per the `db_backup.json.example` file.
4. Run `go run main.go` to execute the program


### db_backup.json file example
     ```json
     [
        {
        "database_name": "haha_db",
        "db_host": "localhost",
        "db_port": "3306",
        "db_username": "root",
        "db_password": "password123(password may not empty)"
        }
     ]
     ```
