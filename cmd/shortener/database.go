package main

import (
	"fmt"
)

type DatabaseManager struct {
	data map[string]string
}

type MapDatabase struct {
	db *DatabaseManager
}

func (d *DatabaseManager) Insert(key, value string) {
	if d.data == nil {
		d.data = make(map[string]string)
	}
	d.data[key] = value
}

func (d *DatabaseManager) Select(id string) (string, error) {
	if value, ok := d.data[id]; ok {
		return value, nil
	} else {
		return "", fmt.Errorf("key %s not found", id)
	}
}

func (m *MapDatabase) Create(key, value string) {
	m.db.Insert(key, value)
}

func (m *MapDatabase) Select(key string) (string, error) {
	value, err := m.db.Select(key)
	return value, err
}
