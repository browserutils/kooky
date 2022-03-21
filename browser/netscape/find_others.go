//+build windows darwin plan9 android js aix

package netscape

import "errors"

func netscapeRoots() ([]string, error) {
	return nil, errors.New(`not implemented`)
}
