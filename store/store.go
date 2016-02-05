package store

import (
  "github.com/schmich/ward/crypto"
  _ "github.com/mattn/go-sqlite3"
  "database/sql"
  "strings"
  "errors"
  "fmt"
  "os"
)

type Store struct {
  db *sql.DB
  passwordCipher *crypto.Cipher
  keyCipher *crypto.Cipher
}

type Credential struct {
  id int
  Login string `json:"login"`
  Password string `json:"password"`
  Realm string `json:"realm"`
  Note string `json:"note"`
}

func Open(fileName string, password string) (*Store, error) {
  if _, err := os.Stat(fileName); os.IsNotExist(err) {
    return nil, errors.New("Credential database does not exist.")
  }

  db, err := sql.Open("sqlite3", fileName)
  if err != nil {
    return nil, err
  }

  query := `
    SELECT password_salt, password_stretch, password_nonce, encrypted_key, key_nonce, version
    FROM settings
  `

  rows, err := db.Query(query)
  if err != nil {
    return nil, err
  }

  defer rows.Close()

  rows.Next()

  var passwordSalt, passwordNonce, encryptedKey, keyNonce []byte
  var passwordStretch, version int
  rows.Scan(&passwordSalt, &passwordStretch, &passwordNonce, &encryptedKey, &keyNonce, &version)

  if version > 1 {
    return nil, errors.New(fmt.Sprintf("Unsupported version: %d.", version))
  }

  if len(encryptedKey) <= 0 {
    return nil, errors.New("Invalid encrypted key.")
  }

  passwordKey, err := crypto.LoadPasswordKey(password, passwordSalt, passwordStretch)
  if err != nil {
    return nil, err
  }

  passwordCipher, err := crypto.LoadCipher(passwordKey, passwordNonce)
  if err != nil {
    return nil, err
  }

  key, err := passwordCipher.TryDecrypt(encryptedKey)
  if err != nil {
    return nil, err
  }

  keyCipher, err := crypto.LoadCipher(key, keyNonce)
  if err != nil {
    return nil, err
  }

  return &Store {
    db: db,
    passwordCipher: passwordCipher,
    keyCipher: keyCipher,
  }, nil
}

func createCipher(db *sql.DB, password string, passwordStretch int) (*crypto.Cipher, *crypto.Cipher, error) {
  const version = 1

  tx, err := db.Begin()
  if err != nil {
    return nil, nil, err
  }

  defer func() {
    if err != nil {
      tx.Rollback()
    } else {
      tx.Commit()
    }
  }()

  passwordKey, passwordSalt, err := crypto.NewPasswordKey(password, passwordStretch)
  if err != nil {
    return nil, nil, err
  }

  passwordCipher, err := crypto.NewCipher(passwordKey)
  if err != nil {
    return nil, nil, err
  }

  key := crypto.NewKey()
  keyCipher, err := crypto.NewCipher(key)
  if err != nil {
    return nil, nil, err
  }

  insert, err := tx.Prepare(`
    INSERT INTO settings (password_salt, password_stretch, password_nonce, encrypted_key, key_nonce, version)
    VALUES (?, ?, ?, ?, ?, ?)
  `)

  if err != nil {
    return nil, nil, err
  }

  defer insert.Close()

  encryptedKey := passwordCipher.Encrypt(key)

  insert.Exec(
    passwordSalt,
    passwordStretch,
    passwordCipher.GetNonce(),
    encryptedKey,
    keyCipher.GetNonce(),
    version)

  return passwordCipher, keyCipher, nil
}

func Create(fileName string, password string, passwordStretch int) (*Store, error) {
  if _, err := os.Stat(fileName); err == nil {
    return nil, errors.New("Credential database already exists.")
  }

  db, err := sql.Open("sqlite3", fileName)
  if err != nil {
    return nil, err
  }

  defer func() {
    if err != nil {
      db.Close()
      os.Remove(fileName)
    }
  }()

  create := `
    CREATE TABLE credentials (
      id INTEGER NOT NULL PRIMARY KEY,
      login BLOB,
      password BLOB,
      realm BLOB,
      note BLOB
    );

    CREATE TABLE settings (
      password_salt BLOB,
      password_stretch INTEGER,
      password_nonce BLOB,
      encrypted_key BLOB,
      key_nonce BLOB,
      version INTEGER
    );
	`

  _, err = db.Exec(create)
  if err != nil {
    return nil, err
  }

  passwordCipher, keyCipher, err := createCipher(db, password, passwordStretch)
  if err != nil {
    return nil, err
  }

  return &Store {
    db: db,
    passwordCipher: passwordCipher,
    keyCipher: keyCipher,
  }, nil
}

func (store *Store) updateNonce(passwordNonce, keyNonce []byte, tx *sql.Tx) error {
  update, err := tx.Prepare("UPDATE settings SET password_nonce=?, key_nonce=?")
  if err != nil {
    return err
  }

  defer update.Close()

  update.Exec(passwordNonce, keyNonce)

  return nil
}

