package client

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/iov-one/weave/crypto"
	"github.com/iov-one/weave/errors"
)

// KeyPerm is the file permissions for saved private keys
const KeyPerm = 0600

// GenPrivateKey creates a new random key.
// Alias to simplify usage.
func GenPrivateKey() *crypto.PrivateKey {
	return crypto.GenPrivKeyEd25519()
}

// DecodePrivateKeyFromSeed decodes private key from seed string
func DecodePrivateKeyFromSeed(hexSeed string) (*crypto.PrivateKey, error) {
	data, err := hex.DecodeString(hexSeed)
	if err != nil {
		return nil, err
	}
	if len(data) != 64 {
		return nil, errors.Wrap(ErrInvalid, "key")
	}
	key := &crypto.PrivateKey{Priv: &crypto.PrivateKey_Ed25519{Ed25519: data}}
	return key, nil
}

// DecodePrivateKey reads a hex string created by EncodePrivateKey
// and returns the original PrivateKey
func DecodePrivateKey(hexKey string) (*crypto.PrivateKey, error) {
	data, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}
	var key crypto.PrivateKey
	err = key.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// EncodePrivateKey stores the private key as a hex string
// that can be saved and later loaded
func EncodePrivateKey(key *crypto.PrivateKey) (string, error) {
	data, err := key.Marshal()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

// LoadPrivateKey will load a private key from a file,
// Which was previously written by SavePrivateKey
func LoadPrivateKey(filename string) (*crypto.PrivateKey, error) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return DecodePrivateKey(string(raw))
}

// SavePrivateKey will encode the private key in hex and write to
// the named file
//
// Refuses to overwrite a file unless force is true
func SavePrivateKey(key *crypto.PrivateKey, filename string, force bool) error {
	if force {
		if fileExists, err := fileExists(filename); fileExists && err != nil {
			return err
		}
		// actually do the write
		hexKey, err := EncodePrivateKey(key)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(filename, []byte(hexKey), KeyPerm)
	}

	return errors.Wrapf(ErrPermission, "file: %s", filename)
}

// LoadPrivateKeys will load an array of private keys from a file,
// Which was previously written by SavePrivateKeys
func LoadPrivateKeys(filename string) ([]*crypto.PrivateKey, error) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var encoded []string
	err = json.Unmarshal(raw, &encoded)
	if err != nil {
		return nil, err
	}

	keys := make([]*crypto.PrivateKey, len(encoded))
	for i, hexKey := range encoded {
		keys[i], err = DecodePrivateKey(hexKey)
		if err != nil {
			return nil, err
		}
	}

	return keys, nil
}

// SavePrivateKeys will encode an array of private keys
// as a json array of hex strings and
// write to the named file
//
// Refuses to overwrite a file unless force is true
func SavePrivateKeys(keys []*crypto.PrivateKey, filename string, force bool) error {
	if force {
		var err error
		encoded := make([]string, len(keys))
		for i, k := range keys {
			encoded[i], err = EncodePrivateKey(k)
			if err != nil {
				return err
			}
		}
		data, err := json.Marshal(encoded)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(filename, data, KeyPerm)
	}

	return errors.Wrapf(ErrPermission, "file: %s", filename)
}

// KeysByAddress takes a list of keys and creates a map
// to look up private keys by their (hex-encoded) address
func KeysByAddress(keys []*crypto.PrivateKey) map[string]*crypto.PrivateKey {
	res := make(map[string]*crypto.PrivateKey, len(keys))
	for _, k := range keys {
		addr := k.PublicKey().Address()
		res[addr.String()] = k
	}
	return res
}

// fileExists returns error if file under given path does not exists.
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return false, errors.Wrapf(ErrPermission, "path: %s", path)
	}
	return true, nil
}
