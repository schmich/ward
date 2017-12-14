package main

import (
  "github.com/jawher/mow.cli"
  "encoding/json"
  "os"
)

func (app *App) exportCommand(cmd *cli.Cmd) {
  cmd.Spec = "[--compact] [FILE]"

  file := cmd.StringArg("FILE", "", "Destination file. Otherwise, output written to stdout.")
  compact := cmd.BoolOpt("compact", false, "Generate compact JSON output.")

  cmd.Action = func() {
    app.runExport(*file, *compact)
  }
}

func (app *App) runExport(fileName string, compact bool) {
  db := app.openStore()
  defer db.Close()

  var err error
  var output *os.File

  if fileName == "" {
    output = os.Stdout
  } else {
    output, err = os.Create(fileName)
    if err != nil {
      panic(err)
    }

    defer output.Close()
  }

  credentials := db.AllCredentials()

  var jsonData []byte
  if compact {
    jsonData, err = json.Marshal(credentials)
  } else {
    jsonData, err = json.MarshalIndent(credentials, "", "  ")
  }

  if err != nil {
    printError("%s\n", err)
    return
  }

  output.Write(jsonData)

  if fileName != "" {
    printSuccess("Exported credentials to %s.\n", fileName)
  }
}
