package kooky_test

import (
	"fmt"
	"os"

	"github.com/browserutils/kooky/browser/chrome"
)

// on macOS:
var cookieStorePath = "/Google/Chrome/Default/Cookies"

func Example_chromeSimpleMacOS() {
	// construct file path for the sqlite database containing the cookies
	dir, _ := os.UserConfigDir() // on macOS: "/<USER>/Library/Application Support/"
	cookieStoreFile := dir + cookieStorePath

	// read the cookies from the file
	// decryption is handled automatically
	cookies, err := chrome.ReadCookies(cookieStoreFile)
	if err != nil {
		// TODO: handle the error
		return
	}

	for _, cookie := range cookies {
		fmt.Println(cookie)
	}
}
