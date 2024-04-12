// Package to automate some database discovery tasks on the openm model databases.
// Some things we could try to do:
// Print out table definitions for tables common to all models (for a given set of databases).
// Print out all (non system) table definitions for a specified database and create a database schema diagram. 

// Start with executing sqlite tables list from inside go and capturing its output.
// Then executing sqlite with a specific table name argument (for all tables) to produce the table schemas.

// Then we can start implementing a scanner and AST for the SQL table schemas.
// And then functions to output the dot file.
package main

import (
    "fmt"
    "os"
    "os/exec"
    //"io/ioutil"
    //"path/filepath"
)

func main () {
    wd, err := os.Getwd()
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(wd)

    // Obtain a sqlite database filename (passed as command line argument).
    // Run sqlite command (SQL query) on the database.

    // We may need to add another pair of double quotes and backspace them.
    query := "SELECT name FROM sqlite_schema WHERE type='table' AND name NOT LIKE 'sqlite?'"

    cmd := exec.Command("sqlite3", os.Args[1], query)
    fmt.Println("Created command object.")

    output, err := cmd.Output()
    fmt.Println("Called Output() on command object.")

    if err != nil {
        fmt.Println(err)
        return
    }

    fmt.Println(output)
    fmt.Println("Successfully captured standard output from sqlite3.")

    // Create a directory to store the table-schema files.
}
