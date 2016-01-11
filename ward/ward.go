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

func runInit() {
  password := readPassword("Master password: ")
  confirm := readPassword("Master password (confirm): ")

  if password != confirm {
    panic("Passwords do not match.")
  }

  db := store.Create("test.db", password)
  defer db.Close()
}

func readInput(prompt string) string {
  print(color.CyanString(prompt))
  input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
  return input
}

func readPassword(prompt string) string {
  print(color.CyanString(prompt))
  password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
  println()
  return string(password)
}

func runAdd() {
  master := readPassword("Master password: ")

  db := store.Open("test.db", master)
  defer db.Close()

  login := readInput("Username: ")

  password := readPassword("Password (enter to generate): ")
  if len(password) > 0 {
    confirm := readPassword("Password (confirm): ")

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

  website := readInput("Website: ")
  note := readInput("Note: ")

  db.AddCredential(&store.Credential {
    Login: login,
    Password: string(password),
    Website: website,
    Note: note,
  })

  println("Credential added.")
}

func runCopy(query string) {
  master := readPassword("Master password: ")

  db := store.Open("test.db", master)
  defer db.Close()

  credentials := db.FindCredentials(query)
  for _, credential := range credentials {
    fmt.Println(credential)
  }
}

func runApp(args []string) {
  app := kingpin.New("ward", "Password manager.")
  init := app.Command("init", "Create a new password database.")
  add := app.Command("add", "Add a new credential.")
  copy := app.Command("copy", "Copy password.")
  query := copy.Arg("query", "Text to match.").Required().String()

  switch kingpin.MustParse(app.Parse(args[1:])) {
  case init.FullCommand():
    runInit()
  case add.FullCommand():
    runAdd()
  case copy.FullCommand():
    runCopy(*query)
  }
}

func main() {
  runApp(os.Args)
}
