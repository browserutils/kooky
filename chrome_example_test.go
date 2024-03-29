package kooky_test

import (
	"fmt"
	"os"

	"github.com/browserutils/kooky/browser/chrome"
	"github.com/browserutils/kooky/browser/edge"
)

// on macOS:
var cookieStorePathChrome = "/Google/Chrome/Default/Cookies"
var cookieStorePathMsEdge = "/Microsoft Edge/Default/Cookies"

func Example_chromeSimpleMacOS() {
	// construct file path for the sqlite database containing the cookies
	dir, _ := os.UserConfigDir() // on macOS: "/<USER>/Library/Application Support/"
	cookieStoreFile := dir + cookieStorePathChrome

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

func Example_msEdgeSimpleMacOS() {
	// construct file path for the sqlite database containing the cookies
	dir, _ := os.UserConfigDir() // on macOS: "/<USER>/Library/Application Support/"
	cookieStoreFile := dir + cookieStorePathMsEdge

	// read the cookies from the file
	// decryption is handled automatically
	cookies, err := edge.ReadCookies(cookieStoreFile)
	if err != nil {
		// TODO: handle the error
		return
	}

	for _, cookie := range cookies {
		fmt.Println(cookie)
	}
}
