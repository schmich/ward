package main

import (
  "github.com/jawher/mow.cli"
)

func (app *App) delCommand(cmd *cli.Cmd) {
  cmd.Spec = "QUERY..."

  query := cmd.Strings(cli.StringsArg {
    Name: "QUERY",
    Desc: "Criteria to match.",
    Value: []string{},
    EnvVar: "",
  })

  cmd.Action = func() {
    app.runDel(*query)
  }
}

func (app *App) runDel(query []string) {
  db := app.openStore()
  defer db.Close()

  credential := app.findCredential(db, query)
  if credential == nil {
    return
  }

  identifier := formatCredential(credential)
  if confirm := app.readYesNo("Delete " + identifier); confirm {
    db.DeleteCredential(credential)
    app.printSuccess("Credential deleted.\n")
  } else {
    app.printError("Canceled.\n")
  }
}
