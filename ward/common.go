package main

import (
  "github.com/schmich/ward/store"
  "github.com/schmich/ward/crypto"
  "github.com/fatih/color"
  "golang.org/x/crypto/ssh/terminal"
  "strings"
  "strconv"
  "fmt"
  "os"
)

func (app *App) printSuccess(format string, args ...interface {}) {
  fmt.Printf(color.GreenString("✓ ") + format, args...)
}

func (app *App) printError(format string, args ...interface {}) {
  fmt.Printf(color.RedString("✗ ") + format, args...)
}

func (app *App) readInput(prompt string) string {
  fmt.Fprint(os.Stderr, prompt)
  color.Set(color.FgHiBlack)
  defer color.Unset()
  app.scanner.Scan()
  return app.scanner.Text()
}

func (app *App) readPassword(prompt string) string {
  fmt.Fprint(os.Stderr, prompt)
  password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
  println()
  return string(password)
}

func (app *App) readPasswordConfirm(prompt string) string {
  for {
    password := app.readPassword(prompt + ": ")
    confirm := app.readPassword(prompt + " (confirm): ")

    if password != confirm {
      app.printError("Passwords do not match.\n")
    } else {
      return password
    }
  }
}

func (app *App) readEditPasswordConfirm() (bool, string) {
  for {
    password := app.readPassword("Password (blank to keep current): ")
    if password == "" {
      return false, ""
    }

    confirm := app.readPassword("Password (confirm): ")

    if password != confirm {
      app.printError("Passwords do not match.\n")
    } else {
      return true, password
    }
  }
}

func (app *App) openStore() *store.Store {
  for {
    master := app.readPassword("Master password: ")
    db, err := store.Open(app.storeFileName, master)
    if err == nil {
      return db
    }

    app.printError("%s\n", err)

    if _, ok := err.(crypto.IncorrectPasswordError); !ok {
      if _, ok = err.(crypto.InvalidPasswordError); !ok {
        os.Exit(1)
      }
    }
  }
}

func (app *App) readIndex(low, high int, prompt string) int {
  for {
    input := app.readInput(prompt)
    index, err := strconv.Atoi(input)
    if (err != nil) || (index < low) || (index > high) {
      app.printError("Invalid choice.\n")
    } else {
      return index
    }
  }
}

func (app *App) selectCredential(credentials []*store.Credential) *store.Credential {
  for i, credential := range credentials {
    fmt.Fprintf(os.Stderr, "%d. %s\n", i + 1, getIdentifier(credential))
  }

  index := app.readIndex(1, len(credentials), "> ")
  return credentials[index - 1]
}

func (app *App) findCredential(db *store.Store, query []string) *store.Credential {
  credentials := db.FindCredentials(query)
  if len(credentials) == 0 {
    queryString := strings.Join(query, " ")
    app.printError("No credentials match \"%s\".\n", queryString)
    return nil
  } else if len(credentials) == 1 {
    return credentials[0]
  } else {
    queryString := strings.Join(query, " ")
    fmt.Fprintf(os.Stderr, "Found multiple credentials matching \"%s\":\n", queryString)
    return app.selectCredential(credentials)
  }
}

func getIdentifier(credential *store.Credential) string {
  loginRealm := ""
  if len(credential.Login) > 0 && len(credential.Realm) > 0 {
    loginRealm = credential.Login + "@" + credential.Realm
  } else {
    loginRealm = credential.Login + credential.Realm
  }

  if len(credential.Note) == 0 {
    return loginRealm
  }

  if len(loginRealm) == 0 {
    return credential.Note
  } else {
    return loginRealm + " (" + credential.Note + ")"
  }
}
