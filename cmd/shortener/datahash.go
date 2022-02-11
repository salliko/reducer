package main

import (
	"crypto/md5"
	"fmt"
)

type Hasing interface {
	Hash([]byte) string
}

type Md5HashData struct {
}

func (m *Md5HashData) Hash(val []byte) string {
	hash := fmt.Sprintf("%x", md5.Sum(val))
	return hash[:6]
}
