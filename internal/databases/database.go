package databases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/salliko/reducer/config"
	"os"
)

var ErrConflict = errors.New(`conflict`)
var ErrGone = errors.New(`Gone`)

type Database interface {
	Create(key, value, userID string) error
	Select(key string) (string, error)
	SelectAll(string) ([]URL, error)
	Close()
	Ping() error
	CreateMany(URL) error
	Flush() error
	Delete(string, string)
}

type URL struct {
	Hash     string `json:"hash"`
	Original string `json:"original"`
	UserID   string `json:"user_id"`
}

type InputURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type OutputURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func hasKey(key string, db []URL) bool {
	for _, row := range db {
		if row.Hash == key {
			return true
		}
	}
	return false
}

func getOriginal(key string, db []URL) (string, error) {
	for _, row := range db {
		if row.Hash == key {
			return row.Original, nil
		}
	}

	return "", fmt.Errorf("key %s not found", key)
}

type MapDatabase struct {
	db []URL
}

func NewMapDatabase() *MapDatabase {
	return &MapDatabase{}
}

func (m *MapDatabase) Close() {
	// Заглушка
}

func (m *MapDatabase) Ping() error {
	return nil
}

func (m *MapDatabase) CreateMany(v URL) error {
	return nil
}

func (m *MapDatabase) Flush() error {
	return nil
}

func (m *MapDatabase) Create(key, value, userID string) error {
	if hasKey(key, m.db) {
		return ErrConflict
	}
	m.db = append(m.db, URL{Hash: key, Original: value, UserID: userID})
	return nil
}

func (m *MapDatabase) Select(key string) (string, error) {
	original, err := getOriginal(key, m.db)
	if err != nil {
		return original, err
	}
	return original, nil
}

func (m *MapDatabase) SelectAll(userID string) ([]URL, error) {
	var data []URL
	for _, val := range m.db {
		if val.UserID == userID {
			data = append(data, val)
		}
	}
	return data, nil
}

func (m *MapDatabase) Delete(key, userID string) {}

type FileDatabase struct {
	path string
	db   []URL
}

func NewFileDatabase(fileName string) (*FileDatabase, error) {
	fileDatabase := &FileDatabase{path: fileName}
	err := fileDatabase.sync()
	if err != nil {
		return nil, err
	}
	return fileDatabase, nil
}

func (f *FileDatabase) Close() {
	// Заглушка
}

func (f *FileDatabase) Ping() error {
	return nil
}

func (f *FileDatabase) CreateMany(v URL) error {
	return nil
}

func (f *FileDatabase) Flush() error {
	return nil
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

func (f *FileDatabase) Create(key, value, userID string) error {
	if hasKey(key, f.db) {
		return ErrConflict
	}

	file, err := os.OpenFile(f.path, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	defer file.Close()

	f.db = append(f.db, URL{Hash: key, Original: value, UserID: userID})

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

	original, err := getOriginal(key, f.db)
	if err != nil {
		return original, err
	}
	return original, nil

}

func (f *FileDatabase) SelectAll(userID string) ([]URL, error) {
	var data []URL
	for _, val := range f.db {
		if val.UserID == userID {
			data = append(data, val)
		}
	}
	return data, nil
}

func (f *FileDatabase) Delete(key, userID string) {}

type PostgresqlDatabase struct {
	conn   *pgxpool.Pool
	buffer []URL
}

func NewPostgresqlDatabase(cfg config.Config) (*PostgresqlDatabase, error) {
	conn, err := pgxpool.Connect(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(context.Background(), createTable)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return &PostgresqlDatabase{conn: conn, buffer: make([]URL, 0, 500)}, nil
}

func (p *PostgresqlDatabase) Close() {
	p.conn.Close()
}

func (p *PostgresqlDatabase) Ping() error {
	return p.conn.Ping(context.Background())
}

func (p *PostgresqlDatabase) Create(key, value, userID string) error {
	original, err := p.Select(key)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	if original != "" {
		return ErrConflict
	}

	rows, err := p.conn.Query(context.Background(), insert, key, value, userID)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func (p *PostgresqlDatabase) CreateMany(value URL) error {
	p.buffer = append(p.buffer, value)

	if cap(p.buffer) == len(p.buffer) {
		err := p.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresqlDatabase) Flush() error {
	tx, err := p.conn.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Запрос prepare
	stmt, err := tx.Prepare(context.Background(), "insert", insert)
	if err != nil {
		return err
	}
	// Закрываем запрос
	for _, v := range p.buffer {
		if _, err := tx.Exec(context.Background(), stmt.SQL, v.Hash, v.Original, v.UserID); err != nil {
			if err = tx.Rollback(context.Background()); err != nil {
				return err
			}
			return err
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	p.buffer = p.buffer[:0]

	return nil
}

func (p *PostgresqlDatabase) Select(key string) (string, error) {
	var original string
	var isDeleted bool
	err := p.conn.QueryRow(context.Background(), selectOriginal, key).Scan(&original, &isDeleted)
	if err != nil {
		return "", err
	}
	if isDeleted {
		return "", ErrGone
	}
	return original, nil
}

func (p *PostgresqlDatabase) SelectAll(userID string) ([]URL, error) {
	var data []URL
	rows, err := p.conn.Query(context.Background(), selectAllUserRows, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var u URL
		err := rows.Scan(&u.Hash, &u.Original, &u.UserID)
		if err != nil {
			return nil, err
		}
		data = append(data, u)
	}
	return data, nil
}

func (p *PostgresqlDatabase) Delete(key, userID string) {
	rows, _ := p.conn.Query(context.Background(), delete, key, userID)
	defer rows.Close()
}
