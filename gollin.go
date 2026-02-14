package main

import (
    "fmt"
    "os"
    "log"
    "strings"
)

func main() {
    // reads the contents of gollin at the given path and generates a .go file
    fmt.Println("Begin parsing");

    if len(os.Args) < 2 {
        log.Fatal("gollin file path required.")
        os.Exit(1)
    }
    filePath := os.Args[1]

    var gollinPath string

    if strings.HasSuffix(filePath, ".gol") {
       gollinPath = filePath
    } else {
        // Build the full filepath
        // Should probably check for extensions before doing this
        var gollinBuilder strings.Builder
        gollinBuilder.WriteString(filePath)
        gollinBuilder.WriteString(".gol")
        gollinPath = gollinBuilder.String()
    }

    // Get gollin code
    code, err := os.ReadFile(gollinPath)

    if err != nil {
        log.Fatal(err)
    }

    // Manipulate the code here
    // parse()

    // we know gollinPath has suffix .gol at this point
    newFilePath := strings.Split(gollinPath, ".")[0]
    var goPath strings.Builder
    goPath.WriteString(newFilePath)
    goPath.WriteString(".go")

    // Spit out the go code
    os.WriteFile(goPath.String(), code, 0777);
}

func parse(filePath string) {
    return
}
