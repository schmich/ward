package main

import (
  "github.com/schmich/ward/store"
  "github.com/schmich/ward/passgen"
  "github.com/jawher/mow.cli"
  "github.com/atotto/clipboard"
  "fmt"
)

func (app *App) addCommand(cmd *cli.Cmd) {
  const SimilarChars = "5SB8|1IiLl0Oo"

  cmd.Spec = "[--login] [--realm] [--note] [--no-copy] [--gen [--length] [--min-length] [--max-length] [--no-upper] [--no-lower] [--no-digit] [--no-symbol] [--no-similar] [--min-upper] [--max-upper] [--min-lower] [--max-lower] [--min-digit] [--max-digit] [--min-symbol] [--max-symbol] [--exclude]]"

  login := cmd.StringOpt("login", "", "Login for credential, e.g. username or email.")
  realm := cmd.StringOpt("realm", "", "Realm for credential, e.g. website or WiFi AP name.")
  note := cmd.StringOpt("note", "", "Note for credential.")
  noCopy := cmd.BoolOpt("no-copy", false, "Do not copy password to the clipboard.")

  gen := cmd.BoolOpt("gen", false, "Generate a password.")
  length := cmd.IntOpt("length", 0, "Password length.")
  minLength := cmd.IntOpt("min-length", 30, "Minimum length password.")
  maxLength := cmd.IntOpt("max-length", 40, "Maximum length password.")

  noUpper := cmd.BoolOpt("no-upper", false, "Exclude uppercase characters from password.")
  noLower := cmd.BoolOpt("no-lower", false, "Exclude lowercase characters from password.")
  noDigit := cmd.BoolOpt("no-digit", false, "Exclude digit characters from password.")
  noSymbol := cmd.BoolOpt("no-symbol", false, "Exclude symbol characters from password.")
  noSimilar := cmd.BoolOpt("no-similar", false, "Exclude similar characters from password: " + SimilarChars + ".")

  minUpper := cmd.IntOpt("min-upper", 0, "Minimum number of uppercase characters in password.")
  maxUpper := cmd.IntOpt("max-upper", -1, "Maximum number of uppercase characters in password.")
  minLower := cmd.IntOpt("min-lower", 0, "Minimum number of lowercase characters in password.")
  maxLower := cmd.IntOpt("max-lower", -1, "Maximum number of lowercase characters in password.")
  minDigit := cmd.IntOpt("min-digit", 0, "Minimum number of digit characters in password.")
  maxDigit := cmd.IntOpt("max-digit", -1, "Maximum number of digit characters in password.")
  minSymbol := cmd.IntOpt("min-symbol", 0, "Minimum number of symbol characters in password.")
  maxSymbol := cmd.IntOpt("max-symbol", -1, "Maximum number of symbol characters in password.")

  exclude := cmd.StringOpt("exclude", "", "Exclude specific characters from password.")

  cmd.Action = func() {
    if !*gen {
      app.runAdd(*login, *realm, *note, !*noCopy)
    } else {
      generator := passgen.New()
      if *length == 0 {
        generator.SetLength(*minLength, *maxLength)
      } else {
        generator.SetLength(*length, *length)
      }
      upper := generator.AddAlphabet("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
      lower := generator.AddAlphabet("abcdefghijklmnopqrstuvwxyz")
      digit := generator.AddAlphabet("0123456789")
      symbol := generator.AddAlphabet("`~!@#$%^&*()-_=+[{]}\\|;:'\",<.>/?")
      upper.SetMinMax(*minUpper, *maxUpper)
      lower.SetMinMax(*minLower, *maxLower)
      digit.SetMinMax(*minDigit, *maxDigit)
      symbol.SetMinMax(*minSymbol, *maxSymbol)
      if (*noUpper) {
        upper.SetMinMax(0, 0)
      }
      if (*noLower) {
        lower.SetMinMax(0, 0)
      }
      if (*noDigit) {
        digit.SetMinMax(0, 0)
      }
      if (*noSymbol) {
        symbol.SetMinMax(0, 0)
      }
      generator.Exclude = *exclude
      if (*noSimilar) {
        generator.Exclude += SimilarChars
      }
      app.runGen(*login, *realm, *note, !*noCopy, generator)
    }
  }
}

func (app *App) runAdd(login, realm, note string, copyPassword bool) {
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

  app.printSuccess("Credential added. ")

  if copyPassword {
    clipboard.WriteAll(password)
    fmt.Println("Password copied to the clipboard.")
  } else {
    fmt.Println()
  }
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
