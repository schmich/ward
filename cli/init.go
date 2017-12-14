package main

import (
  "github.com/schmich/ward/store"
  "github.com/jawher/mow.cli"
  "path/filepath"
  "fmt"
  "os"
)

func (app *App) initCommand(cmd *cli.Cmd) {
  cmd.Spec = "[--stretch|--link=<file>]"

  stretch := cmd.IntOpt("stretch", 200000, "Password key stretch iterations.")
  file := cmd.StringOpt("link", "", "Link to an existing credential database.")

  cmd.Action = func() {
    if *file == "" {
      app.runInit(*stretch)
    } else {
      app.runLink(*file)
    }
  }
}

func (app *App) runInit(keyStretch int) {
  fmt.Println("Creating new credential database.")
  password := readPasswordConfirm("Master password")

  db, err := store.Create(app.storeFileName, password, keyStretch)
  if err != nil {
    printError("Failed to create database: %s\n", err.Error())
    return
  }

  defer db.Close()

  printSuccess("Credential database created at %s.\n", app.storeFileName)
}

func (app *App) runLink(existingFileName string) {
  existingFullPath, _ := filepath.Abs(existingFileName)
  err := os.Symlink(existingFullPath, app.storeFileName)
  if err != nil {
    printError("Could not use existing database: %s\n", err)
  } else {
    printSuccess(
      "Linked to existing database %s -> %s.\n",
      app.storeFileName,
      existingFullPath)
  }
}
