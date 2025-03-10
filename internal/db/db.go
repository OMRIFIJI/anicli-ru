package db

import (
	"anicliru/internal/api/models"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	bolt "go.etcd.io/bbolt"
)

const animeBucket = "anime"

type DBHandler struct {
	db *bolt.DB
	b  *bolt.Bucket
}

func getDBPath() (string, error) {
	dataHome, err := xdg.DataFile("anicli-ru")
	if err != nil {
		return "", err
	}

	dbPath := filepath.Join(dataHome, "anime.db")
	return dbPath, nil
}

func NewDBHandler() (*DBHandler, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(dbPath)
	// БД не существует - создаем
	if os.IsNotExist(err) {
		db, err := initDB(dbPath)
		if err != nil {
			return nil, err
		}
		return &DBHandler{db: db}, nil
	}

	// БД существует - открываем
	db, err := openDB(dbPath)
	if err != nil {
		return nil, err
	}
	return &DBHandler{db: db}, nil
}

func openDB(dbPath string) (*bolt.DB, error) {
	opts := &bolt.Options{Timeout: 1}
	db, err := bolt.Open(dbPath, 0600, opts)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (dbh *DBHandler) CloseDB() {
	dbh.db.Close()
}

func initDB(dbPath string) (*bolt.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("не удалось создать директорию для базы данных. %s", err)
	}

	db, err := openDB(dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(animeBucket))
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func (dbh *DBHandler) DeleteAnime(title string) error {
	if err := dbh.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(animeBucket))
		if err := b.Delete([]byte(title)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// Вызывается в конце работы программы. Сохраняет информацию об аниме,
// если пользователь его не досмотрел. Если же пользователь досмотрел аниме,
// то функция удалит это аниме из бд.
func (dbh *DBHandler) UpdateAnime(anime *models.Anime) error {
	// Если просмотрено
	if anime.EpCtx.Cur == anime.EpCtx.TotalEpCount {
		if err := dbh.DeleteAnime(anime.Title); err != nil {
			return err
		}
		return nil
	}

	if err := dbh.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(animeBucket))

		// Ничего не делает, если удалять нечего
		if err := b.Delete([]byte(anime.Title)); err != nil {
			return err
		}

		animeJson, err := json.Marshal(anime)
		if err != nil {
			return err
		}

		if err := b.Put([]byte(anime.Title), animeJson); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (dbh *DBHandler) GetAnimeSlice() ([]models.Anime, error) {
	var animeSlice []models.Anime
	if err := dbh.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(animeBucket))

		if err := b.ForEach(func(title, animeJson []byte) error {
			var anime models.Anime
			if err := json.Unmarshal(animeJson, &anime); err != nil {
				return err
			}

			anime.Title = string(title)
			animeSlice = append(animeSlice, anime)
			return nil
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return animeSlice, nil
}

// Возвращает не nil указатель на аниме с названием title из базы данных.
// Если такого аниме нет в базе данных или произошла ошибка, то возвращает nil и ошибку.
func (dbh *DBHandler) GetAnime(title string) (*models.Anime, error) {
	var anime *models.Anime

	if err := dbh.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(animeBucket))
		animeJson := b.Get([]byte(title))
		if animeJson == nil {
			return errors.New("данное аниме не найдено в базе данных")
		}

		if err := json.Unmarshal(animeJson, &anime); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return anime, nil
}

func (dbh *DBHandler) DeleteAllAnime() error {
	if err := dbh.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(animeBucket))

		cursor := bucket.Cursor()
		for key, _ := cursor.First(); key != nil; key, _ = cursor.Next() {
			err := bucket.Delete(key)
			if err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
