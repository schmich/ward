package passgen

import "crypto/rand"

type Options struct {
  Length int
  Upper bool
  Lower bool
  Number bool
  Symbol bool
}

func NewPassword(options *Options) string {
  upper := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
  lower := []byte("abcdefghijklmnopqrstuvwxyz")
  number := []byte("0123456789")
  symbol := []byte("`!@#$%^&*()-_=+[{]}\\|;:'\",<.>/?")

  alphabet := make([]byte, 0)
  if options.Upper {
    alphabet = append(alphabet, upper...)
  }

  if options.Lower {
    alphabet = append(alphabet, lower...)
  }

  if options.Number {
    alphabet = append(alphabet, number...)
  }

  if options.Symbol {
    alphabet = append(alphabet, symbol...)
  }

  password := make([]byte, options.Length)

  random := make([]byte, options.Length)
  _, err := rand.Read(random)
  if err != nil {
    panic(err)
  }

  for i := 0; i < len(password); i++ {
    index := int(random[i]) % len(alphabet)
    password[i] = alphabet[index]
  }

  return string(password)
}
