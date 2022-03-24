package edge

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/zellyn/kooky"
	"github.com/zellyn/kooky/internal"
	"github.com/zellyn/kooky/internal/chrome"
	"www.velocidex.com/golang/go-ese/parser"

	"github.com/go-sqlite/sqlite3"
)

type edgeCookieStore struct {
	internal.DefaultCookieStore
	chrome.CookieStore
	ESECatalog *parser.Catalog
}

var _ kooky.CookieStore = (*edgeCookieStore)(nil)

func (s *edgeCookieStore) Open() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}
	if s.File != nil {
		s.File.Seek(0, 0)
		return nil
	}
	if s.ESECatalog != nil || s.Database != nil {
		return nil
	}
	if f, err := os.Open(s.FileNameStr); err != nil {
		return err
	} else {
		s.File = f
	}

	signature := make([]byte, 16)
	if _, err := s.File.Read(signature); err != nil {
		return err
	}
	if _, err := s.File.Seek(0, 0); err != nil {
		return err
	}
	if s.ESECatalog == nil {
		// In the file header of the database we find that the first 4 bytes are a XOR checksum.
		// The following 4 bytes after the checksum is a file signature. The file signature has
		// offset 4, and the value is EF CD AB 89.

		var signatureESEdatabase = []byte{239, 205, 171, 137} // EF CD AB 89
		if bytes.Equal(signature[4:8], signatureESEdatabase) {
			// file is an ESE database

			// TODO: create temporary copy of the file on Windows - a service on Windows has a permanent lock on it
			// TODO: remove temporary copy in Close()

			ese_ctx, err := parser.NewESEContext(s.File)
			if err != nil {
				return err
			}

			catalog, err := parser.ReadCatalog(ese_ctx)
			if err != nil {
				return err
			}

			s.ESECatalog = catalog

			return nil
		}
	}

	if s.Database == nil {
		var signatureSQLite3database = []byte("SQLite format 3\x00")
		if bytes.Equal(signature[0:16], signatureSQLite3database) {
			// file is an SQLite 3 database

			db, err := sqlite3.Open(s.FileNameStr)
			if err != nil {
				return err
			}

			s.Database = db

			return nil
		}
	}

	// file is probably a cookie text file (*.txt)

	return nil
}

func (s *edgeCookieStore) Close() error {
	if s == nil {
		return errors.New(`cookie store is nil`)
	}

	var err, errFile, errDB error

	if s.File != nil {
		errFile = s.File.Close()
		s.File = nil
	}
	if s.Database != nil {
		errDB = s.Database.Close()
		s.Database = nil
	}
	s.ESECatalog = nil

	if errFile != nil && errDB == nil {
		err = errFile
	} else if errFile == nil && errDB != nil {
		err = errDB
	} else if errFile != nil && errDB != nil {
		err = fmt.Errorf("os.File.Close() error \"%v\" and github.com/go-sqlite/sqlite3.DbFile.Close() error \"%v\" occurred", errFile, errDB)
	}

	return err
}
