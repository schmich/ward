package main

import (
  "github.com/schmich/ward/store"
  "github.com/schmich/ward/passgen"
  "golang.org/x/crypto/ssh/terminal"
  "gopkg.in/alecthomas/kingpin.v2"
  "github.com/fatih/color"
  "bufio"
  "fmt"
  "os"
)

type App struct {
  app *kingpin.Application
  init *kingpin.CmdClause
  new *kingpin.CmdClause
  copy *kingpin.CmdClause
  query *string
}

func NewApp() *App {
  app := kingpin.New("ward", "Password manager.")
  init := app.Command("init", "Create a new password database.")
  new := app.Command("new", "Add a new credential.")
  copy := app.Command("copy", "Copy password.")
  query := copy.Arg("query", "Text to match.").Required().String()

  return &App {
    app: app,
    init: init,
    new: new,
    copy: copy,
    query: query,
  }
}

func (app *App) runInit() {
  password := app.readPassword("Master password: ")
  confirm := app.readPassword("Master password (confirm): ")

  if password != confirm {
    panic("Passwords do not match.")
  }

  db := store.Create("test.db", password)
  defer db.Close()
}

func (app *App) readInput(prompt string) string {
  print(color.CyanString(prompt))
  input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
  return input
}

func (app *App) readPassword(prompt string) string {
  print(color.CyanString(prompt))
  password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
  println()
  return string(password)
}

func (app *App) runNew() {
  master := app.readPassword("Master password: ")

  db := store.Open("test.db", master)
  defer db.Close()

  login := app.readInput("Username: ")

  password := app.readPassword("Password (enter to generate): ")
  if len(password) > 0 {
    confirm := app.readPassword("Password (confirm): ")

    if confirm != password {
      panic("Passwords do not match.")
    }
  } else {
    password = passgen.NewPassword(&passgen.Options {
      Length: 30,
      Upper: true,
      Lower: true,
      Number: true,
      Symbol: true,
    })
  }

  website := app.readInput("Website: ")
  note := app.readInput("Note: ")

  db.AddCredential(&store.Credential {
    Login: login,
    Password: string(password),
    Website: website,
    Note: note,
  })

  println("Credential added.")
}

func (app *App) runCopy(query string) {
  master := app.readPassword("Master password: ")

  db := store.Open("test.db", master)
  defer db.Close()

  credentials := db.FindCredentials(query)
  for _, credential := range credentials {
    fmt.Println(credential)
  }
}

func (app *App) Run(args []string) {
  switch kingpin.MustParse(app.app.Parse(args[1:])) {
  case app.init.FullCommand():
    app.runInit()
  case app.new.FullCommand():
    app.runNew()
  case app.copy.FullCommand():
    app.runCopy(*app.query)
  }
}

func main() {
  app := NewApp()
  app.Run(os.Args)
}
