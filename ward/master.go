package main

import (
  "github.com/jawher/mow.cli"
)

func (app *App) masterCommand(cmd *cli.Cmd) {
  stretch := cmd.IntOpt("stretch", 200000, "Password key stretch iterations.")

  cmd.Action = func() {
    app.runUpdateMasterPassword(*stretch)
  }
}

func (app *App) runUpdateMasterPassword(keyStretch int) {
  db := app.openStore()
  defer db.Close()

  password := app.readPasswordConfirm("New master password")
  err := db.UpdateMasterPassword(password, keyStretch)
  if err != nil {
    app.printError("%s\n", err)
    return
  }

  app.printSuccess("Master password updated.\n")
}
