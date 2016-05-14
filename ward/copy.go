package main

import (
  "github.com/jawher/mow.cli"
  "github.com/atotto/clipboard"
)

func (app *App) copyCommand(cmd *cli.Cmd) {
  cmd.Spec = "QUERY..."

  query := cmd.Strings(cli.StringsArg {
    Name: "QUERY",
    Desc: "Criteria to match.",
    Value: []string{},
    EnvVar: "",
  })

  cmd.Action = func() {
    app.runCopy(*query)
  }
}

func (app *App) runCopy(query []string) {
  db := app.openStore()
  defer db.Close()

  credential := app.findCredential(db, query)
  if credential == nil {
    return
  }

  clipboard.WriteAll(credential.Password)
  identifier := formatCredential(credential)

  app.printSuccess("Password for %s copied to the clipboard.\n", identifier)
}
