package app

import (
	"github.com/OMRIFIJI/anicli-ru/internal/db"
	"github.com/OMRIFIJI/anicli-ru/internal/logger"
	"flag"
	"fmt"

	"rsc.io/getopt"
)

func RunApp() error {
	logger.Init()
	dbh, err := db.NewDBHandler()
	if err != nil {
		return err
	}
	defer dbh.CloseDB()

	flag.Usage = func() {
		fmt.Println("Доступные аргументы командной строки для anicli-ru:")
		getopt.PrintDefaults()
	}

	helpPtr := flag.Bool("help", false, "выводит сообщение, которое вы сейчас видите")
	continuePtr := flag.Bool("continue", false, "продолжить просмотр аниме")
	deletePtr := flag.Bool("delete", false, "удалить запись из базы данных, просматриваемых аниме")
	deleteAllPtr := flag.Bool("delete-all", false, "удалить все записи из базы данных, просматриваемых аниме")
	checkProvidersPtr := flag.Bool("check-providers", false, "проверить, какие источники из конфига доступны")

	getopt.Aliases(
		"c", "continue",
		"h", "help",
		"d", "delete",
	)
	getopt.Parse()

	if *helpPtr {
		flag.Usage()
		return nil
	}

	if *continuePtr {
		return continuePipe(dbh)
	}

	if *deletePtr {
		return deletePipe(dbh)
	}

	if *deleteAllPtr {
		return deleteAllPipe(dbh)
	}

	if *checkProvidersPtr {
		return checkProvidersPipe()
	}

	if flag.Parsed() {
		return defaultPipe(dbh)
	}

	return nil
}
