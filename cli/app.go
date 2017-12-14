package main

import (
  "github.com/schmich/ward/store"
  "github.com/schmich/ward/crypto"
  "github.com/jawher/mow.cli"
  "path/filepath"
  "os"
)

type App struct {
  storeFileName string
}

func NewApp(fileName string) *App {
  fullPath, _ := filepath.Abs(fileName)

  return &App {
    storeFileName: fullPath,
  }
}

func (app *App) openStore() *store.Store {
  for {
    master := readPassword("Master password: ")
    db, err := store.Open(app.storeFileName, master)
    if err == nil {
      return db
    }

    printError("%s\n", err)

    if _, ok := err.(crypto.IncorrectPasswordError); !ok {
      if _, ok = err.(crypto.InvalidPasswordError); !ok {
        os.Exit(1)
      }
    }
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
  ward.Command("list", "Print a table-formatted list of credentials.", app.listCommand)
  ward.Command("import", "Import JSON-formatted credentials.", app.importCommand)
  ward.Command("export", "Export JSON-formatted credentials.", app.exportCommand)
  ward.Command("master", "Update master password.", app.masterCommand)
  ward.Run(args)
}
