package model

type DatabaseBackup struct {
	DatabaseName string `json:"database_name"`
	DBHost       string `json:"db_host"`
	DBPort       string `json:"db_port"`
	DBUsername   string `json:"db_username"`
	DBPassword   string `json:"db_password"`
	DBFileSQL    string `json:"db_file_sql"`
	DBFileZip    string `json:"db_file_zip"`
	DBError      error  `json:"db_error"`
}
