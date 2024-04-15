package cmd

import (
	"github.com/spf13/cobra"
	"go.mills.io/bitcask/v2"
	"log"
)

var removeCmd = &cobra.Command{
	Use:     "remove [alias] || [alias] [alias]...",
	Aliases: []string{"rm"},
	Short:   "Remove alias to the playlist link",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		aliasesDB, err := getDB("alias", false)
		if err != nil {
			log.Fatal(err)
		}
		defer aliasesDB.Close()

		for index := range args {
			aliasName := args[index]
			err = aliasesDB.Delete(bitcask.Key(aliasName))
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
