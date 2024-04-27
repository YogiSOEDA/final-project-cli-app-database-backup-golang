package controllers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"sekolahbeta/hacker/cli-app-database-backup/model"
	"sync"
	"time"

	"github.com/google/uuid"
	// "github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
)

// type dbFile struct {
// 	dbName     string
// 	dbFileName string
// }

func dumpDB(chIn <-chan model.DatabaseBackup, chOut chan dbFile, wg *sync.WaitGroup) {
// func dumpDB(chIn <-chan model.DatabaseBackup, chOut chan<- dbFile, errChan chan<- error, wg *sync.WaitGroup) {
	// var resultErr error

	newUUID := uuid.New()
	now := time.Now()

	defer wg.Done()

	for db := range chIn {
		fileName := fmt.Sprintf("msql-%s-%s-%s.sql", now.Format("2006-01-02-15-04-05"), db.DatabaseName, newUUID.String())
		file, err := os.Create(fmt.Sprintf("resources/sql/%s", fileName))
		if err != nil {
			logrus.Println(err)
			// resultErr = multierror.Append(resultErr, err)
		}

		cmd := exec.Command("mysqldump", "-h", db.DBHost, "-P", db.DBPort, "-u", db.DBUsername, fmt.Sprintf("-p%s", db.DBPassword), db.DatabaseName)
		cmd.Stdout = file

		err = cmd.Run()
		if err != nil {
			logrus.Println(err)
			// resultErr = multierror.Append(resultErr, err)
		}

		chOut <- dbFile{
			dbName:     db.DatabaseName,
			dbFileName: fileName,
		}
	}

	// if resultErr != nil {
	// 	errChan <- resultErr
	// }

	// close(chOut)
	// close(errChan)
}

func zipFileSQL(chIn <-chan dbFile, chOut chan dbFile, wg *sync.WaitGroup) {
// func zipFileSQL(chIn <-chan dbFile, chOut chan<- dbFile, errChan chan<- error, wg *sync.WaitGroup) {
	// var resultErr error

	defer wg.Done()

	for sql := range chIn {
		fileName := fmt.Sprintf("%s.zip", sql.dbFileName)
		file, err := os.Create(fmt.Sprintf("resources/archive/%s", fileName))
		if err != nil {
			logrus.Println(err)
			// resultErr = multierror.Append(resultErr, err)
		}

		zipWriter := zip.NewWriter(file)

		f1, err := os.Open(fmt.Sprintf("resources/sql/%s", sql.dbFileName))
		if err != nil {
			logrus.Println(err)
			// resultErr = multierror.Append(resultErr, err)
		}

		w1, err := zipWriter.Create(sql.dbFileName)
		if err != nil {
			logrus.Println(err)
			// resultErr = multierror.Append(resultErr, err)
		}

		if _, err := io.Copy(w1, f1); err != nil {
			logrus.Println(err)
			// resultErr = multierror.Append(resultErr, err)
		}

		zipWriter.Close()

		chOut <- dbFile{
			dbFileName: fileName,
			dbName:     sql.dbName,
		}
	}

	// if resultErr != nil {
	// 	errChan <- resultErr
	// }

	// close(chOut)
	// close(errChan)
}

func uploadZip(chIn <-chan dbFile, wg *sync.WaitGroup) {
// func uploadZip(chIn <-chan dbFile, errChan chan<- error, wg *sync.WaitGroup) {
	// var resultErr error

	defer wg.Done()

	for zip := range chIn {
		file, err := os.Open(fmt.Sprintf("resources/archive/%s", zip.dbFileName))
		if err != nil {
			logrus.Println(err)

			// resultErr = multierror.Append(resultErr, err)
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file_name", file.Name())
		if err != nil {
			logrus.Println(err)

			// resultErr = multierror.Append(resultErr, err)
		}

		io.Copy(part, file)

		err = writer.WriteField("database_name", zip.dbName)
		if err != nil {
			logrus.Println(err)

			// resultErr = multierror.Append(resultErr, err)
		}

		writer.Close()

		r, err := http.NewRequest("POST", fmt.Sprintf("%s:%s/%s", os.Getenv("API_URL"), os.Getenv("API_PORT"), zip.dbName), body)
		if err != nil {
			logrus.Println(err)

			// resultErr = multierror.Append(resultErr, err)
		}

		r.Header.Add("Content-Type", writer.FormDataContentType())
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("SECRET_KEY")))

		client := &http.Client{}
		client.Do(r)
	}

	// if resultErr != nil {
	// 	errChan <- resultErr
	// }

	// close(errChan)
}

func BackupDB() {
	var listDB []model.DatabaseBackup

	dataJson, err := os.ReadFile("./resources/json/db_backup.json")
	if err != nil {
		logrus.Fatalf("Gagal memuat file json, error : %s", err.Error())
	}

	err = json.Unmarshal(dataJson, &listDB)
	if err != nil {
		logrus.Printf("Gagal encoded json, error : %s", err.Error())
	}

	fmt.Println(listDB)

	_ = os.Mkdir("resources/sql", 0777)
	_ = os.Mkdir("resources/archive", 0777)

	ch := make(chan model.DatabaseBackup)
	chSql := make(chan dbFile)
	chZip := make(chan dbFile)
	// errors := make(chan error)
	wg := sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		// go dumpDB(ch, chSql, errors, &wg)
		go dumpDB(ch, chSql, &wg)
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		// go zipFileSQL(chSql, chZip, errors, &wg)
		go zipFileSQL(chSql, chZip, &wg)
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		// go uploadZip(chZip, errors, &wg)
		go uploadZip(chZip, &wg)
	}

	// go func() {
	// 	defer close(ch)
	// 	defer close(errors)

	for _, db := range listDB {
		ch <- db
	}
	// }()
	close(ch)
	wg.Wait()
	close(chSql)
	close(chZip)
	// close(errors)
	// for err := range errors {
	// 	logrus.Println(err)
	// }
}
