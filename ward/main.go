package main

import (
  "github.com/mitchellh/go-homedir"
  "path/filepath"
  "os"
)

func main() {
  wardFile := os.Getenv("WARDFILE")
  if wardFile == "" {
    homeDir, _ := homedir.Dir()
    wardFile = filepath.Join(homeDir, ".ward")
  }

  app := NewApp(wardFile)
  app.Run(os.Args)
}
