package crypto

import (
	"crypto/rsa"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	var err error
	c := NewDefaultConfig()
	t.Run("GenerateKeyPair", func(t *testing.T) {
		var keyPair *rsa.PrivateKey
		if keyPair, err = c.GenerateKeyPair(); err != nil {
			t.Fatalf("GenerateKeyPairs() failed, err: %v", err)
		}
		if err = c.SaveKeyPair(keyPair); err != nil {
			t.Fatalf("SaveKeyPair() failed, err: %v", err)
		}
		if _, err = os.Stat(c.IDRsa); err != nil {
			t.Fatalf("private key is not exist, err: %v", err)
		}
		if _, err = os.Stat(c.IDRsaPub); err != nil {
			t.Fatalf("publick key is not exist, err: %v", err)
		}
	})

	var privateKey *KeyPair
	t.Run("LoadPrivateKey", func(t *testing.T) {
		if privateKey, err = c.LoadPrivateKey(); err != nil {
			t.Fatalf("LoadPrivateKey() failed, err: %v", err)
		}
	})
	var publicKey *KeyPair
	t.Run("LoadPublicKey", func(t *testing.T) {
		if publicKey, err = c.LoadPublicKey(); err != nil {
			t.Fatalf("publicKey() failed, err: %v", err)
		}
	})

	t.Run("encrypt and decrypt", func(t *testing.T) {
		plainText := "foo+PKCS1v15"
		var cipherText string

		// Encrypt
		if cipherText, err = publicKey.EncryptPKCS1v15(plainText); err != nil {
			t.Fatalf("EncryptPKCS1v15() failed, err: %v", err)
		}
		t.Logf("Encrypt success! result is: %s\n", cipherText)

		// Decrypt
		var decrypted string
		if decrypted, err = privateKey.DecryptPKCS1v15(cipherText); err != nil {
			t.Fatalf("DecryptPKCS1v15() failed, err: %v", err)
		}
		if decrypted != plainText {
			t.Fatalf("plainText(%s) and decrypted(%s) are not same", plainText, decrypted)
		}
	})

	t.Run("sign and validate", func(t *testing.T) {
		payload := "this is a secret"
		var signature string
		// Sign
		if signature, err = privateKey.SignPKCS1v15(payload); err != nil {
			t.Fatalf("SignPKCS1v15() failed, err: %v", err)
		}
		t.Logf("Sign suceess, result is: %s", signature)

		// Verify
		if err = publicKey.VerifyPKCS1v15(payload, signature); err != nil {
			t.Fatalf("VerifyPKCS1v15() failed, err: %v", err)
		}
		t.Logf("Verify passed")
	})
}

func Test_generateKeyPair(t *testing.T) {
	t.Run("keySize is negative", func(t *testing.T) {
		_, err := generateKeyPair(-1)
		assert.NotNil(t, err)
	})
	t.Run("keySize is 0", func(t *testing.T) {
		_, err := generateKeyPair(0)
		assert.NotNil(t, err)
	})
	t.Run("keySize is 123", func(t *testing.T) {
		_, err := generateKeyPair(123)
		assert.Nil(t, err)
	})
	t.Run("keySize is 1024", func(t *testing.T) {
		_, err := generateKeyPair(1024)
		assert.Nil(t, err)
	})
}
