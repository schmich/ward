package main

import (
  "github.com/jawher/mow.cli"
  "github.com/rodaine/table"
  "github.com/fatih/color"
)

func (app *App) listCommand(cmd *cli.Cmd) {
  cmd.Action = func() {
    app.runList()
  }
}

func (app *App) runList() {
  db := app.openStore()
  defer db.Close()

  headerFmt := color.New(color.FgCyan, color.Underline).SprintfFunc()

  table := table.New("Login", "Realm")
  table.WithHeaderFormatter(headerFmt)

  credentials := db.AllCredentials()
  for _, credential := range credentials {
    table.AddRow(credential.Login, credential.Realm)
  }

  table.Print()
}
