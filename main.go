package main

import (
	"fmt"
	"sekolahbeta/hacker/cli-app-database-backup/controllers/backup"

	"github.com/joho/godotenv"
)

func Init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("env not found")
	}
}

func main() {
	Init()

	backup.AyokBackup()
}
