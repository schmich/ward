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

  credential := findCredential(db, query)
  if credential == nil {
    return
  }

  identifier := formatCredential(credential)
  if confirm := readYesNo("Delete " + identifier); confirm {
    db.DeleteCredential(credential)
    printSuccess("Credential deleted.\n")
  } else {
    printError("Canceled.\n")
  }
}
