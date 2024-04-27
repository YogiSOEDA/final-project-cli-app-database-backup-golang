package dump

import (
	"fmt"
	"os"
	"os/exec"
	"sekolahbeta/hacker/cli-app-database-backup/model"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func DumpDatabase(ch <-chan model.DatabaseBackup) <-chan model.DatabaseBackup {
	out := make(chan model.DatabaseBackup)

	newUUID := uuid.New()
	now := time.Now()

	go func() {
		for db := range ch {
			db.DBFileSQL = fmt.Sprintf("msql-%s-%s-%s.sql", now.Format("2006-01-02-15-04-05"), db.DatabaseName, newUUID)
			file, err := os.Create(fmt.Sprintf("resources/sql/%s", db.DBFileSQL))
			if err != nil {
				logrus.Println(err)
			}

			cmd := exec.Command("mysqldump", "-h", db.DBHost, "-P", db.DBPort, "-u", db.DBUsername, fmt.Sprintf("-p%s", db.DBPassword), db.DatabaseName)
			cmd.Stdout = file

			err = cmd.Run()
			if err != nil {
				logrus.Println(err)
			}
			
			file.Close()

			out <- db
		}

		close(out)
	}()

	return out
}

func AyokDump(ch <-chan model.DatabaseBackup, worker int) <-chan model.DatabaseBackup {
	out := make(chan model.DatabaseBackup)
	var chIns []<-chan model.DatabaseBackup

	wg := sync.WaitGroup{}
	wg.Add(worker)

	for i := 0; i < worker; i++ {
		chIns = append(chIns, DumpDatabase(ch))
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