package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"log"
	"os"
)

var ErrConflict = errors.New(`conflict`)

type MapDatabase struct {
	db map[string]map[string]string
}

func NewMapDatabase() *MapDatabase {
	return &MapDatabase{db: make(map[string]map[string]string)}
}

func (m *MapDatabase) Create(key, value, userID string) error {
	if _, ok := m.db[key]; ok {
		return ErrConflict
	}
	m.db[key] = map[string]string{
		"URL":    value,
		"userID": userID,
	}
	return nil
}

func (m *MapDatabase) Select(key string) (string, error) {
	if value, ok := m.db[key]; ok {
		return value["URL"], nil
	} else {
		return "", fmt.Errorf("key %s not found", key)
	}
}

func (m *MapDatabase) SelectAll(userID string) map[string]string {
	data := make(map[string]string)
	for key, val := range m.db {
		if val["userID"] == userID {
			data[key] = val["URL"]
		}
	}
	return data
}

type FileDatabase struct {
	path string
	db   RowsFileDatabase
}

type RowFileDatabase struct {
	Hash   string `json:"hash"`
	URL    string `json:"url"`
	UserID string `json:"user_id"`
}

type RowsFileDatabase struct {
	Rows []RowFileDatabase `json:"rows"`
}

func NewFileDatabase(fileName string) (*FileDatabase, error) {
	fileDatabase := &FileDatabase{path: fileName, db: RowsFileDatabase{}}
	err := fileDatabase.sync()
	if err != nil {
		return nil, err
	}
	return fileDatabase, nil
}

func (f *FileDatabase) sync() error {
	file, err := os.OpenFile(f.path, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	if fileInfo.Size() != 0 {
		err = json.NewDecoder(file).Decode(&f.db)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FileDatabase) hasKey(key string) bool {
	for _, rowFile := range f.db.Rows {
		if rowFile.Hash == key {
			return true
		}
	}
	return false
}

func (f *FileDatabase) Create(key, value, userID string) error {
	if f.hasKey(key) {
		return ErrConflict
	}

	file, err := os.OpenFile(f.path, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	defer file.Close()

	f.db.Rows = append(f.db.Rows, RowFileDatabase{Hash: key, URL: value, UserID: userID})

	err = json.NewEncoder(file).Encode(f.db)
	if err != nil {
		return err
	}

	return nil
}

func (f *FileDatabase) Select(key string) (string, error) {
	err := f.sync()
	if err != nil {
		return "", err
	}

	for _, rowVal := range f.db.Rows {
		if rowVal.Hash == key {
			return rowVal.URL, nil
		}
	}

	return "", fmt.Errorf("key %s not found", key)

}

func (f *FileDatabase) SelectAll(userID string) map[string]string {
	m := make(map[string]string)
	for _, userRow := range f.db.Rows {
		if userRow.UserID == userID {
			m[userRow.Hash] = userRow.URL
		}
	}
	return m
}

type PostgresqlDatabase struct {
	cfg Config
}

func NewPostgresqlDatabase(cfg Config) (*PostgresqlDatabase, error) {
	conn, err := pgx.Connect(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())

	_, err = conn.Query(context.Background(), createTable)
	if err != nil {
		return nil, err
	}

	return &PostgresqlDatabase{cfg: cfg}, nil
}

func (p *PostgresqlDatabase) Create(key, value, userID string) error {
	conn, err := pgx.Connect(context.Background(), p.cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	original, err := p.Select(key)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	if original != "" {
		return ErrConflict
	}

	_, err = conn.Query(context.Background(), insert, key, value, userID)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresqlDatabase) Select(key string) (string, error) {
	conn, err := pgx.Connect(context.Background(), p.cfg.DatabaseDSN)
	if err != nil {
		return "", err
	}
	defer conn.Close(context.Background())

	var original string
	err = conn.QueryRow(context.Background(), selectOriginal, key).Scan(&original)
	if err != nil {
		return "", err
	}
	return original, nil
}

func (p *PostgresqlDatabase) SelectAll(userID string) map[string]string {
	conn, err := pgx.Connect(context.Background(), p.cfg.DatabaseDSN)
	if err != nil {
		log.Println(err.Error())
	}
	defer conn.Close(context.Background())

	m := make(map[string]string)
	rows, _ := conn.Query(context.Background(), selectAllUserRows, userID)
	for rows.Next() {
		var hash string
		var original string
		err := rows.Scan(&hash, &original)
		if err != nil {
			log.Println(err.Error())
		}
		m[hash] = original
	}
	return m
}
