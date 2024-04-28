package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
	"go.mills.io/bitcask/v2"
)

var (
	newUser bool

	sortAllAliases bool
	progressPrint  = true
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "greyris [link] || [alias] || all || [link] all",
	Short: "Sorts you Spotify playlists",
	Long: `This console utility will help you sort your Spotify playlists

Sorting rules: by author name -> by album release date -> by track number in the album

Requires: The Redirect URI of your Spotify App should be "http://localhost:8080/callback"
`,

	Version: "v1.4.0",

	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}

		aliasesDB, err := getDB("alias", false)
		if err != nil {
			log.Fatal(err)
		}
		defer aliasesDB.Close()

		for index, value := range args {
			if value == "all" {
				sortAllAliases = true
			} else if !strings.HasPrefix(value, "https://open.spotify.com/playlist/") {
				link, err := aliasesDB.Get(bitcask.Key(value))
				if err != nil {
					return errors.New("invalid link or undefined alias")
				}
				args[index] = string(getIdByLink(string(link)))
			}
		}

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		userInfoDB, err := getDB("login", newUser)
		if err != nil {
			log.Fatal(err)
		}
		defer userInfoDB.Close()

		client, err := login(userInfoDB)
		if err != nil {
			log.Fatal(err)
		}

		if sortAllAliases {
			aliasesDB, err := getDB("alias", false)
			if err != nil {
				log.Fatal(err)
			}
			defer aliasesDB.Close()

			aliases, err := getAllIDsFromDB(aliasesDB)
			if err != nil {
				log.Fatal(err)
			}

			args = append(args, aliases...)
			args = removeValue(args, "all")
		}

		if len(args) > 1 || sortAllAliases {
			progressPrint = false
		}

		for _, value := range args {
			playlistID := spotify.ID(value)

			if len(args) > 1 || sortAllAliases {
				playlistDetails, err := client.GetPlaylist(context.Background(), playlistID)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("Sorts %s\n", playlistDetails.Name)
			}

			if progressPrint {
				fmt.Println("Fetching tracks list...")
			}
			items, err := getAllItemsList(client, playlistID)
			if err != nil {
				log.Fatal(err)
			}

			tracks := itemsToTracks(items)
			if progressPrint {
				fmt.Println("Sorting...")
			}
			sorted := sortTrackList(append([]Track(nil), tracks...))

			if progressPrint {
				fmt.Println("Reordering playlist...")
			}
			err = reorderPlaylist(client, playlistID, tracks, sorted)
			if err != nil {
				log.Fatal(err)
			}

			if progressPrint {
				fmt.Println()
			}
		}

		fmt.Println("Completed")
	},
}

func init() {
	rootCmd.Flags().BoolVar(&newUser, "new-user", false, "log in as a new user")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
