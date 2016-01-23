package main

import (
  "github.com/schmich/ward/store"
  "github.com/schmich/ward/crypto"
  "github.com/schmich/ward/passgen"
  "golang.org/x/crypto/ssh/terminal"
  "github.com/mitchellh/go-homedir"
  "github.com/fatih/color"
  "github.com/jawher/mow.cli"
  "github.com/atotto/clipboard"
  "encoding/json"
  "path/filepath"
  "io/ioutil"
  "bufio"
  "strings"
  "strconv"
  "fmt"
  "os"
)

type App struct {
  scanner *bufio.Scanner
  storeFileName string
}

func NewApp(fileName string) *App {
  fullPath, _ := filepath.Abs(fileName)

  return &App {
    scanner: bufio.NewScanner(os.Stdin),
    storeFileName: fullPath,
  }
}

func (app *App) printSuccess(format string, args ...interface {}) {
  fmt.Printf(color.GreenString("✓ ") + format, args...)
}

func (app *App) printError(format string, args ...interface {}) {
  fmt.Printf(color.RedString("✗ ") + format, args...)
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

func (app *App) readPasswordConfirm(prompt string) string {
  for {
    password := app.readPassword(prompt + ": ")
    confirm := app.readPassword(prompt + " (confirm): ")

    if password != confirm {
      app.printError("Passwords do not match.\n")
    } else {
      return password
    }
  }
}

func (app *App) runInit() {
  fmt.Println("Creating new credential database.")
  password := app.readPasswordConfirm("Master password")

  db, err := store.Create(app.storeFileName, password)
  if err != nil {
    app.printError("Failed to create database: %s\n", err.Error())
    return
  }

  defer db.Close()

  app.printSuccess("Credential database created at %s.\n", app.storeFileName)
}

func (app *App) openStore() *store.Store {
  for {
    master := app.readPassword("Master password: ")
    db, err := store.Open(app.storeFileName, master)
    if err == nil {
      return db
    }

    app.printError("%s\n", err)

    if _, ok := err.(crypto.IncorrectPasswordError); !ok {
      os.Exit(1)
    }
  }
}

func (app *App) runAdd(login, realm, note string) {
  db := app.openStore()
  defer db.Close()

  if login == "" {
    login = app.readInput("Login: ")
  }

  password := app.readPasswordConfirm("Password")

  if realm == "" {
    realm = app.readInput("Realm: ")
  }

  if note == "" {
    note = app.readInput("Note: ")
  }

  db.AddCredential(&store.Credential {
    Login: login,
    Password: password,
    Realm: realm,
    Note: note,
  })

  app.printSuccess("Credential added.\n")
}

type passwordResult struct {
  password string
  err error
}

func (app *App) runGen(login, realm, note string, copyPassword bool, generator *passgen.Generator) {
  db := app.openStore()
  defer db.Close()

  passwordChan := make(chan *passwordResult)
  go func() {
    password, err := generator.Generate()
    passwordChan <- &passwordResult { password: password, err: err }
  }()

  if login == "" {
    login = app.readInput("Login: ")
  }

  if realm == "" {
    realm = app.readInput("Realm: ")
  }

  if note == "" {
    note = app.readInput("Note: ")
  }

  result := <-passwordChan
  if result.err != nil {
    app.printError("%s\n", result.err)
    return
  }

  db.AddCredential(&store.Credential {
    Login: login,
    Password: result.password,
    Realm: realm,
    Note: note,
  })

  app.printSuccess("Credential added. ")

  if copyPassword {
    clipboard.WriteAll(result.password)
    fmt.Println("Generated password copied to the clipboard.")
  } else {
    fmt.Println()
  }
}

func filter(arr []string, pred func(string) bool) []string {
  result := make([]string, 0)
  for _, str := range arr {
    if pred(str) {
      result = append(result, str)
    }
  }

  return result
}

func (app *App) readIndex(low, high int, prompt string) int {
  for {
    input := app.readInput(prompt)
    index, err := strconv.Atoi(input)
    if (err != nil) || (index < low) || (index > high) {
      app.printError("Invalid choice.\n")
    } else {
      return index
    }
  }
}

func (app *App) selectCredential(credentials []*store.Credential) *store.Credential {
  for i, credential := range credentials {
    parts := []string { credential.Login, credential.Realm, credential.Note }
    parts = filter(parts, func(s string) bool { return s != "" })
    fmt.Printf("%d. %s\n", i + 1, strings.Join(parts, ", "))
  }

  index := app.readIndex(1, len(credentials), "> ")
  return credentials[index - 1]
}

func (app *App) findCredential(db *store.Store, query []string) *store.Credential {
  credentials := db.FindCredentials(query)
  if len(credentials) == 0 {
    queryString := strings.Join(query, " ")
    app.printError("No credentials match \"%s\".\n", queryString)
    return nil
  } else if len(credentials) == 1 {
    return credentials[0]
  } else {
    queryString := strings.Join(query, " ")
    fmt.Printf("Found multiple credentials matching \"%s\":\n", queryString)
    return app.selectCredential(credentials)
  }
}

func (app *App) runCopy(query []string) {
  db := app.openStore()
  defer db.Close()

  credential := app.findCredential(db, query)
  if credential == nil {
    return
  }

  clipboard.WriteAll(credential.Password)
  app.printSuccess("Password copied to the clipboard.\n")
}

func (app *App) runEdit(query []string) {
  db := app.openStore()
  defer db.Close()

  credential := app.findCredential(db, query)
  if credential == nil {
    return
  }

  fmt.Printf("Login: %s\n", credential.Login)
  fmt.Printf("Password: %s\n", credential.Password)
  fmt.Printf("Realm: %s\n", credential.Realm)
  fmt.Printf("Note: %s\n", credential.Note)

  if login := app.readInput("Login (blank to keep current): "); login != "" {
    credential.Login = login
  }

  if password := app.readInput("Password (blank to keep current): "); password != "" {
    credential.Password = password
  }

  if realm := app.readInput("Realm (blank to keep current): "); realm != "" {
    credential.Realm = realm
  }

  if note := app.readInput("Note (blank to keep current): "); note != "" {
    credential.Note = note
  }

  db.UpdateCredential(credential)
  app.printSuccess("Credential updated.\n")
}

func (app *App) runDel(query []string) {
  db := app.openStore()
  defer db.Close()

  credential := app.findCredential(db, query)
  if credential == nil {
    return
  }

  db.DeleteCredential(credential)
  app.printSuccess("Credential deleted.\n")
}

func (app *App) runExport(fileName string, indent bool) {
  db := app.openStore()
  defer db.Close()

  var err error
  var output *os.File

  if fileName == "" {
    output = os.Stdout
  } else {
    output, err = os.Create(fileName)
    if err != nil {
      panic(err)
    }

    defer output.Close()
  }

  credentials := db.AllCredentials()

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

  if fileName != "" {
    app.printSuccess("Exported credentials to %s.\n", fileName)
  }
}

func (app *App) runImport(fileName string) {
  db := app.openStore()
  defer db.Close()

  input, err := os.Open(fileName)
  if err != nil {
    app.printError("Failed to open %s: %s\n", fileName, err)
    return
  }

  defer input.Close()

  contents, err := ioutil.ReadAll(input)
  if err != nil {
    app.printError("%s\n", err)
    return
  }

  var credentials []store.Credential
  err = json.Unmarshal(contents, &credentials)

  if err != nil {
    app.printError("%s\n", err)
    return
  }

  fmt.Printf("Importing %d credentials.\n", len(credentials))
  for _, credential := range credentials {
    db.AddCredential(&credential)
  }

  app.printSuccess("Imported credentials from %s.\n", fileName)
}

func (app *App) runUpdateMasterPassword() {
  db := app.openStore()
  defer db.Close()

  password := app.readPasswordConfirm("New master password")
  err := db.UpdateMasterPassword(password)
  if err != nil {
    app.printError("%s\n", err)
    return
  }

  app.printSuccess("Master password updated.\n")
}

func (app *App) runLink(existingFileName string) {
  existingFullPath, _ := filepath.Abs(existingFileName)
  err := os.Link(existingFullPath, app.storeFileName)
  if err != nil {
    app.printError("Could not use existing database: %s\n", err)
  } else {
    app.printSuccess(
      "Linked to existing database %s -> %s.\n",
      app.storeFileName,
      existingFullPath)
  }
}

func (app *App) Run(args []string) {
  ward := cli.App("ward", "Secure password manager - https://github.com/schmich/ward")

  ward.Version("v version", "ward 0.0.1")

  ward.Command("init", "Create a new credential database.", func(cmd *cli.Cmd) {
    file := cmd.StringOpt("link", "", "Link to an existing credential database.")

    cmd.Action = func() {
      if *file == "" {
        app.runInit()
      } else {
        app.runLink(*file)
      }
    }
  })

  ward.Command("add", "Add a credential with a known password.", func(cmd *cli.Cmd) {
    login := cmd.StringOpt("login", "", "Login for credential, e.g. username or email.")
    realm := cmd.StringOpt("realm", "", "Realm for credential, e.g. website or WiFi AP name.")
    note := cmd.StringOpt("note", "", "Note for credential.")

    cmd.Action = func() {
      app.runAdd(*login, *realm, *note)
    }
  })

  ward.Command("gen", "Add a credential with a generated password.", func(cmd *cli.Cmd) {
    login := cmd.StringOpt("login", "", "Login for credential, e.g. username or email.")
    realm := cmd.StringOpt("realm", "", "Realm for credential, e.g. website or WiFi AP name.")
    note := cmd.StringOpt("note", "", "Note for credential.")

    length := cmd.IntOpt("length", 0, "Password length.")
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

    noCopy := cmd.BoolOpt("no-copy", false, "Do not copy generated password to the clipboard.")

    cmd.Action = func() {
      generator := passgen.New()
      generator.AddAlphabet("upper", "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
      generator.AddAlphabet("lower", "abcdefghijklmnopqrstuvwxyz")
      generator.AddAlphabet("digit", "0123456789")
      generator.AddAlphabet("symbol", "`~!@#$%^&*()-_=+[{]}\\|;:'\",<.>/?")
      if *length == 0 {
        generator.SetLength(*minLength, *maxLength)
      } else {
        generator.SetLength(*length, *length)
      }
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
      generator.Exclude = *exclude
      if (*noSimilar) {
        generator.Exclude += "B8|1IiLl0Oo"
      }
      app.runGen(*login, *realm, *note, !*noCopy, generator)
    }
  })

  ward.Command("copy", "Copy a password to the clipboard.", func(cmd *cli.Cmd) {
    cmd.Spec = "QUERY..."

    query := cmd.Strings(cli.StringsArg {
      Name: "QUERY",
      Desc: "Criteria to match.",
      Value: []string{},
      EnvVar: "",
    })

    cmd.Action = func() {
      app.runCopy(*query)
    }
  })

  ward.Command("edit", "Edit an existing credential.", func(cmd *cli.Cmd) {
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
  })

  ward.Command("del", "Delete a stored credential.", func(cmd *cli.Cmd) {
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
  })

  ward.Command("show", "Show a stored credential.", func(cmd *cli.Cmd) {
    // TODO
    fmt.Println("show")
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
    file := cmd.StringArg("FILE", "", "File to import.")

    cmd.Action = func() {
      app.runImport(*file)
    }
  })

  ward.Command("master", "Update master password.", func(cmd *cli.Cmd) {
    cmd.Action = app.runUpdateMasterPassword
  })

  ward.Run(args)
}

func main() {
  homeDir, _ := homedir.Dir()
  wardPath := filepath.Join(homeDir, ".ward")

  app := NewApp(wardPath)
  app.Run(os.Args)
}
