package crypto

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"kusionstack.io/kusion/pkg/util/kfile"
)

const (
	DefaultDotKeys  = ".keys"
	DefaultIDRsa    = "id_rsa"
	DefaultIDRsaPub = "id_rsa.pub"
	DefaultKeySize  = 1024
)

var DefaultKeyPath = os.Getenv(kfile.EnvKusionPath)

// Config stores key pair file path and key size.
type Config struct {
	DotKeys  string // DotKeys represents the parent dir of IDRsa and IDRsaPub
	IDRsa    string // IDRsa represents private key file
	IDRsaPub string // IDRsaPub represents public key file
	KeySize  int    // KeySize represents the bit size of key pair
}

// NewDefaultConfig return Config with default params.
func NewDefaultConfig() *Config {
	return &Config{
		DotKeys:  filepath.Join(DefaultKeyPath, DefaultDotKeys),
		IDRsa:    filepath.Join(DefaultKeyPath, DefaultDotKeys, DefaultIDRsa),
		IDRsaPub: filepath.Join(DefaultKeyPath, DefaultDotKeys, DefaultIDRsaPub),
		KeySize:  DefaultKeySize,
	}
}

// GenerateKeyPair generates a pair of rsa key pair with the given key size defined in Config.
func (c *Config) GenerateKeyPair() (*rsa.PrivateKey, error) {
	// Generate key pair
	privateKey, err := generateKeyPair(c.KeySize)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// SaveKeyPair save key pair to local path defined in Config.
func (c *Config) SaveKeyPair(keyPair *rsa.PrivateKey) error {
	// Make parent dir
	if err := os.MkdirAll(c.DotKeys, os.ModePerm); err != nil {
		return err
	}
	// Save private key
	if err := saveIDRsa(c.IDRsa, keyPair); err != nil {
		return err
	}
	// Save public key
	if err := saveIDRsaPub(c.IDRsaPub, keyPair); err != nil {
		return err
	}
	return nil
}

type KeyPair struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// LoadPublicKey return a KeyPair object by loading key files defined in Config.
func (c *Config) LoadPublicKey() (*KeyPair, error) {
	publicKey, err := getIDRsaPub(c.IDRsaPub)
	if err != nil {
		return nil, err
	}
	return &KeyPair{
		publicKey: publicKey,
	}, nil
}

// LoadPrivateKey return a KeyPair object by loading key files defined in Config.
func (c *Config) LoadPrivateKey() (*KeyPair, error) {
	privateKey, err := getIDRsa(c.IDRsa)
	if err != nil {
		return nil, err
	}
	return &KeyPair{
		privateKey: privateKey,
	}, nil
}

// EncryptPKCS1v15 encrypts with public key.
func (p *KeyPair) EncryptPKCS1v15(plainText string) (string, error) {
	cipherText, err := encryptPKCS1v15(plainText, p.publicKey)
	if err != nil {
		return "", err
	}
	return cipherText, nil
}

// DecryptPKCS1v15 decrypts with private key.
func (p *KeyPair) DecryptPKCS1v15(cipherText string) (string, error) {
	plainText, err := decryptPKCS1v15(cipherText, p.privateKey)
	if err != nil {
		return "", err
	}
	return plainText, nil
}

// SignPKCS1v15 signs with private key.
func (p *KeyPair) SignPKCS1v15(payload string) (string, error) {
	signature, err := signPKCS1v15(payload, p.privateKey)
	if err != nil {
		return "", err
	}
	return signature, nil
}

// VerifyPKCS1v15 verifies with public key.
func (p *KeyPair) VerifyPKCS1v15(payload, signature64 string) error {
	return verifyPKCS1v15(payload, signature64, p.publicKey)
}

// generateKeyPair generates an RSA keypair of the given bit size.
//
// NOTE: keySize must be positive and a multiple of 1024 is recommended.
func generateKeyPair(keySize int) (*rsa.PrivateKey, error) {
	// Generate key pair
	keyPair, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, err
	}

	// Validate key
	err = keyPair.Validate()
	if err != nil {
		return nil, err
	}

	return keyPair, nil
}

