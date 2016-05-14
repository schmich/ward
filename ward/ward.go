package main

import (
  "github.com/mitchellh/go-homedir"
  "github.com/jawher/mow.cli"
  "path/filepath"
  "bufio"
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
  ward.Command("list", "Print a table-formatted list of credentials.", app.listCommand)
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
