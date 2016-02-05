package crypto

import (
  "golang.org/x/crypto/pbkdf2"
  "golang.org/x/crypto/sha3"
  gocipher "crypto/cipher"
  "crypto/aes"
  "crypto/rand"
  "math/big"
  "errors"
)

type IncorrectPasswordError string

func (s IncorrectPasswordError) Error() string {
  return "Incorrect password."
}

type InvalidPasswordError string

func (s InvalidPasswordError) Error() string {
  return "Invalid password."
}

type Cipher struct {
  aead gocipher.AEAD
  nonce *big.Int
}

func NewKey() []byte {
  key := make([]byte, aes.BlockSize)
  count, err := rand.Read(key)

  if (err != nil) || (count != len(key)) {
    panic("Failed to generate random key.")
  }

  return key
}

func NewPasswordKey(password string, stretch int) ([]byte, []byte, error) {
  salt := make([]byte, 64)
  count, err := rand.Read(salt)
  if err != nil {
    return nil, nil, err
  }

  if count != len(salt) {
    return nil, nil, errors.New("Failed to generate random salt.")
  }

  key, err := LoadPasswordKey(password, salt, stretch)
  if err != nil {
    return nil, nil, err
  }

  return key, salt, err
}

func LoadPasswordKey(password string, salt []byte, stretch int) ([]byte, error) {
  if len(password) == 0 {
    var e InvalidPasswordError
    return nil, e
  }

  if len(salt) < 64 {
    return nil, errors.New("Invalid salt.")
  }

  if stretch < 1 {
    return nil, errors.New("Key stretch must be at least 1.")
  }

  return pbkdf2.Key([]byte(password), salt, stretch, aes.BlockSize, sha3.New512), nil
}

func NewCipher(key []byte) (*Cipher, error) {
  nonce := make([]byte, 12)
  cipher, err := LoadCipher(key, nonce)
  if err != nil {
    return nil, err
  }

  return cipher, nil
}

func LoadCipher(key, nonce []byte) (*Cipher, error) {
  if len(key) != aes.BlockSize {
    return nil, errors.New("Invalid key.")
  }

  if len(nonce) < 12 {
    return nil, errors.New("Invalid nonce.")
  }

  block, err := aes.NewCipher(key)
  if err != nil {
    return nil, err
  }

  aead, err := gocipher.NewGCM(block)
  if err != nil {
    return nil, err
  }

  nonceInt := big.NewInt(0)
  nonceInt.SetBytes(nonce)

  return &Cipher {
    aead: aead,
    nonce: nonceInt,
  }, nil
}

func (cipher *Cipher) GetNonce() []byte {
  nonceBytes := cipher.nonce.Bytes()
  nonce := make([]byte, cipher.aead.NonceSize())
  copy(nonce[len(nonce) - len(nonceBytes):], nonceBytes)
  return nonce
}

func (cipher *Cipher) Encrypt(plaintext []byte) []byte {
  plaintextBuffer := pad(plaintext)

  nonce := cipher.GetNonce()
  cipher.nonce = cipher.nonce.Add(cipher.nonce, big.NewInt(1))

  ciphertext := make([]byte, 0)
  ciphertext = cipher.aead.Seal(ciphertext, nonce, plaintextBuffer, []byte{})

  return append(ciphertext, nonce...)
}

func (cipher *Cipher) TryDecrypt(ciphertext []byte) ([]byte, error) {
  nonceStart := len(ciphertext) - cipher.aead.NonceSize()
  nonce := ciphertext[nonceStart:]
  ciphertext = ciphertext[:nonceStart]

  plaintext := make([]byte, 0)
  plaintext, err := cipher.aead.Open(plaintext, nonce, ciphertext, []byte{})
  if err != nil {
    var e IncorrectPasswordError
    return []byte{}, e
  }

  return depad(plaintext), nil
}

func (cipher *Cipher) Decrypt(ciphertext []byte) []byte {
  plaintext, _ := cipher.TryDecrypt(ciphertext)
  return plaintext
}

func pad(buffer []byte) []byte {
  // See http://tools.ietf.org/html/rfc5652#section-6.3
  padLength := aes.BlockSize - (len(buffer) % aes.BlockSize)

  padBuffer := make([]byte, padLength)
  for i := 0; i < padLength; i++ {
    padBuffer[i] = byte(padLength)
  }

  return append(buffer, padBuffer...)
}

func depad(buffer []byte) []byte {
  // See http://tools.ietf.org/html/rfc5652#section-6.3
  padLength := int(buffer[len(buffer) - 1])
  return buffer[:len(buffer) - padLength]
}