func (store *Store) update(updateFn func(*sql.Tx) error) error {
  tx, err := store.db.Begin()
  if err != nil {
    return err
  }

  defer func() {
    if err != nil {
      tx.Rollback()
    } else {
      tx.Commit()
    }
  }()

  if err = updateFn(tx); err != nil {
    return err
  }

  return store.updateNonce(
    store.passwordCipher.GetNonce(),
    store.keyCipher.GetNonce(),
    tx)
}

func (store *Store) AddCredential(credential *Credential) {
  store.update(func(tx *sql.Tx) error {
    insert, err := tx.Prepare(`
      INSERT INTO credentials (login, password, realm, note)
      VALUES (?, ?, ?, ?)
    `)

    if err != nil {
      return err
    }

    defer insert.Close()

    insert.Exec(
      store.keyCipher.Encrypt([]byte(credential.Login)),
      store.keyCipher.Encrypt([]byte(credential.Password)),
      store.keyCipher.Encrypt([]byte(credential.Realm)),
      store.keyCipher.Encrypt([]byte(credential.Note)),
    )

    return nil
  })
}

func (store *Store) eachCredential() chan *Credential {
  yield := make(chan *Credential)

  go func() {
    defer close(yield)

    rows, err := store.db.Query(`
      SELECT id, login, password, realm, note
      FROM credentials
    `)

    if err != nil {
      panic(err)
    }

    defer rows.Close()

    for rows.Next() {
      var id int
      var cipherLogin, cipherPassword, cipherRealm, cipherNote []byte
      rows.Scan(&id, &cipherLogin, &cipherPassword, &cipherRealm, &cipherNote)

      credential := &Credential {
        id: id,
        Login: string(store.keyCipher.Decrypt(cipherLogin)),
        Password: string(store.keyCipher.Decrypt(cipherPassword)),
        Realm: string(store.keyCipher.Decrypt(cipherRealm)),
        Note: string(store.keyCipher.Decrypt(cipherNote)),
      }

      yield <- credential
    }
  }()

  return yield
}

func (store *Store) AllCredentials() []*Credential {
  credentials := make([]*Credential, 0)

  for credential := range store.eachCredential() {
    credentials = append(credentials, credential)
  }

  return credentials
}

func (store *Store) FindCredentials(query []string) []*Credential {
  matches := make([]*Credential, 0)

  patterns := make([]string, len(query))
  for i, queryString := range query {
    patterns[i] = strings.ToLower(queryString)
  }

  for credential := range store.eachCredential() {
    valid := true
    llogin := strings.ToLower(credential.Login)
    lrealm := strings.ToLower(credential.Realm)
    lnote := strings.ToLower(credential.Note)
    for _, pattern := range patterns {
      if !strings.Contains(llogin, pattern) && !strings.Contains(lrealm, pattern) && !strings.Contains(lnote, pattern) {
        valid = false
        break
      }
    }
    if (valid) {
      matches = append(matches, credential)
    }
  }

  return matches
}

func (store *Store) UpdateCredential(credential *Credential) {
  if credential.id == 0 {
    panic("Invalid credential ID.")
  }

  store.update(func(tx *sql.Tx) error {
    update, err := tx.Prepare(`
      UPDATE credentials
      SET login=?, password=?, realm=?, note=?
      WHERE id=?
    `)

    if err != nil {
      return err
    }

    defer update.Close()

    update.Exec(
      store.keyCipher.Encrypt([]byte(credential.Login)),
      store.keyCipher.Encrypt([]byte(credential.Password)),
      store.keyCipher.Encrypt([]byte(credential.Realm)),
      store.keyCipher.Encrypt([]byte(credential.Note)),
      credential.id,
    )

    return nil
  })
}

func (store *Store) DeleteCredential(credential *Credential) {
  if credential.id == 0 {
    panic("Invalid credential ID.")
  }

  store.update(func(tx *sql.Tx) error {
    delete, err := tx.Prepare("DELETE FROM credentials WHERE id=?")
    if err != nil {
      return err
    }

    defer delete.Close()

    delete.Exec(credential.id)

    return nil
  })
}

func (store *Store) UpdateMasterPassword(password string, passwordStretch int) error {
  return store.update(func(tx *sql.Tx) error {
    query := "SELECT encrypted_key FROM settings"

    rows, err := tx.Query(query)
    if err != nil {
      return err
    }

    defer rows.Close()
    rows.Next()

    var encryptedKey []byte
    rows.Scan(&encryptedKey)

    key := store.passwordCipher.Decrypt(encryptedKey)

    passwordKey, passwordSalt, err := crypto.NewPasswordKey(password, passwordStretch)
    if err != nil {
      return err
    }

    passwordCipher, err := crypto.NewCipher(passwordKey)
    if err != nil {
      return err
    }

    update, err := tx.Prepare(`
      UPDATE settings
      SET password_salt=?, password_stretch=?, password_nonce=?, encrypted_key=?
    `)

    if err != nil {
      return err
    }

    defer update.Close()

    update.Exec(
      passwordSalt,
      passwordStretch,
      passwordCipher.GetNonce(),
      passwordCipher.Encrypt(key))

    store.passwordCipher = passwordCipher

    return nil
  })
}

func (store *Store) Close() {
  store.db.Close()
}
