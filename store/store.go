package store

import (
  "github.com/schmich/ward/crypto"
  _ "github.com/mattn/go-sqlite3"
  "crypto/rand"
  "database/sql"
  "strings"
  "errors"
)

type Store struct {
  db *sql.DB
  cipher *crypto.Cipher
}

type Credential struct {
  id int
  Login string `json:"login"`
  Password string `json:"password"`
  Website string `json:"website"`
  Note string `json:"note"`
}

func Open(filename string, password string) (*Store, error) {
  db, err := sql.Open("sqlite3", filename)
  if err != nil {
    return nil, err
  }

  query := "select salt, stretch, nonce, sentinel, version from settings"
  rows, err := db.Query(query)
  if err != nil {
    return nil, err
  }

  defer rows.Close()

  rows.Next()

  var salt []byte
  var stretch int
  var nonce []byte
  var sentinel []byte
  var version int
  rows.Scan(&salt, &stretch, &nonce, &sentinel, &version)

  if version != 1 {
    return nil, errors.New("Invalid version.")
  }

  if stretch < 1 {
    return nil, errors.New("Invalid key stretch.")
  }

  if len(salt) < 64 {
    return nil, errors.New("Invalid salt.")
  }

  if len(nonce) < 12 {
    return nil, errors.New("Invalid nonce.")
  }

  if len(sentinel) <= 0 {
    return nil, errors.New("Invalid sentinel.")
  }

  cipher := crypto.LoadCipher(password, salt, stretch, nonce)

  _, err = cipher.TryDecrypt(sentinel)
  if err != nil {
    return nil, err
  }

  return &Store {
    db: db,
    cipher: cipher,
  }, nil
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
    sentinel blob,
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

  insert, err := db.Prepare("insert into settings (salt, stretch, nonce, sentinel, version) values (?, ?, ?, ?, ?)")
  if err != nil {
    panic(err)
  }

  defer insert.Close()

  sentinel := make([]byte, 16)
  _, err = rand.Read(sentinel)
  if err != nil {
    panic(err)
  }

  insert.Exec(
    cipher.GetSalt(),
    stretch,
    cipher.GetNonce(),
    cipher.Encrypt(sentinel),
    version)

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

func (store *Store) eachCredential() chan *Credential {
  yield := make(chan *Credential)

  go func() {
    rows, err := store.db.Query("select id, login, password, website, note from credentials")
    if err != nil {
      panic(err)
    }

    defer rows.Close()

    for rows.Next() {
      var id int
      var cipherLogin, cipherPassword, cipherWebsite, cipherNote []byte
      rows.Scan(&id, &cipherLogin, &cipherPassword, &cipherWebsite, &cipherNote)

      credential := &Credential {
        id: id,
        Login: string(store.cipher.Decrypt(cipherLogin)),
        Password: string(store.cipher.Decrypt(cipherPassword)),
        Website: string(store.cipher.Decrypt(cipherWebsite)),
        Note: string(store.cipher.Decrypt(cipherNote)),
      }

      yield <- credential
    }

    close(yield)
  }()

  return yield
}

func (store *Store) GetCredentials() []*Credential {
  credentials := make([]*Credential, 0)

  for credential := range store.eachCredential() {
    credentials = append(credentials, credential)
  }

  return credentials
}

func (store *Store) FindCredentials(query string) []*Credential {
  matches := make([]*Credential, 0)

  for credential := range store.eachCredential() {
    if strings.Contains(credential.Login, query) || strings.Contains(credential.Website, query) || strings.Contains(credential.Note, query) {
      matches = append(matches, credential)
    }
  }

  return matches
}

func (store *Store) UpdateCredential(credential *Credential) {
  if credential.id == 0 {
    panic("Invalid credential ID.")
  }

  store.update(func(tx *sql.Tx) {
    update, err := tx.Prepare("update credentials set login=?, password=?, website=?, note=? where id=?")
    if err != nil {
      panic(err)
    }

    defer update.Close()

    update.Exec(
      store.cipher.Encrypt([]byte(credential.Login)),
      store.cipher.Encrypt([]byte(credential.Password)),
      store.cipher.Encrypt([]byte(credential.Website)),
      store.cipher.Encrypt([]byte(credential.Note)),
      credential.id,
    )
  })
}

func (store *Store) DeleteCredential(credential *Credential) {
  if credential.id == 0 {
    panic("Invalid credential ID.")
  }

  store.update(func(tx *sql.Tx) {
    delete, err := tx.Prepare("delete from credentials where id=?")
    if err != nil {
      panic(err)
    }

    defer delete.Close()

    delete.Exec(credential.id)
  })
}

func (store *Store) Close() {
  store.db.Close()
}
