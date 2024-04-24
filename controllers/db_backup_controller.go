package controllers

import (
	"archive/zip"
	"bytes"

	"encoding/json"

	// "encoding/json"
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
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func Init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("env not found")
	}
}

// func openFile(path string) (<-chan model.DatabaseBackup, error) {
// 	var listDB []model.DatabaseBackup

// 	dbChan := make(chan model.DatabaseBackup)

// 	records, err := os.ReadFile(path)
// 	if err != nil {
// 		return dbChan, err
// 	}

// 	err = json.Unmarshal(records, &listDB)
// 	if err != nil {
// 		return dbChan, err
// 	}

// 	go func() {
// 		for _, db := range listDB {
// 			dbChan <- db
// 		}
// 	}()

// 	return dbChan, nil
// }

func dumpDatabase(listDB <-chan model.DatabaseBackup) <-chan dbFile {
	// func dumpDatabase(listDB <-chan model.DatabaseBackup) (<-chan model.DatabaseBackup, <-chan string) {
	// dbChan := make(chan model.DatabaseBackup)
	dbChan := make(chan dbFile)

	// fileChan := make(chan string)

	newUUID := uuid.New()
	now := time.Now()

	go func() {
		for db := range listDB {
			fileName := fmt.Sprintf("msql-%s-%s-%s.sql", now.Format("2006-01-02-15-04-05"), db.DatabaseName, newUUID.String())
			file, err := os.Create(fmt.Sprintf("resources/sql/%s", fileName))
			if err != nil {
				// logrus.Panic()
				logrus.Println(err)
				// fmt.Println(err)
			}

			// defer file.Close()

			cmd := exec.Command("mysqldump", "-h", db.DBHost, "-P", db.DBPort, "-u", db.DBUsername, fmt.Sprintf("-p%s", db.DBPassword), db.DatabaseName)
			cmd.Stdout = file

			err = cmd.Run()
			if err != nil {
				fmt.Println(err)
			}

			dbChan <- dbFile{
				dbName:     db.DatabaseName,
				dbFileName: fileName,
			}
			// fileChan <- fileName
			// fileChan <- fileName
		}

		close(dbChan)
		// close(fileChan)
	}()

	// go func() {
	// 	for _, db := range listDB {
	// 		dbChan <- db
	// 	}

	// 	close(dbChan)
	// }()

	return dbChan //, fileChan
}

func appendDB(dbChanMany ...<-chan dbFile) <-chan dbFile {
	// func appendDB(dbChanMany ...<-chan model.DatabaseBackup) <-chan model.DatabaseBackup {
	wg := sync.WaitGroup{}

	mergedChan := make(chan dbFile)
	// mergedChan := make(chan model.DatabaseBackup)

	wg.Add(len(dbChanMany))
	for _, ch := range dbChanMany {
		go func(ch <-chan dbFile) {
			// go func(ch <-chan model.DatabaseBackup) {
			for db := range ch {
				mergedChan <- db
			}
			wg.Done()
		}(ch)
	}

	go func() {
		wg.Wait()
		close(mergedChan)
	}()

	return mergedChan
}

func zipFile(sqlChan <-chan dbFile) <-chan dbFile {

	dbChan := make(chan dbFile)

	go func() {
		for sql := range sqlChan {
			fileName := fmt.Sprintf("%s.zip", sql.dbFileName)
			file, err := os.Create(fmt.Sprintf("resources/archive/%s", fileName))
			if err != nil {
				logrus.Println(err)
			}

			zipWriter := zip.NewWriter(file)

			f1, err := os.Open(fmt.Sprintf("resources/sql/%s", sql.dbFileName))
			if err != nil {
				logrus.Println(err)
			}

			w1, err := zipWriter.Create(sql.dbFileName)
			if err != nil {
				logrus.Println(err)
			}

			if _, err := io.Copy(w1, f1); err != nil {
				logrus.Println(err)
			}

			zipWriter.Close()

			dbChan <- dbFile{
				dbFileName: fileName,
				dbName:     sql.dbName,
			}
		}

		close(dbChan)

	}()

	return dbChan
	// func zipFile(sqlChan <-chan model.DatabaseBackup) <-chan model.DatabaseBackup {
	// dbChan := make(chan model.DatabaseBackup)

	// newUUID := uuid.New()
	// now := time.Now()

	// // go func() {
	// 	fileName := fmt.Sprintf("msql-%s-%s-%s.sql", now.Format("2006-01-02-15-04-05"), sqlChan.DatabaseName, newUUID.String())
	// 	file, err := os.Create(fmt.Sprintf("./resources/sql/%s", fileName))
	// 	if err != nil {
	// 		// logrus.Panic()
	// 		logrus.Println(err)
	// 		// fmt.Println(err)
	// 	}
	// }()

}

// func ahay(awo <-chan model.DatabaseBackup)  {

// }

type dbFile struct {
	dbName     string
	dbFileName string
}

func sendDataAPI(zipChan <-chan dbFile, wg *sync.WaitGroup) {
// func sendDataAPI(zipChan <-chan dbFile) {

	go func() {
		for zip := range zipChan {
			file, err := os.Open(fmt.Sprintf("resources/archive/%s", zip.dbFileName))
			if err != nil {
				logrus.Println(err)
			}

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, err := writer.CreateFormFile("file_name", file.Name())
			if err != nil {
				logrus.Println(err)
			}

			io.Copy(part, file)
			err = writer.WriteField("database_name", zip.dbName)
			if err != nil {
				logrus.Println(err)
			}
			writer.Close()

			// r, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:3000/%s", zip.dbName), body)
			r, err := http.NewRequest("POST", fmt.Sprintf("%s:%s/%s", os.Getenv("API_URL"), os.Getenv("API_PORT"), zip.dbName), body)
			if err != nil {
				logrus.Println(err)
			}

			r.Header.Add("Content-Type", writer.FormDataContentType())
			r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("SECRET_KEY")))

			client := &http.Client{}
			client.Do(r)
		}

		wg.Done()
	}()
}

