package cleanup

import (
	"fmt"
	"os"
	"sekolahbeta/hacker/cli-app-database-backup/model"
	"sync"

	"github.com/sirupsen/logrus"
)

func CleanUpFile(ch <-chan model.DatabaseBackup) <-chan model.DatabaseBackup {
	out := make(chan model.DatabaseBackup)

	go func() {
		for db := range ch {

			removeFileSQL(db.DBFileSQL)

			removeFileZip(db.DBFileZip)

			out <- db
		}

		close(out)
	}()

	return out
}

func removeFileSQL(path string) {
	err := os.Remove(fmt.Sprintf("resources/sql/%s", path))
	if err != nil {
		logrus.Println(err)
	}
}

func removeFileZip(path string) {
	err := os.Remove(fmt.Sprintf("resources/archive/%s", path))
	if err != nil {
		logrus.Println(err)
	}
}

func AyokClean(ch <-chan model.DatabaseBackup, worker int) <-chan model.DatabaseBackup {
	out := make(chan model.DatabaseBackup)
	var chIns []<-chan model.DatabaseBackup

	wg := sync.WaitGroup{}
	wg.Add(worker)

	for i := 0; i < worker; i++ {
		chIns = append(chIns, CleanUpFile(ch))
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
