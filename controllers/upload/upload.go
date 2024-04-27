package upload

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"sekolahbeta/hacker/cli-app-database-backup/model"
	"sync"

	"github.com/sirupsen/logrus"
)

func UploadDatabase(ch <-chan model.DatabaseBackup) <-chan model.DatabaseBackup{
	out := make(chan model.DatabaseBackup)

	go func() {
		for db := range ch {
			file, err := os.Open(fmt.Sprintf("resources/archive/%s", db.DBFileZip))
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

			err = writer.WriteField("database_name", db.DatabaseName)
			if err != nil {
				logrus.Println(err)
			}

			writer.Close()

			r, err := http.NewRequest("POST", fmt.Sprintf("%s:%s/%s", os.Getenv("API_URL"), os.Getenv("API_PORT"), db.DatabaseName), body)
			if err != nil {
				logrus.Println(err)
			}

			r.Header.Add("Content-Type", writer.FormDataContentType())
			r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("SECRET_KEY")))

			client := &http.Client{}
			client.Do(r)

			file.Close()

			out <- db
		}

		close(out)
	}()

	return out
}

func AyokUpload(ch <-chan model.DatabaseBackup, worker int) <-chan model.DatabaseBackup {
	out := make(chan model.DatabaseBackup)
	var chIns []<-chan model.DatabaseBackup

	wg := sync.WaitGroup{}
	wg.Add(worker)

	for i := 0; i < worker; i++ {
		chIns = append(chIns, UploadDatabase(ch))
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	for _, in := range chIns {
		go func(c <-chan model.DatabaseBackup) {
			for cc := range c {
				out <- cc
			}

			wg.Done()
		}(in)
	}
	return out
}
