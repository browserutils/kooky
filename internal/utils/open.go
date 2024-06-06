package utils

// OpenFile is like os.Open but might be able to open some locked files.
// The windows implementation sets FILE_SHARE_DELETE in addition.
var OpenFile = openFile
