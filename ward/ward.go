package main

import (
  "github.com/schmich/ward/store"
  "github.com/schmich/ward/crypto"
  "golang.org/x/crypto/ssh/terminal"
  "github.com/mitchellh/go-homedir"
  "github.com/fatih/color"
  "github.com/jawher/mow.cli"
  "path/filepath"
  "bufio"
  "strings"
  "strconv"
  "fmt"
  "os"
)

type App struct {
  scanner *bufio.Scanner
  storeFileName string
}

func NewApp(fileName string) *App {
  fullPath, _ := filepath.Abs(fileName)

  return &App {
    scanner: bufio.NewScanner(os.Stdin),
    storeFileName: fullPath,
  }
}

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

func (app *App) Run(args []string) {
  ward := cli.App("ward", "Secure password manager - https://github.com/schmich/ward")
  ward.Version("v version", "ward " + Version)
  ward.Command("init", "Create a new credential database.", app.initCommand)
  ward.Command("add", "Add a new credential.", app.addCommand)
  ward.Command("copy", "Copy a password to the clipboard.", app.copyCommand)
  ward.Command("edit", "Edit an existing credential.", app.editCommand)
  ward.Command("del", "Delete a stored credential.", app.delCommand)
  ward.Command("qr", "Print password formatted as a QR code.", app.qrCommand)
  ward.Command("import", "Import JSON-formatted credentials.", app.importCommand)
  ward.Command("export", "Export JSON-formatted credentials.", app.exportCommand)
  ward.Command("master", "Update master password.", app.masterCommand)
  ward.Run(args)
}

func main() {
  wardFile := os.Getenv("WARDFILE")
  if wardFile == "" {
    homeDir, _ := homedir.Dir()
    wardFile = filepath.Join(homeDir, ".ward")
  }

  app := NewApp(wardFile)
  app.Run(os.Args)
}
