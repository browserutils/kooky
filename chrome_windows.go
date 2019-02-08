package kooky

/*
#cgo LDFLAGS: -lCrypt32
#define NOMINMAX
#include <windows.h>
#include <Wincrypt.h>
char* decrypt(byte* in, int len, int *outLen) {
	DATA_BLOB input, output;
	LPWSTR pDescrOut =  NULL;
	input.cbData = len;
	input.pbData = in;
	CryptUnprotectData(
		&input,
		&pDescrOut,
		NULL,                 // Optional entropy
		NULL,                 // Reserved
		NULL,                 // Here, the optional
                              // prompt structure is not
                              // used.
		0,
		&output);
	*outLen = output.cbData;
	return output.pbData;
}
void doFree(char* ptr) {
	free(ptr);
}
*/
import "C"

func setChromeKeychainPassword(password []byte) []byte {
	return password
}

func decryptValue(input []byte) (string, error) {
	var length C.int
	decruptedC := C.decrypt((*C.byte)(&input[0]), C.int(len(input)), &length)
	decrypted := C.GoStringN(decruptedC, length)
	return decrypted, nil
}
