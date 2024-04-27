package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"sekolahbeta/hacker/cli-app-database-backup/model"
	"sync"

	"github.com/sirupsen/logrus"
)

func ZipDatabase(ch <-chan model.DatabaseBackup) <-chan model.DatabaseBackup {
	out := make(chan model.DatabaseBackup)

	go func() {
		for db := range ch {
			db.DBFileZip = fmt.Sprintf("%s.zip", db.DBFileSQL)
			file, err := os.Create(fmt.Sprintf("resources/archive/%s", db.DBFileZip))
			if err != nil {
				logrus.Println(err)
			}

			zipWriter := zip.NewWriter(file)

			f1, err := os.Open(fmt.Sprintf("resources/sql/%s", db.DBFileSQL))
			if err != nil {
				logrus.Println(err)
			}

			w1, err := zipWriter.Create(db.DBFileZip)
			if err != nil {
				logrus.Println(err)
			}

			if _, err := io.Copy(w1, f1); err != nil {
				logrus.Println(err)
			}

			zipWriter.Close()
			f1.Close()
			file.Close()

			out <- db
		}

		close(out)
	}()

	return out
}

func AyoZip(ch <- chan model.DatabaseBackup, worker int) <-chan model.DatabaseBackup {
	out := make(chan model.DatabaseBackup)
	var chIns []<-chan model.DatabaseBackup

	wg := sync.WaitGroup{}
	wg.Add(worker)

	for i := 0; i < worker; i++ {
		chIns = append(chIns, ZipDatabase(ch))
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