package firefox

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/pierrec/lz4/v4"
)

var mozLz4Magic = [8]byte{'m', 'o', 'z', 'L', 'z', '4', '0', 0}

func decompressMozLz4(r io.Reader) ([]byte, error) {
	var magic [8]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		return nil, fmt.Errorf("reading mozlz4 magic: %w", err)
	}
	if magic != mozLz4Magic {
		return nil, errors.New("not a mozlz4 file")
	}

	var size uint32
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, fmt.Errorf("reading uncompressed size: %w", err)
	}

	compressed, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading compressed data: %w", err)
	}

	out := make([]byte, size)
	n, err := lz4.UncompressBlock(compressed, out)
	if err != nil {
		return nil, fmt.Errorf("lz4 decompress: %w", err)
	}

	return out[:n], nil
}
