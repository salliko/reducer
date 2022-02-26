package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type MapDatabase struct {
	db map[string]string
}

func NewMapDatabase() *MapDatabase {
	return &MapDatabase{db: make(map[string]string)}
}

func (m *MapDatabase) Create(key, value string) error {
	m.db[key] = value
	return nil
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

func (f *FileDatabase) Create(key, value string) error {
	if f.hasKey(key) {
		return nil
	}

	file, err := os.OpenFile(f.path, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	defer file.Close()

	f.db.Rows = append(f.db.Rows, RowFileDatabase{Hash: key, URL: value})

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
