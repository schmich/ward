package tests

import (
  "github.com/schmich/ward/crypto"
  . "gopkg.in/check.v1"
)

type CryptoSuite struct {
}

var _ = Suite(&CryptoSuite{})

func (s *CryptoSuite) TestNew(c *C) {
  cipher, _ := crypto.NewCipher("pass", 1)
  c.Assert(cipher, NotNil)
  cipher, _ = crypto.NewCipher("pass", 5000)
  c.Assert(cipher, NotNil)
  cipher, _ = crypto.NewCipher("", 1)
  c.Assert(cipher, IsNil)
  cipher, _ = crypto.NewCipher("pass", 0)
  c.Assert(cipher, IsNil)
}

func (s *CryptoSuite) TestLoad(c *C) {
  cipher, _ := crypto.LoadCipher("pass", make([]byte, 64), 1, make([]byte, 12))
  c.Assert(cipher, NotNil)
  cipher, _ = crypto.LoadCipher("pass", make([]byte, 64), 5000, make([]byte, 12))
  c.Assert(cipher, NotNil)
  cipher, _ = crypto.LoadCipher("", make([]byte, 64), 1, make([]byte, 12))
  c.Assert(cipher, IsNil)
  cipher, _ = crypto.LoadCipher("pass", nil, 1, make([]byte, 12))
  c.Assert(cipher, IsNil)
  cipher, _ = crypto.LoadCipher("pass", []byte{}, 1, make([]byte, 12))
  c.Assert(cipher, IsNil)
  cipher, _ = crypto.LoadCipher("pass", make([]byte, 64), 0, make([]byte, 12))
  c.Assert(cipher, IsNil)
  cipher, _ = crypto.LoadCipher("pass", make([]byte, 64), 1, nil)
  c.Assert(cipher, IsNil)
  cipher, _ = crypto.LoadCipher("pass", make([]byte, 64), 1, []byte{})
  c.Assert(cipher, IsNil)
}

func (s *CryptoSuite) TestGetNonce(c *C) {
  cipher, _ := crypto.NewCipher("pass", 1)
  nonce := cipher.GetNonce()
  c.Assert(len(nonce), Equals, 12)
}

func (s *CryptoSuite) TestGetSalt(c *C) {
  cipher, _ := crypto.NewCipher("pass", 1)
  salt := cipher.GetSalt()
  c.Assert(len(salt), Equals, 64)
}

func (s *CryptoSuite) TestNewLoad(c *C) {
  password := "pass"
  stretch := 1
  cipher, _ := crypto.NewCipher(password, stretch)
  nonce := cipher.GetNonce()
  salt := cipher.GetSalt()
  cipher, _ = crypto.LoadCipher(password, salt, stretch, nonce)
  c.Assert(cipher, NotNil)
}

func (s *CryptoSuite) TestEncrypt(c *C) {
  cipher, _ := crypto.NewCipher("pass", 1)
  salt := cipher.GetSalt()
  nonce0 := cipher.GetNonce()
  plaintext := []byte { 1, 2, 3, 4, 5 }
  ciphertext1 := cipher.Encrypt(plaintext)
  nonce1 := cipher.GetNonce()
  c.Assert(len(ciphertext1), Not(Equals), 0)
  c.Assert(ciphertext1, Not(DeepEquals), plaintext)
  c.Assert(nonce0, Not(DeepEquals), nonce1)
  ciphertext2 := cipher.Encrypt(plaintext)
  nonce2 := cipher.GetNonce()
  c.Assert(ciphertext2, Not(DeepEquals), plaintext)
  c.Assert(ciphertext2, Not(DeepEquals), ciphertext1)
  c.Assert(nonce1, Not(DeepEquals), nonce2)
  c.Assert(salt, DeepEquals, cipher.GetSalt())
}

func (s *CryptoSuite) TestEncryptDecrypt(c *C) {
  password := "pass"
  stretch := 1
  encipher, _ := crypto.NewCipher(password, stretch)
  plaintext := []byte { 1, 2, 3, 4, 5 }
  ciphertext := encipher.Encrypt(plaintext)
  salt := encipher.GetSalt()
  nonce := encipher.GetNonce()
  decipher, _ := crypto.LoadCipher(password, salt, stretch, nonce)
  plaintextVerify := decipher.Decrypt(ciphertext)
  c.Assert(plaintext, DeepEquals, plaintextVerify)
}

func (s *CryptoSuite) TestEncryptBadDecrypt(c *C) {
  encipher, _ := crypto.NewCipher("pass", 1)
  plaintext := []byte { 1, 2, 3, 4, 5 }
  ciphertext := encipher.Encrypt(plaintext)
  decipher, _ := crypto.NewCipher("wrongpass", 1)
  _, err := decipher.TryDecrypt(ciphertext)
  c.Assert(err, NotNil)
  decipher, _ = crypto.NewCipher("pass", 100)
  _, err = decipher.TryDecrypt(ciphertext)
  c.Assert(err, NotNil)
}
