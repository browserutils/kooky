//+build !windows

package chrome

import (
	"errors"
)

// TODO: DPAPI offline decryption
// https://elie.net/talk/reversing-dpapi-and-stealing-windows-secrets-offline/
// https://raw.githubusercontent.com/comaeio/OPCDE/master/2017/The%20Blackbox%20of%20DPAPI%20the%20gift%20that%20keeps%20on%20giving%20-%20Bartosz%20Inglot/The%20Blackbox%20of%20DPAPI%20-%20Bart%20Inglot.pdf

func decryptDPAPI(data []byte) ([]byte, error) {
	return nil, errors.New(`DPAPI method not implemented on this platform`)
}
