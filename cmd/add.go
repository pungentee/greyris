package cmd

import (
	"errors"
	"go.mills.io/bitcask/v2"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [name] [link] || [name] [link] [name] [link]...",
	Short: "Add or update alias to the playlist link",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MinimumNArgs(2)(cmd, args); err != nil {
			return err
		}
		if len(args)%2 == 1 {
			return errors.New("wrong number of arguments")
		}
		if !strings.HasPrefix(args[1], "https://open.spotify.com/playlist/") {
			return errors.New("invalid link")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		aliasesDB, err := getDB("alias", false)
		defer aliasesDB.Close()
		if err != nil {
			log.Fatal(err)
		}

		for index := 0; index < len(args); index += 2 {
			aliasName := args[index]
			aliasID := getIdByLink(args[index+1])

			err = aliasesDB.Put(bitcask.Key(aliasName), bitcask.Value(aliasID))
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
