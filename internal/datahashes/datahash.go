package datahashes

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
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

func RandBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return ``, err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
