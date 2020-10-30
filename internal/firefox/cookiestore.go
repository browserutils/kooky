package firefox

import (
	"github.com/go-sqlite/sqlite3"
)

type CookieStore struct {
	Filename         string
	Database         *sqlite3.DbFile
	Browser          string
	Profile          string
	IsDefaultProfile bool
}
