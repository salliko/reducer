package main

import (
	"fmt"
)

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
