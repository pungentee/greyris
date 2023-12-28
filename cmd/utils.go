package cmd

import (
	"go.mills.io/bitcask/v2"
	"log"
	"os"
	"path/filepath"
)

// getDB returns a bitcask.DB that locate in ~/.greyris/<name>
// don't forgot call `defer db.Close()`
func getDB(name string, new bool) (bitcask.DB, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(homeDir, ".greyris", name)

	if new {
		err = os.RemoveAll(path)
		if err != nil {
			return nil, err
		}
	}

	db, err := bitcask.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	return db, nil
}
