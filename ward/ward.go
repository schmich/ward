package main

import (
  "github.com/schmich/ward/store"
  "github.com/schmich/ward/passgen"
  "golang.org/x/crypto/ssh/terminal"
  "github.com/fatih/color"
  "github.com/jawher/mow.cli"
  "encoding/json"
  "path/filepath"
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
  fmt.Fprint(os.Stderr, color.CyanString(prompt))
  app.scanner.Scan()
  return app.scanner.Text()
}

func (app *App) readPassword(prompt string) string {
  fmt.Fprint(os.Stderr, color.CyanString(prompt))
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

  filename := "test.db"

  db := store.Create(filename, password)
  defer db.Close()

  fullPath, _ := filepath.Abs(filename)
  fmt.Printf("Credential database created at %s.\n", fullPath)
}

func (app *App) runAdd(login, website, note string) {
  master := app.readPassword("Master password: ")
  db := store.Open("test.db", master)
  defer db.Close()

  if login == "" {
    login = app.readInput("Login: ")
  }

  password := app.readPassword("Password: ")
  confirm := app.readPassword("Password (confirm): ")

  if confirm != password {
    // Loop.
    panic("Passwords do not match.")
  }

  if website == "" {
    website = app.readInput("Website: ")
  }

  if note == "" {
    note = app.readInput("Note: ")
  }

  db.AddCredential(&store.Credential {
    Login: login,
    Password: password,
    Website: website,
    Note: note,
  })

  println("Credential added.")
}

