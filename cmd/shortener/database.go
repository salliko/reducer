package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type MapDatabase struct {
	db map[string]map[string]string
}

//func NewMapDatabase() *MapDatabase {
//	return &MapDatabase{db: make(map[string]map[string]string)}
//}

//func (m *MapDatabase) Create(userID, key, value string) error {
//	if URLS, ok := m.db[userID]; ok {
//		URLS[key] = value
//	} else {
//		m.db[userID] = map[string]string{key: value}
//	}
//	return nil
//}
//
//func (m *MapDatabase) Select(userID, key string) (string, error) {
//	data := m.db[userID]
//	if value, ok := data[key]; ok {
//		return value, nil
//	} else {
//		return "", fmt.Errorf("key %s not found", key)
//	}
//}
//
//func (m *MapDatabase) SelectAll(userID string) map[string]string {
//	return m.db[userID]
//}
//
//type FileDatabase struct {
//	path string
//	db   RowsFileDatabase
//}
//
//type RowURL struct {
//	Hash string `json:"hash"`
//	URL  string `json:"url"`
//}
//
//type RowUser struct {
//	UserID string   `json:"user_id"`
//	URLS   []RowURL `json:"urls"`
//}
//
//type RowsFileDatabase struct {
//	Rows []RowUser `json:"rows"`
//}
//
//func NewFileDatabase(fileName string) (*FileDatabase, error) {
//	fileDatabase := &FileDatabase{path: fileName, db: RowsFileDatabase{}}
//	err := fileDatabase.sync()
//	if err != nil {
//		return nil, err
//	}
//	return fileDatabase, nil
//}
//
//func (f *FileDatabase) sync() error {
//	file, err := os.OpenFile(f.path, os.O_RDONLY|os.O_CREATE, 0777)
//	if err != nil {
//		return err
//	}
//
//	defer file.Close()
//
//	fileInfo, err := file.Stat()
//	if err != nil {
//		return err
//	}
//
//	if fileInfo.Size() != 0 {
//		err = json.NewDecoder(file).Decode(&f.db)
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (f *FileDatabase) hasKey(userID, key string) bool {
//	for _, rowFile := range f.db.Rows {
//		if rowFile.UserID == userID {
//			for _, row := range rowFile.URLS {
//				if row.Hash == key {
//					return true
//				}
//			}
//		}
//	}
//	return false
//}
//
//func (f *FileDatabase) Create(userID, key, value string) error {
//	if f.hasKey(userID, key) {
//		return nil
//	}
//	file, err := os.OpenFile(f.path, os.O_WRONLY|os.O_CREATE, 0777)
//	if err != nil {
//		return err
//	}
//
//	defer file.Close()
//
//	//f.db.Rows = append(f.db.Rows, RowFileDatabase{Hash: key, URL: value})
//	found := false
//	for index, dataUsers := range f.db.Rows {
//		if dataUsers.UserID == userID {
//			f.db.Rows[index].URLS = append(dataUsers.URLS, RowURL{Hash: key, URL: value})
//			log.Println(dataUsers.URLS)
//			found = true
//			break
//		}
//	}
//
//	log.Println(found)
//
//	if !found {
//		f.db.Rows = append(f.db.Rows, RowUser{
//			UserID: userID,
//			URLS: []RowURL{
//				{Hash: key, URL: value},
//			},
//		})
//	}
//
//	err = json.NewEncoder(file).Encode(f.db)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (f *FileDatabase) Select(userID, key string) (string, error) {
//	err := f.sync()
//	if err != nil {
//		return "", err
//	}
//
//	for _, userRow := range f.db.Rows {
//		if userRow.UserID == userID {
//			for _, row := range userRow.URLS {
//				if row.Hash == key {
//					return row.URL, nil
//				}
//			}
//		}
//	}
//
//	return "", fmt.Errorf("key %s not found", key)
//
//}
//
//func (f *FileDatabase) SelectAll(userID string) map[string]string {
//	m := make(map[string]string)
//	for _, userRow := range f.db.Rows {
//		if userRow.UserID == userID {
//			for _, row := range userRow.URLS {
//				m[row.Hash] = row.URL
//			}
//		}
//	}
//	return m
//}

func NewMapDatabase() *MapDatabase {
	return &MapDatabase{db: make(map[string]map[string]string)}
}

func (m *MapDatabase) Create(key, value, userID string) error {
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
		return nil
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
