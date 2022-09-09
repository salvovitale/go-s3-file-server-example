package postgres

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/salvovitale/go-s3-file-server-example/internal/store"
)

func NewFileStore(db *sqlx.DB) *FileStore {
	return &FileStore{DB: db}
}

type FileStore struct {
	// embedded structure so we inherit all the methods from it
	*sqlx.DB
}

func (s *FileStore) Files() ([]store.File, error) {
	var fs []store.File
	if err := s.Select(&fs, "SELECT * FROM files"); err != nil {
		return []store.File{}, fmt.Errorf("error getting Files: %w", err)
	}
	return fs, nil
}

func (s *FileStore) File(id uuid.UUID) (store.File, error) {
	var f store.File
	if err := s.Get(&f, "SELECT * FROM files WHERE id = $1", id); err != nil {
		return store.File{}, fmt.Errorf("error getting File: %w", err)
	}
	return f, nil
}

func (s *FileStore) StoreFile(f *store.File) error {
	if err := s.Get(f, "INSERT INTO files VALUES ($1, $2) RETURNING *",
		f.ID,
		f.FileName); err != nil {
		return fmt.Errorf("error creating File: %w", err)
	}
	return nil
}

func (s *FileStore) DeleteFile(id uuid.UUID) error {
	if _, err := s.Exec("DELETE FROM files WHERE id = $1", id); err != nil {
		return fmt.Errorf("error deleting File: %w", err)
	}
	return nil
}
