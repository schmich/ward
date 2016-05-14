package main

import (
  "github.com/jawher/mow.cli"
  "fmt"
)

func (app *App) editCommand(cmd *cli.Cmd) {
  cmd.Spec = "QUERY..."

  query := cmd.Strings(cli.StringsArg {
    Name: "QUERY",
    Desc: "Criteria to match.",
    Value: []string{},
    EnvVar: "",
  })

  cmd.Action = func() {
    app.runEdit(*query)
  }
}

func (app *App) runEdit(query []string) {
  db := app.openStore()
  defer db.Close()

  credential := app.findCredential(db, query)
  if credential == nil {
    return
  }

  fmt.Println("\nCurrent credential:")
  fmt.Printf("Login: %s\n", credential.Login)
  fmt.Printf("Password: %s\n", credential.Password)
  fmt.Printf("Realm: %s\n", credential.Realm)
  fmt.Printf("Note: %s\n", credential.Note)
  fmt.Println("\nEdit credential:")

  if login := app.readInput("Login (blank to keep current): "); login != "" {
    credential.Login = login
  }

  if updated, password := app.readEditPasswordConfirm(); updated {
    credential.Password = password
  }

  if realm := app.readInput("Realm (blank to keep current): "); realm != "" {
    credential.Realm = realm
  }

  if note := app.readInput("Note (blank to keep current): "); note != "" {
    credential.Note = note
  }

  db.UpdateCredential(credential)
  fmt.Println()
  app.printSuccess("Credential updated.\n")
}