// saveIDRsa saves private key to filename.
func saveIDRsa(fileName string, keyPair *rsa.PrivateKey) error {
	// Private key stream
	privateKeyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(keyPair),
	}

	// Create file
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	return pem.Encode(f, privateKeyBlock)
}

// saveIDRsaPub saves public key to filename.
func saveIDRsaPub(fileName string, keyPair *rsa.PrivateKey) error {
	// Public key stream
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&keyPair.PublicKey)
	if err != nil {
		return err
	}

	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

	// Create file
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	return pem.Encode(f, publicKeyBlock)
}

// getIDRsa gets private key from filename.
func getIDRsa(filename string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil {
		return nil, errors.New("ERROR: fail get rsa private key, invalid key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// getIDRsaPub gets public key from filename.
func getIDRsaPub(filename string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil {
		return nil, errors.New("ERROR: fail get public key, invalid key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(keyBlock.Bytes)
	if err != nil {
		return nil, err
	}
	if key, ok := publicKey.(*rsa.PublicKey); !ok {
		return nil, fmt.Errorf("input file: %s is not public key", filename)
	} else {
		return key, nil
	}
}

// encryptPKCS1v15 encrypts the given message with RSA and the padding scheme from PKCS #1 v1.5.
func encryptPKCS1v15(plainText string, key *rsa.PublicKey) (string, error) {
	partLen := key.Size() - 11
	chunks := split([]byte(plainText), partLen)

	buffer := bytes.NewBufferString("")
	for _, chunk := range chunks {
		encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, key, chunk)
		if err != nil {
			return "", err
		}
		buffer.Write(encrypted)
	}

	return base64.RawURLEncoding.EncodeToString(buffer.Bytes()), nil
}

// decryptPKCS1v15 decrypts a plaintext using RSA and the padding scheme from PKCS #1 v1.5.
func decryptPKCS1v15(cipherText string, key *rsa.PrivateKey) (string, error) {
	partLen := key.Size()
	raw, err := base64.RawURLEncoding.DecodeString(cipherText)
	chunks := split(raw, partLen)

	buffer := bytes.NewBufferString("")
	for _, chunk := range chunks {
		decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, key, chunk)
		if err != nil {
			return "", err
		}
		buffer.Write(decrypted)
	}

	return buffer.String(), err
}

// signPKCS1v15 calculates the signature of hashed using RSASSA-PKCS1-V1_5-SIGN from RSA PKCS #1 v1.5.
func signPKCS1v15(payload string, key *rsa.PrivateKey) (string, error) {
	// Remove unwanted characters and get sha256 hash of the payload
	replacer := strings.NewReplacer("\n", "", "\r", "", " ", "")
	msg := strings.TrimSpace(strings.ToLower(replacer.Replace(payload)))
	hashed := sha256.Sum256([]byte(msg))

	// Sign the hashed payload
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	// Return base64 encoded string
	return base64.StdEncoding.EncodeToString(signature), nil
}

// verifyPKCS1v15 verifies an RSA PKCS #1 v1.5 signature.
func verifyPKCS1v15(payload string, signature64 string, key *rsa.PublicKey) error {
	// Decode base64 encoded signature
	signature, err := base64.StdEncoding.DecodeString(signature64)
	if err != nil {
		return err
	}

	// Remove unwanted characters and get sha256 hash of the payload
	replacer := strings.NewReplacer("\n", "", "\r", "", " ", "")
	msg := strings.TrimSpace(strings.ToLower(replacer.Replace(payload)))
	hashed := sha256.Sum256([]byte(msg))

	return rsa.VerifyPKCS1v15(key, crypto.SHA256, hashed[:], signature)
}

func split(buf []byte, limit int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/limit+1)
	for len(buf) >= limit {
		chunk, buf = buf[:limit], buf[limit:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf)
	}
	return chunks
}
