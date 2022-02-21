package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Database interface {
	Create(key, value string)
	Select(key string) (string, error)
}

type MapDatabase struct {
	db map[string]string
}

func NewMapDatabase() *MapDatabase {
	return &MapDatabase{db: make(map[string]string)}
}

func (m *MapDatabase) Create(key, value string) {
	m.db[key] = value
}

func (m *MapDatabase) Select(key string) (string, error) {
	if value, ok := m.db[key]; ok {
		return value, nil
	} else {
		return "", fmt.Errorf("key %s not found", key)
	}
}

type FileDatabase struct {
	path string
	db   RowsFileDatabase
}

type RowFileDatabase struct {
	Hash string `json:"hash"`
	URL  string `json:"url"`
}

type RowsFileDatabase struct {
	Rows []RowFileDatabase `json:"rows"`
}

func NewFileDatabase(fileName string) *FileDatabase {
	return &FileDatabase{path: fileName, db: RowsFileDatabase{}}
}

func (f *FileDatabase) Create(key, value string) {
	file, err := os.OpenFile(f.path, os.O_WRONLY|os.O_CREATE, 0777)
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}

	for _, rowFile := range f.db.Rows {
		if rowFile.Hash == key {
			return
		}
	}

	f.db.Rows = append(f.db.Rows, RowFileDatabase{Hash: key, URL: value})

	err = json.NewEncoder(file).Encode(f.db)
	if err != nil {
		log.Fatal(err)
	}
}

func (f *FileDatabase) Select(key string) (string, error) {
	file, err := os.OpenFile(f.path, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return "", err
	}

	err = json.NewDecoder(file).Decode(&f.db)
	if err != nil {
		log.Fatal(err)
	}

	for _, rowVal := range f.db.Rows {
		if rowVal.Hash == key {
			return rowVal.URL, nil
		}
	}

	return "", fmt.Errorf("key %s not found", key)

}
