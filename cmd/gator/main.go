package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/devgiri0082/gator/internal/config"
	"github.com/devgiri0082/gator/internal/database"
	"github.com/devgiri0082/gator/internal/registry"
	rssfetcher "github.com/devgiri0082/gator/internal/rssFetcher"
	_ "github.com/lib/pq"
)

func main() {
	c, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	db, err := sql.Open("postgres", c.DB_URL)
	if err != nil {
		fmt.Printf("Unable to connect to database: %s", c.DB_URL)
		return
	}
	dbQueries := database.New(db)
	client := rssfetcher.New()
	r := registry.New(c, dbQueries, client)
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Invalid number of arguments")
		os.Exit(1)
	}
	key := args[1]
	err = r.Run(key, args[2:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
