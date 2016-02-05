package crypto_test

import (
  "testing"
  "github.com/schmich/ward/crypto"
  . "gopkg.in/check.v1"
)

func Test(t *testing.T) {
  TestingT(t)
}

type CryptoSuite struct {
}

var _ = Suite(&CryptoSuite{})

func (s *CryptoSuite) TestNewKey(c *C) {
  key := crypto.NewKey()
  c.Assert(key, NotNil)
}

func (s *CryptoSuite) TestNewPasswordKey(c *C) {
  key, salt, err := crypto.NewPasswordKey("pass", 1)
  c.Assert(key, NotNil)
  c.Assert(salt, NotNil)
  c.Assert(err, IsNil)
  key, salt, err = crypto.NewPasswordKey("pass", 100)
  c.Assert(key, NotNil)
  c.Assert(salt, NotNil)
  c.Assert(err, IsNil)
}

func (s *CryptoSuite) TestNewPasswordKeyFail(c *C) {
  key, salt, err := crypto.NewPasswordKey("", 1)
  c.Assert(key, IsNil)
  c.Assert(salt, IsNil)
  c.Assert(err, NotNil)
  key, salt, err = crypto.NewPasswordKey("pass", 0)
  c.Assert(key, IsNil)
  c.Assert(salt, IsNil)
  c.Assert(err, NotNil)
}

func (s *CryptoSuite) TestLoadPasswordKey(c *C) {
  salt := make([]byte, 64)
  key, err := crypto.LoadPasswordKey("pass", salt, 1)
  c.Assert(key, NotNil)
  c.Assert(err, IsNil)
}

func (s *CryptoSuite) TestNewLoadPasswordKey(c *C) {
  password := "pass"
  stretch := 1
  newKey, salt, _ := crypto.NewPasswordKey(password, stretch)
  loadKey, _ := crypto.LoadPasswordKey(password, salt, stretch)
  c.Assert(newKey, DeepEquals, loadKey)
}

func (s *CryptoSuite) TestNewCipher(c *C) {
  cipher, err := crypto.NewCipher(crypto.NewKey())
  c.Assert(cipher, NotNil)
  c.Assert(err, IsNil)
}

func (s *CryptoSuite) TestNewCipherFail(c *C) {
  cipher, err := crypto.NewCipher([]byte{})
  c.Assert(cipher, IsNil)
  c.Assert(err, NotNil)
}

func (s *CryptoSuite) TestLoadCipher(c *C) {
  nonce := make([]byte, 12)
  cipher, err := crypto.LoadCipher(crypto.NewKey(), nonce)
  c.Assert(cipher, NotNil)
  c.Assert(err, IsNil)
}

func (s *CryptoSuite) TestNewLoadCipher(c *C) {
  key := crypto.NewKey()
  newCipher, _ := crypto.NewCipher(key)
  loadCipher, _ := crypto.LoadCipher(key, newCipher.GetNonce())
  c.Assert(loadCipher, NotNil)
  c.Assert(loadCipher.GetNonce(), DeepEquals, newCipher.GetNonce())
}

func (s *CryptoSuite) TestGetNonce(c *C) {
  cipher, _ := crypto.NewCipher(crypto.NewKey())
  nonce := cipher.GetNonce()
  c.Assert(nonce, NotNil)
}

func (s *CryptoSuite) TestEncrypt(c *C) {
  cipher, _ := crypto.NewCipher(crypto.NewKey())
  nonce0 := cipher.GetNonce()
  plaintext := []byte { 1, 2, 3, 4, 5 }
  ciphertext1 := cipher.Encrypt(plaintext)
  nonce1 := cipher.GetNonce()
  c.Assert(ciphertext1, NotNil)
  c.Assert(len(ciphertext1), Not(Equals), 0)
  c.Assert(ciphertext1, Not(DeepEquals), plaintext)
  c.Assert(nonce0, Not(DeepEquals), nonce1)
  ciphertext2 := cipher.Encrypt(plaintext)
  nonce2 := cipher.GetNonce()
  c.Assert(ciphertext2, Not(DeepEquals), plaintext)
  c.Assert(ciphertext2, Not(DeepEquals), ciphertext1)
  c.Assert(nonce1, Not(DeepEquals), nonce2)
}

func (s *CryptoSuite) TestTryDecrypt(c *C) {
  cipher, _ := crypto.NewCipher(crypto.NewKey())
  plaintext := []byte { 1, 2, 3, 4, 5 }
  ciphertext := cipher.Encrypt(plaintext)
  plaintextVerify, err := cipher.TryDecrypt(ciphertext)
  c.Assert(err, IsNil)
  c.Assert(plaintextVerify, NotNil)
  c.Assert(plaintextVerify, DeepEquals, plaintext)
}

func (s *CryptoSuite) TestTryDecryptFail(c *C) {
  key := crypto.NewKey()
  encipher, _ := crypto.NewCipher(key)
  plaintext := []byte { 1, 2, 3, 4, 5 }
  ciphertext := encipher.Encrypt(plaintext)
  newKey := crypto.NewKey()
  c.Assert(key, Not(DeepEquals), newKey)
  decipher, _ := crypto.LoadCipher(newKey, encipher.GetNonce())
  plaintextVerify, err := decipher.TryDecrypt(ciphertext)
  c.Assert(err, NotNil)
  c.Assert(len(plaintextVerify), Equals, 0)
}

func (s *CryptoSuite) TestDecrypt(c *C) {
  key := crypto.NewKey()
  encipher, _ := crypto.NewCipher(key)
  plaintext := []byte { 1, 2, 3, 4, 5 }
  ciphertext := encipher.Encrypt(plaintext)
  decipher, _ := crypto.LoadCipher(key, encipher.GetNonce())
  plaintextVerify := decipher.Decrypt(ciphertext)
  c.Assert(plaintextVerify, NotNil)
  c.Assert(plaintextVerify, DeepEquals, plaintext)
}
