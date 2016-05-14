package main

import (
  "github.com/schmich/ward/store"
  "github.com/fatih/color"
  "golang.org/x/crypto/ssh/terminal"
  "strings"
  "strconv"
  "bufio"
  "fmt"
  "os"
)

var scanner = bufio.NewScanner(os.Stdin)

func printSuccess(format string, args ...interface {}) {
  fmt.Printf(color.GreenString("✓ ") + format, args...)
}

func printError(format string, args ...interface {}) {
  fmt.Printf(color.RedString("✗ ") + format, args...)
}

func readInput(prompt string) string {
  fmt.Fprint(os.Stderr, prompt)
  color.Set(color.FgHiBlack)
  defer color.Unset()
  scanner.Scan()
  return scanner.Text()
}

func readPassword(prompt string) string {
  fmt.Fprint(os.Stderr, prompt)
  password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
  println()
  return string(password)
}

func readPasswordConfirm(prompt string) string {
  for {
    password := readPassword(prompt + ": ")
    confirm := readPassword(prompt + " (confirm): ")

    if password != confirm {
      printError("Passwords do not match.\n")
    } else {
      return password
    }
  }
}

func readChar(prompt string, allowedRunes string) byte {
  for {
    response := strings.ToLower(strings.TrimSpace(readInput(prompt)))

    if len(response) == 0 || !strings.Contains(allowedRunes, string(response[0])) {
      printError("Invalid response.\n")
    } else {
      return response[0]
    }
  }
}

func readYesNo(prompt string) bool {
  response := readChar(prompt + " (y/n)? ", "yn")
  return response == 'y'
}

func readIndex(low, high int, prompt string) int {
  for {
    input := strings.TrimSpace(readInput(prompt))
    index, err := strconv.Atoi(input)
    if (err != nil) || (index < low) || (index > high) {
      printError("Invalid choice.\n")
    } else {
      return index
    }
  }
}

func selectCredential(credentials []*store.Credential) *store.Credential {
  for i, credential := range credentials {
    fmt.Fprintf(os.Stderr, "%d. %s\n", i + 1, formatCredential(credential))
  }

  index := readIndex(1, len(credentials), "> ")
  return credentials[index - 1]
}

func findCredential(db *store.Store, query []string) *store.Credential {
  credentials := db.FindCredentials(query)
  if len(credentials) == 0 {
    queryString := strings.Join(query, " ")
    printError("No credentials match \"%s\".\n", queryString)
    return nil
  } else if len(credentials) == 1 {
    return credentials[0]
  } else {
    queryString := strings.Join(query, " ")
    fmt.Fprintf(os.Stderr, "Found multiple credentials matching \"%s\":\n", queryString)
    return selectCredential(credentials)
  }
}

func formatCredential(credential *store.Credential) string {
  loginRealm := ""
  if len(credential.Login) > 0 && len(credential.Realm) > 0 {
    loginRealm = credential.Login + "@" + credential.Realm
  } else {
    loginRealm = credential.Login + credential.Realm
  }

  if len(credential.Note) == 0 {
    return loginRealm
  }

  if len(loginRealm) == 0 {
    return credential.Note
  } else {
    return loginRealm + " (" + credential.Note + ")"
  }
}
