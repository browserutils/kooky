package kooky_test

import (
	"os"

	"github.com/zellyn/kooky"
)

var cookieFile = `cookies.txt`

func ExampleExportCookies() {
	var cookies = []*kooky.Cookie{{Domain: `.test.com`, Name: `test`, Value: `dGVzdA==`}}

	file, err := os.OpenFile(cookieFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		// TODO: handle error
		return
	}
	defer file.Close()

	kooky.ExportCookies(file, cookies)
}
