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

  credential := findCredential(db, query)
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
    response := readChar("Edit login, password, realm, note, or quit (l/p/r/n/q)? ", "lprnq")
    if response == 'q' {
      break
    }

    if response == 'l' {
      credential.Login = readInput("New login: ")
    } else if response == 'p' {
      credential.Password = readPasswordConfirm("New password")
    } else if response == 'r' {
      credential.Realm = readInput("New realm: ")
    } else if response == 'n' {
      credential.Note = readInput("New note: ")
    }

    update = true
  }

  if update {
    db.UpdateCredential(credential)
    printSuccess("Credential updated.\n")
  } else {
    printError("No changes made.\n")
  }
}