func (app *App) runGen(login, website, note string, generator *passgen.Generator) {
  master := app.readPassword("Master password: ")
  db := store.Open("test.db", master)
  defer db.Close()

  password := make(chan string)
  go func() {
    password <- generator.Generate()
  }()

  if login == "" {
    login = app.readInput("Login: ")
  }

  if website == "" {
    website = app.readInput("Website: ")
  }

  if note == "" {
    note = app.readInput("Note: ")
  }

  db.AddCredential(&store.Credential {
    Login: login,
    Password: <-password,
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

func (app *App) runExport(filename string, indent bool) {
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

  var jsonData []byte
  if indent {
    jsonData, err = json.MarshalIndent(credentials, "", "  ")
  } else {
    jsonData, err = json.Marshal(credentials)
  }

  if err != nil {
    panic(err)
  }

  output.Write(jsonData)

  if filename != "" {
    fmt.Printf("Exported credentials to %s.\n", filename)
  }
}

func (app *App) Run(args []string) {
  ward := cli.App("ward", "Secure password manager - https://github.com/schmich/ward")

  ward.Version("v version", "ward 0.0.1")

  ward.Command("init", "Create a new credential database.", func(cmd *cli.Cmd) {
    cmd.Action = func() {
      app.runInit()
    }
  })

  ward.Command("add", "Add a credential with a known password.", func(cmd *cli.Cmd) {
    login := cmd.StringOpt("login", "", "Login for credential, e.g. username or email.")
    website := cmd.StringOpt("website", "", "Website for credential.")
    note := cmd.StringOpt("note", "", "Note for credential.")

    cmd.Action = func() {
      app.runAdd(*login, *website, *note)
    }
  })

  ward.Command("gen", "Add a credential with a generated password.", func(cmd *cli.Cmd) {
    login := cmd.StringOpt("login", "", "Login for credential, e.g. username or email.")
    website := cmd.StringOpt("website", "", "Website for credential.")
    note := cmd.StringOpt("note", "", "Note for credential.")

    minLength := cmd.IntOpt("min-length", 30, "Minimum length password.")
    maxLength := cmd.IntOpt("max-length", 40, "Maximum length password.")

    noUpper := cmd.BoolOpt("no-upper", false, "Exclude uppercase characters in password.")
    noLower := cmd.BoolOpt("no-lower", false, "Exclude lowercase characters in password.")
    noNumeric := cmd.BoolOpt("no-numeric", false, "Exclude numeric characters in password.")
    noSymbol := cmd.BoolOpt("no-symbol", false, "Exclude symbol characters in password.")
    noSimilar := cmd.BoolOpt("no-similar", false, "Exclude similar characters in password.")

    minUpper := cmd.IntOpt("min-upper", 0, "Minimum number of uppercase characters in password.")
    maxUpper := cmd.IntOpt("max-upper", -1, "Maximum number of uppercase characters in password.")
    minLower := cmd.IntOpt("min-lower", 0, "Minimum number of lowercase characters in password.")
    maxLower := cmd.IntOpt("max-lower", -1, "Maximum number of lowercase characters in password.")
    minNumeric := cmd.IntOpt("min-numeric", 0, "Minimum number of numeric characters in password.")
    maxNumeric := cmd.IntOpt("max-numeric", -1, "Maximum number of numeric characters in password.")
    minSymbol := cmd.IntOpt("min-symbol", 0, "Minimum number of symbol characters in password.")
    maxSymbol := cmd.IntOpt("max-symbol", -1, "Maximum number of symbol characters in password.")

    exclude := cmd.StringOpt("exclude", "", "Exclude specific characters from password.")

    cmd.Action = func() {
      generator := passgen.New()
      generator.AddAlphabet("upper", "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
      generator.AddAlphabet("lower", "abcdefghijklmnopqrstuvwxyz")
      generator.AddAlphabet("digit", "0123456789")
      generator.AddAlphabet("symbol", "`~!@#$%^&*()-_=+[{]}\\|;:'\",<.>/?")
      generator.Exclude = *exclude
      generator.SetLength(*minLength, *maxLength)
      generator.SetMinMax("upper", *minUpper, *maxUpper)
      generator.SetMinMax("lower", *minLower, *maxLower)
      generator.SetMinMax("digit", *minNumeric, *maxNumeric)
      generator.SetMinMax("symbol", *minSymbol, *maxSymbol)
      if (*noUpper) {
        generator.SetMinMax("upper", 0, 0)
      }
      if (*noLower) {
        generator.SetMinMax("lower", 0, 0)
      }
      if (*noNumeric) {
        generator.SetMinMax("digit", 0, 0)
      }
      if (*noSymbol) {
        generator.SetMinMax("symbol", 0, 0)
      }
      if (*noSimilar) {
        generator.Exclude += "B8|1IiLl0Oo"
      }
      app.runGen(*login, *website, *note, generator)
    }
  })

  ward.Command("copy", "Copy a password to the clipboard.", func(cmd *cli.Cmd) {
    query := cmd.StringArg("QUERY", "", "Criteria to match.")

    cmd.Action = func() {
      app.runCopy(*query)
    }
  })

  ward.Command("edit", "Edit existing credentials.", func(cmd *cli.Cmd) {
    fmt.Println("edit")
  })

  ward.Command("del", "Delete a stored credential.", func(cmd *cli.Cmd) {
    fmt.Println("del")
  })

  ward.Command("show", "Show a stored credential.", func(cmd *cli.Cmd) {
    fmt.Println("show")
  })

  ward.Command("use", "Use an existing credential database.", func(cmd *cli.Cmd) {
    fmt.Println("use")
  })

  ward.Command("export", "Export JSON-formatted credentials.", func(cmd *cli.Cmd) {
    cmd.Spec = "[--indent] [FILE]"

    file := cmd.StringArg("FILE", "", "Destination file. Otherwise, output written to stdout.")
    indent := cmd.BoolOpt("indent", false, "Indent JSON output.")

    cmd.Action = func() {
      app.runExport(*file, *indent)
    }
  })

  ward.Command("import", "Import JSON-formatted credentials.", func(cmd *cli.Cmd) {
    fmt.Println("import")
  })

  ward.Run(args)
}

func main() {
  app := NewApp()
  app.Run(os.Args)
}
