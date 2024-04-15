package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"go.mills.io/bitcask/v2"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Print list of all aliases",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(0)(cmd, args); err != nil {
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

		err = aliasesDB.ForEach(func(key bitcask.Key) error {
			value, err := aliasesDB.Get(key)
			if err != nil {
				return err
			}
			fmt.Printf("%s: %s\n", string(key), string(value))

			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
