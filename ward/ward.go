package main

import (
  "github.com/schmich/ward/store"
  "github.com/schmich/ward/passgen"
  "golang.org/x/crypto/ssh/terminal"
  "gopkg.in/alecthomas/kingpin.v2"
  "github.com/fatih/color"
  "encoding/json"
  "bufio"
  "fmt"
  "os"
)

type App struct {
  scanner *bufio.Scanner
}

func NewApp() *App {
  return &App {
    scanner: bufio.NewScanner(os.Stdin),
  }
}

func (app *App) readInput(prompt string) string {
  fmt.Println(os.Stderr, color.CyanString(prompt))
  app.scanner.Scan()
  return app.scanner.Text()
}

func (app *App) readPassword(prompt string) string {
  fmt.Println(os.Stderr, color.CyanString(prompt))
  password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
  println()
  return string(password)
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

func (app *App) runAdd() {
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

func (app *App) runExport(filename string) {
  master := app.readPassword("Master password: ")
  db := store.Open("test.db", master)
  defer db.Close()

  var err error
  var output *os.File

  if filename == "" {
    output = os.Stdout
  } else {
    output, err = os.Create(filename)
    if err != nil {
      panic(err)
    }
  }

  credentials := db.GetCredentials()

  json, err := json.Marshal(credentials)
  if err != nil {
    panic(err)
  }

  output.Write(json)

  if filename != "" {
    fmt.Printf("Exported credentials to %s.\n", filename)
  }
}

func (app *App) Run(args []string) {
  ward := kingpin.New("ward", "Password manager.")

  init := ward.Command("init", "Create a new password database.")

  add := ward.Command("add", "Add a new credential.")

  copy := ward.Command("copy", "Copy password.")
  copyQuery := copy.Arg("query", "Text to match.").Required().String()

  export := ward.Command("export", "Export credentials to JSON file.")
  exportFile := export.Arg("file", "Destination file.").String()

  switch kingpin.MustParse(ward.Parse(args[1:])) {
  case init.FullCommand():
    app.runInit()
  case add.FullCommand():
    app.runAdd()
  case copy.FullCommand():
    app.runCopy(*copyQuery)
  case export.FullCommand():
    app.runExport(*exportFile)
  }
}

func main() {
  app := NewApp()
  app.Run(os.Args)
}
