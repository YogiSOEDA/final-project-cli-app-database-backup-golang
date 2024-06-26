package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"sekolahbeta/hacker/cli-app-database-backup/controllers/cleanup"
	"sekolahbeta/hacker/cli-app-database-backup/controllers/dump"
	"sekolahbeta/hacker/cli-app-database-backup/controllers/upload"
	"sekolahbeta/hacker/cli-app-database-backup/controllers/zip"
	"sekolahbeta/hacker/cli-app-database-backup/model"

	"github.com/sirupsen/logrus"
)

func AyokBackup() {
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
	_ = os.Mkdir("resources/archive", 0777)

	dump := dump.AyokDump(dumpAja(listDB), 3)

	zip := zip.AyoZip(dump, 3)

	upload := upload.AyokUpload(zip, 5)

	clean := cleanup.AyokClean(upload, 1)

	for v := range clean {
		fmt.Println(v)
	}
}

func dumpAja(list []model.DatabaseBackup) <-chan model.DatabaseBackup {
	ch := make(chan model.DatabaseBackup)

	go func() {
		for _, source := range list {
			ch <- source
		}

		close(ch)
	}()

	return ch
}
