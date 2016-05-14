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

  fmt.Println("Current credential:")
  fmt.Printf("Login: %s\n", credential.Login)
  fmt.Println("Password: (not shown)")
  fmt.Printf("Realm: %s\n", credential.Realm)
  fmt.Printf("Note: %s\n", credential.Note)

  update := false

  for {
    response := app.readChar("Edit login, password, realm, note, or quit (l/p/r/n/q)? ", "lprnq")
    if response == 'q' {
      break
    }

    if response == 'l' {
      credential.Login = app.readInput("New login: ")
    } else if response == 'p' {
      credential.Password = app.readPasswordConfirm("New password")
    } else if response == 'r' {
      credential.Realm = app.readInput("New realm: ")
    } else if response == 'n' {
      credential.Note = app.readInput("New note: ")
    }

    update = true
  }

  if update {
    db.UpdateCredential(credential)
    app.printSuccess("Credential updated.\n")
  } else {
    app.printError("No changes made.\n")
  }
}
