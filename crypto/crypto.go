package crypto

import "crypto/aes"
import "crypto/rand"
import gocipher "crypto/cipher"
import "golang.org/x/crypto/pbkdf2"
import "golang.org/x/crypto/sha3"
import "math/big"

type Cipher struct {
  aead gocipher.AEAD
  nonce *big.Int
  salt []byte
}

func LoadCipher(password string, salt []byte, stretch int, nonce []byte) *Cipher {
  passwordBuffer := []byte(password)

  derivedKey := pbkdf2.Key(passwordBuffer, salt, stretch, 32, sha3.New512)
  block, err := aes.NewCipher(derivedKey)
  if err != nil {
    panic(err)
  }

  aead, err := gocipher.NewGCM(block)
  if err != nil {
    panic(err)
  }

  nonceInt := big.NewInt(0)
  nonceInt.SetBytes(nonce)

  return &Cipher {
    aead: aead,
    nonce: nonceInt,
    salt: salt,
  }
}

func NewCipher(password string, stretch int) *Cipher {
  salt := make([]byte, 64)
  count, err := rand.Read(salt)

  if err != nil {
    panic(err)
  }

  if count != len(salt) {
    panic("Failed to generate random salt.")
  }

  nonce := big.NewInt(0).Bytes()
  cipher := LoadCipher(password, salt, stretch, nonce)

  return cipher
}

func (cipher *Cipher) GetNonce() []byte {
  nonceBytes := cipher.nonce.Bytes()
  nonce := make([]byte, cipher.aead.NonceSize())
  copy(nonce[len(nonce) - len(nonceBytes):], nonceBytes)
  return nonce
}

func (cipher *Cipher) GetSalt() []byte {
  return cipher.salt
}

func (cipher *Cipher) Encrypt(plaintext []byte) []byte {
  plaintextBuffer := pad(plaintext)

  nonce := cipher.GetNonce()
  cipher.nonce = cipher.nonce.Add(cipher.nonce, big.NewInt(1))

  ciphertext := make([]byte, 0)
  ciphertext = cipher.aead.Seal(ciphertext, nonce, plaintextBuffer, []byte{})

  return append(ciphertext, nonce...)
}

func (cipher *Cipher) Decrypt(ciphertext []byte) []byte {
  nonceStart := len(ciphertext) - cipher.aead.NonceSize()
  nonce := ciphertext[nonceStart:]
  ciphertext = ciphertext[:nonceStart]

  plaintext := make([]byte, 0)
  plaintext, err := cipher.aead.Open(plaintext, nonce, ciphertext, []byte{})
  if err != nil {
    panic(err)
  }

  return depad(plaintext)
}

// TODO: should use http://tools.ietf.org/html/rfc5652#section-6.3
// See https://tools.ietf.org/html/rfc5246#section-6.2.3.2
func pad(buffer []byte) []byte {
  totalLength := len(buffer) + 1
  padLength := aes.BlockSize - (totalLength % aes.BlockSize)

  padBuffer := make([]byte, padLength + 1)
  for i := 0; i < len(padBuffer); i++ {
    padBuffer[i] = byte(padLength)
  }

  return append(buffer, padBuffer...)
}

// TODO: should use http://tools.ietf.org/html/rfc5652#section-6.3
// See https://tools.ietf.org/html/rfc5246#section-6.2.3.2
func depad(buffer []byte) []byte {
  padLength := int(buffer[len(buffer) - 1])
  return buffer[:len(buffer) - padLength - 1]
}