func BackupDatabase() {
	Init()
	// var dbChanTemp <-chan model.DatabaseBackup

	// dataJson, err := openFile("./resources/json/db_backup.json")
	// if err != nil {
	// 	logrus.Fatalf("Gagal memuat file json, error : %s", err.Error())
	// }

	// for i := 0; i < 5; i++ {
	// 	dbChanTemp = append(dbChanTemp, )
	// }
 //DUAAAAAAAAAAAAAAAAAAAAR
	var listDB []model.DatabaseBackup

	dataJson, err := os.ReadFile("./resources/json/db_backup.json")
	if err != nil {
		logrus.Fatalf("Gagal memuat file json, error : %s", err.Error())
	}

	err = json.Unmarshal(dataJson, &listDB)
	if err != nil {
		logrus.Printf("Gagal encoded json, error : %s", err.Error())
	}


	_ = os.Mkdir("resources/sql", 0777)

	var temp []<-chan dbFile

	ch := make(chan model.DatabaseBackup)

	for i := 0; i < 5; i++ {
		temp = append(temp, dumpDatabase(ch))
	}

	for _, db := range listDB {
		ch <- db
	}

	close(ch)

	a := appendDB(temp...)

	var zipChMany []<-chan dbFile
	zipCh := make(chan dbFile)
	for i := 0; i < 5; i++ {
		zipChMany = append(zipChMany, zipFile(zipCh))
	}

	for zip := range a {
		zipCh <- zip
	}

	fsd := appendDB(zipChMany...)

	apiCh := make(chan dbFile)
	wg := sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go sendDataAPI(apiCh, &wg)
	}

	for v := range fsd {
		apiCh <- v
		// fmt.Println(v)
		// fmt.Println(v.dbFileName)
		// fmt.Println(v.dbName)
	}

	close(apiCh)
	wg.Wait()
	
	// ehe := "msql-2024-04-22-21-08-05-db_ar_inv-0311ea05-5155-4b1c-968a-9f644176f0d8.sql.zip"
	// ds := "db_ar_inv"

	// file, err := os.Open(fmt.Sprintf("resources/archive/%s", ehe))
	// if err != nil {
	// 	logrus.Println(err)
	// }

	// defer file.Close()

	// body := &bytes.Buffer{}
	// writer := multipart.NewWriter(body)
	// part, err := writer.CreateFormFile("file_name", file.Name())
	// if err != nil {
	// 	logrus.Println(err)
	// }

	// io.Copy(part, file)
	// err = writer.WriteField("database_name", ds)
	// if err != nil {
	// 	logrus.Println(err)
	// }
	
	// writer.Close()
	// // r, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:3000/%s", zip.dbName), body)
	// r, err := http.NewRequest("POST", fmt.Sprintf("%s:%s/%s", os.Getenv("API_URL"), os.Getenv("API_PORT"), ds), body)
	// // r, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:3000/%s", ds), body)
	// if err != nil {
	// 	logrus.Println(err)
	// }
	
	// r.Header.Add("Content-Type", writer.FormDataContentType())
	// r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("SECRET_KEY")))
	
	// client := &http.Client{}
	// asep, err := client.Do(r)
	// if err != nil {
	// 	logrus.Println(err)
	// }

	// // defer asep.Body.Close()
	
	// fmt.Println(asep)
	close(apiCh)

	// wg.Wait()

	fmt.Println("Udahan Bang")

	// var sd []dbFile
	// var sd []model.DatabaseBackup
	// for v := range fsd {
	// 	sd = append(sd, v)
	// }

	// close(ch)

	// fmt.Println(sd)

	// for _, db := range listDB {
	// 	ch <- db
	// }

	// newUUID := uuid.New()
	// now := time.Now()

	// for _, db := range listDB {

	// 	fileName := fmt.Sprintf("msql-%s-%s-%s.sql", now.Format("2006-01-02-15-04-05"), db.DatabaseName, newUUID.String())
	// 	file, err := os.Create(fmt.Sprintf("./resources/sql/%s", fileName))
	// 	if err != nil {
	// 		// logrus.Panic()
	// 		logrus.Println(err)
	// 		// fmt.Println(err)
	// 	}

	// 	defer file.Close()

	// 	cmd := exec.Command("mysqldump", "-h", db.DBHost, "-P", db.DBPort, "-u", db.DBUsername, fmt.Sprintf("-p%s", db.DBPassword), db.DatabaseName)
	// 	cmd.Stdout = file

	// 	err = cmd.Run()
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }

	// var dbChanTemp []<-chan model.DatabaseBackup

	// for i := 0; i < 5; i++ {
	// 	dbChanTemp = append(dbChanTemp, dumpDatabase(listDB))
	// }

	// mergeCh := appendDB(dbChanTemp...)

	// fmt.Println(mergeCh)

	// for _, v := range listDB {
	// 	fmt.Println(v)
	// }
	// fmt.Println(listDB)

	// var dbChanTemp []<-chan model.DatabaseBackup
}

// func dumpDatabase()  {

// }
