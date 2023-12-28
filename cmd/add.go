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
	Use:   "add [alias name] [playlist link]",
	Short: "Add or update alias to the playlist link",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		if !strings.HasPrefix(args[1], "https://open.spotify.com/playlist/") {
			return errors.New("invalid link")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		aliasName := args[0]
		aliasLink := args[1][:56]

		aliasesDB, err := getDB("alias", false)
		defer aliasesDB.Close()
		if err != nil {
			log.Fatal(err)
		}

		err = aliasesDB.Put(bitcask.Key(aliasName), bitcask.Value(aliasLink))
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
