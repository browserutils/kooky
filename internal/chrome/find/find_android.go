//go:build android
// +build android

package find

import "errors"

var errNotImplemented = errors.New(`not implemented`)

func chromeRoots(yield func(string, error) bool) {
	// https://chromium.googlesource.com/chromium/src.git/+/62.0.3202.58/docs/user_data_dir.md#android
	if !yield(`/data/user/0/com.android.chrome/app_chrome`, nil) { // TODO check
		return
	}
}

func chromiumRoots(yield func(string, error) bool) { _ = yield(``, errNotImplemented) }
