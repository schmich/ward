package main

import (
  "github.com/jawher/mow.cli"
  "github.com/mattn/go-colorable"
  "github.com/qpliu/qrencode-go/qrencode"
  "github.com/fumiyas/qrc/lib"
)

func (app *App) qrCommand(cmd *cli.Cmd) {
  cmd.Spec = "QUERY..."

  query := cmd.Strings(cli.StringsArg {
    Name: "QUERY",
    Desc: "Criteria to match.",
    Value: []string{},
    EnvVar: "",
  })

  cmd.Action = func() {
    app.runQr(*query)
  }
}

func (app *App) runQr(query []string) {
  db := app.openStore()
  defer db.Close()

  credential := app.findCredential(db, query)
  if credential == nil {
    return
  }

  grid, err := qrencode.Encode(credential.Password, qrencode.ECLevelL)
  if err != nil {
    app.printError("%s\n", err)
    return
  }

  identifier := formatCredential(credential)
  app.printSuccess("Password for %s:\n", identifier)

  stdout := colorable.NewColorableStdout()
  qrc.PrintAA(stdout, grid, false)
}
