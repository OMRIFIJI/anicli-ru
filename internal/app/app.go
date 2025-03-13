package app

import (
	"anicliru/internal/db"
	"anicliru/internal/logger"
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
		if err := continuePipe(dbh); err != nil {
			return err
		}
		return nil
	}

	if *deletePtr {
		if err := deletePipe(dbh); err != nil {
			return err
		}
		return nil
	}

	if *deleteAllPtr {
		if err := deleteAllPipe(dbh); err != nil {
			return err
		}
		return nil
	}

	if flag.Parsed() {
		if err := defaultPipe(dbh); err != nil {
			return err
		}
	}
	return nil
}
