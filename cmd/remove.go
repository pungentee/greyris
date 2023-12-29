package cmd

import (
	"github.com/spf13/cobra"
	"go.mills.io/bitcask/v2"
	"log"
)

var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove alias to the playlist link",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		aliasesDB, err := getDB("alias", false)
		defer aliasesDB.Close()
		if err != nil {
			log.Fatal(err)
		}

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
