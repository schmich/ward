package main

import "database/sql"
import _ "github.com/mattn/go-sqlite3"
import "crypto/aes"
import "crypto/rand"
import gocipher "crypto/cipher"
import "golang.org/x/crypto/pbkdf2"
import "golang.org/x/crypto/sha3"
import "golang.org/x/crypto/ssh/terminal"
import "math/big"
import "strings"
import "os"
import "gopkg.in/alecthomas/kingpin.v2"
import "fmt"
import "bufio"
import "github.com/fatih/color"

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

type Ward struct {
  db *sql.DB
  cipher *Cipher
}

func Open(filename string, password string) *Ward {
  db, err := sql.Open("sqlite3", filename)

  if err != nil {
    panic(err)
  }

  query := "select salt, stretch, nonce, version from settings"
  rows, err := db.Query(query)

  if err != nil {
    panic(err)
  }

  defer rows.Close()

  rows.Next()

  var salt []byte
  var stretch int
  var nonce []byte
  var version int
  rows.Scan(&salt, &stretch, &nonce, &version)

  if version != 1 {
    panic("Invalid version.")
  }

  if stretch < 1 {
    panic("Invalid stretch.")
  }

  if len(salt) < 64 {
    panic("Invalid salt.")
  }

  if len(nonce) < 12 {
    panic("Invalid nonce.")
  }

  cipher := LoadCipher(password, salt, stretch, nonce)

  return &Ward {
    db: db,
    cipher: cipher,
  }
}

func Create(filename string, password string) *Ward {
  db, err := sql.Open("sqlite3", filename)

  if err != nil {
    panic(err)
  }

  create := `
	create table credentials (
    id integer not null primary key,
    login blob,
    password blob,
    website blob,
    note blob
  );

  create table settings (
    salt blob,
    stretch integer,
    nonce blob,
    version integer
  );
	`

  _, err = db.Exec(create)
  if err != nil {
    panic(err)
  }

  const version = 1
  const stretch = 100000

  cipher := NewCipher(password, stretch)

  insert, err := db.Prepare("insert into settings (salt, stretch, nonce, version) values (?, ?, ?, ?)")
  if err != nil {
    panic(err)
  }

  defer insert.Close()

  insert.Exec(cipher.GetSalt(), stretch, cipher.GetNonce(), version)

  return &Ward {
    db: db,
    cipher: cipher,
  }
}

func (ward *Ward) updateNonce(nonce []byte, tx *sql.Tx) {
  update, err := tx.Prepare("update settings set nonce = ?")
  if err != nil {
    panic(err)
  }

  defer update.Close()

  update.Exec(nonce)
}

func (ward *Ward) update(updateFn func(*sql.Tx)) {
  tx, err := ward.db.Begin()
  if err != nil {
    panic(err)
  }

  updateFn(tx)
  ward.updateNonce(ward.cipher.GetNonce(), tx)

  tx.Commit()
}

type Credential struct {
  login string
  password string
  website string
  note string
}

func (ward *Ward) AddCredential(credential *Credential) {
  ward.update(func(tx *sql.Tx) {
    insert, err := tx.Prepare("insert into credentials (login, password, website, note) values (?, ?, ?, ?)")
    if err != nil {
      panic(err)
    }

    defer insert.Close()

    insert.Exec(
      ward.cipher.Encrypt([]byte(credential.login)),
      ward.cipher.Encrypt([]byte(credential.password)),
      ward.cipher.Encrypt([]byte(credential.website)),
      ward.cipher.Encrypt([]byte(credential.note)),
    )
  })
}

func (ward *Ward) FindCredentials(query string) []*Credential {
  rows, err := ward.db.Query("select login, password, website, note from credentials")
  if err != nil {
    panic(err)
  }

  defer rows.Close()

  matches := make([]*Credential, 0)

  for rows.Next() {
    var cipherLogin, cipherPassword, cipherWebsite, cipherNote []byte
    rows.Scan(&cipherLogin, &cipherPassword, &cipherWebsite, &cipherNote)

    login := string(ward.cipher.Decrypt(cipherLogin))
    website := string(ward.cipher.Decrypt(cipherWebsite))
    note := string(ward.cipher.Decrypt(cipherNote))

    if strings.Contains(login, query) || strings.Contains(website, query) || strings.Contains(note, query) {
      password := string(ward.cipher.Decrypt(cipherPassword))
      matches = append(matches, &Credential { login: login, password: password, website: website, note: note })
    }
  }

  return matches
}

func (ward *Ward) Close() {
  ward.db.Close()
}

type App struct {
  app *kingpin.Application
  init *kingpin.CmdClause
  new *kingpin.CmdClause
  copy *kingpin.CmdClause
  query *string
}

func NewApp() *App {
  app := kingpin.New("ward", "Password manager.")
  init := app.Command("init", "Create a new password database.")
  new := app.Command("new", "Add a new credential.")
  copy := app.Command("copy", "Copy password.")
  query := copy.Arg("query", "Text to match.").Required().String()

  return &App {
    app: app,
    init: init,
    new: new,
    copy: copy,
    query: query,
  }
}

func (app *App) runInit() {
  master := app.promptMasterPassword("Master password: ")
  confirm := app.promptMasterPassword("Master password (confirm): ")

  if master != confirm {
    panic("Passwords do not match.")
  }

  ward := Create("test.db", master)
  defer ward.Close()
}

func (app *App) promptMasterPassword(prompt string) string {
  print(color.BlueString(prompt))
  password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
  println()
  return string(password)
}

func (app *App) runNew() {
  master := app.promptMasterPassword("Master password: ")
  ward := Open("test.db", master)
  defer ward.Close()

  reader := bufio.NewReader(os.Stdin)

  cyan := color.New(color.FgCyan).PrintfFunc()

  cyan("Username: ")
  login, _ := reader.ReadString('\n')

  cyan("Password (enter to generate): ")
  password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
  println()

  cyan("Website: ")
  website, _ := reader.ReadString('\n')

  cyan("Note: ")
  note, _ := reader.ReadString('\n')

  ward.AddCredential(&Credential {
    login: login,
    password: string(password),
    website: website,
    note: note,
  })

  println("Credential added.")
}

func (app *App) runCopy(query string) {
  master := app.promptMasterPassword("Master password: ")
  ward := Open("test.db", master)
  defer ward.Close()

  credentials := ward.FindCredentials(query)
  for _, credential := range credentials {
    fmt.Println(credential)
  }
}

func (app *App) Run(args []string) {
  switch kingpin.MustParse(app.app.Parse(args[1:])) {
  case app.init.FullCommand():
    app.runInit()
  case app.new.FullCommand():
    app.runNew()
  case app.copy.FullCommand():
    app.runCopy(*app.query)
  }
}

func main() {
  app := NewApp()
  app.Run(os.Args)
}
