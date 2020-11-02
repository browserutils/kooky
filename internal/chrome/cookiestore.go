package chrome

import (
	"github.com/go-sqlite/sqlite3"
)

type CookieStore struct {
	Filename         string
	Database         *sqlite3.DbFile
	KeyringPassword  []byte
	Password         []byte
	OS               string
	Browser          string
	Profile          string
	IsDefaultProfile bool
	DecryptionMethod func(data, password []byte) ([]byte, error)
}
