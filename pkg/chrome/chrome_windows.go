package chrome

// https://groups.google.com/d/msg/golang-nuts/bUetmxErnTw/GHC5obCcmTMJ
// https://play.golang.org/p/fknP9AuLU-
// https://stackoverflow.com/a/60423699

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	CRYPTPROTECT_UI_FORBIDDEN = 0x1
)

var (
	dllcrypt32  = windows.NewLazySystemDLL("Crypt32.dll")
	dllkernel32 = windows.NewLazySystemDLL("Kernel32.dll")

	procDecryptData = dllcrypt32.NewProc("CryptUnprotectData")
	procLocalFree   = dllkernel32.NewProc("LocalFree")
)

type data_blob struct {
	cbData uint32
	pbData *byte
}

func newBlob(d []byte) *data_blob {
	if len(d) == 0 {
		return &data_blob{}
	}
	return &data_blob{
		pbData: &d[0],
		cbData: uint32(len(d)),
	}
}

func (b *data_blob) toByteArray() []byte {
	d := make([]byte, b.cbData)
	copy(d, (*[1 << 30]byte)(unsafe.Pointer(b.pbData))[:])
	return d
}

func decrypt(data []byte) ([]byte, error) {
	// present before Chrome 80
	// 0x01000000D08C9DDF0115D1118C7A00C04FC297EB
	var prefixDPAPI = []byte{1, 0, 0, 0, 208, 140, 157, 223, 1, 21, 209, 17, 140, 122, 0, 192, 79, 194, 151, 235}
	if bytes.HasPrefix(data, prefixDPAPI) {
		return decryptDPAPI(data)
	} else if bytes.HasPrefix(data, []byte(`v10`)) {
		return decryptAES256GCM(data)
	} else {
		return nil, errors.New(`unknown encryption`)
	}
}

func decryptDPAPI(data []byte) ([]byte, error) {
	var outblob data_blob
	r, _, err := procDecryptData.Call(uintptr(unsafe.Pointer(newBlob(data))), 0, 0, 0, 0, CRYPTPROTECT_UI_FORBIDDEN, uintptr(unsafe.Pointer(&outblob)))
	if r == 0 {
		return nil, err
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outblob.pbData)))
	return outblob.toByteArray(), nil
}

func decryptAES256GCM(data []byte) ([]byte, error) {
	// https://stackoverflow.com/a/60423699

	if len(data) < 3+12+16 {
		return nil, errors.New(`encrypted value too short`)
	}

	/* encoded value consists of: {
		"v10" (3 bytes)
		nonce (12 bytes)
		ciphertext (variable size)
		tag (16 bytes)
	}
	*/
	nonce := data[3 : 3+12]
	ciphertextWithTag := data[3+12:]

	if err := getMasterKey(); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}

	// default size for nonce and tag match
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := aesgcm.Open(nil, nonce, ciphertextWithTag, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func decryptValue(encrypted []byte) (string, error) {
	s, err := decrypt(encrypted)
	if err != nil {
		return ``, err
	}
	return string(s), nil
}

func setChromeKeychainPassword(password []byte) []byte {
	return password
}

var masterKey []byte // TODO (?)

func getMasterKey() error {
	// this master key is used globally for all Chrome profiles
	if len(masterKey) > 0 {
		return nil // already set
	}

	// the "Local State" json file is normally one directory above the "Cookies" database
	stateFile := filepath.Join(filepath.Dir(filepath.Dir(dbFile)), `Local State`)

	stateBytes, err := ioutil.ReadFile(stateFile)
	if err != nil {
		return err
	}

	var localState struct {
		OsCrypt struct {
			EncryptedKey string `json:"encrypted_key"`
		} `json:"os_crypt"`
	}
	if err := json.Unmarshal(stateBytes, &localState); err != nil {
		return err
	}

	key, err := base64.StdEncoding.DecodeString(localState.OsCrypt.EncryptedKey)
	if err != nil {
		return err
	}

	if len(key) < 5 || !bytes.HasPrefix(key, []byte(`DPAPI`)) {
		return errors.New(`not an DPAPI key`)
	}
	key = key[5:] // strip "DPAPI"
	key, err = decryptDPAPI(key)
	if err != nil {
		return err
	}
	if len(key) != 32 {
		return errors.New(`master key is not 32 bytes long`)
	}
	masterKey = key

	return nil
}
