package store

import "github.com/google/uuid"

type File struct {
	ID          uuid.UUID `db:"id"`
	FileName    string    `db:"file_name"`
	Description string    `db:"description"`
}

type FileStore interface {
	Files() ([]File, error)
	File(id uuid.UUID) (File, error)
	StoreFile(t *File) error
	DeleteFile(id uuid.UUID) error
}

type Store interface {
	FileStore
}
