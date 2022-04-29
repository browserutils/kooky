package utils

import (
	"bytes"
	"os"
)

type signature struct {
	start int
	sig   []byte
}

var signatures = map[string][]signature{
	`ese`:                {{start: 4, sig: []byte{0xEF, 0xCD, 0xAB, 0x89}}}, // ESE database
	`sqlite`:             {{start: 0, sig: []byte("SQLite format 3\x00")}},  // SQLite 3 database
	`konqueror`:          {{start: 0, sig: []byte("# KDE Cookie File")}},    // Konqueror cookie text file
	`safari`:             {{start: 0, sig: []byte("cook")}},                 // Safari cookie binary file
	`opera_cookies4_1.0`: {{start: 0, sig: []byte{0x00, 0x00, 0x10, 0x00}}}, // Opera Presto cookie binary file (file format v1.0) (cookies4.dat)
	`ie_cache`: {
		{start: 0, sig: []byte("Client UrlCache MMF")}, // Internet Explorer cache file
		{start: 0, sig: []byte("WINE URLCache Ver ")},  // wine index.dat // TODO
	},
	`netscape`: {
		{start: 0, sig: []byte("# Netscape HTTP Cookie File")}, // Netscape cookie text file
		{start: 0, sig: []byte("# HTTP Cookie File")},          // Netscape cookie text file (not strict)
	},
}

func DetectFileType(filename string) (f *os.File, typ string, e error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, ``, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, ``, err
	}
	if fi.IsDir() {
		return f, `text`, nil
	}

	var maxLen int
	for _, sigs := range signatures {
		for _, sig := range sigs {
			l := sig.start + len(sig.sig)
			if l > maxLen {
				maxLen = l
			}
		}
	}
	signature := make([]byte, maxLen)
	if _, err := f.Read(signature); err != nil {
		return nil, ``, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, ``, err
	}

	for name, sigs := range signatures {
		for _, sig := range sigs {
			if bytes.Equal(signature[sig.start:sig.start+len(sig.sig)], sig.sig) {
				return f, name, nil
			}
		}
	}

	return f, `unknown`, nil
}
