package store

import (
  "github.com/schmich/ward/crypto"
  _ "github.com/mattn/go-sqlite3"
  "database/sql"
  "strings"
)

type Store struct {
  db *sql.DB
  cipher *crypto.Cipher
}

type Credential struct {
  Login string
  Password string
  Website string
  Note string
}

func Open(filename string, password string) *Store {
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

  cipher := crypto.LoadCipher(password, salt, stretch, nonce)

  return &Store {
    db: db,
    cipher: cipher,
  }
}

func Create(filename string, password string) *Store {
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

  cipher := crypto.NewCipher(password, stretch)

  insert, err := db.Prepare("insert into settings (salt, stretch, nonce, version) values (?, ?, ?, ?)")
  if err != nil {
    panic(err)
  }

  defer insert.Close()

  insert.Exec(cipher.GetSalt(), stretch, cipher.GetNonce(), version)

  return &Store {
    db: db,
    cipher: cipher,
  }
}

func (store *Store) updateNonce(nonce []byte, tx *sql.Tx) {
  update, err := tx.Prepare("update settings set nonce = ?")
  if err != nil {
    panic(err)
  }

  defer update.Close()

  update.Exec(nonce)
}

func (store *Store) update(updateFn func(*sql.Tx)) {
  tx, err := store.db.Begin()
  if err != nil {
    panic(err)
  }

  updateFn(tx)
  store.updateNonce(store.cipher.GetNonce(), tx)

  tx.Commit()
}

func (store *Store) AddCredential(credential *Credential) {
  store.update(func(tx *sql.Tx) {
    insert, err := tx.Prepare("insert into credentials (login, password, website, note) values (?, ?, ?, ?)")
    if err != nil {
      panic(err)
    }

    defer insert.Close()

    insert.Exec(
      store.cipher.Encrypt([]byte(credential.Login)),
      store.cipher.Encrypt([]byte(credential.Password)),
      store.cipher.Encrypt([]byte(credential.Website)),
      store.cipher.Encrypt([]byte(credential.Note)),
    )
  })
}

func (store *Store) FindCredentials(query string) []*Credential {
  rows, err := store.db.Query("select login, password, website, note from credentials")
  if err != nil {
    panic(err)
  }

  defer rows.Close()

  matches := make([]*Credential, 0)

  for rows.Next() {
    var cipherLogin, cipherPassword, cipherWebsite, cipherNote []byte
    rows.Scan(&cipherLogin, &cipherPassword, &cipherWebsite, &cipherNote)

    login := string(store.cipher.Decrypt(cipherLogin))
    website := string(store.cipher.Decrypt(cipherWebsite))
    note := string(store.cipher.Decrypt(cipherNote))

    if strings.Contains(login, query) || strings.Contains(website, query) || strings.Contains(note, query) {
      password := string(store.cipher.Decrypt(cipherPassword))
      matches = append(matches, &Credential { Login: login, Password: password, Website: website, Note: note })
    }
  }

  return matches
}

func (store *Store) Close() {
  store.db.Close()
}
