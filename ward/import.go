package main

import (
  "github.com/jawher/mow.cli"
  "github.com/schmich/ward/store"
  "encoding/json"
  "io/ioutil"
  "fmt"
  "os"
)

func (app *App) importCommand(cmd *cli.Cmd) {
  file := cmd.StringArg("FILE", "", "File to import.")

  cmd.Action = func() {
    app.runImport(*file)
  }
}

func (app *App) runImport(fileName string) {
  db := app.openStore()
  defer db.Close()

  input, err := os.Open(fileName)
  if err != nil {
    app.printError("Failed to open %s: %s\n", fileName, err)
    return
  }

  defer input.Close()

  contents, err := ioutil.ReadAll(input)
  if err != nil {
    app.printError("%s\n", err)
    return
  }

  var credentials []store.Credential
  err = json.Unmarshal(contents, &credentials)

  if err != nil {
    app.printError("%s\n", err)
    return
  }

  fmt.Printf("Importing %d credentials.\n", len(credentials))
  for _, credential := range credentials {
    db.AddCredential(&credential)
  }

  app.printSuccess("Imported credentials from %s.\n", fileName)
}
