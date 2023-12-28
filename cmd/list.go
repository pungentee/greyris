package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"go.mills.io/bitcask/v2"
	"log"
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
		defer aliasesDB.Close()
		if err != nil {
			log.Fatal(err)
		}

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
